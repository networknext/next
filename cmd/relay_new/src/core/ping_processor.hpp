#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "os/platform.hpp"

#include "relay/relay.hpp"

namespace core
{
  class PingProcessor
  {
   public:
    PingProcessor(relay::relay_t& relay, volatile bool& handle);
    ~PingProcessor() = default;

    void listen(os::Socket& socket);

   private:
    relay::relay_t& mRelay;
    volatile bool& mShouldProcess;
  };
}  // namespace core
#endif