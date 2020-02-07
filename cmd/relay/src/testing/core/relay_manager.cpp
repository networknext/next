#include "testing/test.hpp"

#include "core/relay_manager.hpp"

Test(RelayManager) {}

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