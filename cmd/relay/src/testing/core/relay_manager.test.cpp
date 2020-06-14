#include "includes.h"
#include "testing/test.hpp"

#include "core/relay_manager.hpp"

Test(RelayManager)
{
  const int MaxRelays = MAX_RELAYS;
  const int NumRelays = 32;

  std::array<core::Relay, MAX_RELAYS> incoming;

  for (int i = 0; i < MaxRelays; ++i) {
    auto& relay = incoming[i];
    relay.ID = i;
    std::stringstream ss;
    ss << "127.0.0.1:" << 40000 + i;
    check(relay.Addr.parse(ss.str()) == true);
    check(relay.Addr.Port == 40000 + i);
  }

  core::RelayManager<core::Relay> manager;

  // should be no relays when manager is first created
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == 0);
  }

  // add max relays
  manager.update(NumRelays, incoming);
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(std::find_if(incoming.begin(), incoming.end(), [&](const core::Relay& relay) -> bool {
              return relay.ID == stats.IDs[i];
            }) != incoming.end());
    }
  }

  // remove all relays

  manager.update(0, incoming);
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == 0);
  }

  // add same relay set repeatedly

  for (int j = 0; j < 2; ++j) {
    manager.update(NumRelays, incoming);
    {
      core::RelayStats stats;
      manager.getStats(stats);
      check(stats.NumRelays == NumRelays);
      for (int i = 0; i < NumRelays; ++i) {
        check(incoming[i].ID == stats.IDs[i]);
      }
    }
  }

  // now add a few new relays, while some relays remain the same

  std::array<core::Relay, MAX_RELAYS> diffRelays;
  std::copy(incoming.begin() + 4, incoming.end(), diffRelays.begin());
  manager.update(NumRelays, diffRelays);
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(incoming[i + 4].ID == stats.IDs[i]);
    }
  }
}

Test(legacy_relay_manager)
{
  const int MaxRelays = 64;
  const int NumRelays = 32;

  uint64_t relay_ids[MaxRelays];
  legacy::relay_address_t relay_addresses[MaxRelays];

  for (int i = 0; i < MaxRelays; ++i) {
    relay_ids[i] = i;
    char address_string[256];
    sprintf(address_string, "127.0.0.1:%d", 40000 + i);
    legacy::relay_address_parse(&relay_addresses[i], address_string);
  }

  legacy::relay_manager_t* manager = legacy::relay_manager_create();

  // should be no relays when manager is first created
  {
    legacy::relay_stats_t stats;
    relay_manager_get_stats(manager, &stats);
    check(stats.num_relays == 0);
  }

  // add max relays

  relay_manager_update(manager, NumRelays, relay_ids, relay_addresses);
  {
    legacy::relay_stats_t stats;
    relay_manager_get_stats(manager, &stats);
    check(stats.num_relays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(relay_ids[i] == stats.relay_ids[i]);
    }
  }

  // remove all relays

  relay_manager_update(manager, 0, relay_ids, relay_addresses);
  {
    legacy::relay_stats_t stats;
    relay_manager_get_stats(manager, &stats);
    check(stats.num_relays == 0);
  }

  // add same relay set repeatedly

  for (int j = 0; j < 2; ++j) {
    relay_manager_update(manager, NumRelays, relay_ids, relay_addresses);
    {
      legacy::relay_stats_t stats;
      relay_manager_get_stats(manager, &stats);
      check(stats.num_relays == NumRelays);
      for (int i = 0; i < NumRelays; ++i) {
        check(relay_ids[i] == stats.relay_ids[i]);
      }
    }
  }

  // now add a few new relays, while some relays remain the same

  relay_manager_update(manager, NumRelays, relay_ids + 4, relay_addresses + 4);
  {
    legacy::relay_stats_t stats;
    relay_manager_get_stats(manager, &stats);
    check(stats.num_relays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(relay_ids[i + 4] == stats.relay_ids[i]);
    }
  }

  relay_manager_destroy(manager);
}
