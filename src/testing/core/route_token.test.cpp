#include "includes.h"
#include "testing/test.hpp"

#include "core/route_token.hpp"

#define CRYPTO_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::RouterInfo;
using core::RouteTokenV4;
using net::Address;
using net::AddressType;

TEST(core_RouteTokenV4_general)
{
  RouterInfo info;
  Packet packet;
  packet.length = packet.buffer.size();

  auto [sender_public_key, sender_private_key] = gen_keypair();
  auto [receiver_public_key, receiver_private_key] = gen_keypair();

  std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
  crypto::random_bytes(nonce, crypto_box_NONCEBYTES);

  const auto expire_timestamp = random_whole<uint64_t>();
  const auto session_id = random_whole<uint64_t>();
  const auto session_version = random_whole<uint8_t>();
  const auto kbps_up = random_whole<uint32_t>();
  const auto kbps_down = random_whole<uint32_t>();
  Address next_addr;
  next_addr.type = AddressType::IPv4;
  next_addr.ipv4[0] = random_whole<uint8_t>();
  next_addr.ipv4[1] = random_whole<uint8_t>();
  next_addr.ipv4[2] = random_whole<uint8_t>();
  next_addr.ipv4[3] = random_whole<uint8_t>();
  next_addr.port = random_whole<uint32_t>();
  GenericKey private_key = random_private_key();

  RouteTokenV4 input_token;
  {
    input_token.expire_timestamp = expire_timestamp;
    input_token.session_id = session_id;
    input_token.session_version = session_version;
    input_token.kbps_up = kbps_up;
    input_token.kbps_down = kbps_down;
    input_token.next_addr = next_addr;
    input_token.private_key = private_key;
  }

  {
    size_t index = 0;
    CHECK(input_token.write_encrypted(packet, index, sender_private_key, receiver_public_key));
    CHECK(index == RouteTokenV4::SIZE_OF_SIGNED);
  }

  RouteTokenV4 output_token;
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
  CHECK(input_token.private_key == private_key);
  CHECK(input_token.next_addr == next_addr);

  // make sure input == output
  CHECK(input_token.expire_timestamp == output_token.expire_timestamp);
  CHECK(input_token.session_id == output_token.session_id);
  CHECK(input_token.session_version == output_token.session_version);
  CHECK(input_token.kbps_up == output_token.kbps_up);
  CHECK(input_token.kbps_down == output_token.kbps_down);
  CHECK(input_token.private_key == output_token.private_key);
  CHECK(input_token.next_addr == output_token.next_addr);
}

TEST(core_RouteTokenV4_write)
{
  const auto expire_timestamp = random_whole<uint64_t>();
  const auto session_id = random_whole<uint64_t>();
  const auto session_version = random_whole<uint8_t>();
  const auto kbps_up = random_whole<uint32_t>();
  const auto kbps_down = random_whole<uint32_t>();
  Address next_addr;
  next_addr.type = AddressType::IPv4;
  next_addr.ipv4[0] = random_whole<uint8_t>();
  next_addr.ipv4[1] = random_whole<uint8_t>();
  next_addr.ipv4[2] = random_whole<uint8_t>();
  next_addr.ipv4[3] = random_whole<uint8_t>();
  next_addr.port = random_whole<uint32_t>();
  GenericKey private_key = random_private_key();

  Packet packet;
  RouteTokenV4 token;
  {
    token.expire_timestamp = expire_timestamp;
    token.session_id = session_id;
    token.session_version = session_version;
    token.kbps_up = kbps_up;
    token.kbps_down = kbps_down;
    token.next_addr = next_addr;
    token.private_key = private_key;
  }

  size_t index = 0;
  CHECK(token.write(packet, index));
  CHECK(index == RouteTokenV4::SIZE_OF);
}
