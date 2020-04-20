#ifndef CORE_PACKET_HPP
#define CORE_PACKET_HPP

#include "net/address.hpp"

#include "util/logger.hpp"

#include "util/dump.hpp"

namespace core
{
  const size_t GenericPacketMaxSize = RELAY_MAX_PACKET_BYTES;

  template <size_t BuffSize = GenericPacketMaxSize>
  struct GenericPacket
  {
    GenericPacket() = default;
    GenericPacket(GenericPacket&& other);
    ~GenericPacket() = default;

    GenericPacket<BuffSize>& operator=(GenericPacket<BuffSize>&& other);

    net::Address Addr;
    std::array<uint8_t, BuffSize> Buffer;
    size_t Len;
  };

  // holds BuffSize packets and shares memory between the header and the packet, packet interface is meant to be easy to use
  template <size_t BuffSize, size_t PacketSize = GenericPacketMaxSize>
  class GenericPacketBuffer
  {
   public:
    GenericPacketBuffer();

    // for sending packets
    void push(const net::Address& dest, const uint8_t* data, size_t length);

    // for debugging
    void print();

    // # to send & # sent, or # received
    int Count;

    // wrapper array for received packets
    std::array<GenericPacket<PacketSize>, BuffSize> Packets;

    // c struct needed for sendmmsg & recvmmsg
    std::array<mmsghdr, BuffSize> Headers;

   private:
    // using vectors here to reduce stack memory

    // buffer for sockaddr's
    std::vector<std::array<uint8_t, sizeof(sockaddr_in6)>> mRawAddrBuff;

    // buffer for iovec structs
    std::vector<iovec> mIOVecBuff;
  };

  template <size_t BuffSize>
  GenericPacket<BuffSize>::GenericPacket(GenericPacket&& other)
   : Addr(std::move(other.Addr)), Buffer(std::move(other.Buffer)), Len(std::move(other.Len))
  {}

  template <size_t BuffSize>
  GenericPacket<BuffSize>& GenericPacket<BuffSize>::operator=(GenericPacket<BuffSize>&& other)
  {
    this->Addr = std::move(other.Addr);
    this->Buffer = std::move(other.Buffer);
    this->Len = std::move(other.Len);
    return *this;
  }

  template <size_t BuffSize, size_t PacketSize>
  GenericPacketBuffer<BuffSize, PacketSize>::GenericPacketBuffer(): mRawAddrBuff(BuffSize), mIOVecBuff(BuffSize)
  {
    for (size_t i = 0; i < BuffSize; i++) {
      auto& pkt = Packets[i];

      auto& mhdr = Headers[i];
      auto& hdr = mhdr.msg_hdr;

      // assign the address buffered area to the header
      auto& addr = mRawAddrBuff[i];
      {
        hdr.msg_namelen = addr.size();
        hdr.msg_name = addr.data();
      }

      // assign the iovec to the packet buffer
      auto iov = &mIOVecBuff[i];
      {
        iov->iov_len = pkt.Buffer.size();
        iov->iov_base = pkt.Buffer.data();
      }

      // assign the packet buffered area to the header
      {
        hdr.msg_iovlen = 1;  // Don't change, needs to be 1 to accurately deterimine amount of bytes received afaik
        hdr.msg_iov = iov;
      }
    }
  }

  // TODO make thread safe
  template <size_t BuffSize, size_t PacketSize>
  void GenericPacketBuffer<BuffSize, PacketSize>::push(const net::Address& dest, const uint8_t* data, size_t len)
  {
    assert(len <= PacketSize);

    auto& iov = mIOVecBuff[Count];
    iov.iov_len = len;
    std::copy(data, data + iov.iov_len, reinterpret_cast<uint8_t*>(iov.iov_base));

    dest.to(Headers[Count]);

    Count++;
  }

  template <size_t BuffSize, size_t PacketSize>
  void GenericPacketBuffer<BuffSize, PacketSize>::print()
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