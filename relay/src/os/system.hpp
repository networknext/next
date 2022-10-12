#pragma once

#include "util/logger.hpp"

#if defined(linux) || defined(__linux) || defined(__linux__)

namespace os
{
  INLINE double GetCPU()
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
        return 0.0;
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
        return 0.0;
      }

      fclose(f);
    }

    // read the line and get cpu times
    {
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

      int non_idle = user + nice + system + irq + softirq + steal + guest + guest_nice + iowait;

      int total = idle + non_idle;

      usage = static_cast<double>(total - idle) / static_cast<double>(total);
    }

    return usage * 100.0;
  }
}  // namespace os

#endif // #if defined(linux) || defined(__linux) || defined(__linux__)
