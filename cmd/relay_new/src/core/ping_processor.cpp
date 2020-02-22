#include "includes.h"
#include "ping_processor.hpp"

#include "encoding/write.hpp"

namespace core
{
  PingProcessor::PingProcessor(core::RelayManager& relayManager, volatile bool& shouldProcess, const net::Address& relayAddress)
   : mRelayManager(relayManager), mShouldProcess(shouldProcess), mRelayAddress(relayAddress)
  {}

  void PingProcessor::listen(os::Socket& socket, std::condition_variable& var, std::atomic<bool>& readyToSend)
  {
    readyToSend = true;
    var.notify_one();
    while (mShouldProcess) {
      std::array<core::PingData, MAX_RELAYS> pings;

      auto numPings = mRelayManager.getPingData(pings);

      for (unsigned int i = 0; i < numPings; ++i) {
        size_t index = 0;
        std::array<uint8_t, RELAY_PING_PACKET_BYTES> packetData;
        encoding::WriteUint8(packetData, index, RELAY_PING_PACKET);
        encoding::WriteUint64(packetData, index, pings[i].Seq);
        encoding::WriteAddress(packetData, index, mRelayAddress);
        socket.send(pings[i].Addr, packetData.data(), packetData.size());
      }

      relay::relay_platform_sleep(1.0 / 100.0);
    }
  }
}  // namespace core
