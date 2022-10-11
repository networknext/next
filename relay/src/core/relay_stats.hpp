#pragma once

namespace core
{
  const size_t MAX_RELAYS = 1024;
  struct RelayStats
  {
    unsigned int num_relays;
    std::array<uint64_t, MAX_RELAYS> ids;
    std::array<float, MAX_RELAYS> rtt;
    std::array<float, MAX_RELAYS> jitter;
    std::array<float, MAX_RELAYS> packet_loss;
  };
}
