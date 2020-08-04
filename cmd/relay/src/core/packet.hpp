#ifndef CORE_PACKET_HPP
#define CORE_PACKET_HPP

#include "net/address.hpp"

#include "util/logger.hpp"

#include "util/dump.hpp"

namespace core
{
  const size_t GenericPacketMaxSize = RELAY_MAX_PACKET_BYTES;

  template <typename T>
  struct Packet
  {
    Packet() = default;
    Packet(Packet<T>&& other);
    ~Packet() = default;

    Packet<T>& operator=(Packet<T>&& other);

    net::Address Addr;
    T Buffer;
    size_t Len;
  };

  template <size_t BuffSize = GenericPacketMaxSize>
  using GenericPacket = Packet<std::array<uint8_t, BuffSize>>;

  // holds BuffSize packets and shares memory between the header and the packet, packet interface is meant to be easy to use
  template <size_t BuffSize, size_t PacketSize = GenericPacketMaxSize>
  class GenericPacketBuffer
  {
   public:
    GenericPacketBuffer();

    // for sending packets
    void push(const net::Address& dest, const uint8_t* data, size_t length);
    void push(const GenericPacket<PacketSize>& pkt);

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

    std::mutex mLock;
  };

  template <typename T>
  Packet<T>::Packet(Packet<T>&& other): Addr(std::move(other.Addr)), Buffer(std::move(other.Buffer)), Len(std::move(other.Len))
  {}

  template <typename T>
  Packet<T>& Packet<T>::operator=(Packet<T>&& other)
  {
    this->Addr = std::move(other.Addr);
    this->Buffer = std::move(other.Buffer);
    this->Len = std::move(other.Len);
    return *this;
  }

  template <size_t BuffSize, size_t PacketSize>
  GenericPacketBuffer<BuffSize, PacketSize>::GenericPacketBuffer(): Count(0), mRawAddrBuff(BuffSize), mIOVecBuff(BuffSize)
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

  template <size_t BuffSize, size_t PacketSize>
  void GenericPacketBuffer<BuffSize, PacketSize>::push(const net::Address& dest, const uint8_t* data, size_t len)
  {
    assert(len <= PacketSize);

    // TODO if ever going back to sendmmsg/recvmmsg
    // replace count with an atomic
    // and set by auto count = Count.exchange(Count + 1)
    // or something of the like, it should be possible
    // to prevent a mutex lock within this
    std::lock_guard<std::mutex> lk(mLock);

    auto& pkt = Packets[Count];
    pkt.Len = len;
    auto& iov = mIOVecBuff[Count];
    iov.iov_len = len;
    std::copy(data, data + len, reinterpret_cast<uint8_t*>(iov.iov_base));

    dest.to(Headers[Count]);

    Count++;
  }

  template <size_t BuffSize, size_t PacketSize>
  void GenericPacketBuffer<BuffSize, PacketSize>::push(const GenericPacket<PacketSize>& pkt)
  {
    push(pkt.Addr, pkt.Buffer.data(), pkt.Len);
  }

  template <size_t BuffSize, size_t PacketSize>
  void GenericPacketBuffer<BuffSize, PacketSize>::print()
  {
    Log("Number of packets in buffer: ", Count);
    for (int i = 0; i < Count; i++) {
      auto& packet = Packets[i];
      Log("Sending a packet of size ", packet.Len, " to ", packet.Addr, " with data:");
      util::DumpHex(packet.Buffer.data(), packet.Len);
    }
  }
}  // namespace core
#endif