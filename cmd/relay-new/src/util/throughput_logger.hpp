#ifndef UTIL_THROUGHPUT_LOGGER_HPP
#define UTIL_THROUGHPUT_LOGGER_HPP

using namespace std::chrono_literals;

namespace util
{
  class ThroughputLogger
  {
   public:
    ThroughputLogger();

    void addToTotal(size_t count);

    void stop();

   private:
    std::atomic<bool> mAlive;
    size_t mTotalBytes = 0;
    std::unique_ptr<std::thread> mPrintThread;
    std::mutex mTotalLock;
  };

  inline ThroughputLogger::ThroughputLogger() : mAlive(true)
  {
    mPrintThread = std::make_unique<std::thread>([this] {
      while (this->mAlive) {
        std::this_thread::sleep_for(1s);
        this->mTotalLock.lock();
        std::cout << "Bytes received: " << this->mTotalBytes << "/s\n";
        this->mTotalBytes = 0;
        this->mTotalLock.unlock();
      }
    });
  }

  inline void ThroughputLogger::addToTotal(size_t count)
  {
    mTotalLock.lock();
    mTotalBytes += count;
    mTotalLock.unlock();
  }

  inline void ThroughputLogger::stop() {
    mAlive = false;
    mPrintThread->join();
  }
}  // namespace util
#endif