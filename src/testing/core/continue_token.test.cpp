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
    check(inputToken.write_encrypted(packet, index, sender_private_key, receiver_public_key));
  }

  core::ContinueToken outputToken = std::move(makeToken());
  {
    size_t index = 0;
    check(outputToken.read_encrypted(packet, index, sender_public_key, receiver_private_key));
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
