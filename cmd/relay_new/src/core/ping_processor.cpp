#include "includes.h"
#include "ping_processor.hpp"

#include "encoding/write.hpp"

using namespace std::chrono_literals;
namespace
{
  const size_t MaxPacketsToSend = MAX_RELAYS;
}

namespace core
{
  PingProcessor::PingProcessor(
   const os::Socket& socket,
   core::RelayManager& relayManager,
   const volatile bool& shouldProcess,
   const net::Address& relayAddress,
   util::ThroughputRecorder& recorder,
   legacy::v3::TrafficStats& stats)
   : mSocket(socket),
     mRelayManager(relayManager),
     mShouldProcess(shouldProcess),
     mRelayAddress(relayAddress),
     mRecorder(recorder),
     mStats(stats)
  {
    LogDebug("sending pings using this addr: ", relayAddress);
  }

  void PingProcessor::process(std::condition_variable& var, std::atomic<bool>& readyToSend)
  {
    readyToSend = true;
    var.notify_one();

    GenericPacketBuffer<MaxPacketsToSend, RELAY_PING_PACKET_BYTES> buffer;

    while (mShouldProcess) {
      std::this_thread::sleep_for(10ms);

      std::array<core::PingData, MAX_RELAYS> pings;

      auto numberOfRelaysToPing = mRelayManager.getPingData(pings);

      if (numberOfRelaysToPing == 0) {
        continue;
      }

      for (unsigned int i = 0; i < numberOfRelaysToPing; i++) {
        auto& pkt = buffer.Packets[i];

        auto& mhdr = buffer.Headers[i];
        auto& hdr = mhdr.msg_hdr;

        auto& addr = pings[i].Addr;

        fillMsgHdrWithAddr(hdr, addr);

        size_t index = 0;

        // write data to the buffer
        {
          encoding::WriteUint8(pkt.Buffer, index, RELAY_PING_PACKET);
          encoding::WriteUint64(pkt.Buffer, index, pings[i].Seq);

          // use the recv port addr here so the receiving relay knows where to send it back to
          encoding::WriteAddress(pkt.Buffer, index, mRelayAddress);
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

        // could also just do: (1 + 8 + net::Address::ByteSize) * number of relays to ping to make this faster
        mRecorder.addToSent(wholePacketSize);
        mStats.BytesPerSecManagementTx += wholePacketSize;
      }

      buffer.Count = numberOfRelaysToPing;

      if (!mSocket.multisend(buffer)) {
        Log("failed to send messages, amount to send: ", numberOfRelaysToPing, ", actual sent: ", buffer.Count);
      }
    }
  }

  inline void PingProcessor::fillMsgHdrWithAddr(msghdr& hdr, const net::Address& addr)
  {
    // TODO need error handling here
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
