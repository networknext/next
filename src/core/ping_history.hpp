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

    // helper for testing only
    auto seq() -> uint64_t;

    auto operator[](size_t i) -> const HistoryEntry&;
    auto operator=(const PingHistory& other) -> PingHistory&;

   private:
    uint64_t mSeq = 0;
    std::array<HistoryEntry, RELAY_PING_HISTORY_ENTRY_COUNT> mEntries;

    friend class RouteStats;
  };

  INLINE PingHistory::PingHistory(const PingHistory& other)
  {
    *this = other;
  }

  INLINE void PingHistory::clear()
  {
    GCC_NO_OPT_OUT;
    mSeq = 0;

    mEntries.fill(HistoryEntry());
  }

  INLINE auto PingHistory::pingSent(double time) -> uint64_t
  {
    GCC_NO_OPT_OUT;
    const auto index = mSeq % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = mEntries[index];
    entry.Sequence = mSeq;
    entry.TimePingSent = time;
    entry.TimePongReceived = -1.0;
    mSeq++;
    return entry.Sequence;
  }

  INLINE void PingHistory::pongReceived(uint64_t seq, double time)
  {
    GCC_NO_OPT_OUT;
    const size_t index = seq % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = mEntries[index];
    if (entry.Sequence == seq) {
      entry.TimePongReceived = time;
    }
  }

  INLINE auto PingHistory::seq() -> uint64_t
  {
    return mSeq;
  }

  INLINE auto PingHistory::operator[](size_t i) -> const HistoryEntry&
  {
    return mEntries[i % mEntries.size()];
  }

  INLINE auto PingHistory::operator=(const PingHistory& other) -> PingHistory&
  {
    this->mSeq = other.mSeq;
    std::copy(other.mEntries.begin(), other.mEntries.end(), this->mEntries.begin());
    return *this;
  }
}  // namespace core
