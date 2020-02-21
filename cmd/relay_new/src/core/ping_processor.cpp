#include "includes.h"
#include "ping_processor.hpp"

#include "encoding/write.hpp"

namespace core
{
  PingProcessor::PingProcessor(core::RelayManager& relayManager, volatile bool& shouldProcess)
   : mRelayManager(relayManager), mShouldProcess(shouldProcess)
  {}

  void PingProcessor::listen(os::Socket& socket, std::condition_variable& var, std::atomic<bool>& readyToSend)
  {
    readyToSend = true;
    var.notify_one();
    while (mShouldProcess) {
      std::array<core::PingData, MAX_RELAYS> pings;

      auto numPings = mRelayManager.getPingData(pings);

      for (unsigned int i = 0; i < numPings; ++i) {
        uint8_t packet_data[9];
        packet_data[0] = RELAY_PING_PACKET;
        uint8_t* p = packet_data + 1;
        encoding::write_uint64(&p, pings[i].Seq);
        socket.send(pings[i].Addr, packet_data, 9);
      }

      relay::relay_platform_sleep(1.0 / 100.0);
    }
  }
}  // namespace core
