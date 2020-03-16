#ifndef OS_LINUX_THREAD_HPP
#define OS_LINUX_THREAD_HPP
namespace os
{
  inline bool SetThreadAffinity(std::thread& thread, int cpuNumber, int& error)
  {
    auto handle = thread.native_handle();
    cpu_set_t cpuset;
    CPU_ZERO(&cpuset);
    CPU_SET(cpuNumber, &cpuset);
    return (error = pthread_setaffinity_np(handle, sizeof(cpuset), &cpuset)) == 0;
  }
}  // namespace os
#endif