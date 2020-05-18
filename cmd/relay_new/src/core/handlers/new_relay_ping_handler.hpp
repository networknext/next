#pragma once

#include "base_handler.hpp"
#include "core/packets/new_relay_ping_packet.hpp"
#include "core/packets/types.hpp"
#include "encoding/read.hpp"
#include "legacy/v3/traffic_stats.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class NewRelayPingHandler: public BaseHandler
    {
     public:
      NewRelayPingHandler(
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

    inline NewRelayPingHandler::NewRelayPingHandler(
     GenericPacket<>& packet,
     const os::Socket& socket,
     const net::Address& receivingAddress,
     util::ThroughputRecorder& recorder,
     legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSocket(socket), mRecvAddr(receivingAddress), mRecorder(recorder), mStats(stats)
    {}

    inline void NewRelayPingHandler::handle()
    {
      packets::NewRelayPingPacket packetWrapper(mPacket);

      packetWrapper.Internal.Buffer[0] = static_cast<uint8_t>(packets::Type::NewRelayPong);
      mPacket.Addr = packetWrapper.getFromAddr();
      packetWrapper.writeFromAddr(mRecvAddr);
      mPacket.Len = packets::NewRelayPingPacket::ByteSize;

      mRecorder.addToSent(mPacket.Len);
      mStats.BytesPerSecMeasurementTx += mPacket.Len;

      LogDebug("got new ping from ", mPacket.Addr);

      if (!mSocket.send(mPacket)) {
        Log("failed to send data");
      }
    }
  }  // namespace handlers
}  // namespace core