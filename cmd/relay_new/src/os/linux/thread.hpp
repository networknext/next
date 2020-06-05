#ifndef OS_LINUX_THREAD_HPP
#define OS_LINUX_THREAD_HPP
namespace os
{
  inline auto SetThreadAffinity(std::thread& thread, int cpuNumber) -> std::tuple<bool, std::string>
  {
    cpu_set_t cpuset;
    CPU_ZERO(&cpuset);
    CPU_SET(cpuNumber, &cpuset);
    auto res = pthread_setaffinity_np(thread.native_handle(), sizeof(cpuset), &cpuset);
    if (res == 0) {
      return {true, std::string()};
    } else {
      std::stringstream ss;
      ss << "error setting thread affinity: " << res;
      return {false, ss.str()};
    }
  }

  inline auto SetThreadSchedMax(std::thread& thread) -> std::tuple<bool, std::string>
  {
    struct sched_param param;
    param.sched_priority = sched_get_priority_max(SCHED_FIFO);
    int ret = pthread_setschedparam(thread.native_handle(), SCHED_FIFO, &param);
    if (ret) {
      std::string msg = "unable to increase server thread priority: ";
      msg += strerror(ret);
      return {false, msg};
    }

    return {true, std::string()};
  }
}  // namespace os
#endif