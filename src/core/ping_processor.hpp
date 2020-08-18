#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "core/relay_manager.hpp"
#include "crypto/hash.hpp"
#include "encoding/base64.hpp"
#include "encoding/write.hpp"
#include "os/platform.hpp"
#include "packets/new_relay_ping_packet.hpp"
#include "packets/old_relay_ping_packet.hpp"
#include "packets/types.hpp"
#include "util/throughput_recorder.hpp"

using namespace std::chrono_literals;

namespace core
{
  const size_t MaxPingsToSend = MAX_RELAYS;

  template <typename T>
  class PingProcessor
  {
   public:
    PingProcessor(
     const os::Socket& socket,
     core::RelayManager<T>& relayManger,
     const volatile bool& shouldProcess,
     util::ThroughputRecorder& recorder);
    ~PingProcessor() = default;

    void process(std::atomic<bool>& readyToSend);

   private:
    const os::Socket& mSocket;
    core::RelayManager<T>& mRelayManager;
    const volatile bool& mShouldProcess;
    util::ThroughputRecorder& mRecorder;

    void fillMsgHdrWithAddr(msghdr& hdr, const net::Address& addr);
  };

  template <typename T>
  PingProcessor<T>::PingProcessor(
   const os::Socket& socket,
   core::RelayManager<T>& relayManager,
   const volatile bool& shouldProcess,
   util::ThroughputRecorder& recorder)
   : mSocket(socket), mRelayManager(relayManager), mShouldProcess(shouldProcess), mRecorder(recorder)
  {}

  template <>
  void PingProcessor<Relay>::process(std::atomic<bool>& readyToSend)
  {
    readyToSend = true;
    GenericPacketBuffer<MaxPingsToSend, packets::NewRelayPingPacket::ByteSize> buffer;

    while (!mSocket.closed() && mShouldProcess) {
      // Sleep for 10ms, but the actual ping rate is controlled by RELAY_PING_TIME
      std::this_thread::sleep_for(10ms);

      std::array<core::PingData, MAX_RELAYS> pings;

      auto numberOfRelaysToPing = mRelayManager.getPingData(pings);

      if (numberOfRelaysToPing == 0) {
        continue;
      }

      for (unsigned int i = 0; i < numberOfRelaysToPing; i++) {
        auto& ping = pings[i];
        auto& pkt = buffer.Packets[i];

        auto& mhdr = buffer.Headers[i];
        auto& hdr = mhdr.msg_hdr;

        auto& addr = ping.Addr;

        pkt.Addr = addr;
        fillMsgHdrWithAddr(hdr, addr);

        size_t index = crypto::PacketHashLength;

        // write data to the buffer
        {
          if (!encoding::WriteUint8(pkt.Buffer, index, static_cast<uint8_t>(packets::Type::NewRelayPing))) {
            LOG("could not write packet type");
            assert(false);
          }

          if (!encoding::WriteUint64(pkt.Buffer, index, ping.Seq)) {
            LOG("could not write sequence");
            assert(false);
          }

          crypto::SignNetworkNextPacket(pkt.Buffer, index);
        }

        pkt.Len = index;
        hdr.msg_iov[0].iov_len = index;

        size_t headerSize = 0;
        if (addr.Type == net::AddressType::IPv4) {
          headerSize = net::IPv4UDPHeaderSize;
        } else if (addr.Type == net::AddressType::IPv6) {
          headerSize = net::IPv6UDPHeaderSize;
        }

        size_t wholePacketSize = headerSize + pkt.Len;

        // could also just do: (1 + 8) * number of relays to ping to make this faster
        mRecorder.OutboundPingTx.add(wholePacketSize);

        if (mSocket.closed() || !mShouldProcess) {
          break;
        }

#ifndef RELAY_MULTISEND
        if (!mSocket.send(pkt)) {
          LOG("failed to send new ping to ", pkt.Addr);
        }
#endif
      }

#ifdef RELAY_MULTISEND
      if (!mSocket.closed() && mShouldProcess) {
        buffer.Count = numberOfRelaysToPing;
        if (!mSocket.multisend(buffer)) {
          Log("failed to send messages, amount to send: ", numberOfRelaysToPing, ", actual sent: ", buffer.Count);
        }
      }
#endif
    }
  }

  template <typename T>
  inline void PingProcessor<T>::fillMsgHdrWithAddr(msghdr& hdr, const net::Address& addr)
  {
    assert(addr.Type != net::AddressType::None);
    if (addr.Type == net::AddressType::IPv4) {
      auto sin = reinterpret_cast<sockaddr_in*>(hdr.msg_name);
      addr.to(*sin);
      hdr.msg_namelen = sizeof(*sin);
    } else if (addr.Type == net::AddressType::IPv6) {
      auto sin = reinterpret_cast<sockaddr_in6*>(hdr.msg_name);
      addr.to(*sin);
      hdr.msg_namelen = sizeof(*sin);
    }
  }
}  // namespace core
#endif