#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/route_request_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::RouterInfo;
using core::RouteToken;
using core::SessionMap;
using crypto::Keychain;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_route_request_handler_sdk4)
{
  Packet packet;
  Keychain keychain = make_keychain();
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket from_socket, socket, next_socket;

  router_info.set_timestamp(0);

  Address from, addr, next;
  SocketConfig config = default_socket_config();

  CHECK(from.parse("127.0.0.1"));
  CHECK(from_socket.create(from, config));  // only to assign a port

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  CHECK(next.parse("127.0.0.1"));
  CHECK(next_socket.create(next, config));

  packet.length = 1 + RouteToken::SIZE_OF_SIGNED * 2;
  packet.addr = from;

  RouteToken token;
  token.kbps_up = random_whole<uint32_t>();
  token.kbps_down = random_whole<uint32_t>();
  token.next_addr = next;
  token.private_key = random_private_key();
  token.session_id = 123456789;
  token.session_version = 123;
  token.expire_timestamp = 10;

  size_t index = 1;
  CHECK(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));

  CHECK(map.get(token.hash()) == nullptr);

  core::handlers::route_request_handler_sdk4(packet, keychain, map, recorder, router_info, socket);

  CHECK(map.get(token.hash()) != nullptr);

  auto session = map.get(token.hash());

  CHECK(session->expire_timestamp == token.expire_timestamp);
  CHECK(session->session_id == token.session_id);
  CHECK(session->session_version == token.session_version);
  CHECK(session->kbps_up == token.kbps_up);
  CHECK(session->kbps_down == token.kbps_down);
  CHECK(session->prev_addr == from);
  CHECK(session->next_addr == token.next_addr);
  CHECK(session->private_key == token.private_key);

  CHECK(recorder.route_request_tx.num_packets == 1);
  CHECK(recorder.route_request_tx.num_bytes == packet.length - RouteToken::SIZE_OF_SIGNED).on_fail([&] {
    std::cout << "len = " << packet.length << '\n';
    std::cout << "bytes = " << recorder.route_request_tx.num_bytes << '\n';
  });

  size_t prev_len = packet.length;
  CHECK(next_socket.recv(packet));
  CHECK(packet.addr == addr).on_fail([&] {
    std::cout << "addr = " << addr << '\n';
    std::cout << "next = " << next << '\n';
    std::cout << "packet = " << packet.addr << '\n';
  });
  CHECK(packet.length == prev_len - RouteToken::SIZE_OF_SIGNED);
}
