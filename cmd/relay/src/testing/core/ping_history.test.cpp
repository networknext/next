#include "testing/test.hpp"

#include <random>
#include <functional>
#include <iostream>

#include "core/ping_history.hpp"

Test(PingHistory_general)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<double> dist(1.0, 1000.0);
  auto rand = std::bind(dist, gen);

  core::PingHistory ph;
  for (size_t i = 0; i < ph.Entries.size(); i++) {
    auto& entry = ph.Entries[i];

    // by default, all entries start out like this
    check(entry.Seq == INVALID_SEQUENCE_NUMBER);
    check(entry.TimePingSent == -1.0);
    check(entry.TimePongRecieved == -1.0);

    auto lastSeq = ph.seq();
    auto pingTime = rand();
    auto pongTime = rand();

    // record the ping data
    check(ph.pingSent(pingTime) == lastSeq);
    check(ph.seq() == lastSeq + 1);
    check(entry.Seq == lastSeq);
    check(entry.TimePingSent == pingTime);
    check(entry.TimePongRecieved == -1.0);

    // record the pong data
    ph.pongReceived(lastSeq, pongTime);

    check(entry.TimePongRecieved == pongTime);
  }
}