#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "os/platform.hpp"

#include "core/relay_manager.hpp"

namespace core
{
  class PingProcessor
  {
   public:
    PingProcessor(core::RelayManager& relayManger, volatile bool& shouldProcess);
    ~PingProcessor() = default;

    void listen(os::Socket& socket, std::condition_variable& var, std::atomic<bool>& readyToSend);

   private:
    core::RelayManager& mRelayManager;
    volatile bool& mShouldProcess;
  };
}  // namespace core
#endif