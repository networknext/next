#ifndef CORE_PING_HISTORY_HPP
#define CORE_PING_HISTORY_HPP

#include <cinttypes>
#include <array>

#include "config.hpp"

namespace core
{
  struct HistoryEntry
  {
    uint64_t Seq = INVALID_SEQUENCE_NUMBER;
    double TimePingSent = -1.0;
    double TimePongRecieved = -1.0;
  };

  class PingHistory
  {
   public:
    PingHistory() = default;
    ~PingHistory() = default;

    void clear();

    uint64_t pingSent(double time);

    void pongReceived(uint64_t seq, double time);

    std::array<HistoryEntry, RELAY_PING_HISTORY_ENTRY_COUNT> Entries;

    // helper for testing only
    uint64_t seq();

   private:
    uint64_t mSeq = 0;
  };

  inline uint64_t PingHistory::seq()
  {
    return mSeq;
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