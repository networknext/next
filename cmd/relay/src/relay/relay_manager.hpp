#ifndef RELAY_RELAY_MANAGER_HPP
#define RELAY_RELAY_MANAGER_HPP

#include <cinttypes>

#include "config.hpp"
#include "net/address.hpp"
#include "relay_ping_history.hpp"

namespace relay
{
  struct relay_stats_t
  {
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    float relay_rtt[MAX_RELAYS];
    float relay_jitter[MAX_RELAYS];
    float relay_packet_loss[MAX_RELAYS];
  };

  struct relay_manager_t
  {
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    double relay_last_ping_time[MAX_RELAYS];
    legacy::relay_address_t relay_addresses[MAX_RELAYS];
    relay_ping_history_t* relay_ping_history[MAX_RELAYS];
    relay_ping_history_t ping_history_array[MAX_RELAYS];
  };

  relay_manager_t* relay_manager_create();

  void relay_manager_reset(relay_manager_t* manager);

  void relay_manager_update(
   relay_manager_t* manager, int num_relays, const uint64_t* relay_ids, const legacy::relay_address_t* relay_addresses);

  bool relay_manager_process_pong(relay_manager_t* manager, const legacy::relay_address_t* from, uint64_t sequence);

  void relay_manager_get_stats(relay_manager_t* manager, relay_stats_t* stats);

  void relay_manager_destroy(relay_manager_t* manager);
}  // namespace relay
#endif