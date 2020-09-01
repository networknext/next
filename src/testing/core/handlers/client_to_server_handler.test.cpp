#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/client_to_server_handler.hpp"
#include "crypto/bytes.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::packets::Direction;
using core::packets::Header;
using core::packets::Type;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_client_to_server_handler_unsigned_packet)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.setTimestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = Header::ByteSize + 100;
  packet.addr = addr;

  Header header = {
   .type = Type::ClientToServer,
   .sequence = 123123130131LL,
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>();
  session->next_addr = addr;
  session->expire_timestamp = 10;
  session->private_key = private_key;
  session->client_to_server_sequence = 0;
  legacy::relay_replay_protection_reset(&session->client_to_server_protection);
  legacy::relay_replay_protection_reset(&session->server_to_client_protection);

  map.set(header.hash(), session);

  size_t index = 0;

  check(header.write(packet, index, Direction::ClientToServer, private_key));
  check(index == Header::ByteSize);

  core::handlers::client_to_server_handler(packet, map, recorder, router_info, socket, false);
  size_t prev_len = packet.length;
  check(socket.recv(packet));
  check(prev_len == packet.length);

  check(recorder.client_to_server_tx.num_packets == 1);
  check(recorder.client_to_server_tx.num_bytes == packet.length).onFail([&] {
    std::cout << "packet len = " << packet.length << std::endl;
    std::cout << "byte count = " << recorder.client_to_server_tx.num_bytes << std::endl;
  });

  core::handlers::client_to_server_handler(packet, map, recorder, router_info, socket, false);
  // check already received
  check(!socket.recv(packet));
}

Test(core_handlers_client_to_server_handler_signed_packet)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.setTimestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = crypto::PACKET_HASH_LENGTH + Header::ByteSize + 100;
  packet.addr = addr;

  Header header = {
   .type = Type::ClientToServer,
   .sequence = 123123130131LL,
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>();
  session->next_addr = addr;
  session->expire_timestamp = 10;
  session->private_key = private_key;
  session->session_id = header.session_id;
  session->session_version = header.session_version;
  session->client_to_server_sequence = 0;
  legacy::relay_replay_protection_reset(&session->client_to_server_protection);
  legacy::relay_replay_protection_reset(&session->server_to_client_protection);

  map.set(header.hash(), session);

  size_t index = crypto::PACKET_HASH_LENGTH;

  check(header.write(packet, index, Direction::ClientToServer, private_key));
  check(index == crypto::PACKET_HASH_LENGTH + Header::ByteSize);

  core::handlers::client_to_server_handler(packet, map, recorder, router_info, socket, true);
  check(socket.recv(packet));

  core::handlers::client_to_server_handler(packet, map, recorder, router_info, socket, true);
  // check already received
  check(!socket.recv(packet));
}
