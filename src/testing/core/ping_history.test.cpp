#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "core/ping_history.hpp"

using core::INVALID_SEQUENCE_NUMBER;
using core::PING_HISTORY_ENTRY_COUNT;

TEST(core_PingHistory_general)
{
  core::PingHistory ph;
  for (size_t i = 0; i < PING_HISTORY_ENTRY_COUNT * 2; i++) {
    auto& entry = ph[i];

    // by default, all entries start out like this
    if (i < PING_HISTORY_ENTRY_COUNT) {
      CHECK(entry.sequence_number == INVALID_SEQUENCE_NUMBER);
      CHECK(entry.time_ping_sent == -1.0);
      CHECK(entry.time_pong_received == -1.0);
    } else {
      CHECK(entry.sequence_number != INVALID_SEQUENCE_NUMBER);
      CHECK(entry.time_ping_sent != -1.0);
      CHECK(entry.time_pong_received != -1.0);
    }

    auto last_seq = ph.most_recent_sequence;
    auto ping_time = 1 + random_decimal<double>() * 999.0;
    auto pong_time = ping_time + 1 + random_decimal<double>() * 999.0;

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

TEST(core_PingHistory_into_RouteStats)
{
  core::PingHistory ph;

  float total_ping = 0.0;
  float total_rtt = 0.0;

  for (size_t i = 0; i < PING_HISTORY_ENTRY_COUNT; i++) {
    auto ping = random_decimal<double>() * 999.0;

    if ((i & 1) == 1) {
      // every odd entry will have received a valid pong
      auto pong = 1 + random_decimal<double>() * 2.0;

      // new
      auto seq = ph.ping_sent(ping);
      ph.pong_received(seq, ping + pong);

      total_ping += ping;
      total_rtt += 1000.0f * pong;
    } else {
      // every even entry will have not received a pong
      ph.ping_sent(ping);
    }
  }

  double mean_rtt = total_rtt / 128;
  double std_dev_rtt = 0.0;
  size_t num_jitter_samples = 0;

  for (size_t i = 1; i < PING_HISTORY_ENTRY_COUNT; i += 2) {
    auto& entry = ph[i];
    double rtt = 1000.0 * (entry.time_pong_received - entry.time_ping_sent);
    if (rtt > mean_rtt) {
      double error = rtt - mean_rtt;
      std_dev_rtt += error * error;
      num_jitter_samples++;
    }
  }

  auto expected_rtt = total_rtt / 128;  // every odd is added, so half the relay ping history entry count
  auto expected_jitter = 3.0f * static_cast<float>(std::sqrt(std_dev_rtt / num_jitter_samples));
  auto expected_packet_loss = 50.0;  // half the pings never get a pong, so 50%

  // account for floating point imprecision
  expected_rtt = (expected_rtt * 100.0) / 100.0;
  expected_jitter = (expected_jitter * 100.0) / 100.0;

  core::RouteStats stats;
  ph.into(stats, 0.0, 1000.0, 0.0);

  // of the three fields, skipping jitter, too much time to check, relying on legacy code being correct

  double actual_rtt = (stats.rtt * 100.0) / 100.0;
  double actual_jitter = (stats.jitter * 100.0) / 100.0;
  double actual_packet_loss = (stats.packet_loss * 100.0) / 100.0;

  CHECK(std::abs(actual_rtt - expected_rtt) < 0.001).on_fail([&] {
    std::cout << "actual rtt = " << actual_rtt << '\n';
    std::cout << "expected rtt = " << expected_rtt << '\n';
  });
  CHECK(std::abs(actual_jitter - expected_jitter) < 0.001).on_fail([&] {
    std::cout << "actual jitter = " << actual_jitter << '\n';
    std::cout << "expected jitter = " << expected_jitter << '\n';
  });
  CHECK(std::abs(actual_packet_loss - expected_packet_loss) < 0.001);
}
