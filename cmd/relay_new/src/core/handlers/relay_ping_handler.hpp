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
      RelayPingHandler(GenericPacket<>& packet, const int size, const os::Socket& socket, const net::Address& mRecvAddr);

      void handle();

     private:
      const os::Socket& mSocket;
      const net::Address& mRecvAddr;
    };

    inline RelayPingHandler::RelayPingHandler(
     GenericPacket<>& packet, const int size, const os::Socket& socket, const net::Address& receivingAddress)
     : BaseHandler(packet, size), mSocket(socket), mRecvAddr(receivingAddress)
    {}

    inline void RelayPingHandler::handle()
    {
      net::Address sendingAddr;  // where it actually came from
      packets::RelayPingPacket packet(mPacket, mPacketSize);

      packet.Internal.Buffer[0] = RELAY_PONG_PACKET;
      sendingAddr = packet.getFromAddr();
      packet.writeFromAddr(mRecvAddr);

      LogDebug("got ping packet from ", sendingAddr);

      // ? probably want to send immediately than use sendmmsg here?
      if (!mSocket.send(sendingAddr, packet.Internal.Buffer.data(), RELAY_PING_PACKET_BYTES)) {
        Log("failed to send data");
      }
    }
  }  // namespace handlers
}  // namespace core
#endif