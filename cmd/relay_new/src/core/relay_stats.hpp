#ifndef CORE_RELAY_STATS_HPP
#define CORE_RELAY_STATS_HPP

namespace core
{

  // TODO make this an array of structs composed of id, rtt, jitter, and pl, would be cache friendly
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
