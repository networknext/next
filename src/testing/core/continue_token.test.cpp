#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "core/continue_token.hpp"

using core::ContinueToken;
using core::ContinueTokenV4;
using core::Packet;
using core::RouterInfo;

Test(core_ContinueTokenV4_general)
{
  Packet packet;
  packet.length = packet.buffer.size();

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> sender_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> sender_private_key;
  crypto_box_keypair(sender_public_key.data(), sender_private_key.data());

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> receiver_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> receiver_private_key;
  crypto_box_keypair(receiver_public_key.data(), receiver_private_key.data());

  std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
  crypto::RandomBytes(nonce, crypto_box_NONCEBYTES);

  const auto ExpireTimestamp = random_whole<uint64_t>();
  const auto SessionID = random_whole<uint64_t>();
  const auto SessionVersion = random_whole<uint8_t>();

  ContinueTokenV4 input_token;
  {
    input_token.expire_timestamp = ExpireTimestamp;
    input_token.session_id = SessionID;
    input_token.session_version = SessionVersion;
  }

  {
    size_t index = 0;
    check(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
    check(index == ContinueTokenV4::SIZE_OF_ENCRYPTED);
  }

  ContinueTokenV4 output_token;
  {
    size_t index = 0;
    check(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  check(input_token.expire_timestamp == ExpireTimestamp);
  check(input_token.session_id == SessionID);
  check(input_token.session_version == SessionVersion);

  // assert input == output
  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
}

Test(core_ContinueTokenV4_write)
{
  Packet packet;
  ContinueTokenV4 token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;

  size_t index = 0;
  check(token.write(packet, index));
  check(index == ContinueTokenV4::SIZE_OF);

  uint64_t expire;
  uint64_t id;
  uint8_t ver;

  index = 0;
  check(encoding::read_uint64(packet.buffer, index, expire));
  check(encoding::read_uint64(packet.buffer, index, id));
  check(encoding::read_uint8(packet.buffer, index, ver));

  check(expire == 6);
  check(id == 1);
  check(ver == 2);
}

Test(core_ContinueTokenV4_read)
{
  Packet packet;
  ContinueTokenV4 token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;

  size_t index = 0;
  check(token.write(packet, index));

  ContinueTokenV4 other;

  index = 0;
  check(other.read(packet, index));
  check(index == ContinueTokenV4::SIZE_OF);

  check(token == other);
}

Test(core_ContinueToken_general)
{
  Packet packet;
  packet.length = packet.buffer.size();

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> sender_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> sender_private_key;
  crypto_box_keypair(sender_public_key.data(), sender_private_key.data());

  std::array<uint8_t, crypto_box_PUBLICKEYBYTES> receiver_public_key;
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> receiver_private_key;
  crypto_box_keypair(receiver_public_key.data(), receiver_private_key.data());

  std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
  crypto::RandomBytes(nonce, crypto_box_NONCEBYTES);

  const auto ExpireTimestamp = random_whole<uint64_t>();
  const auto SessionID = random_whole<uint64_t>();
  const auto SessionVersion = random_whole<uint8_t>();
  const auto SessionFlags = random_whole<uint8_t>();

  ContinueToken input_token;
  {
    input_token.expire_timestamp = ExpireTimestamp;
    input_token.session_id = SessionID;
    input_token.session_version = SessionVersion;
    input_token.session_flags = SessionFlags;
  }

  {
    size_t index = 0;
    check(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
    check(index == ContinueToken::SIZE_OF_ENCRYPTED);
  }

  ContinueToken output_token;
  {
    size_t index = 0;
    check(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  check(input_token.expire_timestamp == ExpireTimestamp);
  check(input_token.session_id == SessionID);
  check(input_token.session_version == SessionVersion);
  check(input_token.session_flags == SessionFlags);

  // assert input == output
  check(input_token.expire_timestamp == output_token.expire_timestamp);
  check(input_token.session_id == output_token.session_id);
  check(input_token.session_version == output_token.session_version);
  check(input_token.session_flags == output_token.session_flags);
}

Test(core_ContinueToken_write)
{
  Packet packet;
  ContinueToken token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;
  token.session_flags = 3;

  size_t index = 0;
  check(token.write(packet, index));
  check(index == ContinueToken::SIZE_OF);

  uint64_t expire;
  uint64_t id;
  uint8_t ver;
  uint8_t flags;

  index = 0;
  check(encoding::read_uint64(packet.buffer, index, expire));
  check(encoding::read_uint64(packet.buffer, index, id));
  check(encoding::read_uint8(packet.buffer, index, ver));
  check(encoding::read_uint8(packet.buffer, index, flags));

  check(expire == 6);
  check(id == 1);
  check(ver == 2);
  check(flags == 3);
}

Test(core_ContinueToken_read)
{
  Packet packet;
  ContinueToken token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;
  token.session_flags = 3;

  size_t index = 0;
  check(token.write(packet, index));

  ContinueToken other;

  index = 0;
  check(other.read(packet, index));
  check(index == ContinueToken::SIZE_OF);

  check(token == other);
}
