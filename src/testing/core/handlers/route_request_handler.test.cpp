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
  Socket socket, next_socket;

  router_info.setTimestamp(0);

  Address addr, next;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  check(next.parse("127.0.0.1"));
  check(socket.create(next, config));

  packet.Len = 1 + RouteToken::EncryptedByteSize * 2;
  packet.Addr = addr;

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
  check(session->PrevAddr == addr);
  check(session->NextAddr == token.NextAddr);
  check(session->PrivateKey == token.PrivateKey);

  check(recorder.RouteRequestTx.PacketCount == 1);
  check(recorder.RouteRequestTx.ByteCount == packet.Len - RouteToken::EncryptedByteSize);

  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(packet.Addr == next);
  check(packet.Len == prev_len - RouteToken::EncryptedByteSize);
}