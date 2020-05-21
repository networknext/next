#pragma once

namespace crypto
{
  // fnv1a 64
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
  [[gnu::always_inline]] inline void IsNetworkNextPacket(T& buffer, int bytes)
  {
    // TODO
  }

  template <typename T>
  [[gnu::always_inline]] inline void SignNetworkNextPacket(T& buffer, int bytes)
  {
    // TODO
  }
}  // namespace crypto

namespace legacy
{
//  inline int relay_is_network_next_packet(const uint8_t* packet_data, int packet_bytes)
//  {
//    if (packet_bytes <= RELAY_PACKET_HASH_BYTES)
//      return 0;
//
//    if (packet_bytes > RELAY_MAX_PACKET_BYTES)
//      return false;
//
//    const uint8_t* message = packet_data + RELAY_PACKET_HASH_BYTES;
//    const int message_length = packet_bytes - RELAY_PACKET_HASH_BYTES;
//
//    assert(message_length > 0);
//
//    uint8_t hash[RELAY_PACKET_HASH_BYTES];
//    crypto_generichash(
//     hash, RELAY_PACKET_HASH_BYTES, message, message_length, relay_packet_hash_key, crypto_generichash_KEYBYTES);
//
//    return memcmp(hash, packet_data, RELAY_PACKET_HASH_BYTES) == 0;
//  }
//
//  inline void relay_sign_network_next_packet(uint8_t* packet_data, int packet_bytes)
//  {
//    assert(packet_bytes > RELAY_PACKET_HASH_BYTES);
//    crypto_generichash(
//     packet_data,
//     RELAY_PACKET_HASH_BYTES,
//     packet_data + RELAY_PACKET_HASH_BYTES,
//     packet_bytes - RELAY_PACKET_HASH_BYTES,
//     relay_packet_hash_key,
//     crypto_generichash_KEYBYTES);
//  }
}  // namespace legacy
