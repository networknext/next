#pragma once

#include "core/route_stats.hpp"
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
  class _test_core_RelayManager_reset_;
  class _test_core_RelayManager_process_pong_;
  class _test_core_RelayManager_copy_existing_relays_;
  class _test_core_RelayManager_copy_new_relays_;
}  // namespace testing

namespace core
{
  // how long between relay to relay pings, in seconds
  const double PING_RATE = 0.1;
  // how many seconds before a packet is considered as lost
  const double PING_SAFETY = 1.0;
  // default value for stats
  const double INVALID_PING_TIME = -10000.0;
  // how long to sample for relay stats, in seconds
  const double STATS_WINDOW = 10.0;

  struct PingData
  {
    uint64_t sequence = 0;
    Address address;
  };

  struct RelayPingInfo
  {
    uint64_t id = 0;
    Address address;
  };

  struct Relay
  {
    uint64_t id = 0;
    Address address;
    double last_ping_time = INVALID_PING_TIME;
    PingHistory* history = nullptr;

    // Used only in tests, asserts using the history's pointed to address, not the value of it
    auto operator==(const Relay& other) const -> bool;
  };

  INLINE auto Relay::operator==(const Relay& other) const -> bool
  {
    return this->id == other.id && this->address == other.address && this->last_ping_time == other.last_ping_time &&
           this->history == other.history;
  }

  class RelayManager
  {
    friend testing::_test_core_handlers_relay_pong_handler_;
    friend testing::_test_core_RelayManager_reset_;
    friend testing::_test_core_RelayManager_process_pong_;
    friend testing::_test_core_RelayManager_copy_existing_relays_;
    friend testing::_test_core_RelayManager_copy_new_relays_;

   public:
    RelayManager();
    ~RelayManager() = default;

    void reset();

    auto update(size_t num_incoming_relays, const std::array<RelayPingInfo, MAX_RELAYS>& incoming_relays) -> bool;

    auto get_ping_targets(std::array<PingData, MAX_RELAYS>& data) -> size_t;

    auto process_pong(const Address& from, uint64_t seq) -> bool;

    void get_stats(RelayStats& stats);

   private:
    std::mutex mutex;
    unsigned int num_relays;
    std::array<Relay, MAX_RELAYS> relays;
    std::vector<PingHistory> ping_history;
    util::Clock clock;

    auto copy_existing_relays(
     size_t& index,
     size_t num_incoming_relays,
     const std::array<RelayPingInfo, MAX_RELAYS>& incoming,
     std::array<Relay, MAX_RELAYS>& new_relays,
     std::array<bool, MAX_RELAYS>& found,
     std::array<bool, MAX_RELAYS>& history_slot_taken) -> bool;

    auto copy_new_relays(
     size_t& index,
     size_t num_incoming_relays,
     const std::array<RelayPingInfo, MAX_RELAYS>& incoming,
     std::array<Relay, MAX_RELAYS>& new_relays,
     std::array<bool, MAX_RELAYS>& found,
     std::array<bool, MAX_RELAYS>& history_slot_taken) -> bool;
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

  INLINE auto RelayManager::update(size_t num_incoming_relays, const std::array<RelayPingInfo, MAX_RELAYS>& incoming) -> bool
  {
    if (num_incoming_relays > MAX_RELAYS) {
      LOG(ERROR, "updating with too many relays: ", num_incoming_relays);
      return false;
    }

    std::array<bool, MAX_RELAYS> history_slot_taken{};
    std::array<bool, MAX_RELAYS> found{};
    std::array<Relay, MAX_RELAYS> new_relays{};

    size_t index = 0;

    {
      std::lock_guard<std::mutex> lk(this->mutex);

      // preserve existing relay entries if they exist in the new data set
      copy_existing_relays(index, num_incoming_relays, incoming, new_relays, found, history_slot_taken);

      // copy all new relays not found in the current relay list
      copy_new_relays(index, num_incoming_relays, incoming, new_relays, found, history_slot_taken);

      // commit the updated relay array, move semantics won't make a difference here
      this->num_relays = index;
      std::copy(new_relays.begin(), new_relays.begin() + index, this->relays.begin());

      // make sure all the ping times are evenly distributed to avoid clusters of ping packets (prevent relays from ddos-ing
      // eachother)

      auto current_time = this->clock.elapsed<Second>();

      for (unsigned int i = 0; i < this->num_relays; i++) {
        this->relays[i].last_ping_time = current_time - PING_RATE + i * PING_RATE / this->num_relays;
      }

#ifndef NDEBUG

      // make sure everything is correct

      // TODO move these to tests

      assert(this->num_relays == index);

      unsigned int num_found = 0;
      for (unsigned int i = 0; i < num_incoming_relays; i++) {
        for (unsigned int j = 0; j < this->num_relays; j++) {
          const auto& incoming_relay = incoming[i];
          const auto& relay = this->relays[j];
          if (incoming_relay.id == relay.id && incoming_relay.address == relay.address) {
            num_found++;
            break;
          }
        }
      }

      assert(num_found == this->num_relays);

      for (unsigned int i = 0; i < num_incoming_relays; i++) {
        for (unsigned int j = 0; j < num_incoming_relays; j++) {
          if (i == j) {
            continue;
          }
          assert(this->relays[i].history != this->relays[j].history);
        }
      }
#endif
    }

    return true;
  }

