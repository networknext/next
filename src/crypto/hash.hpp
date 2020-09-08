#pragma once

#include "util/macros.hpp"
#include "core/packet_types.hpp"

using core::RELAY_MAX_PACKET_BYTES;

namespace crypto
{
  const size_t PACKET_HASH_LENGTH = 8;
  const size_t HASH_KEY_SIZE = crypto_generichash_KEYBYTES;

  INLINE auto hash_key() -> const std::array<uint8_t, HASH_KEY_SIZE>&
  {
    static const std::array<uint8_t, HASH_KEY_SIZE> key = {
     0xe3, 0x18, 0x61, 0x72, 0xee, 0x70, 0x62, 0x37, 0x40, 0xf6, 0x0a, 0xea, 0xe0, 0xb5, 0x1a, 0x2c,
     0x2a, 0x47, 0x98, 0x8f, 0x27, 0xec, 0x63, 0x2c, 0x25, 0x04, 0x74, 0x89, 0xaf, 0x5a, 0xeb, 0x24,
    };
    return key;
  }

  // index is the start of the packet, including the area reserved for the hash
  // length is the total length of the packet, including the hash
  template <typename T>
  auto is_network_next_packet(const T& buffer, size_t& index, size_t length) -> bool;

  // index is the start of the packet, including the area reserved for the hash
  // length is the total length of the packet, including the hash
  template <typename T>
  auto is_network_next_packet_sdk4(const T& buffer, size_t& index, size_t length) -> bool;

  // index is the start of the packet, including the area reserved for the hash
  // length is the total length of the packet, including the hash
  template <typename T>
  auto sign_network_next_packet(T& buffer, size_t& index, size_t length) -> bool;

  // index is the start of the packet, including the area reserved for the hash
  // length is the total length of the packet, including the hash
  template <typename T>
  auto sign_network_next_packet_sdk4(T& buffer, size_t& index, size_t length) -> bool;

  // fnv1a 64
  auto fnv(const std::string& str) -> uint64_t;

  template <typename T>
  INLINE auto is_network_next_packet(const T& buffer, size_t& index, size_t length) -> bool
  {
    if (length <= PACKET_HASH_LENGTH) {
      return false;
    }

    if (length > RELAY_MAX_PACKET_BYTES) {
      return false;
    }

    std::array<uint8_t, PACKET_HASH_LENGTH> hash;
    crypto_generichash(
     hash.data(),
     PACKET_HASH_LENGTH,
     &buffer[index + PACKET_HASH_LENGTH],
     length - PACKET_HASH_LENGTH,
     hash_key().data(),
     crypto_generichash_KEYBYTES);

    return memcmp(hash.data(), &buffer[index], PACKET_HASH_LENGTH) == 0;
  }

  template <typename T>
  auto is_network_next_packet_sdk4(const T& buffer, size_t& index, size_t length) -> bool
  {
    if (length <= PACKET_HASH_LENGTH) {
      return false;
    }

    if (length > RELAY_MAX_PACKET_BYTES) {
      return false;
    }

    length -= PACKET_HASH_LENGTH;

    if (length > 32) {
      length = 32;
    }

    std::array<uint8_t, PACKET_HASH_LENGTH> hash;
    crypto_generichash(
     hash.data(),
     PACKET_HASH_LENGTH,
     &buffer[index + PACKET_HASH_LENGTH],
     length,
     hash_key().data(),
     crypto_generichash_KEYBYTES);

    return memcmp(hash.data(), &buffer[index], PACKET_HASH_LENGTH) == 0;
  }

  template <typename T>
  INLINE auto sign_network_next_packet(T& buffer, size_t& index, size_t length) -> bool
  {
    if (length <= PACKET_HASH_LENGTH) {
      return false;
    }

    if (length > RELAY_MAX_PACKET_BYTES) {
      return false;
    }

    crypto_generichash(
     &buffer[index],
     PACKET_HASH_LENGTH,
     &buffer[index + PACKET_HASH_LENGTH],
     length - PACKET_HASH_LENGTH,
     hash_key().data(),
     crypto_generichash_KEYBYTES);

    return true;
  }

  template <typename T>
  auto sign_network_next_packet_sdk4(T& buffer, size_t& index, size_t length) -> bool
  {
    if (length <= PACKET_HASH_LENGTH) {
      return false;
    }

    if (length > RELAY_MAX_PACKET_BYTES) {
      return false;
    }

    length -= PACKET_HASH_LENGTH;
    if (length > 32) {
      length = 32;
    }

    crypto_generichash(
     &buffer[index],
     PACKET_HASH_LENGTH,
     &buffer[index + PACKET_HASH_LENGTH],
     length,
     hash_key().data(),
     crypto_generichash_KEYBYTES);

    return true;
  }

  INLINE auto fnv(const std::string& str) -> uint64_t
  {
    uint64_t fnv = 0xCBF29CE484222325;
    for (const auto& chr : str) {
      fnv ^= chr;
      fnv *= 0x00000100000001B3;
    }
    return fnv;
  }
}  // namespace crypto
