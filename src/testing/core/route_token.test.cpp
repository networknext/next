#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "core/route_token.hpp"

TEST(core_RouteTokenV4_general)
{
  core::RouterInfo info;
  core::Packet packet;
  packet.length = packet.buffer.size();

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> sender_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> sender_private_key;
  crypto_box_keypair(sender_public_key.data(), sender_private_key.data());

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> receiver_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> receiver_private_key;
  crypto_box_keypair(receiver_public_key.data(), receiver_private_key.data());

  std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
  crypto::random_bytes(nonce, crypto_box_NONCEBYTES);

  const auto expire_timestamp = random_whole<uint64_t>();
  const auto session_id = random_whole<uint64_t>();
  const auto session_version = random_whole<uint8_t>();
  const auto kbps_up = random_whole<uint32_t>();
  const auto kbps_down = random_whole<uint32_t>();
  net::Address NextAddr;
  NextAddr.type = net::AddressType::IPv4;
  NextAddr.ipv4[0] = random_whole<uint8_t>();
  NextAddr.ipv4[1] = random_whole<uint8_t>();
  NextAddr.ipv4[2] = random_whole<uint8_t>();
  NextAddr.ipv4[3] = random_whole<uint8_t>();
  NextAddr.port = random_whole<uint32_t>();
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
  crypto::random_bytes(PrivateKey, PrivateKey.size());

  core::RouteTokenV4 input_token;
  {
    input_token.expire_timestamp = expire_timestamp;
    input_token.session_id = session_id;
    input_token.session_version = session_version;
    input_token.kbps_up = kbps_up;
    input_token.kbps_down = kbps_down;
    input_token.next_addr = NextAddr;
    input_token.private_key = PrivateKey;
  }

  {
    size_t index = 0;
    CHECK(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
  }

  core::RouteTokenV4 output_token;
  {
    size_t index = 0;
    CHECK(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  CHECK(input_token.expire_timestamp == expire_timestamp);
  CHECK(input_token.session_id == session_id);
  CHECK(input_token.session_version == session_version);
  CHECK(input_token.kbps_up == kbps_up);
  CHECK(input_token.kbps_down == kbps_down);
  CHECK(input_token.private_key == PrivateKey);
  CHECK(input_token.next_addr == NextAddr);

  // make sure input == output
  CHECK(input_token.expire_timestamp == output_token.expire_timestamp);
  CHECK(input_token.session_id == output_token.session_id);
  CHECK(input_token.session_version == output_token.session_version);
  CHECK(input_token.kbps_up == output_token.kbps_up);
  CHECK(input_token.kbps_down == output_token.kbps_down);
  CHECK(input_token.private_key == output_token.private_key);
  CHECK(input_token.next_addr == output_token.next_addr);
}

TEST(core_RouteToken_general)
{
  core::RouterInfo info;
  core::Packet packet;
  packet.length = packet.buffer.size();

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> sender_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> sender_private_key;
  crypto_box_keypair(sender_public_key.data(), sender_private_key.data());

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> receiver_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> receiver_private_key;
  crypto_box_keypair(receiver_public_key.data(), receiver_private_key.data());

  std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
  crypto::random_bytes(nonce, crypto_box_NONCEBYTES);

  const auto ExpireTimestamp = random_whole<uint64_t>();
  const auto SessionID = random_whole<uint64_t>();
  const auto SessionVersion = random_whole<uint8_t>();
  const auto SessionFlags = random_whole<uint8_t>();
  const auto KbpsUp = random_whole<uint32_t>();
  const auto KbpsDown = random_whole<uint32_t>();
  net::Address NextAddr;
  NextAddr.type = net::AddressType::IPv4;
  NextAddr.ipv4[0] = random_whole<uint8_t>();
  NextAddr.ipv4[1] = random_whole<uint8_t>();
  NextAddr.ipv4[2] = random_whole<uint8_t>();
  NextAddr.ipv4[3] = random_whole<uint8_t>();
  NextAddr.port = random_whole<uint32_t>();
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
  crypto::random_bytes(PrivateKey, PrivateKey.size());

  core::RouteToken input_token;
  {
    input_token.expire_timestamp = ExpireTimestamp;
    input_token.session_id = SessionID;
    input_token.session_version = SessionVersion;
    input_token.session_flags = SessionFlags;
    input_token.kbps_up = KbpsUp;
    input_token.kbps_down = KbpsDown;
    input_token.next_addr = NextAddr;
    input_token.private_key = PrivateKey;
  }

  {
    size_t index = 0;
    CHECK(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
  }

  core::RouteToken output_token;
  {
    size_t index = 0;
    CHECK(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  CHECK(input_token.expire_timestamp == ExpireTimestamp);
  CHECK(input_token.session_id == SessionID);
  CHECK(input_token.session_version == SessionVersion);
  CHECK(input_token.session_flags == SessionFlags);
  CHECK(input_token.kbps_up == KbpsUp);
  CHECK(input_token.kbps_down == KbpsDown);
  CHECK(input_token.private_key == PrivateKey);
  CHECK(input_token.next_addr == NextAddr);

  // make sure input == output
  CHECK(input_token.expire_timestamp == output_token.expire_timestamp);
  CHECK(input_token.session_id == output_token.session_id);
  CHECK(input_token.session_version == output_token.session_version);
  CHECK(input_token.session_flags == output_token.session_flags);
  CHECK(input_token.kbps_up == output_token.kbps_up);
  CHECK(input_token.kbps_down == output_token.kbps_down);
  CHECK(input_token.private_key == output_token.private_key);
  CHECK(input_token.next_addr == output_token.next_addr);
}
