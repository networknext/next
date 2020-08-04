#include "includes.h"
#include "ping_history.hpp"

namespace core
{
  auto PingHistory::pingSent(double time) -> uint64_t
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

  void PingHistory::pongReceived(uint64_t seq, double time)
  {
    GCC_NO_OPT_OUT;
    const auto index = seq % RELAY_PING_HISTORY_ENTRY_COUNT;
    auto& entry = mEntries[index];
    if (entry.Sequence == seq) {
      entry.TimePongReceived = time;
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
