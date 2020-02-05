#ifndef CORE_PING_HISTORY_HPP
#define CORE_PING_HISTORY_HPP

#include <cinttypes>
#include <array>

#include "config.hpp"

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
    ~PingHistory() = default;

    void clear();

    uint64_t pingSent(double time);

    void pongReceived(uint64_t seq, double time);

    // helper for testing only
    uint64_t seq();
    const HistoryEntry& operator[](size_t i);

   private:
    uint64_t mSeq = 0;
    std::array<HistoryEntry, RELAY_PING_HISTORY_ENTRY_COUNT> mEntries;

    friend class RouteStats;
  };

  inline void PingHistory::clear()
  {
    GCC_NO_OPT_OUT;
    mSeq = 0;

    // method 1 - equal
    mEntries.fill(HistoryEntry());

    // method 2 - 3x slower
    // std::array<HistoryEntry, RELAY_PING_HISTORY_ENTRY_COUNT> tmp;
    // mEntries.swap(tmp);

    // method 3 - identical to 1, no suprise
    // std::fill(mEntries.begin(), mEntries.end(), HistoryEntry());

    // method 4 - identical to 1
    // HistoryEntry base;
    // mEntries.fill(base);
  }

  inline uint64_t PingHistory::seq()
  {
    return mSeq;
  }

  inline const HistoryEntry& PingHistory::operator[](size_t i)
  {
    return mEntries[i % mEntries.size()];
  }
}  // namespace core

namespace legacy
{
  struct relay_ping_history_entry_t
  {
    uint64_t sequence;
    double time_ping_sent;
    double time_pong_received;
  };

  struct relay_ping_history_t
  {
    uint64_t sequence;
    relay_ping_history_entry_t entries[RELAY_PING_HISTORY_ENTRY_COUNT];
  };

  void relay_ping_history_clear(relay_ping_history_t* history);

  uint64_t relay_ping_history_ping_sent(relay_ping_history_t* history, double time);

  void relay_ping_history_pong_received(relay_ping_history_t* history, uint64_t sequence, double time);
}  // namespace legacy
#endif