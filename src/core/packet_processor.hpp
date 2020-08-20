#pragma once

#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "packet.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "token.hpp"
#include "util/macros.hpp"

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
     const volatile bool& handle,
     util::ThroughputRecorder& recorder,
     const RouterInfo& routerInfo);
    ~PacketProcessor() = default;

    void process(std::atomic<bool>& readyToReceive);

   private:
    const std::atomic<bool>& mShouldReceive;
    const os::Socket& mSocket;
    const crypto::Keychain& mKeychain;
    SessionMap& mSessionMap;
    RelayManager<Relay>& mRelayManager;
    const volatile bool& mShouldProcess;
    util::ThroughputRecorder& mRecorder;
    const RouterInfo& mRouterInfo;

    void processPacket(GenericPacket<>& packet, GenericPacketBuffer<MaxPacketsToSend>& outputBuff);

    bool getAddrFromMsgHdr(net::Address& addr, const msghdr& hdr) const;
  };

  INLINE bool PacketProcessor::getAddrFromMsgHdr(net::Address& addr, const msghdr& hdr) const
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
