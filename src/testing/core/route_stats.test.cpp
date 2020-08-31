#include "includes.h"
#include "testing/test.hpp"

#include "core/route_stats.hpp"

Test(core_route_stats_basic_test)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<float> pingDist(0.0, 1000.0);  // pings will happen in this range
  std::uniform_real_distribution<float> pongDist(0.0, 3.0);     // pongs will happen in ping + this range
  auto pingRand = std::bind(pingDist, gen);
  auto pongRand = std::bind(pongDist, gen);

  core::PingHistory ph;

  float totalPing = 0.0;
  float totalRTT = 0.0;

  for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++) {
    auto ping = pingRand();

    if ((i & 1) == 1) {
      // every odd entry will have received a valid pong
      auto pong = pongRand();

      // new
      auto seq = ph.pingSent(ping);
      ph.pongReceived(seq, ping + pong);

      totalPing += ping;
      totalRTT += 1000.0f * pong;
    } else {
      // every even entry will have not received a pong
      ph.pingSent(ping);
    }
  }

  double mean_rtt = totalRTT / 128;
  double std_dev_rtt = 0.0;
  size_t num_jitter_samples = 0;

  for (int i = 1; i < RELAY_PING_HISTORY_ENTRY_COUNT; i += 2) {
    auto& entry = ph[i];
    double rtt = 1000.0 * (entry.TimePongReceived - entry.TimePingSent);
    if (rtt > mean_rtt) {
      double error = rtt - mean_rtt;
      std_dev_rtt += error * error;
      num_jitter_samples++;
    }
  }

  auto expectedRTT = totalRTT / 128;  // every odd is added, so half the relay ping history entry count
  auto expectedJitter = 3.0f * static_cast<float>(std::sqrt(std_dev_rtt / num_jitter_samples));
  auto expectedPacketloss = 50.0;  // half the pings never get a pong, so 50%

  // account for floating point imprecision
  expectedRTT = (expectedRTT * 100.0) / 100.0;
  expectedJitter = (expectedJitter * 100.0) / 100.0;

  core::RouteStats stats(ph, 0.0, 1000.0, 0.0);

  // of the three fields, skipping jitter, too much time to check, relying on legacy code being correct

  double actual_rtt = (stats.getRTT() * 100.0) / 100.0;
  double actual_jitter = (stats.getJitter() * 100.0) / 100.0;
  double actual_packet_loss = (stats.getPacketLoss() * 100.0) / 100.0;

  check(std::abs(actual_rtt - expectedRTT) < 0.001).onFail([&] {
    std::cout << "actual rtt = " << actual_rtt << '\n';
    std::cout << "expected rtt = " << expectedRTT << '\n';
  });
  check(std::abs(actual_jitter - expectedJitter) < 0.001).onFail([&] {
    std::cout << "actual jitter = " << actual_jitter << '\n';
    std::cout << "expected jitter = " << expectedJitter << '\n';
  });
  check(std::abs(actual_packet_loss - expectedPacketloss) < 0.001);
}
