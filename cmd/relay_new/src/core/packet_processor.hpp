#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "crypto/keychain.hpp"
#include "os/platform.hpp"
#include "packet.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "token.hpp"
#include "util/throughput_recorder.hpp"

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
     const util::Clock& relayClock,
     const crypto::Keychain& keychain,
     core::SessionMap& sessions,
     core::RelayManager& relayManager,
     const volatile bool& handle,
     util::ThroughputRecorder& recorder,
     const net::Address& receivingAddr);
    ~PacketProcessor() = default;

    void process(std::condition_variable& var, std::atomic<bool>& readyToReceive);

   private:
    const std::atomic<bool>& mShouldReceive;
    const os::Socket& mSocket;
    const util::Clock& mRelayClock;
    const crypto::Keychain& mKeychain;
    core::SessionMap& mSessionMap;
    core::RelayManager& mRelayManager;
    const volatile bool& mShouldProcess;
    util::ThroughputRecorder& mRecorder;
    const net::Address& mRecvAddr;

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