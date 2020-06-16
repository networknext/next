#include "includes.h"
#include "testing/test.hpp"

#include "core/route_stats.hpp"

Test(new_and_legacy_route_stats_basic)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<float> pingDist(0.0, 1000.0);  // pings will happen in this range
  std::uniform_real_distribution<float> pongDist(0.0, 3.0);     // pongs will happen in ping + this range
  auto pingRand = std::bind(pingDist, gen);
  auto pongRand = std::bind(pongDist, gen);

  core::PingHistory ph;

  legacy::relay_route_stats_t l_stats;
  legacy::relay_ping_history_t l_ph;

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

      // legacy
      seq = legacy::relay_ping_history_ping_sent(&l_ph, ping);
      legacy::relay_ping_history_pong_received(&l_ph, seq, ping + pong);

      totalPing += ping;
      totalRTT += 1000.0f * pong;
    } else {
      // every even entry will have not received a pong
      ph.pingSent(ping);
      legacy::relay_ping_history_ping_sent(&l_ph, ping);
    }
  }

  legacy::relay_route_stats_from_ping_history(&l_ph, 0.0, 1000.0, &l_stats, 0.0);

  auto expectedPacketloss = 50.0;     // half the pings never get a pong, so 50%
  auto expectedRTT = totalRTT / 128;  // every odd is added, so half the relay ping history entry count

  // account for floating point imprecision
  expectedRTT = std::floor(expectedRTT * 100) / 100;
  l_stats.rtt = std::floor(l_stats.rtt * 100) / 100;

  core::RouteStats stats(ph, 0.0, 1000.0, 0.0);

  // of the three fields, skipping jitter, too much time to check, relying on legacy code being correct

  check(l_stats.packet_loss == expectedPacketloss);
  check(l_stats.rtt == expectedRTT);


  check(fabs(stats.getRTT() - l_stats.rtt) <= 0.1).onFail([&] {
    std::cout << "stats rtt = " << stats.getRTT() << ", legacy stats = " << l_stats.rtt << std::endl;
  });
  check(stats.getPacketLoss() == l_stats.packet_loss);
  check(stats.getJitter() == l_stats.jitter);
}
