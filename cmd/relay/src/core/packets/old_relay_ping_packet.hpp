#pragma once

#include "core/packet.hpp"
#include "encoding/read.hpp"

namespace core
{
  namespace packets
  {
    class OldRelayPingPacket
    {
     public:
      static const size_t MACByteSize = 32;
      static const size_t PingKeyBytes = 32;

      // type | token [expire timestamp | relay id | ping mac] | sequence
      static const size_t ByteSize = 1 + 8 + 8 + MACByteSize + 8;

      OldRelayPingPacket(GenericPacket<>& packet);

      GenericPacket<>& Internal;

      auto getTimestamp() -> uint64_t;
      auto getID() -> uint64_t;
      auto getMAC() -> std::vector<uint64_t>;
      auto getSequence() -> uint64_t;
    };

    inline OldRelayPingPacket::OldRelayPingPacket(GenericPacket<>& packet): Internal(packet) {}

    inline auto OldRelayPingPacket::getTimestamp() -> uint64_t
    {
      size_t index = 1;
      return encoding::ReadUint64(Internal.Buffer, index);
    }

    inline auto OldRelayPingPacket::getID() -> uint64_t
    {
      size_t index = 9;
      return encoding::ReadUint64(Internal.Buffer, index);
    }

    inline auto OldRelayPingPacket::getMAC() -> std::vector<uint64_t>
    {
      size_t index = 17;
      std::vector<uint64_t> mac(32);
      encoding::ReadBytes(Internal.Buffer, index, mac, 32);
      return mac;
    }

    inline auto OldRelayPingPacket::getSequence() -> uint64_t
    {
      size_t index = 49;
      return encoding::ReadUint64(Internal.Buffer, index);
    }
  }  // namespace packets
}  // namespace core