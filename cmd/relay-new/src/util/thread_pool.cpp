#include "includes.h"
#include "thread_pool.hpp"

namespace util
{
  ThreadPool::ThreadPool(unsigned int count): mAlive(true)
  {
    mWorkers.resize(count);
    for (unsigned int i = 0; i < count; i++) {
      auto worker = std::make_shared<WaiterThread>();
      worker->onFinish([this, worker] {
        mWorkerLock.lock();
        {
          mFreeWorkers.push(worker);
          mWaitVar.notify_one();
        }
        mWorkerLock.unlock();
      });
      mWorkers[i] = worker;
      mFreeWorkers.push(worker);
    }

    mJobDispatcher = std::make_unique<std::thread>([this] {
      while (mAlive) {
        // wait while no workers or no jobs
        std::unique_lock<std::mutex> lock(mWaitLock);
        mWaitVar.wait(lock, [this] {
          return (mFreeWorkers.size() > 0 && mJobs.size() > 0) || !mAlive;
        });

        // return if dead
        if (!mAlive) {
          return;
        }

        mWorkerLock.lock();
        mJobLock.lock();
        {
          const auto count = mFreeWorkers.size();
          for (size_t i = 0; i < count; i++) {
            auto worker = mFreeWorkers.front();
            mFreeWorkers.pop();

            worker->run(mJobs.front());
            mJobs.pop();

            if (mJobs.size() == 0) {
              break;
            }
          }
        }
        mJobLock.unlock();
        mWorkerLock.unlock();
      }
    });
  }

  ThreadPool::~ThreadPool()
  {
    terminate();
  }

  void ThreadPool::push(ThreadFunc job)
  {
    mJobLock.lock();
    {
      mJobs.push(job);
      mWaitVar.notify_one();
    }
    mJobLock.unlock();
  }

  void ThreadPool::terminate()
  {
    if (mAlive) {
      mAlive = false;
      mWaitVar.notify_one();
      mJobDispatcher->join();
      mJobDispatcher.reset();
      for (auto& worker : mWorkers) {
        worker->terminate();
        worker.reset();
      }
    }
  }
}  // namespace util