#pragma once

#include "util/logger.hpp"

namespace os
{
  class LibTopWrapper
  {
   public:
    LibTopWrapper();

    auto getCPU() -> double;
    auto getRAM() -> double;
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

  inline auto LibTopWrapper::getRAM() -> double
  {
    glibtop_mem mem;
    glibtop_get_mem(&mem);
    return static_cast<double>(mem.user) / static_cast<double>(mem.total);
  }

  struct SysUsage
  {
    double CPU;
    double RAM;
  };

  inline auto GetUsage() -> SysUsage
  {
    static LibTopWrapper wrapper;
    return SysUsage{
     .CPU = wrapper.getCPU(),
     .RAM = wrapper.getRAM(),
    };
  }
}  // namespace os
