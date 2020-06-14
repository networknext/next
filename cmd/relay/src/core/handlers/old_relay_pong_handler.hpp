#pragma once

#include "base_handler.hpp"
#include "core/packets/old_relay_pong_packet.hpp"
#include "core/relay_manager.hpp"
#include "util/logger.hpp"

namespace core
{
  namespace handlers
  {
    class OldRelayPongHandler: public BaseHandler
    {
     public:
      OldRelayPongHandler(GenericPacket<>& packet, RelayManager<V3Relay>& manager);

      void handle();

     private:
      RelayManager<V3Relay>& mRelayManager;
    };

    inline OldRelayPongHandler::OldRelayPongHandler(GenericPacket<>& packet, RelayManager<V3Relay>& manager)
     : BaseHandler(packet), mRelayManager(manager)
    {}

    inline void OldRelayPongHandler::handle()
    {
      packets::OldRelayPongPacket packetWrapper(mPacket);
      LogDebug("got old pong from ", mPacket.Addr, " with sequence ", packetWrapper.getSequence());
      mRelayManager.processPong(mPacket.Addr, packetWrapper.getSequence());
    }
  }  // namespace handlers
}  // namespace core
