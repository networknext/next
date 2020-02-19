#include "includes.h"
#include "testing/test.hpp"

#include "core/ping_history.hpp"

Test(PingHistory_general)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<double> dist(1.0, 1000.0);
  auto rand = std::bind(dist, gen);

  core::PingHistory ph;
  for (size_t i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT * 2; i++) {
    auto& entry = ph[i];

    // by default, all entries start out like this
    if (i < RELAY_PING_HISTORY_ENTRY_COUNT) {
      check(entry.Sequence == INVALID_SEQUENCE_NUMBER);
      check(entry.TimePingSent == -1.0);
      check(entry.TimePongReceived == -1.0);
    } else {
      check(entry.Sequence != INVALID_SEQUENCE_NUMBER);
      check(entry.TimePingSent != -1.0);
      check(entry.TimePongReceived != -1.0);
    }

    auto lastSeq = ph.seq();
    auto pingTime = rand();
    auto pongTime = rand();

    // record the ping data
    check(ph.pingSent(pingTime) == lastSeq);
    check(ph.seq() == lastSeq + 1);
    check(entry.Sequence == lastSeq);
    check(entry.TimePingSent == pingTime);
    check(entry.TimePongReceived == -1.0);

    // record the pong data
    ph.pongReceived(lastSeq, pongTime);

    check(entry.TimePongReceived == pongTime);
  }
}

Test(legacy_ping_history_t_general)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<double> dist(1.0, 1000.0);
  auto rand = std::bind(dist, gen);

  legacy::relay_ping_history_t ph;
  legacy::relay_ping_history_clear(&ph);
  for (size_t i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT * 2; i++) {
    auto& entry = ph.entries[i % RELAY_PING_HISTORY_ENTRY_COUNT];

    // by default, all entries start out like this
    if (i < RELAY_PING_HISTORY_ENTRY_COUNT) {
      check(entry.sequence == INVALID_SEQUENCE_NUMBER);
      check(entry.time_ping_sent == -1.0);
      check(entry.time_pong_received == -1.0);
    } else {
      check(entry.sequence != INVALID_SEQUENCE_NUMBER);
      check(entry.time_ping_sent != -1.0);
      check(entry.time_pong_received != -1.0);
    }

    auto lastSeq = ph.sequence;
    auto pingTime = rand();
    auto pongTime = rand();

    // record the ping data
    check(legacy::relay_ping_history_ping_sent(&ph, pingTime) == lastSeq);
    check(ph.sequence == lastSeq + 1);
    check(entry.sequence == lastSeq);
    check(entry.time_ping_sent == pingTime);
    check(entry.time_pong_received == -1.0);

    // record the pong data
    legacy::relay_ping_history_pong_received(&ph, lastSeq, pongTime);

    check(entry.time_pong_received == pongTime);
  }
}
