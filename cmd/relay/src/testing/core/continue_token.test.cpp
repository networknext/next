#include "includes.h"
#include "testing/test.hpp"

#include "crypto/bytes.hpp"

#include "core/continue_token.hpp"

namespace
{
  core::ContinueToken makeToken()
  {
    static core::RouterInfo info;
    return core::ContinueToken(info);
  }
}  // namespace

Test(core_ContinueToken_general)
{
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

  core::ContinueToken inputToken = std::move(makeToken());
  {
    inputToken.ExpireTimestamp = ExpireTimestamp;
    inputToken.SessionID = SessionID;
    inputToken.SessionVersion = SessionVersion;
    inputToken.SessionFlags = SessionFlags;
  }

  {
    size_t index = 0;
    check(
     inputToken.writeEncrypted(packet.Buffer.data(), packet.Buffer.size(), index, sender_private_key, receiver_public_key));
  }

  core::ContinueToken outputToken = std::move(makeToken());
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

  // assert input == output
  check(inputToken.ExpireTimestamp == outputToken.ExpireTimestamp);
  check(inputToken.SessionID == outputToken.SessionID);
  check(inputToken.SessionVersion == outputToken.SessionVersion);
  check(inputToken.SessionFlags == outputToken.SessionFlags);
}

Test(legacy_relay_continue_token_t_general)
{
  uint8_t buffer[core::ContinueToken::EncryptedByteSize];

  legacy::relay_continue_token_t input_token;
  memset(&input_token, 0, sizeof(input_token));

  input_token.expire_timestamp = 1234241431241LL;
  input_token.session_id = 1234241431241LL;
  input_token.session_version = 5;
  input_token.session_flags = 1;

  legacy::relay_write_continue_token(&input_token, buffer, core::ContinueToken::ByteSize);

  unsigned char sender_public_key[crypto_box_PUBLICKEYBYTES];
  unsigned char sender_private_key[crypto_box_SECRETKEYBYTES];
  crypto_box_keypair(sender_public_key, sender_private_key);

  unsigned char receiver_public_key[crypto_box_PUBLICKEYBYTES];
  unsigned char receiver_private_key[crypto_box_SECRETKEYBYTES];
  crypto_box_keypair(receiver_public_key, receiver_private_key);

  unsigned char nonce[crypto_box_NONCEBYTES];
  legacy::relay_random_bytes(nonce, crypto_box_NONCEBYTES);

  check(
   legacy::relay_encrypt_continue_token(sender_private_key, receiver_public_key, nonce, buffer, sizeof(buffer)) == RELAY_OK);

  check(legacy::relay_decrypt_continue_token(sender_public_key, receiver_private_key, nonce, buffer) == RELAY_OK);

  legacy::relay_continue_token_t output_token;

  legacy::relay_read_continue_token(&output_token, buffer);

  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
  check(input_token.session_flags == output_token.session_flags);

  uint8_t* p = buffer;

  check(relay_write_encrypted_continue_token(&p, &input_token, sender_private_key, receiver_public_key) == RELAY_OK);

  p = buffer;

  memset(&output_token, 0, sizeof(output_token));

  check(relay_read_encrypted_continue_token(&p, &output_token, sender_public_key, receiver_private_key) == RELAY_OK);

  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_flags == output_token.session_flags);
}