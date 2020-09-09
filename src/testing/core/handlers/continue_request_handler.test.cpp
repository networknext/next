#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/continue_request_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::ContinueToken;
using core::ContinueTokenV4;
using core::Packet;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::PacketType;
using crypto::Keychain;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_continue_request_handler_sdk4_unsigned)
{
  Packet packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.set_timestamp(0);
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.buffer[0] = static_cast<uint8_t>(PacketType::ContinueRequest4);
  packet.length = 1 + ContinueTokenV4::SIZE_OF_ENCRYPTED * 2;

  ContinueTokenV4 token;
  token.expire_timestamp = 20;
  token.session_id = 0x13;
  token.session_version = 3;

  size_t index = 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));
  check(packet.buffer[0] == static_cast<uint8_t>(PacketType::ContinueRequest4));
  check(index == 1 + ContinueTokenV4::SIZE_OF_ENCRYPTED).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>();
  session->expire_timestamp = 10;
  session->session_id = token.session_id;
  session->session_version = token.session_version;
  session->next_addr = addr;
  session->client_to_server_sequence = 0;
  map.set(token.hash(), session);

  size_t prev_len = packet.length;

  core::handlers::continue_request_handler_sdk4(packet, map, keychain, recorder, info, socket, false);

  check(socket.recv(packet));
  check(packet.length == prev_len - ContinueTokenV4::SIZE_OF_ENCRYPTED);
  check(session->expire_timestamp == token.expire_timestamp);

  index = 0;
  check(!crypto::is_network_next_packet_sdk4(packet.buffer, index, packet.length));
}

Test(core_handlers_continue_request_handler_sdk4_signed)
{
  Packet packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.set_timestamp(0);
  Socket socket;

  packet.buffer[crypto::PACKET_HASH_LENGTH] = static_cast<uint8_t>(PacketType::ContinueRequest);
  packet.length = crypto::PACKET_HASH_LENGTH + 1 + ContinueTokenV4::SIZE_OF_ENCRYPTED * 2;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  ContinueTokenV4 token;
  token.expire_timestamp = 20;
  token.session_id = 0x13;
  token.session_version = 3;

  size_t index = crypto::PACKET_HASH_LENGTH + 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));
  check(packet.buffer[crypto::PACKET_HASH_LENGTH] == static_cast<uint8_t>(PacketType::ContinueRequest));
  check(index == crypto::PACKET_HASH_LENGTH + 1 + ContinueTokenV4::SIZE_OF_ENCRYPTED).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>();
  session->expire_timestamp = 10;
  session->session_id = token.session_id;
  session->session_version = token.session_version;
  session->next_addr = addr;
  session->client_to_server_sequence = 0;
  map.set(token.hash(), session);

  size_t prev_len = packet.length;

  core::handlers::continue_request_handler_sdk4(packet, map, keychain, recorder, info, socket, true);

  check(socket.recv(packet));
  check(packet.length == prev_len - ContinueTokenV4::SIZE_OF_ENCRYPTED);
  check(session->expire_timestamp == token.expire_timestamp);

  index = 0;
  check(crypto::is_network_next_packet_sdk4(packet.buffer, index, packet.length));
}

Test(core_handlers_continue_request_handler_unsigned)
{
  Packet packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.set_timestamp(0);
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.buffer[0] = static_cast<uint8_t>(PacketType::ContinueRequest);
  packet.length = 1 + ContinueToken::SIZE_OF_ENCRYPTED * 2;

  ContinueToken token;
  token.expire_timestamp = 20;
  token.session_id = 0x13;
  token.session_version = 3;
  token.session_flags = 0;

  size_t index = 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));
  check(packet.buffer[0] == static_cast<uint8_t>(PacketType::ContinueRequest));
  check(index == 1 + ContinueToken::SIZE_OF_ENCRYPTED).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>();
  session->expire_timestamp = 10;
  session->session_id = token.session_id;
  session->session_version = token.session_version;
  session->next_addr = addr;
  session->client_to_server_sequence = 0;
  map.set(token.hash(), session);

  size_t prev_len = packet.length;

  core::handlers::continue_request_handler(packet, map, keychain, recorder, info, socket, false);

  check(socket.recv(packet));
  check(packet.length == prev_len - ContinueToken::SIZE_OF_ENCRYPTED);
  check(session->expire_timestamp == token.expire_timestamp);

  index = 0;
  check(!crypto::is_network_next_packet(packet.buffer, index, packet.length));
}

Test(core_handlers_continue_request_handler_signed)
{
  Packet packet;
  SessionMap map;
  Keychain keychain = testing::make_keychain();
  ThroughputRecorder recorder;
  RouterInfo info;
  info.set_timestamp(0);
  Socket socket;

  packet.buffer[crypto::PACKET_HASH_LENGTH] = static_cast<uint8_t>(PacketType::ContinueRequest);
  packet.length = crypto::PACKET_HASH_LENGTH + 1 + ContinueToken::SIZE_OF_ENCRYPTED * 2;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  ContinueToken token;
  token.expire_timestamp = 20;
  token.session_id = 0x13;
  token.session_version = 3;
  token.session_flags = 0;

  size_t index = crypto::PACKET_HASH_LENGTH + 1;
  check(token.write_encrypted(packet, index, router_private_key(), keychain.relay_public_key));
  check(packet.buffer[crypto::PACKET_HASH_LENGTH] == static_cast<uint8_t>(PacketType::ContinueRequest));
  check(index == crypto::PACKET_HASH_LENGTH + 1 + ContinueToken::SIZE_OF_ENCRYPTED).onFail([&] {
    std::cout << index << '\n';
  });

  auto session = std::make_shared<Session>();
  session->expire_timestamp = 10;
  session->session_id = token.session_id;
  session->session_version = token.session_version;
  session->next_addr = addr;
  session->client_to_server_sequence = 0;
  map.set(token.hash(), session);

  size_t prev_len = packet.length;

  core::handlers::continue_request_handler(packet, map, keychain, recorder, info, socket, true);

  check(socket.recv(packet));
  check(packet.length == prev_len - ContinueToken::SIZE_OF_ENCRYPTED);
  check(session->expire_timestamp == token.expire_timestamp);

  index = 0;
  check(crypto::is_network_next_packet(packet.buffer, index, packet.length));
}
