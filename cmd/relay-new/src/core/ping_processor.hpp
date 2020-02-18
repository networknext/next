#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "os/platform.hpp"

#include "relay/relay.hpp"

namespace core
{
  class PingProcessor
  {
   public:
    PingProcessor(os::Socket& socket, relay::relay_t& relay, volatile bool& handle);
    ~PingProcessor();

    void listen();

    void stop();

   private:
    os::Socket& mSocket;
    relay::relay_t mRelay;
    volatile bool& mHandle;
  };
}  // namespace core
#endif