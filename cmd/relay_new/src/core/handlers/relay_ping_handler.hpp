#ifndef CORE_HANDLERS_RELAY_PING_HANDLER
#define CORE_HANDLERS_RELAY_PING_HANDLER

#include "base_handler.hpp"

#include "encoding/read.hpp"

#include "net/address.hpp"

#include "os/platform.hpp"

#include "core/packets/relay_ping_packet.hpp"

namespace core
{
  namespace handlers
  {
    class RelayPingHandler: public BaseHandler
    {
     public:
      RelayPingHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket& packet,
       const int size,
       const os::Socket& socket);

      void handle();

     private:
      const os::Socket& mSocket;
    };

    inline RelayPingHandler::RelayPingHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket& packet,
     const int size,
     const os::Socket& socket)
     : BaseHandler(relayClock, routerInfo, packet, size), mSocket(socket)
    {}

    inline void RelayPingHandler::handle()
    {
      net::Address sendingAddr;                   // where it actually came from
      auto& recevingAddr = mSocket.getAddress();  // where the sender should talk to this relay
      packets::RelayPingPacket packet(mPacket, mPacketSize);

      packet.Data[0] = RELAY_PONG_PACKET;
      sendingAddr = packet.getFromAddr();
      packet.writeFromAddr(recevingAddr);

      LogDebug("got ping packet from ", sendingAddr);

      // TODO probably want to send immediately than use sendmmsg here?
      if (!mSocket.send(sendingAddr, packet.Data.data(), RELAY_PING_PACKET_BYTES)) {
        Log("failed to send data");
      }
    }
  }  // namespace handlers
}  // namespace core
#endif