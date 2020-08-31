#pragma once

#include "core/relay_manager.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "encoding/base64.hpp"
#include "encoding/write.hpp"
#include "os/socket.hpp"
#include "packets/types.hpp"

using namespace std::chrono_literals;

using core::Packet;
using core::packets::RELAY_PING_PACKET_SIZE;
using core::packets::Type;
using net::Address;
using net::AddressType;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  const size_t MaxPingsToSend = MAX_RELAYS;

  class Pinger
  {
   public:
    Pinger(
     const Socket& socket, RelayManager& relay_manager, const volatile bool& should_process, ThroughputRecorder& recorder);
    ~Pinger() = default;

    void process();

   private:
    const Socket& socket;
    RelayManager& relay_manager;
    const volatile bool& should_process;
    ThroughputRecorder& recorder;
  };

  INLINE Pinger::Pinger(
   const Socket& socket, RelayManager& relay_manager, const volatile bool& should_process, ThroughputRecorder& recorder)
   : socket(socket), relay_manager(relay_manager), should_process(should_process), recorder(recorder)
  {}

  INLINE void Pinger::process()
  {
    Packet pkt;

    while (!socket.closed() && should_process) {
      // Sleep for 10ms, but the actual ping rate is controlled by RELAY_PING_TIME
      std::this_thread::sleep_for(10ms);

      std::array<core::PingData, MAX_RELAYS> pings;

      auto numberOfRelaysToPing = relay_manager.getPingData(pings);

      if (numberOfRelaysToPing == 0) {
        continue;
      }

      for (unsigned int i = 0; i < numberOfRelaysToPing; i++) {
        const auto& ping = pings[i];

        pkt.Addr = ping.Addr;

        size_t index = crypto::PacketHashLength;

        // write data to the buffer
        {
          if (!encoding::WriteUint8(pkt.Buffer, index, static_cast<uint8_t>(Type::RelayPing))) {
            LOG(ERROR, "could not write packet type");
            assert(false);
          }

          if (!encoding::WriteUint64(pkt.Buffer, index, ping.Seq)) {
            LOG(ERROR, "could not write sequence");
            assert(false);
          }

          crypto::SignNetworkNextPacket(pkt.Buffer, index);
        }

        pkt.Len = index;

        if (socket.closed() || !should_process) {
          break;
        }

        if (!socket.send(pkt)) {
          LOG(ERROR, "failed to send new ping to ", pkt.Addr);
        }

        size_t headerSize = 0;
        if (pkt.Addr.Type == AddressType::IPv4) {
          headerSize = net::IPv4UDPHeaderSize;
        } else if (pkt.Addr.Type == AddressType::IPv6) {
          headerSize = net::IPv6UDPHeaderSize;
        }

        size_t wholePacketSize = headerSize + pkt.Len;

        // could also just do: (1 + 8) * number of relays to ping to make this faster
        recorder.OutboundPingTx.add(wholePacketSize);
      }
    }
  }
}  // namespace core