  INLINE auto RelayManager::get_ping_targets(std::array<PingData, MAX_RELAYS>& data) -> size_t
  {
    double current_time = this->clock.elapsed<Second>();
    size_t num_pings = 0;

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(this->mutex);
      for (unsigned int i = 0; i < this->num_relays; ++i) {
        if (this->relays[i].last_ping_time + PING_RATE <= current_time) {
          auto& relay = this->relays[i];
          auto& ping_data = data[num_pings++];

          ping_data.sequence = relay.history->ping_sent(current_time);
          ping_data.address = relay.address;
          relay.last_ping_time = current_time;
        }
      }
    }

    return num_pings;
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
          relay.history->pong_received(seq, this->clock.elapsed<Second>());
          pong_received = true;
          break;
        }
      }
    }

    return pong_received;
  }

  INLINE void RelayManager::get_stats(RelayStats& stats)
  {
    auto current_time = this->clock.elapsed<Second>();

    // locked mutex scope
    {
      std::lock_guard<std::mutex> lk(this->mutex);
      stats.num_relays = this->num_relays;

      for (unsigned int i = 0; i < this->num_relays; i++) {
        auto& relay = this->relays[i];
        RouteStats rs;
        relay.history->into(rs, current_time - STATS_WINDOW, current_time, PING_SAFETY);
        stats.ids[i] = relay.id;
        stats.rtt[i] = rs.rtt;
        stats.jitter[i] = rs.jitter;
        stats.packet_loss[i] = rs.packet_loss;
      }
    }
  }

  INLINE auto RelayManager::copy_existing_relays(
   size_t& index,
   size_t num_incoming_relays,
   const std::array<RelayPingInfo, MAX_RELAYS>& incoming,
   std::array<Relay, MAX_RELAYS>& new_relays,
   std::array<bool, MAX_RELAYS>& found,
   std::array<bool, MAX_RELAYS>& history_slot_taken) -> bool
  {
    for (unsigned int i = 0; i < this->num_relays; i++) {
      const auto& relay = this->relays[i];
      for (unsigned int j = 0; j < num_incoming_relays; j++) {
        if (relay.id == incoming[j].id) {
          found[j] = true;
          new_relays[index++] = relay;

          const long slot = relay.history - this->ping_history.data();

          if (slot < 0) {
            LOG(ERROR, "history slot index invalid (< 0)");
            return false;
          }

          if (slot >= static_cast<long>(MAX_RELAYS)) {
            LOG(ERROR, "history slot index invalid (>= MAX_RELAYS)");
            return false;
          }

          history_slot_taken[slot] = true;
        }
      }
    }

    return true;
  }

  INLINE auto RelayManager::copy_new_relays(
   size_t& index,
   size_t num_incoming_relays,
   const std::array<RelayPingInfo, MAX_RELAYS>& incoming,
   std::array<Relay, MAX_RELAYS>& new_relays,
   std::array<bool, MAX_RELAYS>& found,
   std::array<bool, MAX_RELAYS>& history_slot_taken) -> bool
  {
    for (unsigned int i = 0; i < num_incoming_relays; i++) {
      if (!found[i]) {
        auto& new_relay = new_relays[index];
        new_relay.id = incoming[i].id;
        new_relay.address = incoming[i].address;

        for (size_t j = 0; j < MAX_RELAYS; j++) {
          if (!history_slot_taken[j]) {
            new_relay.history = &this->ping_history[j];
            new_relay.history->clear();
            history_slot_taken[j] = true;
            break;
          }
        }

        if (new_relay.history == nullptr) {
          return false;
        }

        index++;
      }
    }

    return true;
  }
}  // namespace core
