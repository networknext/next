#pragma once

#include "base_handler.hpp"
#include "core/packets/new_relay_ping_packet.hpp"
#include "core/relay_manager.hpp"
#include "util/logger.hpp"

namespace core
{
  namespace handlers
  {
    class NewRelayPongHandler: public BaseHandler
    {
     public:
      NewRelayPongHandler(GenericPacket<>& packet, RelayManager<Relay>& manager);

      void handle();

     private:
      RelayManager<Relay>& mRelayManager;
    };

    inline NewRelayPongHandler::NewRelayPongHandler(GenericPacket<>& packet, RelayManager<Relay>& manager)
     : BaseHandler(packet), mRelayManager(manager)
    {}

    inline void NewRelayPongHandler::handle()
    {
      packets::NewRelayPingPacket packet(mPacket);
      mRelayManager.processPong(mPacket.Addr, packet.getSeqNum());
    }
  }  // namespace handlers
}  // namespace core
