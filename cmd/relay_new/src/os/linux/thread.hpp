#ifndef OS_LINUX_THREAD_HPP
#define OS_LINUX_THREAD_HPP
namespace os
{
  inline auto SetThreadAffinity(std::thread& thread, int cpuNumber, int& error) -> bool
  {
    cpu_set_t cpuset;
    CPU_ZERO(&cpuset);
    CPU_SET(cpuNumber, &cpuset);
    return (error = pthread_setaffinity_np(thread.native_handle(), sizeof(cpuset), &cpuset)) == 0;
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