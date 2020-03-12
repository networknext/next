#ifndef CORE_HANDLERS_NEAR_PING_HANDLER_HPP
#define CORE_HANDLERS_NEAR_PING_HANDLER_HPP

#include "base_handler.hpp"

#include "core/session.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class NearPingHandler: public BaseHandler
    {
     public:
      NearPingHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int packetSize,
       const net::Address& from,
       const os::Socket& socket);

      void handle();

     private:
      const net::Address& mFrom;
      const os::Socket& mSocket;
    };

    inline NearPingHandler::NearPingHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     const net::Address& from,
     const os::Socket& socket)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mFrom(from), mSocket(socket)
    {}

    inline void NearPingHandler::handle()
    {
      if (mPacketSize != 1 + 8 + 8 + 8 + 8) {
        return;
      }

      mPacket.Buffer[0] = RELAY_NEAR_PONG_PACKET;
      mSocket.send(mFrom, mPacket.Buffer.data(), mPacketSize - 16);  // ? why 16?
    }
  }  // namespace handlers
}  // namespace core
#endif