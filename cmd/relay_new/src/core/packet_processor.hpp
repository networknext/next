#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "session.hpp"

#include "core/packet.hpp"
#include "core/relay_manager.hpp"
#include "core/router_info.hpp"
#include "core/token.hpp"

#include "crypto/keychain.hpp"

#include "os/platform.hpp"

#include "util/throughput_logger.hpp"

#include "net/buffered_sender.hpp"

namespace core
{
  const size_t MaxPacketsToReceive = 1024;

  // same as receive, the amount of output depends on the input
  const size_t MaxPacketsToSend = MaxPacketsToReceive;

  class PacketProcessor
  {
   public:
    PacketProcessor(os::Socket& socket,
     const util::Clock& relayClock,
     const crypto::Keychain& keychain,
     const core::RouterInfo& routerInfo,
     core::SessionMap& sessions,
     core::RelayManager& relayManager,
     const volatile bool& handle,
     util::ThroughputLogger& logger,
     const net::Address& receivingAddr);
    ~PacketProcessor() = default;

    void process(std::condition_variable& var, std::atomic<bool>& readyToReceive);

   private:
    const os::Socket& mSocket;
    const util::Clock& mRelayClock;
    const crypto::Keychain& mKeychain;
    const core::RouterInfo mRouterInfo;
    core::SessionMap& mSessionMap;
    core::RelayManager& mRelayManager;
    const volatile bool& mShouldProcess;
    util::ThroughputLogger& mLogger;
    const net::Address& mRecvAddr;

    // perf based on using 2 packet processors, original benchmark (using sendto()) is 72 Mb/s

    // basicaly a slightly less effecient sento(), no noticable Mb/s diff
    // net::BufferedSender<1, 0> mSender;

    // caused a decrease in perf, probably timing out too often, down to 56 Mb/s
    // net::BufferedSender<60, 40> mSender;

    // caused a gain, but only because the timeout wasn't present,
    // further adds to the timeout being responsable for the perf decrease, about 80-83 Mb/s
    // net::BufferedSender<40, 0> mSender;

    // similar gain, but ranges from 56 Mb/s to 84 Mb/s
    // net::BufferedSender<400, 0> mSender;

    // massive decrease, don't even bother
    // net::BufferedSender<100, 100> mSender;

    // Stable gain to 84 Mb/s, ideal but test func fails due to receiving ~80 less packets than expected
    // net::BufferedSender<10, 1000> mSender;

    // gets test func to pass
    // net::BufferedSender<3, 750> mSender;

    void processPacket(GenericPacket<>& packet, mmsghdr& header, GenericPacketBuffer<MaxPacketsToSend>& outputBuff);

    bool getAddrFromMsgHdr(net::Address& addr, const msghdr& hdr) const;
  };

  [[gnu::always_inline]] inline bool PacketProcessor::getAddrFromMsgHdr(net::Address& addr, const msghdr& hdr) const
  {
    bool retval = false;
    auto sockad = reinterpret_cast<sockaddr*>(hdr.msg_name);

    switch (sockad->sa_family) {
      case AF_INET: {
        auto sin = reinterpret_cast<sockaddr_in*>(sockad);
        addr = *sin;
        retval = true;
      } break;
      case AF_INET6: {
        auto sin = reinterpret_cast<sockaddr_in6*>(sockad);
        addr = *sin;
        retval = true;
      } break;
    }

    return retval;
  }
}  // namespace core
#endif