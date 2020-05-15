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
    class RelayPingPacket
    {
     public:
      static const size_t ByteSize = 1 + 8 + net::Address::ByteSize + 1; // type | sequence | addr

      RelayPingPacket(GenericPacket<>& packet);

      // getters do no cache, just make the indexes of the packet clearer
      auto getSeqNum() -> uint64_t;
      auto getFromAddr() -> net::Address;
      auto isV3() -> bool;

      // write the addr to the buffer
      void writeFromAddr(const net::Address& addr);

      GenericPacket<>& Internal;
    };

    inline RelayPingPacket::RelayPingPacket(GenericPacket<>& packet): Internal(packet) {}

    inline auto RelayPingPacket::getSeqNum() -> uint64_t
    {
      size_t index = 1;
      return encoding::ReadUint64(Internal.Buffer, index);
    }

    inline auto RelayPingPacket::getFromAddr() -> net::Address
    {
      size_t index = 9;
      net::Address addr;
      encoding::ReadAddress(Internal.Buffer, index, addr);
      return addr;
    }

    inline auto RelayPingPacket::isV3() -> bool
    {
      size_t index = 28;
      return encoding::ReadUint8(Internal.Buffer, index);
    }

    inline void RelayPingPacket::writeFromAddr(const net::Address& addr)
    {
      size_t index = 9;
      encoding::WriteAddress(Internal.Buffer, index, addr);
    }
  }  // namespace packets
}  // namespace core
#endif