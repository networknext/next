#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/session_pong_handler.hpp"

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
using crypto::PACKET_HASH_LENGTH;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_session_pong_handler_sdk4)
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

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = PacketHeaderV4::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeaderV4 header;
  {
    header.type = PacketType::ServerToClient4;
    header.sequence = 123123130131LL | (1ULL << 63);
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

  size_t index = 0;

  check(header.write(packet, index, PacketDirection::ServerToClient, private_key));
  check(index == PacketHeaderV4::SIZE_OF_SIGNED);

  core::handlers::session_pong_handler_sdk4(packet, map, recorder, router_info, socket);

  size_t prev_len = packet.length;
  check(socket.recv(packet));
  check(prev_len == packet.length);

  core::handlers::session_pong_handler_sdk4(packet, map, recorder, router_info, socket);
  check(!socket.recv(packet));
}

Test(core_handlers_session_pong_handler_unsigned)
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

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = PacketHeader::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::ServerToClient;
    header.sequence = 123123130131LL | (1ULL << 63);
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

  size_t index = 0;

  check(header.write(packet, index, PacketDirection::ServerToClient, private_key));
  check(index == PacketHeader::SIZE_OF_SIGNED);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, false);

  size_t prev_len = packet.length;
  check(socket.recv(packet));
  check(prev_len == packet.length);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, false);
  check(!socket.recv(packet));
}

Test(core_handlers_session_pong_handler_signed)
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

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = crypto::PACKET_HASH_LENGTH + PacketHeader::SIZE_OF_SIGNED + 32;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::ServerToClient;
    header.sequence = 123123130131LL | (1ULL << 63);
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

  check(header.write(packet, index, PacketDirection::ServerToClient, private_key));
  check(index == PACKET_HASH_LENGTH + PacketHeader::SIZE_OF_SIGNED);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, true);

  size_t prev_len = packet.length;
  check(socket.recv(packet));
  check(prev_len == packet.length);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, true);
  check(!socket.recv(packet));
}
