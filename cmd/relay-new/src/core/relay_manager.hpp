#ifndef CORE_RELAY_MANAGER_HPP
#define CORE_RELAY_MANAGER_HPP

#include "relay_stats.hpp"
#include "ping_history.hpp"

#include "net/address.hpp"

namespace core
{
  class RelayManager
  {
   public:
    RelayManager();
    ~RelayManager();

    void reset();

    void update(const std::vector<uint64_t>& relayIDs, const net::Address& relayAddr);

    void processPong(const net::Address& from, uint64_t seq);

    void getStats(RelayStats& stats);

   private:
    int mNumRelays;
    std::array<uint64_t, MAX_RELAYS> mRelayIDs;
    std::array<double, MAX_RELAYS> mLastRelayPingTime;
    std::array<net::Address, MAX_RELAYS> mRelayAddresses;
    std::array<PingHistory*, MAX_RELAYS> mRelayPingHistory;
    std::array<PingHistory, MAX_RELAYS> mPingHistoryArray;
  };
}  // namespace core
#endif
