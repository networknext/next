#pragma once

#include "util/logger.hpp"

namespace os
{
  class LibTopWrapper
  {
   public:
    LibTopWrapper();

    auto get_cpu() -> double;
    auto get_mem() -> double;
  };

  INLINE LibTopWrapper::LibTopWrapper()
  {
    glibtop_init();
  }

  INLINE auto LibTopWrapper::get_cpu() -> double
  {
    static struct
    {
      uint64_t total = 0;
      uint64_t idle = 0;
    } last_cpu, curr_cpu;
    glibtop_cpu cpu;
    glibtop_get_cpu(&cpu);

    curr_cpu.total = cpu.total - last_cpu.total;
    curr_cpu.idle = cpu.idle - last_cpu.idle;

    last_cpu.total = cpu.total;
    last_cpu.idle = cpu.idle;

    return static_cast<double>(curr_cpu.total - curr_cpu.idle) / static_cast<double>(curr_cpu.total);
  }

  INLINE auto LibTopWrapper::get_mem() -> double
  {
    glibtop_mem mem;
    glibtop_get_mem(&mem);
    return static_cast<double>(mem.user) / static_cast<double>(mem.total);
  }

  struct SysUsage
  {
    double cpu;
    double mem;
  };

  INLINE auto GetUsage() -> SysUsage
  {
    static LibTopWrapper wrapper;
    return SysUsage{
     .cpu = wrapper.get_cpu(),
     .mem = wrapper.get_mem(),
    };
  }

  struct CPUUsageCache
  {
    int Idle;
    int Total;
  };

  // This should basically do what libtop does
  // mainly a sanity check in case libtop behaves weird
  // after the binary is deployed
  INLINE auto GetUsageAlt() -> std::tuple<double, bool>
  {
    double usage = 0.0;
    // get the first line of /proc/stat
    std::vector<char> line_buff;
    {
      FILE* f = nullptr;

      f = fopen("/proc/stat", "r");
      if (f == nullptr) {
        LOG(ERROR, "could not open /proc/stat");
        perror("OS msg:");
        return {0, false};
      }

      size_t line_length = 0;

      while (true) {
        char c = fgetc(f);
        if (c == EOF || c == '\n') {
          break;
        }
        line_length++;
      }

      line_buff.resize(line_length + 1);

      rewind(f);

      if (fgets(line_buff.data(), line_buff.size(), f) == nullptr) {
        LOG(ERROR, "could not read first line of /proc/stat");
        perror("OS msg:");
        return {0, false};
      }

      fclose(f);
    }

    // read the line and get cpu times
    {
      static CPUUsageCache prev;
      CPUUsageCache curr;
      int user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice;

      sscanf(
       line_buff.data(),
       // [cpu] user nice system idle iowait irq softirq steal guest guest_nice
       "%*s %d %d %d %d %d %d %d %d %d %d",
       &user,
       &nice,
       &system,
       &idle,
       &iowait,
       &irq,
       &softirq,
       &steal,
       &guest,
       &guest_nice);

      // iowait is added to non-idle because the relay is basically the only thing running on the servers
      // thus waiting is consumed cpu time since threads are locked to cores
      curr.Idle = idle;
      int non_idle = user + nice + system + irq + softirq + steal + guest + guest_nice + iowait;
      curr.Total = curr.Idle + non_idle;

      int total = curr.Total - prev.Total;
      idle = curr.Idle - prev.Idle;

      usage = static_cast<double>(total - idle) / static_cast<double>(total);

      prev = curr;
    }

    return {usage, true};
  }
}  // namespace os
