#ifndef CORE_ROUTE_STATS_HPP
#define CORE_ROUTE_STATS_HPP

#include "ping_history.hpp"

namespace core
{
  class RouteStats
  {
   public:
   // start, end, & safety are in seconds
    RouteStats(const PingHistory& ph, double start, double end, double safety);

    auto getRTT() const -> float;
    auto getJitter() const -> float;
    auto getPacketLoss() const -> float;

   private:
    float mRTT;
    float mJitter;
    float mPacketLoss;
  };

  inline auto RouteStats::getRTT() const -> float
  {
    return mRTT;
  }

  inline auto RouteStats::getJitter() const -> float
  {
    return mJitter;
  }

  inline auto RouteStats::getPacketLoss() const -> float
  {
    return mPacketLoss;
  }
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