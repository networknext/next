#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/continue_response_handler.hpp"

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
using crypto::GenericKey;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_continue_response_handler_unsigned)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  Socket socket;

  const GenericKey private_key = random_private_key();

  RouterInfo router_info;
  router_info.setTimestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = Header::ByteSize;

  Header header = {
   .type = Type::ContinueResponse,
   .sequence = 123123130131LL | (1ULL << 63) | (1ULL << 62),
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>();
  session->session_id = header.session_id;
  session->session_version = header.session_version;
  session->private_key = private_key;
  session->prev_addr = addr;
  session->expire_timestamp = 10;
  session->server_to_client_sequence = 0;

  map.set(header.hash(), session);

  size_t index = 0;
  check(header.write(packet, index, Direction::ServerToClient, private_key));

  core::handlers::continue_response_handler(packet, map, recorder, router_info, socket, false);
  size_t prev_len = packet.length;
  check(socket.recv(packet)).onFail([&] {
    std::cout << "unable to receive packet\n";
  });
  check(prev_len == packet.length);

  check(recorder.continue_response_tx.num_packets == 1);
  check(recorder.continue_response_tx.num_bytes == packet.length).onFail([&] {
    std::cout << "packet len = " << packet.length << '\n';
    std::cout << "byte count = " << recorder.continue_response_rx.num_bytes << '\n';
  });

  core::handlers::continue_response_handler(packet, map, recorder, router_info, socket, false);
  check(!socket.recv(packet));
}

Test(core_handlers_continue_response_handler_signed)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  Socket socket;

  const GenericKey private_key = random_private_key();
  RouterInfo router_info;
  router_info.setTimestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.length = crypto::PACKET_HASH_LENGTH + Header::ByteSize;

  Header header = {
   .type = Type::ContinueResponse,
   .sequence = 123123130131LL | (1ULL << 63) | (1ULL << 62),
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>();
  session->session_id = header.session_id;
  session->session_version = header.session_version;
  session->private_key = private_key;
  session->prev_addr = addr;
  session->expire_timestamp = 10;
  session->server_to_client_sequence = 0;

  map.set(header.hash(), session);

  size_t index = crypto::PACKET_HASH_LENGTH;
  check(header.write(packet, index, Direction::ServerToClient, private_key));

  core::handlers::continue_response_handler(packet, map, recorder, router_info, socket, true);
  size_t prev_len = packet.length;
  check(socket.recv(packet)).onFail([&] {
    std::cout << "unable to receive packet\n";
  });
  check(prev_len == packet.length);

  check(recorder.continue_response_tx.num_packets == 1);
  check(recorder.continue_response_tx.num_bytes == packet.length).onFail([&] {
    std::cout << "packet len = " << packet.length << '\n';
    std::cout << "byte count = " << recorder.continue_response_rx.num_bytes << '\n';
  });

  core::handlers::continue_response_handler(packet, map, recorder, router_info, socket, true);
  check(!socket.recv(packet));
}
