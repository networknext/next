#pragma once

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
