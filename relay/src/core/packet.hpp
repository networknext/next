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

   private:

    std::mutex mutex;
  };

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
