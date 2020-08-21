#pragma once

#include "core/packet.hpp"
#include "crypto/hash.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "net/address.hpp"

namespace core
{
  namespace packets
  {
    class RelayPingPacket
    {
     public:
      static const size_t BYTE_SIZE = crypto::PacketHashLength + 1 + 8;  // hash | type | sequence

      RelayPingPacket(GenericPacket<>& packet);

      // getters do no cache, just make the indexes of the packet clearer
      auto getSeqNum() -> uint64_t;

      GenericPacket<>& Internal;
    };

    inline RelayPingPacket::RelayPingPacket(GenericPacket<>& packet): Internal(packet) {}

    inline auto RelayPingPacket::getSeqNum() -> uint64_t
    {
      size_t index = crypto::PacketHashLength + 1;
      return encoding::ReadUint64(Internal.Buffer, index);
    }
  }  // namespace packets
}  // namespace core
