#ifndef CORE_HANDLERS_BASE_HANDLER_HPP
#define CORE_HANDLERS_BASE_HANDLER_HPP

#include "core/packet.hpp"

namespace core
{
  namespace handlers
  {
    class BaseHandler
    {
     protected:
      BaseHandler(GenericPacket<>& packet, const int packetSize);

      GenericPacket<>& mPacket;
      const int mPacketSize;
    };

    inline BaseHandler::BaseHandler(GenericPacket<>& packet, const int packetSize): mPacket(packet), mPacketSize(packetSize) {}
  }  // namespace handlers
}  // namespace core
#endif