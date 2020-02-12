#include "includes.h"
#include "wait_group.hpp"

namespace util
{
  WaitGroup::WaitGroup(unsigned int signalCount): mSignalCount(signalCount), mSignals(0) {}

  void WaitGroup::signal()
  {
    mSignalLock.lock();
    {
      if (++mSignals >= mSignalCount) {
        mWaitVar.notify_one();
      }
    }
    mSignalLock.unlock();
  }

  void WaitGroup::wait()
  {
    std::unique_lock<std::mutex> lock(mWaitLock);
    mWaitVar.wait(lock, [this] {
      return mSignals >= mSignalCount;
    });
  }
}  // namespace util