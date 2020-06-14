#include "includes.h"
#include "hash.hpp"

namespace
{
  const size_t PacketHashKeySize = crypto_generichash_KEYBYTES;
  const std::array<uint8_t, PacketHashKeySize> RelayPacketHashKey = {
   0xe3, 0x18, 0x61, 0x72, 0xee, 0x70, 0x62, 0x37, 0x40, 0xf6, 0x0a, 0xea, 0xe0, 0xb5, 0x1a, 0x2c,
   0x2a, 0x47, 0x98, 0x8f, 0x27, 0xec, 0x63, 0x2c, 0x25, 0x04, 0x74, 0x89, 0xaf, 0x5a, 0xeb, 0x24,
  };
}  // namespace

namespace legacy
{
  bool relay_is_network_next_packet(const uint8_t* packet_data, size_t packet_length)
  {
    if (packet_length <= crypto::PacketHashLength) {
      return 0;
    }

    if (packet_length > RELAY_MAX_PACKET_BYTES) {
      return false;
    }

    const uint8_t* message = packet_data + crypto::PacketHashLength;
    const int message_length = packet_length - crypto::PacketHashLength;

    assert(message_length > 0);

    uint8_t hash[crypto::PacketHashLength];
    crypto_generichash(
     hash, crypto::PacketHashLength, message, message_length, RelayPacketHashKey.data(), crypto_generichash_KEYBYTES);

    return memcmp(hash, packet_data, crypto::PacketHashLength) == 0;
  }

  void relay_sign_network_next_packet(uint8_t* packet_data, size_t packet_length)
  {
    assert(packet_length > crypto::PacketHashLength);
    crypto_generichash(
     packet_data,
     crypto::PacketHashLength,
     packet_data + crypto::PacketHashLength,
     packet_length - crypto::PacketHashLength,
     RelayPacketHashKey.data(),
     crypto_generichash_KEYBYTES);
  }
}  // namespace legacy