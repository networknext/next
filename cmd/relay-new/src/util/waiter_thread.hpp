#ifndef UTIL_WAITER_THREAD
#define UTIL_WAITER_THREAD
namespace util
{
  using ThreadFunc = std::function<void(void)>;

  class WaiterThread
  {
   public:
    WaiterThread();
    ~WaiterThread();

    void run(ThreadFunc job);

    void onFinish(ThreadFunc onfinish);

    void terminate();

   private:
    std::atomic<bool> mAlive;
    std::atomic<bool> mHasJob;
    std::unique_ptr<std::thread> mThread;
    std::mutex mWaitLock;
    std::mutex mJobLock;
    std::condition_variable mWaitVar;
    ThreadFunc mJob;
    ThreadFunc mOnFinish;
  };
}  // namespace util
#endif