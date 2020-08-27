#include "includes.h"
#include "testing/test.hpp"

#include "crypto/bytes.hpp"

#include "core/continue_token.hpp"

using core::ContinueToken;
using core::GenericPacket;
using core::RouterInfo;

namespace
{
  core::ContinueToken make_token()
  {
    static RouterInfo info;
    return ContinueToken(info);
  }
}  // namespace

Test(core_ContinueToken_general)
{
  GenericPacket<> packet;
  packet.Len = packet.Buffer.size();

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

  ContinueToken input_token = std::move(make_token());
  {
    input_token.ExpireTimestamp = ExpireTimestamp;
    input_token.SessionID = SessionID;
    input_token.SessionVersion = SessionVersion;
    input_token.SessionFlags = SessionFlags;
  }

  {
    size_t index = 0;
    check(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
    check(index == ContinueToken::EncryptedByteSize);
  }

  ContinueToken output_token = std::move(make_token());
  {
    size_t index = 0;
    check(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  check(input_token.ExpireTimestamp == ExpireTimestamp);
  check(input_token.SessionID == SessionID);
  check(input_token.SessionVersion == SessionVersion);
  check(input_token.SessionFlags == SessionFlags);

  // assert input == output
  check(input_token.ExpireTimestamp == output_token.ExpireTimestamp);
  check(input_token.SessionID == output_token.SessionID);
  check(input_token.SessionVersion == output_token.SessionVersion);
  check(input_token.SessionFlags == output_token.SessionFlags);
}

Test(core_ContinueToken_write)
{
  GenericPacket<> packet;
  ContinueToken token = std::move(make_token());

  token.ExpireTimestamp = 6;
  token.SessionID = 1;
  token.SessionVersion = 2;
  token.SessionFlags = 3;

  size_t index = 0;
  check(token.write(packet, index));
  check(index == ContinueToken::ByteSize);

  uint64_t expire;
  uint64_t id;
  uint8_t ver;
  uint8_t flags;

  index = 0;
  check(encoding::ReadUint64(packet.Buffer, index, expire));
  check(encoding::ReadUint64(packet.Buffer, index, id));
  check(encoding::ReadUint8(packet.Buffer, index, ver));
  check(encoding::ReadUint8(packet.Buffer, index, flags));

  check(expire == 6);
  check(id == 1);
  check(ver == 2);
  check(flags == 3);
}

Test(core_ContinueToken_read)
{
  GenericPacket<> packet;
  ContinueToken token = std::move(make_token());

  token.ExpireTimestamp = 6;
  token.SessionID = 1;
  token.SessionVersion = 2;
  token.SessionFlags = 3;

  size_t index = 0;
  check(token.write(packet, index));

  ContinueToken other = std::move(make_token());

  index = 0;
  check(other.read(packet, index));
  check(index == ContinueToken::ByteSize);

  check(token == other);
}
