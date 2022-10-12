#pragma once

namespace core
{
  struct RouteStats
  {
    float rtt = 0.0f;
    float jitter = -1.0f;
    float packet_loss = -1.0f;
  };
}  // namespace core
