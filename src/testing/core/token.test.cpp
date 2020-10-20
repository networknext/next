#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "core/token.hpp"

using core::Packet;
using core::TokenV4;

TEST(core_Token_write)
{
  TokenV4 token;

  token.expire_timestamp = random_whole<uint64_t>();
  token.session_id = random_whole<uint64_t>();
  token.session_version = random_whole<uint8_t>();

  Packet packet;

  size_t index = 0;

  CHECK(token.write(packet, index));

  index = 0;

  uint64_t expire_timestamp, id;
  uint8_t version, flags;

  CHECK(encoding::read_uint64(packet.buffer, index, expire_timestamp));
  CHECK(encoding::read_uint64(packet.buffer, index, id));
  CHECK(encoding::read_uint8(packet.buffer, index, version));
  CHECK(encoding::read_uint8(packet.buffer, index, flags));

  CHECK(token.expire_timestamp == expire_timestamp);
  CHECK(token.session_id == id);
  CHECK(token.session_version == version);
}

TEST(core_Token_read) {
  TokenV4 token, other;
  Packet packet;

  token.expire_timestamp = random_whole<uint64_t>();
  token.session_id = random_whole<uint64_t>();
  token.session_version = random_whole<uint8_t>();

  size_t index = 0;
  CHECK(token.write(packet, index));

  index = 0;
  CHECK(other.read(packet, index));

  CHECK(token == other);
}
