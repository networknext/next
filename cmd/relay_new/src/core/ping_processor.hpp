#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "os/platform.hpp"

#include "core/relay_manager.hpp"

namespace core
{
  class PingProcessor
  {
   public:
    PingProcessor(const os::Socket& socket,
     core::RelayManager& relayManger,
     const volatile bool& shouldProcess,
     const net::Address& relayAddr);
    ~PingProcessor() = default;

    void process(std::condition_variable& var, std::atomic<bool>& readyToSend);

   private:
    const os::Socket& mSocket;
    core::RelayManager& mRelayManager;
    const volatile bool& mShouldProcess;
    const net::Address& mRelayAddress;

    void fillMsgHdrWithAddr(msghdr& hdr, const net::Address& addr);
  };
}  // namespace core
#endif