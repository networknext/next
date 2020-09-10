#include "includes.h"
#include "testing/test.hpp"

#include "core/ping_history.hpp"

TEST(core_PingHistory_general)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<double> dist(1.0, 1000.0);
  auto rand = std::bind(dist, gen);

  core::PingHistory ph;
  for (size_t i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT * 2; i++) {
    auto& entry = ph[i];

    // by default, all entries start out like this
    if (i < RELAY_PING_HISTORY_ENTRY_COUNT) {
      CHECK(entry.sequence_number == INVALID_SEQUENCE_NUMBER);
      CHECK(entry.time_ping_sent == -1.0);
      CHECK(entry.time_pong_received == -1.0);
    } else {
      CHECK(entry.sequence_number != INVALID_SEQUENCE_NUMBER);
      CHECK(entry.time_ping_sent != -1.0);
      CHECK(entry.time_pong_received != -1.0);
    }

    auto last_seq = ph.most_recent_sequence;
    auto ping_time = rand();
    auto pong_time = rand();

    // record the ping data
    CHECK(ph.ping_sent(ping_time) == last_seq);
    CHECK(ph.most_recent_sequence == last_seq + 1);
    CHECK(entry.sequence_number == last_seq);
    CHECK(entry.time_ping_sent == ping_time);
    CHECK(entry.time_pong_received == -1.0);

    // record the pong data
    ph.pong_received(last_seq, pong_time);

    CHECK(entry.time_pong_received == pong_time);
  }
}
