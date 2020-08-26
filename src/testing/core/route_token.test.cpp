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
    check(inputToken.write_encrypted(packet, index, sender_private_key, receiver_public_key));
  }

  core::RouteToken outputToken(info);
  {
    size_t index = 0;
    check(outputToken.read_encrypted(packet, index, sender_public_key, receiver_private_key));
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
