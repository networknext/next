#include "ping_history.hpp"

#include <cassert>

namespace core
{
  void PingHistory::clear()
  {
    GCC_NO_OPT_OUT;
    mSeq = 0;
    Entries.fill(HistoryEntry());
  }

  uint64_t PingHistory::pingSent(double time)
  {
    const auto index = mSeq % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = Entries[index];
    entry.Seq = mSeq;
    entry.TimePingSent = time;
    entry.TimePongRecieved = -1.0;
    return ++mSeq;
  }

  void PingHistory::pongReceived(uint64_t seq, double time)
  {
    const auto index = seq % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = Entries[index];
    if (entry.Seq == seq) {
      entry.TimePongRecieved = time;
    }
  }
}  // namespace core

namespace legacy
{
  void relay_ping_history_clear(relay_ping_history_t* history)
  {
    assert(history);
    history->sequence = 0;
    for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; ++i) {
      history->entries[i].sequence = INVALID_SEQUENCE_NUMBER;
      history->entries[i].time_ping_sent = -1.0;
      history->entries[i].time_pong_received = -1.0;
    }
  }

  uint64_t relay_ping_history_ping_sent(relay_ping_history_t* history, double time)
  {
    assert(history);
    const int index = history->sequence % RELAY_PING_HISTORY_ENTRY_COUNT;
    relay_ping_history_entry_t* entry = &history->entries[index];
    entry->sequence = history->sequence;
    entry->time_ping_sent = time;
    entry->time_pong_received = -1.0;
    history->sequence++;
    return entry->sequence;
  }

  void relay_ping_history_pong_received(relay_ping_history_t* history, uint64_t sequence, double time)
  {
    const int index = sequence % RELAY_PING_HISTORY_ENTRY_COUNT;
    relay_ping_history_entry_t* entry = &history->entries[index];
    if (entry->sequence == sequence) {
      entry->time_pong_received = time;
    }
  }
}  // namespace legacy