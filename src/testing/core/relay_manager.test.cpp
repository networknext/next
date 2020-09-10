#include "includes.h"
#include "testing/test.hpp"

#include "core/relay_manager.hpp"

using core::RelayManager;
using core::RelayPingInfo;
using core::RelayStats;
using core::MAX_RELAYS;

TEST(core_RelayManager_general)
{
  const int num_relays = 32;

  std::array<RelayPingInfo, MAX_RELAYS> incoming;

  for (size_t i = 0; i < MAX_RELAYS; ++i) {
    auto& relay = incoming[i];
    relay.id = i;
    std::stringstream ss;
    ss << "127.0.0.1:" << 40000 + i;
    CHECK(relay.address.parse(ss.str()) == true);
    CHECK(relay.address.port == 40000 + i);
  }

  RelayManager manager;

  // should be no relays when manager is first created
  {
    RelayStats stats;
    manager.get_stats(stats);
    CHECK(stats.num_relays == 0);
  }

  // add max relays
  manager.update(num_relays, incoming);
  {
    RelayStats stats;
    manager.get_stats(stats);
    CHECK(stats.num_relays == num_relays);
    for (int i = 0; i < num_relays; ++i) {
      CHECK(std::find_if(incoming.begin(), incoming.end(), [&](const RelayPingInfo& relay) -> bool {
              return relay.id == stats.ids[i];
            }) != incoming.end());
    }
  }

  // remove all relays

  manager.update(0, incoming);
  {
    RelayStats stats;
    manager.get_stats(stats);
    CHECK(stats.num_relays == 0);
  }

  // add same relay set repeatedly

  for (int j = 0; j < 2; ++j) {
    manager.update(num_relays, incoming);
    {
      RelayStats stats;
      manager.get_stats(stats);
      CHECK(stats.num_relays == num_relays);
      for (int i = 0; i < num_relays; ++i) {
        CHECK(incoming[i].id == stats.ids[i]);
      }
    }
  }

  // now add a few new relays, while some relays remain the same

  std::array<RelayPingInfo, MAX_RELAYS> diff_relays;
  std::copy(incoming.begin() + 4, incoming.end(), diff_relays.begin());
  manager.update(num_relays, diff_relays);
  {
    RelayStats stats;
    manager.get_stats(stats);
    CHECK(stats.num_relays == num_relays);
    for (int i = 0; i < num_relays; ++i) {
      CHECK(incoming[i + 4].id == stats.ids[i]);
    }
  }
}
