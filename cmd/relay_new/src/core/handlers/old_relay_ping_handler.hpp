#pragma once

#include "base_handler.hpp"
#include "core/packets/old_relay_ping_packet.hpp"
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
    class OldRelayPingHandler: public BaseHandler
    {
     public:
      OldRelayPingHandler(
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

    inline OldRelayPingHandler::OldRelayPingHandler(
     GenericPacket<>& packet,
     const os::Socket& socket,
     const net::Address& receivingAddress,
     util::ThroughputRecorder& recorder,
     legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSocket(socket), mRecvAddr(receivingAddress), mRecorder(recorder), mStats(stats)
    {}

    inline void OldRelayPingHandler::handle()
    {
      packets::OldRelayPingPacket packetWrapper(mPacket);
      core::GenericPacket<> outgoing;

      size_t index = 0;
      encoding::WriteUint8(outgoing.Buffer, index, static_cast<uint8_t>(packets::Type::OldRelayPong));
      encoding::WriteUint64(outgoing.Buffer, index, packetWrapper.getID());
      encoding::WriteUint64(outgoing.Buffer, index, packetWrapper.getSequence());
      outgoing.Addr = mPacket.Addr;
      outgoing.Len = index;
      mSocket.send(outgoing);
    };
  }  // namespace handlers
}  // namespace core
