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
      BaseHandler(GenericPacket<>& packet);

      GenericPacket<>& mPacket;
    };

    inline BaseHandler::BaseHandler(GenericPacket<>& packet): mPacket(packet) {}
  }  // namespace handlers
}  // namespace core
#endif