#pragma once

#include "core/route_stats.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

namespace testing
{
  class _test_core_PingHistory_general_;
  class _test_core_RelayManager_process_pong_;
}  // namespace testing

namespace core
{
  const size_t PING_HISTORY_ENTRY_COUNT = 256;
  const size_t INVALID_SEQUENCE_NUMBER = 0xFFFFFFFFFFFFFFFFULL;

  struct HistoryEntry
  {
    uint64_t sequence_number = INVALID_SEQUENCE_NUMBER;
    double time_ping_sent = -1.0;
    double time_pong_received = -1.0;
  };

  class PingHistory
  {
    friend testing::_test_core_PingHistory_general_;
    friend testing::_test_core_RelayManager_process_pong_;

   public:
    PingHistory() = default;
    PingHistory(const PingHistory& other);
    ~PingHistory() = default;

    void clear();

    auto ping_sent(double time) -> uint64_t;

    void pong_received(uint64_t seq, double time);

    void into(RouteStats& stats, double start, double end, double safety);

    auto operator[](size_t i) -> const HistoryEntry&;
    auto operator=(const PingHistory& other) -> PingHistory&;

   private:
    uint64_t most_recent_sequence = 0;
    std::array<HistoryEntry, PING_HISTORY_ENTRY_COUNT> entries;
  };

  INLINE PingHistory::PingHistory(const PingHistory& other)
  {
    *this = other;
  }

  INLINE void PingHistory::clear()
  {
    this->most_recent_sequence = 0;
    this->entries.fill(HistoryEntry());
  }

  INLINE auto PingHistory::ping_sent(double time) -> uint64_t
  {
    const auto index = this->most_recent_sequence % PING_HISTORY_ENTRY_COUNT;
    auto& entry = this->entries[index];
    entry.sequence_number = this->most_recent_sequence;
    entry.time_ping_sent = time;
    entry.time_pong_received = -1.0;
    this->most_recent_sequence++;
    return entry.sequence_number;
  }

  INLINE void PingHistory::pong_received(uint64_t seq, double time)
  {
    const size_t index = seq % PING_HISTORY_ENTRY_COUNT;
    auto& entry = this->entries[index];
    if (entry.sequence_number == seq) {
      entry.time_pong_received = time;
    }
  }

  INLINE void PingHistory::into(RouteStats& stats, double start, double end, double safety)
  {
    // Packet loss calc
    // and RTT calc

    size_t num_pings_sent = 0u;
    size_t num_pongs_received = 0u;

    double total_rtt = 0.0;
    int num_pings = 0;
    int num_pongs = 0;

    for (const auto& entry : this->entries) {
      if (entry.time_ping_sent >= start) {
        if (entry.time_ping_sent <= end - safety) {
          num_pings_sent++;

          if (entry.time_pong_received >= entry.time_ping_sent) {
            num_pongs_received++;
          }
        }

        if (entry.time_ping_sent <= end) {
          num_pings++;

          if (entry.time_pong_received > entry.time_ping_sent) {
            total_rtt += 1000.0 * (entry.time_pong_received - entry.time_ping_sent);
            num_pongs++;
          }
        }
      }
    }

    double mean_rtt = (num_pongs > 0) ? (total_rtt / num_pongs) : 10000.0;

    if (num_pings_sent > 0) {
      stats.packet_loss = static_cast<float>(100.0 * (1.0 - (double(num_pongs_received) / double(num_pings_sent))));
    }

    stats.rtt = static_cast<float>(mean_rtt);

    // Jitter calc

    auto num_jitter_samples = 0u;
    auto std_dev_rtt = 0.0;

    for (const auto& entry : this->entries) {
      // if the entry is within the window and the pong was received
      if (entry.time_ping_sent >= start && entry.time_ping_sent <= end && entry.time_pong_received > entry.time_ping_sent) {
        double rtt = 1000.0 * (entry.time_pong_received - entry.time_ping_sent);
        if (rtt >= mean_rtt) {
          double error = rtt - mean_rtt;
          std_dev_rtt += error * error;
          num_jitter_samples++;
        }
      }
    }

    if (num_jitter_samples > 0) {
      stats.jitter = 3.0f * static_cast<float>(std::sqrt(std_dev_rtt / num_jitter_samples));
    }
  }

  INLINE auto PingHistory::operator[](size_t i) -> const HistoryEntry&
  {
    return this->entries[i % this->entries.size()];
  }

  INLINE auto PingHistory::operator=(const PingHistory& other) -> PingHistory&
  {
    this->most_recent_sequence = other.most_recent_sequence;
    std::copy(other.entries.begin(), other.entries.end(), this->entries.begin());
    return *this;
  }
}  // namespace core
