#ifndef CORE_HANDLERS_RELAY_PONG_HANDLER_HPP
#define CORE_HANDLERS_RELAY_PONG_HANDLER_HPP

#include "base_handler.hpp"
#include "core/packets/relay_ping_packet.hpp"
#include "core/relay_manager.hpp"
#include "util/logger.hpp"

namespace core
{
  namespace handlers
  {
    class RelayPongHandler: public BaseHandler
    {
     public:
      RelayPongHandler(GenericPacket<>& packet, int packetSize, RelayManager& manager, RelayManager& v3Manager);

      void handle();

     private:
      RelayManager& mRelayManager;
      RelayManager& mV3RelayManager;
    };

    inline RelayPongHandler::RelayPongHandler(
     GenericPacket<>& packet, int packetSize, RelayManager& manager, RelayManager& v3Manager)
     : BaseHandler(packet, packetSize), mRelayManager(manager), mV3RelayManager(v3Manager)
    {}

    inline void RelayPongHandler::handle()
    {
      packets::RelayPingPacket packet(mPacket, mPacketSize);

      // process the pong time
      if (packet.isV3()) {
        LogDebug("got v3 pong packet from ", packet.getFromAddr());
        mV3RelayManager.processPong(packet.getFromAddr(), packet.getSeqNum());
      } else {
        LogDebug("got pong packet from ", packet.getFromAddr());
        mRelayManager.processPong(packet.getFromAddr(), packet.getSeqNum());
      }
    }
  }  // namespace handlers
}  // namespace core
#endif