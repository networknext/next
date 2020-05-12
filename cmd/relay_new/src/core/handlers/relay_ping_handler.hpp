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
       util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats);

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
     util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSocket(socket), mRecvAddr(receivingAddress), mRecorder(recorder), mStats(stats)
    {}

    inline void RelayPingHandler::handle()
    {
      net::Address sendingAddr;  // where it actually came from
      packets::RelayPingPacket packet(mPacket);

      packet.Internal.Buffer[0] = RELAY_PONG_PACKET;
      sendingAddr = packet.getFromAddr();
      packet.writeFromAddr(mRecvAddr);

      mRecorder.addToSent(RELAY_PING_PACKET_BYTES);
      mStats.BytesPerSecMeasurementTx += RELAY_PING_PACKET_BYTES;

      if (!mSocket.send(sendingAddr, packet.Internal.Buffer.data(), RELAY_PING_PACKET_BYTES)) {
        Log("failed to send data");
      }
    }
  }  // namespace handlers
}  // namespace core
#endif