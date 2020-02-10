#ifndef CORE_RELAY_STATS_HPP
#define CORE_RELAY_STATS_HPP

#include <array>
#include <cinttypes>

#include "config.hpp"

namespace core
{
  struct RelayStats
  {
    unsigned int NumRelays;
    std::array<uint64_t, MAX_RELAYS> IDs;
    std::array<float, MAX_RELAYS> RTT;
    std::array<float, MAX_RELAYS> Jitter;
    std::array<float, MAX_RELAYS> PacketLoss;
  };
}  // namespace core
#endif