#include "includes.h"
#include "testing/test.hpp"

#include "core/relay_manager.hpp"

using core::RelayManager;
using core::RelayPingInfo;
using core::RelayStats;

Test(core_RelayManager_general)
{
  const int MaxRelays = MAX_RELAYS;
  const int NumRelays = 32;

  std::array<RelayPingInfo, MAX_RELAYS> incoming;

  for (int i = 0; i < MaxRelays; ++i) {
    auto& relay = incoming[i];
    relay.id = i;
    std::stringstream ss;
    ss << "127.0.0.1:" << 40000 + i;
    check(relay.address.parse(ss.str()) == true);
    check(relay.address.port == 40000 + i);
  }

  RelayManager manager;

  // should be no relays when manager is first created
  {
    RelayStats stats;
    manager.get_stats(stats);
    check(stats.num_relays == 0);
  }

  // add max relays
  manager.update(NumRelays, incoming);
  {
    RelayStats stats;
    manager.get_stats(stats);
    check(stats.num_relays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(std::find_if(incoming.begin(), incoming.end(), [&](const RelayPingInfo& relay) -> bool {
              return relay.id == stats.ids[i];
            }) != incoming.end());
    }
  }

  // remove all relays

  manager.update(0, incoming);
  {
    RelayStats stats;
    manager.get_stats(stats);
    check(stats.num_relays == 0);
  }

  // add same relay set repeatedly

  for (int j = 0; j < 2; ++j) {
    manager.update(NumRelays, incoming);
    {
      RelayStats stats;
      manager.get_stats(stats);
      check(stats.num_relays == NumRelays);
      for (int i = 0; i < NumRelays; ++i) {
        check(incoming[i].id == stats.ids[i]);
      }
    }
  }

  // now add a few new relays, while some relays remain the same

  std::array<RelayPingInfo, MAX_RELAYS> diff_relays;
  std::copy(incoming.begin() + 4, incoming.end(), diff_relays.begin());
  manager.update(NumRelays, diff_relays);
  {
    RelayStats stats;
    manager.get_stats(stats);
    check(stats.num_relays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(incoming[i + 4].id == stats.ids[i]);
    }
  }
}
