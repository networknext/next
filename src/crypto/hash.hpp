#pragma once

#include "util/macros.hpp"

namespace legacy
{
  bool relay_is_network_next_packet(const uint8_t* packet_data, size_t packet_length);

  void relay_sign_network_next_packet(uint8_t* packet_data, size_t packet_length);
}  // namespace legacy

namespace crypto
{
  const size_t PACKET_HASH_LENGTH = 8;

  // fnv1a 64
  auto fnv(const std::string& str) -> uint64_t;

  // Both signing functions require the buffer to be the whole packet
  // and the length to be the packet's whole length, including the hash

  template <typename T>
  auto is_network_next_packet(const T& buffer, size_t length) -> bool;

  template <typename T>
  void sign_network_next_packet(T& buffer, size_t length);

  INLINE auto fnv(const std::string& str) -> uint64_t
  {
    uint64_t fnv = 0xCBF29CE484222325;
    for (const auto& chr : str) {
      fnv ^= chr;
      fnv *= 0x00000100000001B3;
    }
    return fnv;
  }

  template <typename T>
  INLINE auto is_network_next_packet(const T& buffer, size_t length) -> bool
  {
    return legacy::relay_is_network_next_packet(buffer.data(), length);
  }

  template <typename T>
  INLINE void sign_network_next_packet(T& buffer, size_t length)
  {
    legacy::relay_sign_network_next_packet(buffer.data(), length);
  }
}  // namespace crypto
