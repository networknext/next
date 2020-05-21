#include "includes.h"
#include "relay_manager.hpp"

#include <cstring>

namespace legacy
{
  relay_manager_t* relay_manager_create()
  {
    relay_manager_t* manager = (relay_manager_t*)malloc(sizeof(relay_manager_t));
    if (!manager)
      return NULL;
    relay_manager_reset(manager);
    return manager;
  }

  void relay_manager_reset(relay_manager_t* manager)
  {
    assert(manager);
    manager->num_relays = 0;
    memset(manager->relay_ids, 0, sizeof(manager->relay_ids));
    memset(manager->relay_last_ping_time, 0, sizeof(manager->relay_last_ping_time));
    memset(manager->relay_addresses, 0, sizeof(manager->relay_addresses));
    memset(manager->relay_ping_history, 0, sizeof(manager->relay_ping_history));
    for (int i = 0; i < MAX_RELAYS; ++i) {
      relay_ping_history_clear(&manager->ping_history_array[i]);
    }
  }

  void relay_manager_update(
   relay_manager_t* manager, int num_relays, const uint64_t* relay_ids, const legacy::relay_address_t* relay_addresses)
  {
    assert(manager);
    assert(num_relays >= 0);
    assert(num_relays <= MAX_RELAYS);
    assert(relay_ids);
    assert(relay_addresses);

    // first copy all current relays that are also in the updated relay relay list

    bool history_slot_taken[MAX_RELAYS];
    memset(history_slot_taken, 0, sizeof(history_slot_taken));

    bool found[MAX_RELAYS];
    memset(found, 0, sizeof(found));

    uint64_t new_relay_ids[MAX_RELAYS];
    double new_relay_last_ping_time[MAX_RELAYS];
    legacy::relay_address_t new_relay_addresses[MAX_RELAYS];
    legacy::relay_ping_history_t* new_relay_ping_history[MAX_RELAYS];

    int index = 0;

    for (int i = 0; i < manager->num_relays; ++i) {
      for (int j = 0; j < num_relays; ++j) {
        if (manager->relay_ids[i] == relay_ids[j]) {
          found[j] = true;
          new_relay_ids[index] = manager->relay_ids[i];
          new_relay_last_ping_time[index] = manager->relay_last_ping_time[i];
          new_relay_addresses[index] = manager->relay_addresses[i];
          new_relay_ping_history[index] = manager->relay_ping_history[i];
          const int slot = manager->relay_ping_history[i] - manager->ping_history_array;
          assert(slot >= 0);
          assert(slot < MAX_RELAYS);
          history_slot_taken[slot] = true;
          index++;
          break;
        }
      }
    }

    // now copy all near relays not found in the current relay list

    for (int i = 0; i < num_relays; ++i) {
      if (!found[i]) {
        new_relay_ids[index] = relay_ids[i];
        new_relay_last_ping_time[index] = -10000.0;
        new_relay_addresses[index] = relay_addresses[i];
        new_relay_ping_history[index] = NULL;
        for (int j = 0; j < MAX_RELAYS; ++j) {
          if (!history_slot_taken[j]) {
            new_relay_ping_history[index] = &manager->ping_history_array[j];
            relay_ping_history_clear(new_relay_ping_history[index]);
            history_slot_taken[j] = true;
            break;
          }
        }
        assert(new_relay_ping_history[index]);
        index++;
      }
    }

    // commit the updated relay array

    manager->num_relays = index;
    memcpy(manager->relay_ids, new_relay_ids, 8 * index);
    memcpy(manager->relay_last_ping_time, new_relay_last_ping_time, 8 * index);
    memcpy(manager->relay_addresses, new_relay_addresses, sizeof(legacy::relay_address_t) * index);
    memcpy(manager->relay_ping_history, new_relay_ping_history, sizeof(legacy::relay_ping_history_t*) * index);

    // make sure all ping times are evenly distributed to avoid clusters of ping packets

    double current_time = relay::relay_platform_time();

    if (manager->num_relays > 0) {
      for (int i = 0; i < manager->num_relays; ++i) {
        manager->relay_last_ping_time[i] = current_time - RELAY_PING_TIME + i * RELAY_PING_TIME / manager->num_relays;
      }
    }

#ifndef NDEBUG

    // make sure everything is correct

    assert(num_relays == index);

    int num_found = 0;
    for (int i = 0; i < num_relays; ++i) {
      for (int j = 0; j < manager->num_relays; ++j) {
        if (
         relay_ids[i] == manager->relay_ids[j] && relay_address_equal(&relay_addresses[i], &manager->relay_addresses[j]) == 1) {
          num_found++;
          break;
        }
      }
    }
    assert(num_found == num_relays);

    for (int i = 0; i < num_relays; ++i) {
      for (int j = 0; j < num_relays; ++j) {
        if (i == j)
          continue;
        assert(manager->relay_ping_history[i] != manager->relay_ping_history[j]);
      }
    }

#endif  // #ifndef DEBUG
  }

  bool relay_manager_process_pong(relay_manager_t* manager, const legacy::relay_address_t* from, uint64_t sequence)
  {
    assert(manager);
    assert(from);

    for (int i = 0; i < manager->num_relays; ++i) {
      if (relay_address_equal(from, &manager->relay_addresses[i])) {
        relay_ping_history_pong_received(manager->relay_ping_history[i], sequence, relay::relay_platform_time());
        return true;
      }
    }

    return false;
  }

  void relay_manager_get_stats(relay_manager_t* manager, relay_stats_t* stats)
  {
    assert(manager);
    assert(stats);

    double current_time = relay::relay_platform_time();

    stats->num_relays = manager->num_relays;

    for (int i = 0; i < stats->num_relays; ++i) {
      legacy::relay_route_stats_t route_stats;
      legacy::relay_route_stats_from_ping_history(
       manager->relay_ping_history[i], current_time - RELAY_STATS_WINDOW, current_time, &route_stats, RELAY_PING_SAFETY);
      stats->relay_ids[i] = manager->relay_ids[i];
      stats->relay_rtt[i] = route_stats.rtt;
      stats->relay_jitter[i] = route_stats.jitter;
      stats->relay_packet_loss[i] = route_stats.packet_loss;
    }
  }

  void relay_manager_destroy(relay_manager_t* manager)
  {
    free(manager);
  }
}  // namespace legacy
