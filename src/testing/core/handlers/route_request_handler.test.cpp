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

  packet.Len = 1 + RouteToken::EncryptedByteSize * 2;
  packet.Addr = from;

  RouteToken token;
  token.KbpsUp = crypto::Random<uint32_t>();
  token.KbpsDown = crypto::Random<uint32_t>();
  token.NextAddr = next;
  token.PrivateKey = random_private_key();
  token.SessionID = 123456789;
  token.SessionVersion = 123;
  token.SessionFlags = 234;
  token.ExpireTimestamp = 10;

  size_t index = 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.RelayPublicKey));

  check(map.get(token.hash()) == nullptr);

  core::handlers::route_request_handler(packet, keychain, map, recorder, router_info, socket, false);

  check(map.get(token.hash()) != nullptr);

  auto session = map.get(token.hash());

  check(session->ExpireTimestamp == token.ExpireTimestamp);
  check(session->SessionID == token.SessionID);
  check(session->SessionVersion == token.SessionVersion);
  check(session->KbpsUp == token.KbpsUp);
  check(session->KbpsDown == token.KbpsDown);
  check(session->PrevAddr == from);
  check(session->NextAddr == token.NextAddr);
  check(session->PrivateKey == token.PrivateKey);

  check(recorder.RouteRequestTx.PacketCount == 1);
  check(recorder.RouteRequestTx.ByteCount == packet.Len - RouteToken::EncryptedByteSize);

  size_t prev_len = packet.Len;
  check(next_socket.recv(packet));
  check(packet.Addr == addr).onFail([&] {
    std::cout << "addr = " << addr << '\n';
    std::cout << "next = " << next << '\n';
    std::cout << "packet = " << packet.Addr << '\n';
  });
  check(packet.Len == prev_len - RouteToken::EncryptedByteSize);

  check(!crypto::IsNetworkNextPacket(packet.Buffer, packet.Len));
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

  packet.Len = crypto::PacketHashLength + 1 + RouteToken::EncryptedByteSize * 2;
  packet.Addr = from;

  RouteToken token;
  token.KbpsUp = crypto::Random<uint32_t>();
  token.KbpsDown = crypto::Random<uint32_t>();
  token.NextAddr = next;
  token.PrivateKey = random_private_key();
  token.SessionID = 123456789;
  token.SessionVersion = 123;
  token.SessionFlags = 234;
  token.ExpireTimestamp = 10;

  size_t index = crypto::PacketHashLength + 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.RelayPublicKey));

  check(map.get(token.hash()) == nullptr);

  core::handlers::route_request_handler(packet, keychain, map, recorder, router_info, socket, true);

  check(map.get(token.hash()) != nullptr);

  auto session = map.get(token.hash());

  check(session->ExpireTimestamp == token.ExpireTimestamp);
  check(session->SessionID == token.SessionID);
  check(session->SessionVersion == token.SessionVersion);
  check(session->KbpsUp == token.KbpsUp);
  check(session->KbpsDown == token.KbpsDown);
  check(session->PrevAddr == from);
  check(session->NextAddr == token.NextAddr);
  check(session->PrivateKey == token.PrivateKey);

  check(recorder.RouteRequestTx.PacketCount == 1);
  check(recorder.RouteRequestTx.ByteCount == packet.Len - RouteToken::EncryptedByteSize);

  size_t prev_len = packet.Len;
  check(next_socket.recv(packet));
  check(packet.Addr == addr);
  check(packet.Len == prev_len - RouteToken::EncryptedByteSize);

  check(crypto::IsNetworkNextPacket(packet.Buffer, packet.Len));
}
