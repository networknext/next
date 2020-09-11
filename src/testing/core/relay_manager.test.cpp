#include "includes.h"
#include "testing/test.hpp"

#include "core/relay_manager.hpp"

#define NET_HELPERS
#include "testing/helpers.hpp"

using namespace std::chrono_literals;

using core::INVALID_PING_TIME;
using core::MAX_RELAYS;
using core::PingData;
using core::Relay;
using core::RelayManager;
using core::RelayPingInfo;
using core::RelayStats;
using net::Address;

TEST(core_RelayManager_reset)
{
  RelayManager manager;
  std::array<RelayPingInfo, MAX_RELAYS> incoming;
  incoming[0].id = random_whole<uint64_t>();
  incoming[0].address = random_address();

  CHECK(manager.num_relays == 0);
  CHECK(manager.relays[0].id == 0);
  CHECK(manager.relays[0].address == Address());
  CHECK(manager.relays[0].history == nullptr);
  CHECK(manager.relays[0].last_ping_time == INVALID_PING_TIME);

  manager.update(1, incoming);

  CHECK(manager.num_relays == 1);
  CHECK(manager.relays[0].id == incoming[0].id);
  CHECK(manager.relays[0].address == incoming[0].address);
  CHECK(manager.relays[0].history != nullptr);
  CHECK(manager.relays[0].last_ping_time != INVALID_PING_TIME);

  manager.reset();

  CHECK(manager.num_relays == 0);
  CHECK(manager.relays[0].id == 0);
  CHECK(manager.relays[0].address == Address());
  CHECK(manager.relays[0].history == nullptr);
  CHECK(manager.relays[0].last_ping_time == INVALID_PING_TIME);
}

TEST(core_RelayManager_update)
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

TEST(core_RelayManager_get_ping_targets)
{
  RelayManager manager;
  std::array<RelayPingInfo, MAX_RELAYS> incoming;
  incoming[0].id = random_whole<uint64_t>();
  incoming[0].address = random_address();

  manager.update(1, incoming);

  std::array<PingData, MAX_RELAYS> ping_data;
  CHECK(manager.get_ping_targets(ping_data) == 1);
  CHECK(ping_data[0].sequence == 0);
  CHECK(ping_data[0].address == incoming[0].address);

  std::this_thread::sleep_for(100ms);

  CHECK(manager.get_ping_targets(ping_data) == 1);
  CHECK(ping_data[0].sequence == 1);
  CHECK(ping_data[0].address == incoming[0].address);

  std::this_thread::sleep_for(90ms);
  CHECK(manager.get_ping_targets(ping_data) == 0);
}

TEST(core_RelayManager_process_pong)
{
  RelayManager manager;
  std::array<RelayPingInfo, MAX_RELAYS> incoming;
  incoming[0].id = random_whole<uint64_t>();
  incoming[0].address = random_address();

  manager.update(1, incoming);

  std::array<PingData, MAX_RELAYS> ping_data;
  CHECK(manager.get_ping_targets(ping_data) == 1);

  CHECK(manager.relays[0].last_ping_time != INVALID_PING_TIME);
  CHECK(manager.relays[0].history->entries[ping_data[0].sequence].time_pong_received == -1.0);
  CHECK(manager.process_pong(ping_data[0].address, ping_data[0].sequence));
  CHECK(manager.relays[0].history->entries[ping_data[0].sequence].time_pong_received != -1.0);
}

TEST(core_RelayManager_get_stats)
{
  RelayManager manager;
  std::array<RelayPingInfo, MAX_RELAYS> incoming;
  incoming[0].id = random_whole<uint64_t>();
  incoming[0].address = random_address();

  manager.update(1, incoming);

  RelayStats rs;
  manager.get_stats(rs);

  CHECK(rs.num_relays == 1);
}

TEST(core_RelayManager_copy_existing_relays)
{
  RelayManager manager;

  size_t num_incoming_relays = 1;
  std::array<RelayPingInfo, MAX_RELAYS> incoming;
  incoming[0].id = random_whole<uint64_t>();
  incoming[0].address = random_address();

  CHECK(manager.update(num_incoming_relays, incoming));

  std::array<Relay, MAX_RELAYS> new_relays{};
  std::array<bool, MAX_RELAYS> found{};
  std::array<bool, MAX_RELAYS> history_slot_taken{};

  for (const auto taken : history_slot_taken) {
    CHECK(!taken);
  }

  size_t index = 0;
  CHECK(manager.copy_existing_relays(index, num_incoming_relays, incoming, new_relays, found, history_slot_taken));

  CHECK(new_relays[0] == manager.relays[0]);
  CHECK(history_slot_taken[0]);
}
