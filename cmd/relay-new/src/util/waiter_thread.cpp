#include "includes.h"
#include "waiter_thread.hpp"

namespace util
{
  WaiterThread::WaiterThread(): mAlive(true), mHasJob(false)
  {
    mThread = std::make_unique<std::thread>([this] {
      while (mAlive) {
        std::unique_lock<std::mutex> lock(mWaitLock);
        mWaitVar.wait(lock, [this] {
          return mHasJob || !mAlive;
        });

        if (!mAlive) {
          return;
        }

        mJobLock.lock();
        {
          mJob();
          if (mOnFinish) {
            mOnFinish();
          }
          mHasJob = false;
        }
        mJobLock.unlock();
      }
    });
  }

  WaiterThread::~WaiterThread()
  {
    terminate();
  }

  void WaiterThread::run(ThreadFunc job)
  {
    mJobLock.lock();
    {
      mJob = job;
      mHasJob = true;
      mWaitVar.notify_one();
    }
    mJobLock.unlock();
  }

  void WaiterThread::onFinish(ThreadFunc onfinish)
  {
    mJobLock.lock();
    {
      mOnFinish = onfinish;
    }
    mJobLock.unlock();
  }

  void WaiterThread::terminate()
  {
    if (mAlive) {
      mAlive = false;
      mWaitVar.notify_one();
      mThread->join();
    }
  }
}  // namespace util