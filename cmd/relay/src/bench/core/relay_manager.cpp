#include "bench/bench.hpp"

#include <sstream>

#include "core/relay_manager.hpp"

const auto REPS = 1000000U;

Bench(RelayManager_update, true)
{
  const auto MaxRelays = 128U;

  // new
  {
    std::vector<uint64_t> ids(MaxRelays, 0);
    std::vector<net::Address> addrs(MaxRelays, net::Address());
    util::Clock clock;  // satisfy the constructor dependency
    core::RelayManager manager(clock);

    for (auto i = 0U; i < MaxRelays; i++) {
      ids[i] = i;
      std::stringstream ss;
      ss << "127.0.0.1:" << i;
      addrs[i].parse(ss.str());
    }

    Do(REPS)
    {
      manager.update(MaxRelays, ids, addrs);
    }

    auto elapsed = Timer.elapsed<util::Microsecond>() / REPS;
    std::cout << "new: " << elapsed << '\n';
  }

  // legacy
  {
    uint64_t ids[MaxRelays];
    legacy::relay_address_t addrs[MaxRelays];
    legacy::relay_manager_t* manager = legacy::relay_manager_create();

    for (auto i = 0U; i < MaxRelays; i++) {
      ids[i] = i;
      std::stringstream ss;
      ss << "127.0.0.1:" << i;
      legacy::relay_address_parse(&addrs[i], ss.str().c_str());
    }

    Do(REPS)
    {
      legacy::relay_manager_update(manager, MaxRelays, ids, addrs);
    }

    auto elapsed = Timer.elapsed<util::Microsecond>() / REPS;
    std::cout << "legacy: " << elapsed << '\n';
  }
}