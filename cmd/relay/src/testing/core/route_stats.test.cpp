#include "testing/test.hpp"

#include <random>
#include <functional>

#include "core/route_stats.hpp"

Test(RouteStats_basic) {}

Test(legacy_route_stats_t_basic)
{
  std::default_random_engine gen;
  std::uniform_real_distribution<double> pingDist(0.0, 1000.0);
  auto pingRand = std::bind(pingDist, gen);

  legacy::relay_route_stats_t stats;
  legacy::relay_ping_history_t ph;
  double totalPing = 0.0;

  for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++) {
    auto ping = pingRand();

    if (i & 1) {
      // every odd entry will have received a valid pong
      auto seq = legacy::relay_ping_history_ping_sent(&ph, ping);
      legacy::relay_ping_history_pong_received(&ph, seq, ping + 1.0);
      totalPing += ping;
    } else {
      // every even entry will have not received a pong
      legacy::relay_ping_history_ping_sent(&ph, ping);
    }
  }

  auto expectedPacketloss = 50.0;

  legacy::relay_route_stats_from_ping_history(&ph, 0.0, 1000.0, &stats, 0.0);

  check(stats.packet_loss == expectedPacketloss);
}
