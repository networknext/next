#pragma once

#include "util/logger.hpp"

namespace os
{
  class LibTopWrapper
  {
   public:
    LibTopWrapper();

    auto getCPU() -> double;
    auto getMem() -> double;
  };

  inline LibTopWrapper::LibTopWrapper()
  {
    glibtop_init();
  }

  inline auto LibTopWrapper::getCPU() -> double
  {
    glibtop_cpu cpu;
    glibtop_get_cpu(&cpu);
    return static_cast<double>(cpu.total - cpu.idle) / static_cast<double>(cpu.total);
  }

  inline auto LibTopWrapper::getMem() -> double
  {
    glibtop_mem mem;
    glibtop_get_mem(&mem);
    return static_cast<double>(mem.user) / static_cast<double>(mem.total);
  }

  struct SysUsage
  {
    double CPU;
    double Mem;
  };

  inline auto GetUsage() -> SysUsage
  {
    static LibTopWrapper wrapper;
    return SysUsage{
     .CPU = wrapper.getCPU(),
     .Mem = wrapper.getMem(),
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
  inline auto GetUsageAlt() -> std::tuple<double, bool>
  {
    double usage = 0.0;
    // get the first line of /proc/stat
    std::vector<char> lineBuff;
    {
      FILE* f = nullptr;

      f = fopen("/proc/stat", "r");
      if (f == nullptr) {
        LOG(ERROR, "could not open /proc/stat");
        perror("OS msg:");
        return {0, false};
      }

      size_t lineLength = 0;

      while (true) {
        char c = fgetc(f);
        if (c == EOF || c == '\n') {
          break;
        }
        lineLength++;
      }

      lineBuff.resize(lineLength + 1);

      rewind(f);

      if (fgets(lineBuff.data(), lineBuff.size(), f) == nullptr) {
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
      int user, nice, system, idle, iowait, irq, softirq, steal, guest, guestNice;

      sscanf(
       lineBuff.data(),
       // [cpu] user nice system idle iowait irq softirq steal guest guestNice
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
       &guestNice);

      // iowait is added to non-idle because the relay is basically the only thing running on the servers
      // thus waiting is consumed cpu time since threads are locked to cores
      curr.Idle = idle;
      int nonIdle = user + nice + system + irq + softirq + steal + guest + guestNice + iowait;
      curr.Total = curr.Idle + nonIdle;

      int total = curr.Total - prev.Total;
      idle = curr.Idle - prev.Idle;

      usage = static_cast<double>(total - idle) / static_cast<double>(total);

      prev = curr;
    }

    return {usage, true};
  }
}  // namespace os
