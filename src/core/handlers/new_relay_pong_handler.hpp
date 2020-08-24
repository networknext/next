#pragma once

#include "core/packets/relay_ping_packet.hpp"
#include "core/relay_manager.hpp"

using core::packets::RelayPingPacket;

namespace core
{
  namespace handlers
  {
    inline void relay_pong_handler(GenericPacket<>& packet, RelayManager& manager)
    {
      packets::RelayPingPacket pkt(packet);
      manager.processPong(packet.Addr, pkt.getSeqNum());
    }
  }  // namespace handlers
}  // namespace core
