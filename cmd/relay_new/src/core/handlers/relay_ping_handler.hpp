#ifndef CORE_HANDLERS_RELAY_PING_HANDLER
#define CORE_HANDLERS_RELAY_PING_HANDLER

#include "base_handler.hpp"
#include "core/packets/relay_ping_packet.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"
#include "legacy/v3/traffic_stats.hpp"

namespace core
{
  namespace handlers
  {
    class RelayPingHandler: public BaseHandler
    {
     public:
      RelayPingHandler(
       GenericPacket<>& packet,
       const os::Socket& socket,
       const net::Address& mRecvAddr,
       util::ThroughputRecorder& recorder,
       legacy::v3::TrafficStats& stats);

      void handle();

     private:
      const os::Socket& mSocket;
      const net::Address& mRecvAddr;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline RelayPingHandler::RelayPingHandler(
     GenericPacket<>& packet,
     const os::Socket& socket,
     const net::Address& receivingAddress,
     util::ThroughputRecorder& recorder,
     legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSocket(socket), mRecvAddr(receivingAddress), mRecorder(recorder), mStats(stats)
    {}

    inline void RelayPingHandler::handle()
    {
      packets::RelayPingPacket packetWrapper(mPacket);

      packetWrapper.Internal.Buffer[0] = RELAY_PONG_PACKET;
      mPacket.Addr = packetWrapper.getFromAddr();
      packetWrapper.writeFromAddr(mRecvAddr);
      mPacket.Len = RELAY_PING_PACKET_BYTES;

      mRecorder.addToSent(mPacket.Len);
      mStats.BytesPerSecMeasurementTx += mPacket.Len;

      LogDebug("sending pong to ", mPacket.Addr);
      if (!mSocket.send(mPacket)) {
        Log("failed to send data");
      }
    }
  }  // namespace handlers
}  // namespace core
#endif