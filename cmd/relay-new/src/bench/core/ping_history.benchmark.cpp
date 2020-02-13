#include "includes.h"
#include "bench/bench.hpp"

#include "core/ping_history.hpp"

const auto REPS = 10000000;

Bench(PingHistory_sent)
{
  // new
  {
    core::PingHistory ph;

    Do(REPS)
    {
      ph.pingSent(123.456);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "new: " << elapsed << '\n';
  }

  // old
  {
    legacy::relay_ping_history_t ph;

    Do(REPS)
    {
      legacy::relay_ping_history_ping_sent(&ph, 123.456);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "legacy: " << elapsed << '\n';
  }
}

Bench(PingHistory_clear)
{
  // new
  {
    core::PingHistory ph;

    Do(REPS)
    {
      ph.clear();
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "new: " << elapsed << '\n';
  }

  // legacy
  {
    legacy::relay_ping_history_t ph;

    Do(REPS)
    {
      legacy::relay_ping_history_clear(&ph);
    }

    auto elapsed = Timer.elapsed<util::Nanosecond>() / REPS;
    std::cout << "legacy: " << elapsed << '\n';
  }
}
