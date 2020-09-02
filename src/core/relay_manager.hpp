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
using util::Second;

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
    std::mutex mutex;
    unsigned int num_relays;
    std::array<Relay, MAX_RELAYS> relays;
    std::vector<PingHistory> ping_history;
    util::Clock clock;
  };

  INLINE RelayManager::RelayManager()
  {
    this->ping_history.resize(MAX_RELAYS);
    reset();
  }

  INLINE void RelayManager::reset()
  {
    // locked mutex scope
    std::lock_guard<std::mutex> lk(this->mutex);
    this->num_relays = 0;
    std::fill(this->relays.begin(), this->relays.end(), Relay());
  }

  INLINE auto RelayManager::process_pong(const Address& from, uint64_t seq) -> bool
  {
    bool pong_received = false;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(this->mutex);
      for (unsigned int i = 0; i < this->num_relays; i++) {
        auto& relay = this->relays[i];
        if (from == relay.address) {
          relay.history->pongReceived(seq, this->clock.elapsed<Second>());
          pong_received = true;
          break;
        }
      }
    }

    return pong_received;
  }

  INLINE void RelayManager::get_stats(RelayStats& stats)
  {
    auto currentTime = this->clock.elapsed<Second>();

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(this->mutex);
      stats.num_relays = this->num_relays;

      for (unsigned int i = 0; i < this->num_relays; i++) {
        auto& relay = this->relays[i];

        RouteStats rs(*relay.history, currentTime - RELAY_STATS_WINDOW, currentTime, RELAY_PING_SAFETY);
        stats.ids[i] = relay.id;
        stats.rtt[i] = rs.getRTT();
        stats.jitter[i] = rs.getJitter();
        stats.packet_loss[i] = rs.getPacketLoss();
      }
    }
  }

  INLINE auto RelayManager::get_ping_targets(std::array<PingData, MAX_RELAYS>& data) -> size_t
  {
    double currentTime = this->clock.elapsed<Second>();
    size_t numPings = 0;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(this->mutex);
      for (unsigned int i = 0; i < this->num_relays; ++i) {
        if (this->relays[i].last_ping_time + RELAY_PING_TIME <= currentTime) {
          auto& relay = this->relays[i];
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
      std::lock_guard<std::mutex> lk(this->mutex);
      for (unsigned int i = 0; i < this->num_relays; i++) {
        const auto& relay = this->relays[i];
        for (unsigned int j = 0; j < numRelays; j++) {
          if (relay.id == incoming[j].id) {
            found[j] = true;
            newRelays[index++] = relay;

            const auto slot = relay.history - this->ping_history.data();
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
              newRelay.history = &this->ping_history[j];
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
      this->num_relays = index;
      std::copy(newRelays.begin(), newRelays.begin() + index, this->relays.begin());

      // make sure all the ping times are evenly distributed to avoid clusters of ping packets

      auto currentTime = this->clock.elapsed<Second>();

      for (unsigned int i = 0; i < this->num_relays; i++) {
        this->relays[i].last_ping_time = currentTime - RELAY_PING_TIME + i * RELAY_PING_TIME / this->num_relays;
      }

#ifndef NDEBUG

      // make sure everything is correct

      assert(this->num_relays == index);

      unsigned int numFound = 0;
      for (unsigned int i = 0; i < numRelays; i++) {
        for (unsigned int j = 0; j < this->num_relays; j++) {
          const auto& incomingRelay = incoming[i];
          const auto& relay = this->relays[j];
          if (incomingRelay.id == relay.id && incomingRelay.address == relay.address) {
            numFound++;
            break;
          }
        }
      }

      assert(numFound == this->num_relays);

      for (unsigned int i = 0; i < numRelays; i++) {
        for (unsigned int j = 0; j < numRelays; j++) {
          if (i == j) {
            continue;
          }
          assert(this->relays[i].history != this->relays[j].history);
        }
      }
#endif
    }
  }
}  // namespace core
