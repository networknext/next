#ifndef UTIL_WAIT_GROUP_HPP
#define UTIL_WAIT_GROUP_HPP
namespace util
{
  class WaitGroup
  {
   public:
    WaitGroup(unsigned int signalCount);

    /* Signal to the wait group that a task is done */
    void signal();

    /* Wait until the specified number of signals are reached */
    void wait();

   private:
    const std::atomic<unsigned int> mSignalCount;
    std::atomic<unsigned int> mSignals;
    std::mutex mSignalLock;
    std::mutex mWaitLock;
    std::condition_variable mWaitVar;
  };
}  // namespace uitl
#endif
