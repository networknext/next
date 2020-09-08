#include "includes.h"
#include "testing/test.hpp"

#include "core/route_stats.hpp"
#include "core/ping_history.hpp"

Test(core_route_stats_basic_test)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<float> ping_dist(0.0, 1000.0);  // pings will happen in this range
  std::uniform_real_distribution<float> pong_dist(0.0, 3.0);     // pongs will happen in ping + this range
  auto ping_rand = std::bind(ping_dist, gen);
  auto pong_rand = std::bind(pong_dist, gen);

  core::PingHistory ph;

  float total_ping = 0.0;
  float total_rtt = 0.0;

  for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++) {
    auto ping = ping_rand();

    if ((i & 1) == 1) {
      // every odd entry will have received a valid pong
      auto pong = pong_rand();

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

  for (int i = 1; i < RELAY_PING_HISTORY_ENTRY_COUNT; i += 2) {
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

  check(std::abs(actual_rtt - expected_rtt) < 0.001).onFail([&] {
    std::cout << "actual rtt = " << actual_rtt << '\n';
    std::cout << "expected rtt = " << expected_rtt << '\n';
  });
  check(std::abs(actual_jitter - expected_jitter) < 0.001).onFail([&] {
    std::cout << "actual jitter = " << actual_jitter << '\n';
    std::cout << "expected jitter = " << expected_jitter << '\n';
  });
  check(std::abs(actual_packet_loss - expected_packet_loss) < 0.001);
}
