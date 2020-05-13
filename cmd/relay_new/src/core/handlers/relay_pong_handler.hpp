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
      RelayPongHandler(GenericPacket<>& packet, RelayManager& manager);

      void handle();

     private:
      RelayManager& mRelayManager;
    };

    inline RelayPongHandler::RelayPongHandler(GenericPacket<>& packet, RelayManager& manager)
     : BaseHandler(packet), mRelayManager(manager)
    {}

    inline void RelayPongHandler::handle()
    {
      packets::RelayPingPacket packet(mPacket);

      // process the pong time
      LogDebug("got pong packet from ", packet.getFromAddr());
      mRelayManager.processPong(packet.getFromAddr(), packet.getSeqNum());
    }
  }  // namespace handlers
}  // namespace core
#endif