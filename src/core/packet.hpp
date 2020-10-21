#pragma once

#include "net/address.hpp"
#include "packet_types.hpp"
#include "util/dump.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

namespace core
{
  struct Packet
  {
    Packet() = default;
    Packet(Packet&& other);
    ~Packet() = default;

    Packet& operator=(Packet&& other);

    net::Address addr;
    std::array<uint8_t, RELAY_MAX_PACKET_BYTES> buffer;
    size_t length;
  };

  INLINE Packet::Packet(Packet&& other): addr(std::move(other.addr)), buffer(std::move(other.buffer)), length(std::move(other.length))
  {}

  INLINE Packet& Packet::operator=(Packet&& other)
  {
    this->addr = std::move(other.addr);
    this->buffer = std::move(other.buffer);
    this->length = std::move(other.length);
    return *this;
  }

  // holds BuffSize packets and shares memory between the header and the packet, packet interface is meant to be easy to use
  template <size_t BuffSize>
  class PacketBuffer
  {
   public:
    PacketBuffer();

    // for sending packets
    void push(const net::Address& dest, const uint8_t* data, size_t length);
    void push(const Packet& pkt);

    // for debugging
    void print();

    // # to send & # sent, or # received
    int count;

    // wrapper array for received packets
    std::array<Packet, BuffSize> Packets;

    // c struct needed for sendmmsg & recvmmsg
    std::array<mmsghdr, BuffSize> Headers;

   private:
    // using vectors here to reduce stack memory

    // buffer for sockaddr's
    std::vector<std::array<uint8_t, sizeof(sockaddr_in6)>> raw_address_buffer;

    // buffer for iovec structs
    std::vector<iovec> io_vec_buffer;

    std::mutex mutex;
  };

  template <size_t BuffSize>
  INLINE PacketBuffer<BuffSize>::PacketBuffer(): count(0), raw_address_buffer(BuffSize), io_vec_buffer(BuffSize)
  {
    for (size_t i = 0; i < BuffSize; i++) {
      auto& pkt = Packets[i];

      auto& mhdr = Headers[i];
      auto& hdr = mhdr.msg_hdr;

      // assign the address buffered area to the header
      auto& addr = raw_address_buffer[i];
      {
        hdr.msg_namelen = addr.size();
        hdr.msg_name = addr.data();
      }

      // assign the iovec to the packet buffer
      auto iov = &io_vec_buffer[i];
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

  template <size_t BuffSize>
  INLINE void PacketBuffer<BuffSize>::push(const net::Address& dest, const uint8_t* data, size_t len)
  {
    // if ever going back to sendmmsg/recvmmsg
    // replace count with an atomic
    // and set by auto count = count.exchange(count + 1)
    // or something of the like, it should be possible
    // to prevent a mutex lock within this
    std::lock_guard<std::mutex> lk(mutex);

    auto& pkt = Packets[count];
    pkt.Len = len;
    auto& iov = io_vec_buffer[count];
    iov.iov_len = len;
    std::copy(data, data + len, reinterpret_cast<uint8_t*>(iov.iov_base));

    dest.into(Headers[count]);

    count++;
  }

  template <size_t BuffSize>
  INLINE void PacketBuffer<BuffSize>::push(const Packet& pkt)
  {
    push(pkt.addr, pkt.buffer.data(), pkt.length);
  }

  template <size_t BuffSize>
  INLINE void PacketBuffer<BuffSize>::print()
  {
    LOG(DEBUG, "number of packets in buffer: ", count);
    for (int i = 0; i < count; i++) {
      auto& packet = Packets[i];
      LOG(DEBUG, "sending a packet of size ", packet.Len, " to ", packet.Addr, " with data:");
      util::DumpHex(packet.Buffer.data(), packet.Len);
    }
  }
}  // namespace core
