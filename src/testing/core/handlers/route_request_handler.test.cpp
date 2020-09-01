#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/route_request_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS ;
#include "testing/helpers.hpp"

using core::Packet;
using core::RouterInfo;
using core::RouteToken;
using core::SessionMap;
using crypto::Keychain;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_route_request_handler_unsigned)
{
  Packet packet;
  Keychain keychain = make_keychain();
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket from_socket, socket, next_socket;

  router_info.setTimestamp(0);

  Address from, addr, next;
  SocketConfig config = default_socket_config();

  check(from.parse("127.0.0.1"));
  check(from_socket.create(from, config));  // only to assign a port

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  check(next.parse("127.0.0.1"));
  check(next_socket.create(next, config));

  packet.length = 1 + RouteToken::EncryptedByteSize * 2;
  packet.addr = from;

  RouteToken token;
  token.KbpsUp = crypto::Random<uint32_t>();
  token.KbpsDown = crypto::Random<uint32_t>();
  token.NextAddr = next;
  token.PrivateKey = random_private_key();
  token.SessionID = 123456789;
  token.SessionVersion = 123;
  token.SessionFlags = 234;
  token.expire_timestamp = 10;

  size_t index = 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));

  check(map.get(token.hash()) == nullptr);

  core::handlers::route_request_handler(packet, keychain, map, recorder, router_info, socket, false);

  check(map.get(token.hash()) != nullptr);

  auto session = map.get(token.hash());

  check(session->expire_timestamp == token.expire_timestamp);
  check(session->session_id == token.SessionID);
  check(session->session_version == token.SessionVersion);
  check(session->kbps_up == token.KbpsUp);
  check(session->kbps_down == token.KbpsDown);
  check(session->prev_addr == from);
  check(session->next_addr == token.NextAddr);
  check(session->private_key == token.PrivateKey);

  check(recorder.route_request_tx.num_packets == 1);
  check(recorder.route_request_tx.num_bytes == packet.length - RouteToken::EncryptedByteSize);

  size_t prev_len = packet.length;
  check(next_socket.recv(packet));
  check(packet.addr == addr).onFail([&] {
    std::cout << "addr = " << addr << '\n';
    std::cout << "next = " << next << '\n';
    std::cout << "packet = " << packet.addr << '\n';
  });
  check(packet.length == prev_len - RouteToken::EncryptedByteSize);

  check(!crypto::is_network_next_packet(packet.buffer, packet.length));
}

Test(core_handlers_route_request_handler_signed)
{
  Packet packet;
  Keychain keychain = make_keychain();
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket from_socket, socket, next_socket;

  router_info.setTimestamp(0);

  Address from, addr, next;
  SocketConfig config = default_socket_config();

  check(from.parse("127.0.0.1"));
  check(from_socket.create(from, config));  // only to assign a port

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  check(next.parse("127.0.0.1"));
  check(next_socket.create(next, config));

  packet.length = crypto::PACKET_HASH_LENGTH + 1 + RouteToken::EncryptedByteSize * 2;
  packet.addr = from;

  RouteToken token;
  token.KbpsUp = crypto::Random<uint32_t>();
  token.KbpsDown = crypto::Random<uint32_t>();
  token.NextAddr = next;
  token.PrivateKey = random_private_key();
  token.SessionID = 123456789;
  token.SessionVersion = 123;
  token.SessionFlags = 234;
  token.expire_timestamp = 10;

  size_t index = crypto::PACKET_HASH_LENGTH + 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));

  check(map.get(token.hash()) == nullptr);

  core::handlers::route_request_handler(packet, keychain, map, recorder, router_info, socket, true);

  check(map.get(token.hash()) != nullptr);

  auto session = map.get(token.hash());

  check(session->expire_timestamp == token.expire_timestamp);
  check(session->session_id == token.SessionID);
  check(session->session_version == token.SessionVersion);
  check(session->kbps_up == token.KbpsUp);
  check(session->kbps_down == token.KbpsDown);
  check(session->prev_addr == from);
  check(session->next_addr == token.NextAddr);
  check(session->private_key == token.PrivateKey);

  check(recorder.route_request_tx.num_packets == 1);
  check(recorder.route_request_tx.num_bytes == packet.length - RouteToken::EncryptedByteSize);

  size_t prev_len = packet.length;
  check(next_socket.recv(packet));
  check(packet.addr == addr);
  check(packet.length == prev_len - RouteToken::EncryptedByteSize);

  check(crypto::is_network_next_packet(packet.buffer, packet.length));
}
