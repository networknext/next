#ifndef CORE_PACKET_HPP
#define CORE_PACKET_HPP

#include "net/address.hpp"

#include "util/logger.hpp"

#include "util/dump.hpp"

namespace core
{
  const size_t GenericPacketMaxSize = RELAY_MAX_PACKET_BYTES;
  struct GenericPacket
  {
    net::Address Addr;
    std::array<uint8_t, GenericPacketMaxSize> Buffer;
    size_t Len;
  };

  template <size_t BuffSize>
  class GenericPacketBuffer
  {
   public:
    GenericPacketBuffer();

    void print();  // for debugging

    // # to send & # sent, or # received
    int Count;

    // wrapper array for received packets
    std::array<GenericPacket, BuffSize> Packets;

    // c struct needed for recvmmsg
    std::array<mmsghdr, BuffSize> Headers;

    // buffer for sockaddr's
    std::array<std::array<uint8_t, sizeof(sockaddr_in6)>, BuffSize> RawAddrBuff;

    // buffer for iovec structs
    std::array<iovec, BuffSize> IOVecBuff;
  };

  template <size_t BuffSize>
  GenericPacketBuffer<BuffSize>::GenericPacketBuffer()
  {
    for (size_t i = 0; i < BuffSize; i++) {
      auto& pkt = Packets[i];

      auto& mhdr = Headers[i];
      auto& hdr = mhdr.msg_hdr;

      // assign the address buffered area to the header
      auto& addr = RawAddrBuff[i];
      {
        hdr.msg_namelen = addr.size();
        hdr.msg_name = addr.data();
      }

      // assign the iovec to the packet buffer
      auto iov = &IOVecBuff[i];
      {
        iov->iov_len = pkt.Buffer.size();
        iov->iov_base = pkt.Buffer.data();
      }

      // assign the packet buffered area to the header
      {
        hdr.msg_iovlen = 1;  // Don't change, needs to be 1 to accurately deterimine amount of bytes received
        hdr.msg_iov = iov;
      }
    }
  }

  template <size_t BuffSize>
  void GenericPacketBuffer<BuffSize>::print()
  {
    LogDebug("Number of packets in buffer: ", Count);
    for (int i = 0; i < Count; i++) {
      auto& packet = Packets[i];
      LogDebug("Sending a packet of size ", packet.Len, " to ", packet.Addr, " with data:");
      util::DumpHex(packet.Buffer.data(), packet.Len);
    }
  }
}  // namespace core
#endif