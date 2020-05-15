#pragma once

#include "base_handler.hpp"
#include "core/packets/old_relay_ping_packet.hpp"
#include "core/relay_manager.hpp"
#include "util/logger.hpp"

namespace core
{
  namespace handlers
  {
    class OldRelayPongHandler: public BaseHandler
    {
     public:
      OldRelayPongHandler(GenericPacket<>& packet, RelayManager& manager);

      void handle();

     private:
      RelayManager& mRelayManager;
    };

    inline OldRelayPongHandler::OldRelayPongHandler(GenericPacket<>& packet, RelayManager& manager)
     : BaseHandler(packet), mRelayManager(manager)
    {}

    inline void OldRelayPongHandler::handle() {
      packets::OldRelayPingPacket packet(mPacket);
    }
  }  // namespace handlers
}  // namespace core
