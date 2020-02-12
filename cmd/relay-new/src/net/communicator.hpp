#ifndef NET_COMMUNICATOR_HPP
#define NET_COMMUNICATOR_HPP

#include "relay/relay.hpp"
#include "util/throughput_logger.hpp"
#include "util/thread_pool.hpp"

namespace net
{
  class Communicator
  {
   public:
    Communicator(relay::relay_t& relay, volatile bool& handle);
    ~Communicator();

    void stop();

   private:
    relay::relay_t& mRelay;
    volatile bool& mHandle;

    std::unique_ptr<std::thread> mPingThread;
    std::unique_ptr<std::thread> mRecvThread;
    std::unique_ptr<util::ThreadPool> mThreadPool;

    util::ThroughputLogger mLogger;

    void initPingThread();
    void initRecvThread();
  };
}  // namespace net
#endif