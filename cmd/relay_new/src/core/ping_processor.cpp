#include "includes.h"
#include "ping_processor.hpp"

#include "encoding/write.hpp"

using namespace std::chrono_literals;

namespace core
{
  PingProcessor::PingProcessor(core::RelayManager& relayManager, volatile bool& shouldProcess, const net::Address& relayAddress)
   : mRelayManager(relayManager), mShouldProcess(shouldProcess), mRelayAddress(relayAddress)
  {}

  void PingProcessor::process(os::Socket& socket, std::condition_variable& var, std::atomic<bool>& readyToSend)
  {
    readyToSend = true;
    var.notify_one();
    while (mShouldProcess) {
      std::array<core::PingData, MAX_RELAYS> pings;

      auto numPings = mRelayManager.getPingData(pings);

      std::vector<net::MultiMessage> messages;
      messages.resize(numPings);

      for (unsigned int i = 0; i < messages.size(); i++) {
        auto& msg = messages[i];
        msg.Addr = pings[i].Addr;  // send to pings addr
        msg.Msg.resize(RELAY_PING_PACKET_BYTES);

        size_t index = 0;

        encoding::WriteUint8(msg.Msg, index, RELAY_PING_PACKET);
        encoding::WriteUint64(msg.Msg, index, pings[i].Seq);
        encoding::WriteAddress(
         msg.Msg, index, mRelayAddress);  // use the recv port addr here so the receiving relay knows where to send it back to
      }

      int sentMessages = 0;
      if (!socket.multisend(messages, sentMessages)) {
        Log("failed to send messages, amount to send: ", numPings, ", actual sent: ", sentMessages);
      }

      std::this_thread::sleep_for(10ms);
    }
  }
}  // namespace core
