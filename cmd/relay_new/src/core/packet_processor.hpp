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
  class PacketProcessor
  {
   public:
    PacketProcessor(os::Socket& socket,
     const util::Clock& relayClock,
     const crypto::Keychain& keychain,
     const core::RouterInfo& routerInfo,
     core::SessionMap& sessions,
     core::RelayManager& relayManager,
     volatile bool& handle,
     util::ThroughputLogger* logger);
    ~PacketProcessor() = default;

    void process(std::condition_variable& var, std::atomic<bool>& readyToReceive);

    void flushResponses();

   private:
    const os::Socket& mSocket;
    const util::Clock& mRelayClock;
    const crypto::Keychain& mKeychain;
    const core::RouterInfo mRouterInfo;
    core::SessionMap& mSessionMap;
    core::RelayManager& mRelayManager;
    volatile bool& mShouldProcess;

    util::ThroughputLogger* mLogger;

    net::BufferedSender<1, 1> mSender;

    void processPacket(GenericPacket& packet, mmsghdr& header);

    bool getAddrFromMsgHdr(net::Address& addr, const msghdr& hdr) const;
  };

  inline void PacketProcessor::flushResponses()
  {
    mSender.autoSend();
  }

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