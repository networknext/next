#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "core/continue_token.hpp"

using core::ContinueToken;
using core::ContinueTokenV4;
using core::Packet;
using core::RouterInfo;

TEST(core_ContinueTokenV4_general)
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
  crypto::random_bytes(nonce, crypto_box_NONCEBYTES);

  const auto expire_timestamp = random_whole<uint64_t>();
  const auto session_id = random_whole<uint64_t>();
  const auto session_version = random_whole<uint8_t>();

  ContinueTokenV4 input_token;
  {
    input_token.expire_timestamp = expire_timestamp;
    input_token.session_id = session_id;
    input_token.session_version = session_version;
  }

  {
    size_t index = 0;
    CHECK(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
    CHECK(index == ContinueTokenV4::SIZE_OF_ENCRYPTED);
  }

  ContinueTokenV4 output_token;
  {
    size_t index = 0;
    CHECK(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  CHECK(input_token.expire_timestamp == expire_timestamp);
  CHECK(input_token.session_id == session_id);
  CHECK(input_token.session_version == session_version);

  // assert input == output
  CHECK(input_token.expire_timestamp == output_token.expire_timestamp);
  CHECK(input_token.session_id == output_token.session_id);
  CHECK(input_token.session_version == output_token.session_version);
}

TEST(core_ContinueTokenV4_write)
{
  Packet packet;
  ContinueTokenV4 token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;

  size_t index = 0;
  CHECK(token.write(packet, index));
  CHECK(index == ContinueTokenV4::SIZE_OF);

  uint64_t expire;
  uint64_t id;
  uint8_t ver;

  index = 0;
  CHECK(encoding::read_uint64(packet.buffer, index, expire));
  CHECK(encoding::read_uint64(packet.buffer, index, id));
  CHECK(encoding::read_uint8(packet.buffer, index, ver));

  CHECK(expire == 6);
  CHECK(id == 1);
  CHECK(ver == 2);
}

TEST(core_ContinueTokenV4_read)
{
  Packet packet;
  ContinueTokenV4 token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;

  size_t index = 0;
  CHECK(token.write(packet, index));

  ContinueTokenV4 other;

  index = 0;
  CHECK(other.read(packet, index));
  CHECK(index == ContinueTokenV4::SIZE_OF);

  CHECK(token == other);
}

TEST(core_ContinueToken_general)
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
  crypto::random_bytes(nonce, crypto_box_NONCEBYTES);

  const auto expire_timestamp = random_whole<uint64_t>();
  const auto session_id = random_whole<uint64_t>();
  const auto session_version = random_whole<uint8_t>();
  const auto session_flags = random_whole<uint8_t>();

  ContinueToken input_token;
  {
    input_token.expire_timestamp = expire_timestamp;
    input_token.session_id = session_id;
    input_token.session_version = session_version;
    input_token.session_flags = session_flags;
  }

  {
    size_t index = 0;
    CHECK(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
    CHECK(index == ContinueToken::SIZE_OF_ENCRYPTED);
  }

  ContinueToken output_token;
  {
    size_t index = 0;
    CHECK(output_token.read_encrypted(packet, index, sender_public_key, receiver_private_key));
  }

  // make sure nothing changed
  CHECK(input_token.expire_timestamp == expire_timestamp);
  CHECK(input_token.session_id == session_id);
  CHECK(input_token.session_version == session_version);
  CHECK(input_token.session_flags == session_flags);

  // assert input == output
  CHECK(input_token.expire_timestamp == output_token.expire_timestamp);
  CHECK(input_token.session_id == output_token.session_id);
  CHECK(input_token.session_version == output_token.session_version);
  CHECK(input_token.session_flags == output_token.session_flags);
}

TEST(core_ContinueToken_write)
{
  Packet packet;
  ContinueToken token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;
  token.session_flags = 3;

  size_t index = 0;
  CHECK(token.write(packet, index));
  CHECK(index == ContinueToken::SIZE_OF);

  uint64_t expire;
  uint64_t id;
  uint8_t ver;
  uint8_t flags;

  index = 0;
  CHECK(encoding::read_uint64(packet.buffer, index, expire));
  CHECK(encoding::read_uint64(packet.buffer, index, id));
  CHECK(encoding::read_uint8(packet.buffer, index, ver));
  CHECK(encoding::read_uint8(packet.buffer, index, flags));

  CHECK(expire == 6);
  CHECK(id == 1);
  CHECK(ver == 2);
  CHECK(flags == 3);
}

TEST(core_ContinueToken_read)
{
  Packet packet;
  ContinueToken token;

  token.expire_timestamp = 6;
  token.session_id = 1;
  token.session_version = 2;
  token.session_flags = 3;

  size_t index = 0;
  CHECK(token.write(packet, index));

  ContinueToken other;

  index = 0;
  CHECK(other.read(packet, index));
  CHECK(index == ContinueToken::SIZE_OF);

  CHECK(token == other);
}
