#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/continue_request_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::ContinueToken;
using core::Packet;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::packets::Type;
using crypto::Keychain;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_continue_request_handler_unsigned)
{
  Packet packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.setTimestamp(0);
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.Buffer[0] = static_cast<uint8_t>(Type::ContinueRequest);
  packet.Len = 1 + ContinueToken::EncryptedByteSize * 2;

  ContinueToken token;
  token.ExpireTimestamp = 20;
  token.SessionID = 0x13;
  token.SessionVersion = 3;

  size_t index = 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.RelayPublicKey));
  check(packet.Buffer[0] == static_cast<uint8_t>(Type::ContinueRequest));
  check(index == 1 + ContinueToken::EncryptedByteSize).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>();
  session->ExpireTimestamp = 10;
  session->SessionID = token.SessionID;
  session->SessionVersion = token.SessionVersion;
  session->NextAddr = addr;
  session->ClientToServerSeq = 0;
  map.set(token.hash(), session);

  size_t prev_len = packet.Len;

  core::handlers::continue_request_handler(packet, map, keychain, recorder, info, socket, false);

  check(socket.recv(packet));
  check(packet.Len == prev_len - ContinueToken::EncryptedByteSize);
  check(session->ExpireTimestamp == token.ExpireTimestamp);
}

Test(core_handlers_continue_request_handler_signed)
{
  Packet packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.setTimestamp(0);
  Socket socket;

  packet.Buffer[crypto::PacketHashLength] = static_cast<uint8_t>(Type::ContinueRequest);
  packet.Len = crypto::PacketHashLength + 1 + ContinueToken::EncryptedByteSize * 2;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  ContinueToken token;
  token.ExpireTimestamp = 20;
  token.SessionID = 0x13;
  token.SessionVersion = 3;

  size_t index = crypto::PacketHashLength + 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.RelayPublicKey));
  check(packet.Buffer[crypto::PacketHashLength] == static_cast<uint8_t>(Type::ContinueRequest));
  check(index == crypto::PacketHashLength + 1 + ContinueToken::EncryptedByteSize).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>();
  session->ExpireTimestamp = 10;
  session->SessionID = token.SessionID;
  session->SessionVersion = token.SessionVersion;
  session->NextAddr = addr;
  session->ClientToServerSeq = 0;
  map.set(token.hash(), session);

  size_t prev_len = packet.Len;

  core::handlers::continue_request_handler(packet, map, keychain, recorder, info, socket, true);

  check(socket.recv(packet));
  check(packet.Len == prev_len - ContinueToken::EncryptedByteSize);
  check(session->ExpireTimestamp == token.ExpireTimestamp);
}