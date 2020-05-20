#ifndef CORE_PACKETS_RELAY_PING_PACKET_HPP
#define CORE_PACKETS_RELAY_PING_PACKET_HPP

#include "core/packet.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "net/address.hpp"

namespace core
{
  namespace packets
  {
    class NewRelayPingPacket
    {
     public:
      static const size_t ByteSize = 1 + 8;  // type | sequence | addr

      NewRelayPingPacket(GenericPacket<>& packet);

      // getters do no cache, just make the indexes of the packet clearer
      auto getSeqNum() -> uint64_t;

      GenericPacket<>& Internal;
    };

    inline NewRelayPingPacket::NewRelayPingPacket(GenericPacket<>& packet): Internal(packet) {}

    inline auto NewRelayPingPacket::getSeqNum() -> uint64_t
    {
      size_t index = 1;
      return encoding::ReadUint64(Internal.Buffer, index);
    }
  }  // namespace packets
}  // namespace core
#endif