#ifndef UTIL_THROUGHPUT_LOGGER_HPP
#define UTIL_THROUGHPUT_LOGGER_HPP

#include "util/console.hpp"

using namespace std::chrono_literals;

namespace util
{
  class ThroughputLogger
  {
   public:
    ThroughputLogger();
    ~ThroughputLogger();

    void addToRecvTotal(size_t count);
    void addToSentTotal(size_t count);

    void stop();

   private:
    std::mutex mLock;
    std::atomic<bool> mAlive;
    size_t mTotalRecvBytes;
    size_t mTotalRecvPackets;
    size_t mTotalSentBytes;
    size_t mTotalSentPackets;
    std::unique_ptr<std::thread> mPrintThread;
    std::ofstream mOutput;
    util::Console mConsole;
  };

  inline ThroughputLogger::ThroughputLogger()
   : mAlive(true), mTotalRecvBytes(0), mTotalRecvPackets(0), mTotalSentBytes(0), mTotalSentPackets(0), mConsole(mOutput)
  {
    mOutput.open("log.txt", std::ios::out);

    if (!mOutput) {
      std::cout << "Could not log to file";
      mAlive = false;
      return;
    }

    mPrintThread = std::make_unique<std::thread>([this] {
      while (this->mAlive) {
        std::this_thread::sleep_for(1s);

        size_t totalRecvBytes;
        size_t totalRecvPackets;
        size_t totalSentBytes;
        size_t totalSentPackets;

        mLock.lock();
        {
          totalRecvBytes = this->mTotalRecvBytes;
          totalRecvPackets = this->mTotalRecvPackets;
          totalSentBytes = this->mTotalSentBytes;
          totalSentPackets = this->mTotalSentPackets;
          this->mTotalRecvBytes = 0;
          this->mTotalRecvPackets = 0;
          this->mTotalSentBytes = 0;
          this->mTotalSentPackets = 0;
        }
        mLock.unlock();

        mConsole.log("Bytes received: ", totalRecvBytes, "/s\n");
        mConsole.log("Packets received: ", totalRecvPackets, "/s\n");
        mConsole.log("Bytes sent: ", totalSentBytes, "/s\n");
        mConsole.log("Packets sent: ", totalSentPackets, "/s\n");
      }
    });
  }

  inline ThroughputLogger::~ThroughputLogger()
  {
    stop();
  }

  inline void ThroughputLogger::addToRecvTotal(size_t count)
  {
    mLock.lock();
    mTotalRecvBytes += count;
    mTotalRecvPackets++;
    mLock.unlock();
  }

  inline void ThroughputLogger::addToSentTotal(size_t count)
  {
    mLock.lock();
    mTotalSentBytes += count;
    mTotalSentPackets++;
    mLock.unlock();
  }

  inline void ThroughputLogger::stop()
  {
    mAlive = false;
    if (mPrintThread && mPrintThread->joinable()) {
      mPrintThread->join();
    }
  }
}  // namespace util
#endif