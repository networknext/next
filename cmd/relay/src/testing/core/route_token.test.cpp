#include "includes.h"
#include "testing/test.hpp"

#include "crypto/bytes.hpp"

#include "core/route_token.hpp"

Test(core_RouteToken_general)
{
  core::RouterInfo info;
  core::GenericPacket<> packet;

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> sender_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> sender_private_key;
  crypto_box_keypair(sender_public_key.data(), sender_private_key.data());

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> receiver_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> receiver_private_key;
  crypto_box_keypair(receiver_public_key.data(), receiver_private_key.data());

  std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
  crypto::RandomBytes(nonce, crypto_box_NONCEBYTES);

  const auto ExpireTimestamp = crypto::Random<uint64_t>();
  const auto SessionID = crypto::Random<uint64_t>();
  const auto SessionVersion = crypto::Random<uint8_t>();
  const auto SessionFlags = crypto::Random<uint8_t>();
  const auto KbpsUp = crypto::Random<uint32_t>();
  const auto KbpsDown = crypto::Random<uint32_t>();
  net::Address NextAddr;
  NextAddr.Type = net::AddressType::IPv4;
  NextAddr.IPv4[0] = crypto::Random<uint8_t>();
  NextAddr.IPv4[1] = crypto::Random<uint8_t>();
  NextAddr.IPv4[2] = crypto::Random<uint8_t>();
  NextAddr.IPv4[3] = crypto::Random<uint8_t>();
  NextAddr.Port = crypto::Random<uint32_t>();
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
  crypto::RandomBytes(PrivateKey, PrivateKey.size());

  core::RouteToken inputToken(info);
  {
    inputToken.ExpireTimestamp = ExpireTimestamp;
    inputToken.SessionID = SessionID;
    inputToken.SessionVersion = SessionVersion;
    inputToken.SessionFlags = SessionFlags;
    inputToken.KbpsUp = KbpsUp;
    inputToken.KbpsDown = KbpsDown;
    inputToken.NextAddr = NextAddr;
    inputToken.PrivateKey = PrivateKey;
  }

  {
    size_t index = 0;
    check(
     inputToken.writeEncrypted(packet.Buffer.data(), packet.Buffer.size(), index, sender_private_key, receiver_public_key));
  }

  core::RouteToken outputToken(info);
  {
    size_t index = 0;
    check(
     outputToken.readEncrypted(packet.Buffer.data(), packet.Buffer.size(), index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  check(inputToken.ExpireTimestamp == ExpireTimestamp);
  check(inputToken.SessionID == SessionID);
  check(inputToken.SessionVersion == SessionVersion);
  check(inputToken.SessionFlags == SessionFlags);
  check(inputToken.KbpsUp == KbpsUp);
  check(inputToken.KbpsDown == KbpsDown);
  check(inputToken.PrivateKey == PrivateKey);
  check(inputToken.NextAddr == NextAddr);

  // make sure input == output
  check(inputToken.ExpireTimestamp == outputToken.ExpireTimestamp);
  check(inputToken.SessionID == outputToken.SessionID);
  check(inputToken.SessionVersion == outputToken.SessionVersion);
  check(inputToken.SessionFlags == outputToken.SessionFlags);
  check(inputToken.KbpsUp == outputToken.KbpsUp);
  check(inputToken.KbpsDown == outputToken.KbpsDown);
  check(inputToken.PrivateKey == outputToken.PrivateKey);
  check(inputToken.NextAddr == outputToken.NextAddr);
}

Test(legacy_relay_route_token_t_general)
{
  uint8_t buffer[core::RouteToken::EncryptedByteSize];

  legacy::relay_route_token_t input_token;
  memset(&input_token, 0, sizeof(input_token));

  input_token.expire_timestamp = 1234241431241LL;
  input_token.session_id = 1234241431241LL;
  input_token.session_version = 5;
  input_token.session_flags = 1;
  input_token.next_address.type = static_cast<uint8_t>(net::AddressType::IPv4);
  input_token.next_address.data.ipv4[0] = 127;
  input_token.next_address.data.ipv4[1] = 0;
  input_token.next_address.data.ipv4[2] = 0;
  input_token.next_address.data.ipv4[3] = 1;
  input_token.next_address.port = 40000;

  legacy::relay_write_route_token(&input_token, buffer, core::RouteToken::ByteSize);

  unsigned char sender_public_key[crypto_box_PUBLICKEYBYTES];
  unsigned char sender_private_key[crypto_box_SECRETKEYBYTES];
  crypto_box_keypair(sender_public_key, sender_private_key);

  unsigned char receiver_public_key[crypto_box_PUBLICKEYBYTES];
  unsigned char receiver_private_key[crypto_box_SECRETKEYBYTES];
  crypto_box_keypair(receiver_public_key, receiver_private_key);

  unsigned char nonce[crypto_box_NONCEBYTES];
  legacy::relay_random_bytes(nonce, crypto_box_NONCEBYTES);

  check(legacy::relay_encrypt_route_token(sender_private_key, receiver_public_key, nonce, buffer, sizeof(buffer)) == RELAY_OK);

  check(legacy::relay_decrypt_route_token(sender_public_key, receiver_private_key, nonce, buffer) == RELAY_OK);

  legacy::relay_route_token_t output_token;

  legacy::relay_read_route_token(&output_token, buffer);

  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
  check(input_token.session_flags == output_token.session_flags);
  check(input_token.kbps_up == output_token.kbps_up);
  check(input_token.kbps_down == output_token.kbps_down);
  check(memcmp(input_token.private_key, output_token.private_key, crypto_box_SECRETKEYBYTES) == 0);
  check(legacy::relay_address_equal(&input_token.next_address, &output_token.next_address) == 1);

  uint8_t* p = buffer;

  check(relay_write_encrypted_route_token(&p, &input_token, sender_private_key, receiver_public_key) == RELAY_OK);

  p = buffer;

  check(relay_read_encrypted_route_token(&p, &output_token, sender_public_key, receiver_private_key) == RELAY_OK);

  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
  check(input_token.session_flags == output_token.session_flags);
  check(input_token.kbps_up == output_token.kbps_up);
  check(input_token.kbps_down == output_token.kbps_down);
  check(memcmp(input_token.private_key, output_token.private_key, crypto_box_SECRETKEYBYTES) == 0);
  check(legacy::relay_address_equal(&input_token.next_address, &output_token.next_address) == 1);
}