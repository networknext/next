#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/session_ping_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::PacketHeaderV4;
using core::PacketType;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_session_ping_handler_sdk4_unsigned)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.set_timestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = PacketHeaderV4::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeaderV4 header;
  {
    header.type = PacketType::SessionPing4;
    header.sequence = 123123130131LL | (1ULL << 62);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->next_addr = addr;
  session->expire_timestamp = 10;
  session->private_key = private_key;
  session->client_to_server_sequence = 0;

  map.set(header.hash(), session);

  size_t index = 0;

  CHECK(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  CHECK(index == PacketHeaderV4::SIZE_OF_SIGNED);

  core::handlers::session_ping_handler_sdk4(packet, map, recorder, router_info, socket, false);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(prev_len == packet.length);

  core::handlers::session_ping_handler_sdk4(packet, map, recorder, router_info, socket, false);
  CHECK(!socket.recv(packet));
}

TEST(core_handlers_session_ping_handler_sdk4_signed)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.set_timestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = crypto::PACKET_HASH_LENGTH + PacketHeaderV4::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeaderV4 header;
  {
    header.type = PacketType::SessionPing4;
    header.sequence = 123123130131LL | (1ULL << 62);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->client_to_server_sequence = 0;
  session->expire_timestamp = 10;
  session->next_addr = addr;
  session->prev_addr = addr;
  session->private_key = private_key;
  session->session_id = header.session_id;
  session->session_version = header.session_version;

  map.set(header.hash(), session);

  size_t index = crypto::PACKET_HASH_LENGTH;

  CHECK(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  CHECK(index == crypto::PACKET_HASH_LENGTH + PacketHeaderV4::SIZE_OF_SIGNED);

  core::handlers::session_ping_handler_sdk4(packet, map, recorder, router_info, socket, true);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet)).on_fail([&] {
    std::cout << "session = " << *session << '\n';
  });
  CHECK(prev_len == packet.length);

  core::handlers::session_ping_handler_sdk4(packet, map, recorder, router_info, socket, true);
  CHECK(!socket.recv(packet));
}

TEST(core_handlers_session_ping_handler_unsigned)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.set_timestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = PacketHeader::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::SessionPing;
    header.sequence = 123123130131LL | (1ULL << 62);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->next_addr = addr;
  session->expire_timestamp = 10;
  session->private_key = private_key;
  session->client_to_server_sequence = 0;

  map.set(header.hash(), session);

  size_t index = 0;

  CHECK(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  CHECK(index == PacketHeader::SIZE_OF_SIGNED);

  core::handlers::session_ping_handler(packet, map, recorder, router_info, socket, false);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(prev_len == packet.length);

  core::handlers::session_ping_handler(packet, map, recorder, router_info, socket, false);
  CHECK(!socket.recv(packet));
}

TEST(core_handlers_session_ping_handler_signed)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.set_timestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = crypto::PACKET_HASH_LENGTH + PacketHeader::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::SessionPing;
    header.sequence = 123123130131LL | (1ULL << 62);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->client_to_server_sequence = 0;
  session->expire_timestamp = 10;
  session->next_addr = addr;
  session->prev_addr = addr;
  session->private_key = private_key;
  session->session_id = header.session_id;
  session->session_version = header.session_version;

  map.set(header.hash(), session);

  size_t index = crypto::PACKET_HASH_LENGTH;

  CHECK(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  CHECK(index == crypto::PACKET_HASH_LENGTH + PacketHeader::SIZE_OF_SIGNED);

  core::handlers::session_ping_handler(packet, map, recorder, router_info, socket, true);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet)).on_fail([&] {
    std::cout << "session = " << *session << '\n';
  });
  CHECK(prev_len == packet.length);

  core::handlers::session_ping_handler(packet, map, recorder, router_info, socket, true);
  CHECK(!socket.recv(packet));
}
