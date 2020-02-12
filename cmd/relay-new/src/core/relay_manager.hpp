#ifndef CORE_RELAY_MANAGER_HPP
#define CORE_RELAY_MANAGER_HPP

#include "relay_stats.hpp"
#include "ping_history.hpp"

#include "net/address.hpp"

#include "util/clock.hpp"

namespace core
{
  const auto INVALID_PING_TIME = -10000.0;
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
    unsigned int mNumRelays;
    std::array<uint64_t, MAX_RELAYS> mRelayIDs;
    std::array<double, MAX_RELAYS> mLastRelayPingTime;
    std::array<net::Address, MAX_RELAYS> mRelayAddresses;
    std::array<PingHistory*, MAX_RELAYS> mRelayPingHistory;
    std::array<PingHistory, MAX_RELAYS> mPingHistoryArray;
  };

  /*
    things attempted to make faster:

    - using an unordered_map encapsulating the 5 arrays, and also reducing the n^2 loop to just n, reducing the 5 copy() calls to just a single map.insert(begin, end) call, and combining the 3'rd for loop into the first to reduce overall complexity from O(n^2 + 7n) to just O(3n).
    -- results: the maps iteration at the end took far longer than the amount of time saved by the other optimizations (which would cause getStats() to become much, much slower even if that last loop was brought into the main), if the max relay size ever exceeds a million then maybe the map solution will be more beneficial, but anything under arrays seem to win, even with the bad time complexity

    - using an unordered_set as a relay id cache, reducing the n^2 loop to just n and combinging
    -- results: doing so meant the PingHistory pointer array coudn't be used, and instead the constructor or clear() had to be used to reset the index, and that was slower than the n^2 loop iteration with the pointers

    - End result, this is as fast as it's gonna get until there's spare time for someone to really rework it

    - One last idea, maybe go back to the unordered map concept, and since iterating the first loop is blazingly fast bring the last loop (the one with the time stuff) into the first. And use an array of struct pointers where the struct is the encapsulated info currently in the 5 arrays. So getStats() can be iterated just as fast while the relay data can be updated in just O(2n) time. Doing so would require the relay array to have it's internal's shifted a lot during the function, which may be more expensive than the time saved by reducing all the loops again.
  */

  // it is used in one place throughout the codebase, so always inline it, no sense in doing a function call
  [[gnu::always_inline]] inline void RelayManager::update(unsigned int numRelays,
   const std::array<uint64_t, MAX_RELAYS>& relayIDs,
   const std::array<net::Address, MAX_RELAYS>& relayAddrs)
  {
    assert(numRelays <= MAX_RELAYS);

    // first copy all current relays that are also in the update lists

    std::array<bool, MAX_RELAYS> historySlotToken{false};
    std::array<bool, MAX_RELAYS> found{false};
    std::array<uint64_t, MAX_RELAYS> newRelayIDs{0};
    std::array<double, MAX_RELAYS> newRelayLastPingTime{0};
    std::array<net::Address, MAX_RELAYS> newRelayAddresses;
    std::array<PingHistory*, MAX_RELAYS> newRelayPingHistory{nullptr};

    unsigned int index = 0;

    for (unsigned int i = 0; i < mNumRelays; i++) {
      for (unsigned int j = 0; j < numRelays; j++) {
        if (mRelayIDs[i] == relayIDs[j]) {
          found[j] = true;

          newRelayIDs[index] = mRelayIDs[i];
          newRelayLastPingTime[index] = mLastRelayPingTime[i];
          newRelayAddresses[index] = mRelayAddresses[i];
          newRelayPingHistory[index] = mRelayPingHistory[i];

          const auto slot = mRelayPingHistory[i] - mPingHistoryArray.data();  // TODO make this more readable
          assert(slot >= 0);
          assert(slot < MAX_RELAYS);
          historySlotToken[slot] = true;
          index++;
        }
      }
    }

    // copy all near relays not found in the current relay list

    for (unsigned int i = 0; i < numRelays; i++) {
      if (!found[i]) {
        newRelayIDs[index] = relayIDs[i];
        newRelayLastPingTime[index] = INVALID_PING_TIME;
        newRelayAddresses[index] = relayAddrs[i];
        newRelayPingHistory[index] = nullptr;
        for (int j = 0; j < MAX_RELAYS; j++) {
          if (!historySlotToken[j]) {
            newRelayPingHistory[index] = &mPingHistoryArray[j];
            newRelayPingHistory[index]->clear();
            historySlotToken[j] = true;
            break;
          }
        }
        assert(newRelayPingHistory[index] != nullptr);
        index++;
      }
    }

    // commit the updated relay array
    mNumRelays = index;

    std::copy(newRelayIDs.begin(), newRelayIDs.begin() + index, mRelayIDs.begin());
    std::copy(newRelayLastPingTime.begin(), newRelayLastPingTime.begin() + index, mLastRelayPingTime.begin());
    std::copy(newRelayAddresses.begin(), newRelayAddresses.begin() + index, mRelayAddresses.begin());
    std::copy(newRelayPingHistory.begin(), newRelayPingHistory.begin() + index, mRelayPingHistory.begin());

    // make sure all the ping times are evenly distributed to avoid clusters of ping packets

    auto currentTime = mClock.elapsed<util::Second>();

    for (unsigned int i = 0; i < mNumRelays; i++) {
      mLastRelayPingTime[i] = currentTime - RELAY_PING_TIME + i * RELAY_PING_TIME / mNumRelays;
    }

#ifndef NDEBUG

    // make sure everything is correct

    assert(mNumRelays == index);

    unsigned int numFound = 0;
    for (unsigned int i = 0; i < numRelays; i++) {
      for (unsigned int j = 0; j < mNumRelays; j++) {
        if (relayIDs[i] == mRelayIDs[j] && relayAddrs[i] == mRelayAddresses[j]) {
          numFound++;
          break;
        }
      }
    }

    assert(numFound == mNumRelays);

    for (unsigned int i = 0; i < numRelays; i++) {
      for (unsigned int j = 0; j < numRelays; j++) {
        if (i == j) {
          continue;
        }
        assert(mRelayPingHistory[i] != mRelayPingHistory[j]);
      }
    }
#endif
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
