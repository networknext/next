#pragma once

namespace legacy
{
  bool relay_is_network_next_packet(const uint8_t* packet_data, size_t packet_length);

  void relay_sign_network_next_packet(uint8_t* packet_data, size_t packet_length);
}  // namespace legacy

namespace crypto
{
  const size_t PacketHashLength = 8;

  // fnv1a 64
  auto FNV(const std::string& str) -> uint64_t;

  // Both signing functions require the buffer to be the whole packet
  // and the length to be the packet's whole length, including the hash

  template <typename T>
  auto IsNetworkNextPacket(const T& buffer, size_t length) -> bool;

  template <typename T>
  void SignNetworkNextPacket(T& buffer, size_t length);

  [[gnu::always_inline]] inline auto FNV(const std::string& str) -> uint64_t
  {
    uint64_t fnv = 0xCBF29CE484222325;
    for (const auto& chr : str) {
      fnv ^= chr;
      fnv *= 0x00000100000001B3;
    }
    return fnv;
  }

  template <typename T>
  [[gnu::always_inline]] inline auto IsNetworkNextPacket(const T& buffer, size_t length) -> bool
  {
    return legacy::relay_is_network_next_packet(buffer.data(), length);
  }

  template <typename T>
  [[gnu::always_inline]] inline void SignNetworkNextPacket(T& buffer, size_t length)
  {
    legacy::relay_sign_network_next_packet(buffer.data(), length);
  }
}  // namespace crypto
