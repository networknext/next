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
      OldRelayPingHandler(GenericPacket<>& packet, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats, const uint64_t oldRealyID);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket);

     private:
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
      const uint64_t mOldRelayID;
    };

    inline OldRelayPingHandler::OldRelayPingHandler(
     GenericPacket<>& packet, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats, const uint64_t oldRelayID)
     : BaseHandler(packet), mRecorder(recorder), mStats(stats), mOldRelayID(oldRelayID)
    {}

    template <size_t Size>
    inline void OldRelayPingHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket)
    {
      (void)buff;
      (void)socket;
      packets::OldRelayPingPacket packetWrapper(mPacket);
      core::GenericPacket<> outgoing;

      LogDebug("got ping from ", packetWrapper.getID());

      size_t index = 0;

      if (!encoding::WriteUint8(outgoing.Buffer, index, static_cast<uint8_t>(packets::Type::OldRelayPong))) {
        LogDebug("could not write packet type");
        assert(false);
      }

      if (!encoding::WriteUint64(outgoing.Buffer, index, packetWrapper.getID())) {
        LogDebug("could not write old relay id");
        assert(false);
      }

      if (!encoding::WriteUint64(outgoing.Buffer, index, packetWrapper.getSequence())) {
        LogDebug("could not write sequence");
        assert(false);
      }

      outgoing.Addr = mPacket.Addr;
      outgoing.Len = index;

#ifdef RELAY_MULTISEND
      buff.push(outgoing);
#else
      if (!socket.send(outgoing)) {
        Log("failed to send old pong to ", outgoing.Addr);
      }
#endif
    };
  }  // namespace handlers
}  // namespace core
