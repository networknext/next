#ifndef RELAY_RELAY_ROUTE_STATS_HPP
#define RELAY_RELAY_ROUTE_STATS_HPP

#include "relay_ping_history.hpp"

namespace relay
{
  struct relay_route_stats_t
  {
    float rtt;
    float jitter;
    float packet_loss;
  };

  void relay_route_stats_from_ping_history(
   const relay_ping_history_t* history, double start, double end, relay_route_stats_t* stats, double ping_safety);
}  // namespace relay
#endif
