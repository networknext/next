#ifndef CORE_ROUTE_STATS_HPP
#define CORE_ROUTE_STATS_HPP

#include "ping_history.hpp"

namespace core
{
  class RouteStats
  {
   public:
    RouteStats(const PingHistory& ph, double start, double end, double safety);

    const float RTT;
    const float Jitter;
    const float PacketLoss;
  };
}  // namespace core

namespace legacy
{
  struct relay_route_stats_t
  {
    float rtt;
    float jitter;
    float packet_loss;
  };

  void relay_route_stats_from_ping_history(
   const legacy::relay_ping_history_t* history, double start, double end, relay_route_stats_t* stats, double ping_safety);
}  // namespace legacy
#endif