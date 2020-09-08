#pragma once

#include "util/logger.hpp"
#include "util/macros.hpp"

namespace core
{
  struct HistoryEntry
  {
    uint64_t Sequence = INVALID_SEQUENCE_NUMBER;
    double TimePingSent = -1.0;
    double TimePongReceived = -1.0;
  };

  class PingHistory
  {
   public:
    PingHistory() = default;
    PingHistory(const PingHistory& other);
    ~PingHistory() = default;

    void clear();

    auto pingSent(double time) -> uint64_t;

    void pongReceived(uint64_t seq, double time);

    auto operator[](size_t i) -> const HistoryEntry&;
    auto operator=(const PingHistory& other) -> PingHistory&;

   private:
    uint64_t sequence = 0;
    std::array<HistoryEntry, RELAY_PING_HISTORY_ENTRY_COUNT> entries;

    friend class RouteStats;
  };

  INLINE PingHistory::PingHistory(const PingHistory& other)
  {
    *this = other;
  }

  INLINE void PingHistory::clear()
  {
    GCC_NO_OPT_OUT;
    this->sequence = 0;

    this->entries.fill(HistoryEntry());
  }

  INLINE auto PingHistory::pingSent(double time) -> uint64_t
  {
    GCC_NO_OPT_OUT;
    const auto index = this->sequence % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = this->entries[index];
    entry.Sequence = this->sequence;
    entry.TimePingSent = time;
    entry.TimePongReceived = -1.0;
    this->sequence++;
    return entry.Sequence;
  }

  INLINE void PingHistory::pongReceived(uint64_t seq, double time)
  {
    GCC_NO_OPT_OUT;
    const size_t index = seq % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = this->entries[index];
    if (entry.Sequence == seq) {
      entry.TimePongReceived = time;
    }
  }

  INLINE auto PingHistory::operator[](size_t i) -> const HistoryEntry&
  {
    return this->entries[i % this->entries.size()];
  }

  INLINE auto PingHistory::operator=(const PingHistory& other) -> PingHistory&
  {
    this->sequence = other.sequence;
    std::copy(other.entries.begin(), other.entries.end(), this->entries.begin());
    return *this;
  }
}  // namespace core
