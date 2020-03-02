#ifndef CORE_PACKETS_RELAY_PING_PACKET_HPP
#define CORE_PACKETS_RELAY_PING_PACKET_HPP

#include "core/packet.hpp"

#include "encoding/read.hpp"

#include "net/address.hpp"

namespace core
{
  namespace packets
  {
    class RelayPingPacket
    {
     public:
      RelayPingPacket(GenericPacket<>& packet, int size);

      const uint64_t& getSeqNum();
      const net::Address& getFromAddr();

      void writeFromAddr(const net::Address& addr);

      GenericPacket<>& Internal;
      const int Size;

     private:
      std::optional<uint64_t> mSequenceNumber;
      std::optional<net::Address> mFrom;
    };

    inline RelayPingPacket::RelayPingPacket(GenericPacket<>& packet, int size): Internal(packet), Size(size) {}

    inline const uint64_t& RelayPingPacket::getSeqNum()
    {
      if (!mSequenceNumber.has_value()) {
        size_t index = 1;
        mSequenceNumber = encoding::ReadUint64(Internal.Buffer, index);
      }

      return mSequenceNumber.value();
    }

    inline const net::Address& RelayPingPacket::getFromAddr()
    {
      if (!mFrom.has_value()) {
        size_t index = 9;
        net::Address addr;
        encoding::ReadAddress(Internal.Buffer, index, addr);
        mFrom = addr;
      }

      return mFrom.value();
    }

    inline void RelayPingPacket::writeFromAddr(const net::Address& addr)
    {
      size_t index = 9;
      encoding::WriteAddress(Internal.Buffer, index, addr);
    }
  }  // namespace packets
}  // namespace core
#endif