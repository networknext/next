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
      RelayPingPacket(GenericPacket<>& packet, int size);

      // getters do no cache, just make the indexes of the packet clearer
      auto getSeqNum() -> uint64_t;
      auto getFromAddr() -> net::Address;

      // write the addr to the buffer
      void writeFromAddr(const net::Address& addr);

      GenericPacket<>& Internal;
      const int Size;
    };

    inline RelayPingPacket::RelayPingPacket(GenericPacket<>& packet, int size): Internal(packet), Size(size) {}

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

    inline void RelayPingPacket::writeFromAddr(const net::Address& addr)
    {
      size_t index = 9;
      encoding::WriteAddress(Internal.Buffer, index, addr);
    }
  }  // namespace packets
}  // namespace core
#endif