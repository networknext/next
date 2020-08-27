#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/continue_request_handler.hpp"

#define CRYPTO_HELPERS
#include "testing/helpers.hpp"

using core::GenericPacket;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::packets::Type;
using net::Address;
using os::Socket;
using util::ThroughputRecorder;
using crypto::Keychain;
using core::ContinueToken;
using core::packets::Type;

Test(core_handlers_continue_request_handler_unsigned) {
  GenericPacket<> packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.setTimestamp(0);
  Socket socket;

  Address addr;

  packet.Buffer[0] = static_cast<uint8_t>(Type::ContinueRequest);
  packet.Len = 1 + ContinueToken::EncryptedByteSize * 2;

  check(addr.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, addr, 64 * 1024, 64 * 1024, 0.0, false));

  ContinueToken token(info);
  token.ExpireTimestamp = 20;
  token.SessionID = 0x13;
  token.SessionVersion = 3;

  size_t index = 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.RelayPublicKey));
  check(packet.Buffer[0] == static_cast<uint8_t>(Type::ContinueRequest));
  check(index == 1 + ContinueToken::EncryptedByteSize).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>(info);
  session->ExpireTimestamp = 10;
  session->SessionID = token.SessionID;
  session->SessionVersion = token.SessionVersion;
  session->NextAddr = addr;
  map.set(token.hash(), session);

  size_t prev_len = packet.Len;

  core::handlers::continue_request_handler(packet, map, keychain, recorder, info, socket, false);

  check(socket.recv(packet));
  check(packet.Len == prev_len - ContinueToken::EncryptedByteSize);
  check(session->ExpireTimestamp == token.ExpireTimestamp);
}