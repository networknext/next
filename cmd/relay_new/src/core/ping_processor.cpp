#include "includes.h"
#include "ping_processor.hpp"

#include "encoding/write.hpp"

#include "net/message.hpp"

using namespace std::chrono_literals;

namespace core
{
  PingProcessor::PingProcessor(
   const os::Socket& socket, core::RelayManager& relayManager, volatile bool& shouldProcess, const net::Address& relayAddress)
   : mSocket(socket), mRelayManager(relayManager), mShouldProcess(shouldProcess), mRelayAddress(relayAddress)
  {}

  void PingProcessor::process(std::condition_variable& var, std::atomic<bool>& readyToSend)
  {
    readyToSend = true;
    var.notify_one();
    while (mShouldProcess) {
      std::array<core::PingData, MAX_RELAYS> pings;

      auto numPings = mRelayManager.getPingData(pings);

      std::vector<net::Message> messages;
      messages.resize(numPings);

      for (unsigned int i = 0; i < messages.size(); i++) {
        auto& msg = messages[i];
        msg.Addr = pings[i].Addr;  // send to pings addr
        msg.Len = RELAY_PING_PACKET_BYTES;
        msg.Data.resize(msg.Len);

        size_t index = 0;

        encoding::WriteUint8(msg.Data, index, RELAY_PING_PACKET);
        encoding::WriteUint64(msg.Data, index, pings[i].Seq);

        // use the recv port addr here so the receiving relay knows where to send it back to
        encoding::WriteAddress(msg.Data, index, mRelayAddress);
      }

      int sentMessages = 0;
      if (!mSocket.multisend(messages, numPings, sentMessages)) {
        Log("failed to send messages, amount to send: ", numPings, ", actual sent: ", sentMessages);
      }

      std::this_thread::sleep_for(10ms);
    }
  }
}  // namespace core
