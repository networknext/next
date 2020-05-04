#ifndef CORE_PING_PROCESSOR_HPP
#define CORE_PING_PROCESSOR_HPP

#include "core/relay_manager.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  class PingProcessor
  {
   public:
    PingProcessor(
     const os::Socket& socket,
     core::RelayManager& relayManger,
     const volatile bool& shouldProcess,
     const net::Address& relayAddr,
     util::ThroughputRecorder& recorder);
    ~PingProcessor() = default;

    void process(std::condition_variable& var, std::atomic<bool>& readyToSend);

   private:
    const os::Socket& mSocket;
    core::RelayManager& mRelayManager;
    const volatile bool& mShouldProcess;
    const net::Address& mRelayAddress;
    util::ThroughputRecorder& mRecorder;

    void fillMsgHdrWithAddr(msghdr& hdr, const net::Address& addr);
  };
}  // namespace core
#endif