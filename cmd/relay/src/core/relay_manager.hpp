#ifndef CORE_RELAY_MANAGER_HPP
#define CORE_RELAY_MANAGER_HPP

#include <array>
#include <unordered_map>
#include <cassert>

#include "relay_stats.hpp"
#include "ping_history.hpp"

#include "net/address.hpp"

#include "util/clock.hpp"

namespace core
{
  const auto INVALID_PING_TIME = -10000.0;

  struct RelayData
  {
    double PingTime;
    net::Address Addr;
    PingHistory History;
  };

  class RelayManager
  {
   public:
    RelayManager(const util::Clock& clock);
    ~RelayManager() = default;

    void reset();

    void update(unsigned int numRelays,
     const std::array<uint64_t, MAX_RELAYS>& relayIDs,
     const std::array<net::Address, MAX_RELAYS>& relayAddrs);

    auto processPong(const net::Address& from, uint64_t seq) -> bool;

    void getStats(RelayStats& stats);

   private:
    const util::Clock mClock;
    std::unordered_map<uint64_t, RelayData> mRelayData;
  };

  // it is used in one place throughout the codebase, so always inline it, no sense in doing a function call
  [[gnu::always_inline]] inline void RelayManager::update(unsigned int numRelays,
   const std::array<uint64_t, MAX_RELAYS>& relayIDs,
   const std::array<net::Address, MAX_RELAYS>& relayAddrs)
  {
    assert(numRelays <= MAX_RELAYS);

    // first copy all current relays that are also in the update lists
    std::unordered_map<uint64_t, RelayData> newData;

    unsigned int index = 0;

    for (unsigned int i = 0; i < numRelays; i++) {
      if (mRelayData.find(relayIDs[i]) != mRelayData.end()) {
        const auto& data = mRelayData[relayIDs[i]];
        auto& ndat = newData[relayIDs[i]];

        ndat.PingTime = data.PingTime;
        ndat.Addr = data.Addr;
        ndat.History = data.History;
        index++;
      }
    }

    // copy all near relays not found in the current relay list

    for (unsigned int i = 0; i < numRelays; i++) {
      if (mRelayData.find(relayIDs[i]) == mRelayData.end()) {
        auto& data = newData[relayIDs[i]];
        data.PingTime = INVALID_PING_TIME;
        data.Addr = relayAddrs[i];
        data.History.clear();
        index++;
      }
    }

    // commit the updated relay array

    std::copy(newData.begin(), newData.end(), mRelayData.begin());

    // make sure all the ping times are evenly distributed to avoid clusters of ping packets

    auto currentTime = mClock.elapsed<util::Second>();

    unsigned int i = 0;
    for (auto& pair : mRelayData) {
      auto& data = pair.second;
      data.PingTime = currentTime - RELAY_PING_TIME + i * RELAY_PING_TIME / mRelayData.size();
      i++;
    }

    // #ifndef NDEBUG
    //
    //     // make sure everything is correct
    //
    //    assert(mNumRelays == index);
    //
    //    unsigned int numFound = 0;
    //    for (unsigned int i = 0; i < numRelays; i++) {
    //      for (unsigned int j = 0; j < mNumRelays; j++) {
    //        if (relayIDs[i] == mRelayIDs[j] && relayAddrs[i] == mRelayAddresses[j]) {
    //          numFound++;
    //          break;
    //        }
    //      }
    //    }
    //
    //    assert(numFound == mNumRelays);
    //
    //    for (unsigned int i = 0; i < numRelays; i++) {
    //      for (unsigned int j = 0; j < numRelays; j++) {
    //        if (i == j) {
    //          continue;
    //        }
    //        assert(mRelayPingHistory[i] != mRelayPingHistory[j]);
    //      }
    //    }
    // #endif
  }
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