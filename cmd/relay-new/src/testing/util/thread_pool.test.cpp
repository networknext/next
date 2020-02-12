#include "includes.h"
#include "testing/test.hpp"
#include "util/thread_pool.hpp"
#include "util/wait_group.hpp"

using namespace std::chrono_literals;

Test(ThreadPool_basic)
{
  const int jobs = 6;

  util::ThreadPool tp(jobs - 3);
  util::WaitGroup wg(jobs);
  int inc = 0;

  for (int i = 0; i < jobs; i++) {
    tp.push([&] {
      inc++;
      wg.signal();
    });
  }

  wg.wait();

  tp.terminate();

  check(inc == 6);
}
