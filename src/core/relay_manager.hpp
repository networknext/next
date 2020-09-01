#pragma once

#include "core/route_stats.hpp"
#include "crypto/hash.hpp"
#include "net/address.hpp"
#include "ping_history.hpp"
#include "relay_stats.hpp"
#include "router_info.hpp"
#include "util/clock.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

using net::Address;

namespace testing
{
  class _test_core_handlers_relay_pong_handler_;
}

namespace core
{
  const auto INVALID_PING_TIME = -10000.0;

  struct PingData
  {
    uint64_t sequence;
    Address address;
  };

  struct RelayPingInfo
  {
    uint64_t id;
    Address address;
  };

  struct Relay
  {
    uint64_t id;
    Address address;
    double last_ping_time = INVALID_PING_TIME;
    PingHistory* history = nullptr;
  };

  class RelayManager
  {
    friend class testing::_test_core_handlers_relay_pong_handler_;

   public:
    RelayManager();
    ~RelayManager() = default;

    void reset();

    void update(size_t numRelays, const std::array<RelayPingInfo, MAX_RELAYS>& newRelays);

    auto process_pong(const Address& from, uint64_t seq) -> bool;

    void get_stats(RelayStats& stats);

    auto get_ping_targets(std::array<PingData, MAX_RELAYS>& data) -> size_t;

   private:
    std::mutex mLock;
    unsigned int mNumRelays;
    std::array<Relay, MAX_RELAYS> mRelays;
    std::vector<PingHistory> mPingHistoryBuff;
    util::Clock mClock;
  };

  INLINE RelayManager::RelayManager()
  {
    mPingHistoryBuff.resize(MAX_RELAYS);
    reset();
  }

  INLINE void RelayManager::reset()
  {
    // locked mutex scope
    std::lock_guard<std::mutex> lk(mLock);
    mNumRelays = 0;
    std::fill(mRelays.begin(), mRelays.end(), Relay());
  }

  INLINE auto RelayManager::process_pong(const Address& from, uint64_t seq) -> bool
  {
    bool pongReceived = false;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      for (unsigned int i = 0; i < mNumRelays; i++) {
        auto& relay = mRelays[i];
        if (from == relay.address) {
          relay.history->pongReceived(seq, mClock.elapsed<util::Second>());
          pongReceived = true;
          break;
        }
      }
    }

    return pongReceived;
  }

  INLINE void RelayManager::get_stats(RelayStats& stats)
  {
    auto currentTime = mClock.elapsed<util::Second>();

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      stats.NumRelays = mNumRelays;

      for (unsigned int i = 0; i < mNumRelays; i++) {
        auto& relay = mRelays[i];

        RouteStats rs(*relay.history, currentTime - RELAY_STATS_WINDOW, currentTime, RELAY_PING_SAFETY);
        stats.IDs[i] = relay.id;
        stats.RTT[i] = rs.getRTT();
        stats.Jitter[i] = rs.getJitter();
        stats.PacketLoss[i] = rs.getPacketLoss();
      }
    }
  }

  INLINE auto RelayManager::get_ping_targets(std::array<PingData, MAX_RELAYS>& data) -> size_t
  {
    double currentTime = mClock.elapsed<util::Second>();
    size_t numPings = 0;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(mLock);
      for (unsigned int i = 0; i < mNumRelays; ++i) {
        if (mRelays[i].last_ping_time + RELAY_PING_TIME <= currentTime) {
          auto& relay = mRelays[i];
          auto& pingData = data[numPings++];

          pingData.sequence = relay.history->pingSent(currentTime);
          pingData.address = relay.address;
          relay.last_ping_time = currentTime;
        }
      }
    }

    return numPings;
  }

  // it is used in one place throughout the codebase, so always inline it, no sense in doing a function call
  INLINE void RelayManager::update(size_t numRelays, const std::array<RelayPingInfo, MAX_RELAYS>& incoming)
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
          if (relay.id == incoming[j].id) {
            found[j] = true;
            newRelays[index++] = relay;

            const auto slot = relay.history - mPingHistoryBuff.data();
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
          newRelay.id = incoming[i].id;
          newRelay.address = incoming[i].address;

          // find a history slot for this relay
          // helps when updating and copying in the above loop
          // instead of copying the while ping history array
          // it just copies the pointer
          for (int j = 0; j < MAX_RELAYS; j++) {
            if (!historySlotTaken[j]) {
              newRelay.history = &mPingHistoryBuff[j];
              newRelay.history->clear();
              historySlotTaken[j] = true;
              break;
            }
          }
          assert(newRelay.history != nullptr);
          index++;
        }
      }

      // commit the updated relay array
      mNumRelays = index;
      std::copy(newRelays.begin(), newRelays.begin() + index, mRelays.begin());

      // make sure all the ping times are evenly distributed to avoid clusters of ping packets

      auto currentTime = mClock.elapsed<util::Second>();

      for (unsigned int i = 0; i < mNumRelays; i++) {
        mRelays[i].last_ping_time = currentTime - RELAY_PING_TIME + i * RELAY_PING_TIME / mNumRelays;
      }

#ifndef NDEBUG

      // make sure everything is correct

      assert(mNumRelays == index);

      unsigned int numFound = 0;
      for (unsigned int i = 0; i < numRelays; i++) {
        for (unsigned int j = 0; j < mNumRelays; j++) {
          const auto& incomingRelay = incoming[i];
          const auto& relay = mRelays[j];
          if (incomingRelay.id == relay.id && incomingRelay.address == relay.address) {
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
          assert(mRelays[i].history != mRelays[j].history);
        }
      }
#endif
    }
  }
}  // namespace core
