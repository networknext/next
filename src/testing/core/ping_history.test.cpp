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
