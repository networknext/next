#ifndef CORE_RELAY_MANAGER_HPP
#define CORE_RELAY_MANAGER_HPP

#include "relay_stats.hpp"
#include "ping_history.hpp"

#include "net/address.hpp"

#include "util/clock.hpp"
#include "util/logger.hpp"

namespace core
{
  const auto INVALID_PING_TIME = -10000.0;

  struct PingData
  {
    uint64_t Seq;
    net::Address Addr;
  };

  struct Relay
  {
    uint64_t ID;
    uint64_t V3ID;  // id from old backend
    double LastPingTime = INVALID_PING_TIME;
    net::Address Addr;
    PingHistory* History = nullptr;
  };

  class RelayManager
  {
   public:
    RelayManager(const util::Clock& clock);
    ~RelayManager() = default;

    void reset();

    void update(
     bool v3Update,
     unsigned int numRelays,
     const std::array<uint64_t, MAX_RELAYS>& relayIDs,
     const std::array<net::Address, MAX_RELAYS>& relayAddrs);

    auto processPong(const net::Address& from, uint64_t seq) -> bool;

    void getStats(bool forV3, RelayStats& stats);

    unsigned int getPingData(std::array<PingData, MAX_RELAYS>& data);

   private:
    std::mutex mLock;
    const util::Clock& mClock;
    unsigned int mNumRelays;
    std::array<Relay, MAX_RELAYS> mRelays;
    std::vector<PingHistory> mPingHistoryBuff;
  };

  // it is used in one place throughout the codebase, so always inline it, no sense in doing a function call
  [[gnu::always_inline]] inline void RelayManager::update(
   bool v3Update,
   unsigned int numRelays,
   const std::array<uint64_t, MAX_RELAYS>& relayIDs,
   const std::array<net::Address, MAX_RELAYS>& relayAddrs)
  {
    assert(numRelays <= MAX_RELAYS);

    // first copy all current relays that are also in the update lists

    std::array<bool, MAX_RELAYS> historySlotToken{};
    std::array<bool, MAX_RELAYS> found{};
    std::array<Relay, MAX_RELAYS> newRelays{};

    unsigned int index = 0;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);

      for (unsigned int i = 0; i < mNumRelays; i++) {
        auto& relay = mRelays[i];
        for (unsigned int j = 0; j < numRelays; j++) {
          if (mRelays[i].Addr == relayAddrs[j]) {
            found[j] = true;
            (v3Update ? relay.V3ID : relay.ID) = relayIDs[j];  // always assign the id, needed for first update iteration
            auto& newRelay = newRelays[index++] = relay;

            const auto slot = newRelay.History - mPingHistoryBuff.data();  // TODO make this more readable
            assert(slot >= 0);
            assert(slot < MAX_RELAYS);
            historySlotToken[slot] = true;
          }
        }
      }

      // copy all near relays not found in the current relay list

      for (unsigned int i = 0; i < numRelays; i++) {
        if (!found[i]) {
          auto& newRelay = newRelays[index];
          (v3Update ? newRelay.V3ID : newRelay.ID) = relayIDs[i];
          newRelay.Addr = relayAddrs[i];
          for (int j = 0; j < MAX_RELAYS; j++) {
            if (!historySlotToken[j]) {
              newRelay.History = &mPingHistoryBuff[j];
              newRelay.History->clear();
              historySlotToken[j] = true;
              break;
            }
          }
          assert(newRelays[index].History != nullptr);
          index++;
        }
      }

      // commit the updated relay array
      mNumRelays = index;
      std::copy(newRelays.begin(), newRelays.begin() + index, mRelays.begin());

      // make sure all the ping times are evenly distributed to avoid clusters of ping packets

      auto currentTime = mClock.elapsed<util::Second>();

      for (unsigned int i = 0; i < mNumRelays; i++) {
        mRelays[i].LastPingTime = currentTime - RELAY_PING_TIME + i * RELAY_PING_TIME / mNumRelays;
      }

#ifndef NDEBUG

      // make sure everything is correct

      assert(mNumRelays == index);

      unsigned int numFound = 0;
      for (unsigned int i = 0; i < numRelays; i++) {
        for (unsigned int j = 0; j < mNumRelays; j++) {
          const auto relay = mRelays[j];
          if ((relayIDs[i] == relay.ID || relayIDs[i] == relay.V3ID) && relayAddrs[i] == mRelays[j].Addr) {
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
          assert(mRelays[i].History != mRelays[j].History);
        }
      }
#endif
    }
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
