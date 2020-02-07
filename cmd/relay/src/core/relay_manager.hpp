#ifndef CORE_RELAY_MANAGER_HPP
#define CORE_RELAY_MANAGER_HPP

#include <array>

#include "relay_stats.hpp"
#include "ping_history.hpp"

#include "net/address.hpp"

#include "util/clock.hpp"

namespace core
{
  class RelayManager
  {
   public:
    RelayManager(const util::Clock& clock);
    ~RelayManager() = default;

    void reset();

    // void update(unsigned int numRelays, const std::array<uint64_t, MAX_RELAYS>& relayIDs, const std::array<net::Address,
    // MAX_RELAYS>& relayAddrs);
    void update(unsigned int numRelays, const uint64_t* relayIDs, const net::Address* relayAddrs);

    auto processPong(const net::Address& from, uint64_t seq) -> bool;

    void getStats(RelayStats& stats);

   private:
    const util::Clock mClock;
    unsigned int mNumRelays;
    std::array<uint64_t, MAX_RELAYS> mRelayIDs;
    std::array<double, MAX_RELAYS> mLastRelayPingTime;
    std::array<net::Address, MAX_RELAYS> mRelayAddresses;
    std::array<PingHistory*, MAX_RELAYS> mRelayPingHistory;
    std::array<PingHistory, MAX_RELAYS> mPingHistoryArray;
  };
}  // namespace core

namespace legacy
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
    legacy::relay_ping_history_t* relay_ping_history[MAX_RELAYS];
    legacy::relay_ping_history_t ping_history_array[MAX_RELAYS];
  };

  relay_manager_t* relay_manager_create();

  void relay_manager_reset(relay_manager_t* manager);

  void relay_manager_update(
   relay_manager_t* manager, int num_relays, const uint64_t* relay_ids, const legacy::relay_address_t* relay_addresses);

  bool relay_manager_process_pong(relay_manager_t* manager, const legacy::relay_address_t* from, uint64_t sequence);

  void relay_manager_get_stats(relay_manager_t* manager, relay_stats_t* stats);

  void relay_manager_destroy(relay_manager_t* manager);

}  // namespace legacy
#endif