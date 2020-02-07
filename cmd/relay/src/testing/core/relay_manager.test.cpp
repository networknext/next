#include "testing/test.hpp"

#include <sstream>
#include <iostream>

#include "core/relay_manager.hpp"

Test(RelayManager)
{
  const int MaxRelays = 64;
  const int NumRelays = 32;

  std::vector<uint64_t> relayIDs(MaxRelays, 0);
  std::vector<net::Address> addrs(MaxRelays, net::Address());
  util::Clock clock;
  clock.reset();

  for (int i = 0; i < MaxRelays; ++i) {
    relayIDs[i] = i;
    std::stringstream ss;
    ss << "127.0.0.1:" << 40000 + i;
    check(addrs[i].parse(ss.str()) == true);
    check(addrs[i].Port == 40000 + i);
  }

  core::RelayManager manager(clock);

  // should be no relays when manager is first created
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == 0);
  }

  // add max relays
  std::cout << __FILE__ << __LINE__ << '\n';
  manager.update(NumRelays, relayIDs, addrs);
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(relayIDs[i] == stats.IDs[i]);
    }
  }

  // remove all relays

  std::cout << __FILE__ << __LINE__ << '\n';
  manager.update(0, relayIDs, addrs);
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == 0);
  }

  // add same relay set repeatedly

  for (int j = 0; j < 2; ++j) {
    std::cout << __FILE__ << __LINE__ << '\n';
    manager.update(NumRelays, relayIDs, addrs);
    {
      core::RelayStats stats;
      manager.getStats(stats);
      check(stats.NumRelays == NumRelays);
      for (int i = 0; i < NumRelays; ++i) {
        check(relayIDs[i] == stats.IDs[i]);
      }
    }
  }

  // now add a few new relays, while some relays remain the same

  std::cout << __FILE__ << __LINE__ << '\n';
  std::vector<uint64_t> diffIDs(relayIDs.begin() + 4, relayIDs.end());
  std::vector<net::Address> diffAddrs(addrs.begin() + 4, addrs.end());
  manager.update(NumRelays, diffIDs, diffAddrs);
  {
    core::RelayStats stats;
    manager.getStats(stats);
    check(stats.NumRelays == NumRelays);
    for (int i = 0; i < NumRelays; ++i) {
      check(relayIDs[i + 4] == stats.IDs[i]);
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