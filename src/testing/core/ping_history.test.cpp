#include "includes.h"
#include "testing/test.hpp"

#include "core/ping_history.hpp"

Test(core_PingHistory_general)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<double> dist(1.0, 1000.0);
  auto rand = std::bind(dist, gen);

  core::PingHistory ph;
  for (size_t i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT * 2; i++) {
    auto& entry = ph[i];

    // by default, all entries start out like this
    if (i < RELAY_PING_HISTORY_ENTRY_COUNT) {
      check(entry.sequence_number == INVALID_SEQUENCE_NUMBER);
      check(entry.time_ping_sent == -1.0);
      check(entry.time_pong_received == -1.0);
    } else {
      check(entry.sequence_number != INVALID_SEQUENCE_NUMBER);
      check(entry.time_ping_sent != -1.0);
      check(entry.time_pong_received != -1.0);
    }

    auto lastSeq = ph.most_recent_sequence;
    auto pingTime = rand();
    auto pongTime = rand();

    // record the ping data
    check(ph.ping_sent(pingTime) == lastSeq);
    check(ph.most_recent_sequence == lastSeq + 1);
    check(entry.sequence_number == lastSeq);
    check(entry.time_ping_sent == pingTime);
    check(entry.time_pong_received == -1.0);

    // record the pong data
    ph.pong_received(lastSeq, pongTime);

    check(entry.time_pong_received == pongTime);
  }
}
