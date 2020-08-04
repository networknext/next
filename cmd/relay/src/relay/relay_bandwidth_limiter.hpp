#ifndef RELAY_RELAY_BANDWIDTH_LIMITER_HPP
#define RELAY_RELAY_BANDWIDTH_LIMITER_HPP

namespace relay
{
  struct relay_bandwidth_limiter_t
  {
    uint64_t bits_sent;
    double last_check_time;
    double average_kbps;
  };

  void relay_bandwidth_limiter_reset(relay_bandwidth_limiter_t* bandwidth_limiter);

  bool relay_bandwidth_limiter_add_packet(
   relay_bandwidth_limiter_t* bandwidth_limiter, double current_time, uint32_t kbps_allowed, uint32_t packet_bits);

  void relay_bandwidth_limiter_add_sample(relay_bandwidth_limiter_t* bandwidth_limiter, double kbps);

  double relay_bandwidth_limiter_usage_kbps(relay_bandwidth_limiter_t* bandwidth_limiter, double current_time);

  int relay_wire_packet_bits(int packet_bytes);
}  // namespace relay
#endif
