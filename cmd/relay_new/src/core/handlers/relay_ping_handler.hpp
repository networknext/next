#ifndef CORE_HANDLERS_RELAY_PING_HANDLER
#define CORE_HANDLERS_RELAY_PING_HANDLER

#include "base_handler.hpp"
#include "core/packets/relay_ping_packet.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class RelayPingHandler: public BaseHandler
    {
     public:
      RelayPingHandler(
       GenericPacket<>& packet,
       const int size,
       const os::Socket& socket,
       const net::Address& mRecvAddr,
       util::ThroughputRecorder& recorder);

      void handle();

     private:
      const os::Socket& mSocket;
      const net::Address& mRecvAddr;
      util::ThroughputRecorder& mRecorder;
    };

    inline RelayPingHandler::RelayPingHandler(
     GenericPacket<>& packet,
     const int size,
     const os::Socket& socket,
     const net::Address& receivingAddress,
     util::ThroughputRecorder& recorder)
     : BaseHandler(packet, size), mSocket(socket), mRecvAddr(receivingAddress), mRecorder(recorder)
    {}

    inline void RelayPingHandler::handle()
    {
      net::Address sendingAddr;  // where it actually came from
      packets::RelayPingPacket packet(mPacket, mPacketSize);

      packet.Internal.Buffer[0] = RELAY_PONG_PACKET;
      sendingAddr = packet.getFromAddr();
      packet.writeFromAddr(mRecvAddr);

      mRecorder.addToSent(RELAY_PING_PACKET_BYTES);

      // ? probably want to send immediately than use sendmmsg here?
      if (!mSocket.send(sendingAddr, packet.Internal.Buffer.data(), RELAY_PING_PACKET_BYTES)) {
        Log("failed to send data");
      }
    }
  }  // namespace handlers
}  // namespace core
#endif