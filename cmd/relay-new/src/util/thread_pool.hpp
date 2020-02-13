#ifndef UTIL_THREAD_POOL
#define UTIL_THREAD_POOL

#include "waiter_thread.hpp"

namespace util
{
  using ThreadPtr = std::shared_ptr<WaiterThread>;

  class ThreadPool
  {
   public:
    ThreadPool(unsigned int count);
    ~ThreadPool();

    void push(ThreadFunc job);

    void terminate();

   private:
    std::atomic<bool> mAlive;
    std::unique_ptr<std::thread> mJobDispatcher;
    std::vector<ThreadPtr> mWorkers;
    std::queue<ThreadPtr> mFreeWorkers;
    std::queue<ThreadFunc> mJobs;

    std::mutex mGeneralLock;
    std::mutex mWaitLock;
    std::condition_variable mWaitVar;
  };
}  // namespace util
#endif