#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/server_to_client_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::PacketType;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_server_to_client_handler_unsigned)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  router_info.set_timestamp(0);

  packet.length = PacketHeader::SIZE_OF + 100;
  packet.addr = addr;

  PacketHeader header = {
   .type = PacketType::ServerToClient,
   .sequence = 123123130131LL | (1ULL << 63),
   .session_id = 0x12313131,
   .session_version = 0x12,
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

  core::handlers::server_to_client_handler(packet, map, recorder, router_info, socket, false);

  size_t prev_len = packet.length;
  check(socket.recv(packet));
  check(prev_len == packet.length);

  check(recorder.server_to_client_tx.num_packets == 1);
  check(recorder.server_to_client_tx.num_bytes == packet.length);

  core::handlers::server_to_client_handler(packet, map, recorder, router_info, socket, false);
  check(!socket.recv(packet));
}

Test(core_handlers_server_to_client_handler_signed)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  router_info.set_timestamp(0);

  packet.length = crypto::PACKET_HASH_LENGTH + PacketHeader::SIZE_OF + 100;
  packet.addr = addr;

  PacketHeader header = {
   .type = PacketType::ServerToClient,
   .sequence = 123123130131LL | (1ULL << 63),
   .session_id = 0x12313131,
   .session_version = 0x12,
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

  core::handlers::server_to_client_handler(packet, map, recorder, router_info, socket, true);

  size_t prev_len = packet.length;
  check(socket.recv(packet));
  check(prev_len == packet.length);

  check(recorder.server_to_client_tx.num_packets == 1);
  check(recorder.server_to_client_tx.num_bytes == packet.length);

  core::handlers::server_to_client_handler(packet, map, recorder, router_info, socket, true);
  check(!socket.recv(packet));
}
