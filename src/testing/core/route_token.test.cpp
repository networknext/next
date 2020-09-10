#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "core/route_token.hpp"

Test(core_RouteTokenV4_general)
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
  crypto::RandomBytes(nonce, crypto_box_NONCEBYTES);

  const auto ExpireTimestamp = RandomWhole<uint64_t>();
  const auto SessionID = RandomWhole<uint64_t>();
  const auto SessionVersion = RandomWhole<uint8_t>();
  const auto KbpsUp = RandomWhole<uint32_t>();
  const auto KbpsDown = RandomWhole<uint32_t>();
  net::Address NextAddr;
  NextAddr.type = net::AddressType::IPv4;
  NextAddr.ipv4[0] = RandomWhole<uint8_t>();
  NextAddr.ipv4[1] = RandomWhole<uint8_t>();
  NextAddr.ipv4[2] = RandomWhole<uint8_t>();
  NextAddr.ipv4[3] = RandomWhole<uint8_t>();
  NextAddr.port = RandomWhole<uint32_t>();
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
  crypto::RandomBytes(PrivateKey, PrivateKey.size());

  core::RouteTokenV4 input_token;
  {
    input_token.expire_timestamp = ExpireTimestamp;
    input_token.session_id = SessionID;
    input_token.session_version = SessionVersion;
    input_token.KbpsUp = KbpsUp;
    input_token.KbpsDown = KbpsDown;
    input_token.NextAddr = NextAddr;
    input_token.PrivateKey = PrivateKey;
  }

  {
    size_t index = 0;
    check(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
  }

  core::RouteTokenV4 output_token;
  {
    size_t index = 0;
    check(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  check(input_token.expire_timestamp == ExpireTimestamp);
  check(input_token.session_id == SessionID);
  check(input_token.session_version == SessionVersion);
  check(input_token.KbpsUp == KbpsUp);
  check(input_token.KbpsDown == KbpsDown);
  check(input_token.PrivateKey == PrivateKey);
  check(input_token.NextAddr == NextAddr);

  // make sure input == output
  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
  check(input_token.KbpsUp == output_token.KbpsUp);
  check(input_token.KbpsDown == output_token.KbpsDown);
  check(input_token.PrivateKey == output_token.PrivateKey);
  check(input_token.NextAddr == output_token.NextAddr);
}

Test(core_RouteToken_general)
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
  crypto::RandomBytes(nonce, crypto_box_NONCEBYTES);

  const auto ExpireTimestamp = RandomWhole<uint64_t>();
  const auto SessionID = RandomWhole<uint64_t>();
  const auto SessionVersion = RandomWhole<uint8_t>();
  const auto SessionFlags = RandomWhole<uint8_t>();
  const auto KbpsUp = RandomWhole<uint32_t>();
  const auto KbpsDown = RandomWhole<uint32_t>();
  net::Address NextAddr;
  NextAddr.type = net::AddressType::IPv4;
  NextAddr.ipv4[0] = RandomWhole<uint8_t>();
  NextAddr.ipv4[1] = RandomWhole<uint8_t>();
  NextAddr.ipv4[2] = RandomWhole<uint8_t>();
  NextAddr.ipv4[3] = RandomWhole<uint8_t>();
  NextAddr.port = RandomWhole<uint32_t>();
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
  crypto::RandomBytes(PrivateKey, PrivateKey.size());

  core::RouteToken input_token;
  {
    input_token.expire_timestamp = ExpireTimestamp;
    input_token.session_id = SessionID;
    input_token.session_version = SessionVersion;
    input_token.session_flags = SessionFlags;
    input_token.KbpsUp = KbpsUp;
    input_token.KbpsDown = KbpsDown;
    input_token.NextAddr = NextAddr;
    input_token.PrivateKey = PrivateKey;
  }

  {
    size_t index = 0;
    check(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
  }

  core::RouteToken output_token;
  {
    size_t index = 0;
    check(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  check(input_token.expire_timestamp == ExpireTimestamp);
  check(input_token.session_id == SessionID);
  check(input_token.session_version == SessionVersion);
  check(input_token.session_flags == SessionFlags);
  check(input_token.KbpsUp == KbpsUp);
  check(input_token.KbpsDown == KbpsDown);
  check(input_token.PrivateKey == PrivateKey);
  check(input_token.NextAddr == NextAddr);

  // make sure input == output
  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
  check(input_token.session_flags == output_token.session_flags);
  check(input_token.KbpsUp == output_token.KbpsUp);
  check(input_token.KbpsDown == output_token.KbpsDown);
  check(input_token.PrivateKey == output_token.PrivateKey);
  check(input_token.NextAddr == output_token.NextAddr);
}
