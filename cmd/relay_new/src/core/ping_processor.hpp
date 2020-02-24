#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "os/platform.hpp"

#include "core/relay_manager.hpp"

namespace core
{
  class PingProcessor
  {
   public:
    PingProcessor(core::RelayManager& relayManger, volatile bool& shouldProcess, const net::Address& relayAddr);
    ~PingProcessor() = default;

    void process(os::Socket& socket, std::condition_variable& var, std::atomic<bool>& readyToSend);

   private:
    core::RelayManager& mRelayManager;
    volatile bool& mShouldProcess;
    const net::Address& mRelayAddress;
  };
}  // namespace core
#endif