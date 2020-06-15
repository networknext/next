#include "includes.h"
#include "relay_bandwidth_limiter.hpp"

namespace relay
{
  void relay_bandwidth_limiter_reset(relay_bandwidth_limiter_t* bandwidth_limiter)
  {
    assert(bandwidth_limiter);
    bandwidth_limiter->last_check_time = -100.0;
    bandwidth_limiter->bits_sent = 0;
    bandwidth_limiter->average_kbps = 0.0;
  }

  bool relay_bandwidth_limiter_add_packet(
   relay_bandwidth_limiter_t* bandwidth_limiter, double current_time, uint32_t kbps_allowed, uint32_t packet_bits)
  {
    assert(bandwidth_limiter);
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if (invalid || current_time - bandwidth_limiter->last_check_time >= RELAY_BANDWIDTH_LIMITER_INTERVAL - 0.001f) {
      bandwidth_limiter->bits_sent = 0;
      bandwidth_limiter->last_check_time = current_time;
    }
    bandwidth_limiter->bits_sent += packet_bits;
    return bandwidth_limiter->bits_sent > (uint64_t)(kbps_allowed * 1000 * RELAY_BANDWIDTH_LIMITER_INTERVAL);
  }

  void relay_bandwidth_limiter_add_sample(relay_bandwidth_limiter_t* bandwidth_limiter, double kbps)
  {
    if (bandwidth_limiter->average_kbps == 0.0 && kbps != 0.0) {
      bandwidth_limiter->average_kbps = kbps;
      return;
    }

    if (bandwidth_limiter->average_kbps != 0.0 && kbps == 0.0) {
      bandwidth_limiter->average_kbps = 0.0;
      return;
    }

    const double delta = kbps - bandwidth_limiter->average_kbps;

    if (delta < 0.000001f) {
      bandwidth_limiter->average_kbps = kbps;
      return;
    }

    bandwidth_limiter->average_kbps += delta * 0.1f;
  }

  double relay_bandwidth_limiter_usage_kbps(relay_bandwidth_limiter_t* bandwidth_limiter, double current_time)
  {
    assert(bandwidth_limiter);
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if (!invalid) {
      const double delta_time = current_time - bandwidth_limiter->last_check_time;
      if (delta_time > 0.1f) {
        const double kbps = bandwidth_limiter->bits_sent / delta_time / 1000.0;
        relay_bandwidth_limiter_add_sample(bandwidth_limiter, kbps);
      }
    }
    return bandwidth_limiter->average_kbps;
  }

  int relay_wire_packet_bits(int packet_bytes)
  {
    return (14 + 20 + 8 + packet_bytes + 4) * 8;
  }
}  // namespace relay
