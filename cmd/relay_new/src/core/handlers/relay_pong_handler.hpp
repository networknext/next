#ifndef CORE_HANDLERS_RELAY_PONG_HANDLER_HPP
#define CORE_HANDLERS_RELAY_PONG_HANDLER_HPP

#include "base_handler.hpp"

#include "core/relay_manager.hpp"

namespace core
{
  namespace handlers
  {
    class RelayPongHandler: public BaseHandler
    {
     public:
      RelayPongHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket& packet,
       int packetSize,
       RelayManager& manager);

      void handle();

     private:
      RelayManager& mRelayManager;
    };

    inline RelayPongHandler::RelayPongHandler(
     const util::Clock& relayClock, const RouterInfo& routerInfo, GenericPacket& packet, int packetSize, RelayManager& manager)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mRelayManager(manager)
    {}

    inline void RelayPongHandler::handle()
    {
      net::Address addr;  // the actual from

      size_t index = 1;  // skip the identifier byte
      uint64_t sequence = encoding::ReadUint64(mPacket.Buffer, index);
      // pings are sent on a different port, need to read actual address to stay consistent
      encoding::ReadAddress(mPacket.Buffer, index, addr);
      LogDebug("got pong packet from ", addr);

      // process the pong time
      mRelayManager.processPong(addr, sequence);
    }
  }  // namespace handlers
}  // namespace core
#endif