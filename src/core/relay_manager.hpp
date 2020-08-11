#pragma once

#include "core/route_stats.hpp"
#include "net/address.hpp"
#include "ping_history.hpp"
#include "relay/relay_platform.hpp"
#include "relay_stats.hpp"
#include "util/clock.hpp"
#include "util/logger.hpp"
#include "router_info.hpp"
#include "util/clock.hpp"
#include "crypto/hash.hpp"

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
    double LastPingTime = INVALID_PING_TIME;
    net::Address Addr;
    PingHistory* History = nullptr;
  };

  // where T == Relay || V3Relay
  template <typename T>
  class RelayManager
  {
   public:
    RelayManager();
    ~RelayManager() = default;

    void reset();

    void update(size_t numRelays, const std::array<T, MAX_RELAYS>& newRelays);

    auto processPong(const net::Address& from, uint64_t seq) -> bool;

    void getStats(RelayStats& stats);

    // where U == PingData || V3PingData
    template <typename U>
    auto getPingData(std::array<U, MAX_RELAYS>& data) -> size_t;

   private:
    std::mutex mLock;
    unsigned int mNumRelays;
    std::array<T, MAX_RELAYS> mRelays;
    std::vector<PingHistory> mPingHistoryBuff;
    util::Clock mClock;
  };

  template <typename T>
  RelayManager<T>::RelayManager()
  {
    mPingHistoryBuff.resize(MAX_RELAYS);
    reset();
  }

  template <typename T>
  void RelayManager<T>::reset()
  {
    // locked mutex scope
    std::lock_guard<std::mutex> lk(mLock);
    mNumRelays = 0;
    std::fill(mRelays.begin(), mRelays.end(), T());
  }

  template <typename T>
  auto RelayManager<T>::processPong(const net::Address& from, uint64_t seq) -> bool
  {
    bool pongReceived = false;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      for (unsigned int i = 0; i < mNumRelays; i++) {
        auto& relay = mRelays[i];
        if (from == relay.Addr) {
          relay.History->pongReceived(seq, mClock.elapsed<util::Second>());
          pongReceived = true;
          break;
        }
      }
    }

    return pongReceived;
  }

  template <typename T>
  void RelayManager<T>::getStats(RelayStats& stats)
  {
    auto currentTime = mClock.elapsed<util::Second>();

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      stats.NumRelays = mNumRelays;

      for (unsigned int i = 0; i < mNumRelays; i++) {
        auto& relay = mRelays[i];

        RouteStats rs(*relay.History, currentTime - RELAY_STATS_WINDOW, currentTime, RELAY_PING_SAFETY);
        stats.IDs[i] = relay.ID;
        stats.RTT[i] = rs.getRTT();
        stats.Jitter[i] = rs.getJitter();
        stats.PacketLoss[i] = rs.getPacketLoss();
      }
    }
  }

  template <>
  template <>
  inline auto RelayManager<Relay>::getPingData(std::array<PingData, MAX_RELAYS>& data) -> size_t
  {
    double currentTime = mClock.elapsed<util::Second>();
    size_t numPings = 0;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      for (unsigned int i = 0; i < mNumRelays; ++i) {
        if (mRelays[i].LastPingTime + RELAY_PING_TIME <= currentTime) {
          auto& relay = mRelays[i];
          auto& pingData = data[numPings++];

          pingData.Seq = relay.History->pingSent(currentTime);
          pingData.Addr = relay.Addr;
          relay.LastPingTime = currentTime;
        }
      }
    }

    return numPings;
  }

  // it is used in one place throughout the codebase, so always inline it, no sense in doing a function call
  template <>
  [[gnu::always_inline]] inline void RelayManager<Relay>::update(
   size_t numRelays, const std::array<Relay, MAX_RELAYS>& incoming)
  {
    assert(numRelays <= MAX_RELAYS);

    // first copy all current relays that are also in the update lists

    std::array<bool, MAX_RELAYS> historySlotTaken{};
    std::array<bool, MAX_RELAYS> found{};
    std::array<Relay, MAX_RELAYS> newRelays{};

    unsigned int index = 0;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      for (unsigned int i = 0; i < mNumRelays; i++) {
        const auto& relay = mRelays[i];
        for (unsigned int j = 0; j < numRelays; j++) {
          if (relay.ID == incoming[j].ID) {
            found[j] = true;
            newRelays[index++] = relay;

            const auto slot = relay.History - mPingHistoryBuff.data();
            assert(slot >= 0);
            assert(slot < MAX_RELAYS);
            historySlotTaken[slot] = true;
          }
        }
      }

      // copy all near relays not found in the current relay list

      for (unsigned int i = 0; i < numRelays; i++) {
        if (!found[i]) {
          auto& newRelay = newRelays[index];
          newRelay.ID = incoming[i].ID;
          newRelay.Addr = incoming[i].Addr;

          // find a history slot for this relay
          // helps when updating and copying in the above loop
          // instead of copying the while ping history array
          // it just copies the pointer
          for (int j = 0; j < MAX_RELAYS; j++) {
            if (!historySlotTaken[j]) {
              newRelay.History = &mPingHistoryBuff[j];
              newRelay.History->clear();
              historySlotTaken[j] = true;
              break;
            }
          }
          assert(newRelay.History != nullptr);
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
          const auto& incomingRelay = incoming[i];
          const auto& relay = mRelays[j];
          if (incomingRelay.ID == relay.ID && incomingRelay.Addr == relay.Addr) {
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
