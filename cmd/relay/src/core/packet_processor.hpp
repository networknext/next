#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "crypto/keychain.hpp"
#include "os/platform.hpp"
#include "packet.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "token.hpp"
#include "util/channel.hpp"
#include "util/throughput_recorder.hpp"
#include "legacy/v3/traffic_stats.hpp"
#include "legacy/v3/constants.hpp"

namespace core
{
  const size_t MaxPacketsToReceive = 1024;

  // same as receive, the amount of output depends on the input
  const size_t MaxPacketsToSend = MaxPacketsToReceive;

  class PacketProcessor
  {
   public:
    PacketProcessor(
     const std::atomic<bool>& shouldReceive,
     os::Socket& socket,
     const crypto::Keychain& keychain,
     SessionMap& sessions,
     RelayManager<Relay>& relayManager,
     RelayManager<V3Relay>& v3RelayManager,
     const volatile bool& handle,
     util::ThroughputRecorder& recorder,
     util::Sender<GenericPacket<>>& sender,
     legacy::v3::TrafficStats& stats,
     const uint64_t oldRelayID,
     const std::atomic<legacy::v3::ResponseState>& state,
     const RouterInfo& routerInfo);
    ~PacketProcessor() = default;

    void process(std::atomic<bool>& readyToReceive);

   private:
    const std::atomic<bool>& mShouldReceive;
    const os::Socket& mSocket;
    const crypto::Keychain& mKeychain;
    SessionMap& mSessionMap;
    RelayManager<Relay>& mRelayManager;
    RelayManager<V3Relay>& mV3RelayManager;
    const volatile bool& mShouldProcess;
    util::ThroughputRecorder& mRecorder;
    util::Sender<GenericPacket<>>& mChannel;
    legacy::v3::TrafficStats& mStats;
    const uint64_t mOldRelayID;
    const std::atomic<legacy::v3::ResponseState>& mState;
    const RouterInfo& mRouterInfo;

    void processPacket(GenericPacket<>& packet, GenericPacketBuffer<MaxPacketsToSend>& outputBuff);

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