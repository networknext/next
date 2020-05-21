#ifndef CORE_HANDLERS_NEAR_PING_HANDLER_HPP
#define CORE_HANDLERS_NEAR_PING_HANDLER_HPP

#include "base_handler.hpp"
#include "core/packets/types.hpp"
#include "core/session_map.hpp"
#include "legacy/v3/traffic_stats.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class NearPingHandler: public BaseHandler
    {
     public:
      NearPingHandler(GenericPacket<>& packet, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned);

     private:
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline NearPingHandler::NearPingHandler(
     GenericPacket<>& packet, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mRecorder(recorder), mStats(stats)
    {}

    template <size_t Size>
    inline void NearPingHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
    {
      (void)buff;
      (void)socket;

      size_t length;

      if (isSigned) {
        length = mPacket.Len - crypto::PacketHashLength;
      } else {
        length = mPacket.Len;
      }

      if (length != 1 + 8 + 8 + 8 + 8) {
        Log("ignoring near ping packet, length invalid: ", length);
        return;
      }

      length = mPacket.Len - 16;

      if (isSigned) {
        mPacket.Buffer[crypto::PacketHashLength] = static_cast<uint8_t>(packets::Type::NearPong);
        crypto::SignNetworkNextPacket(mPacket.Buffer, length);
      } else {
        mPacket.Buffer[0] = static_cast<uint8_t>(packets::Type::NearPong);
      }

      mRecorder.addToSent(length);
      mStats.BytesPerSecMeasurementTx += length;

#ifdef RELAY_MULTISEND
      buff.push(mPacket.Addr, mPacket.Buffer.data(), length);
#else
      if (!socket.send(mPacket.Addr, mPacket.Buffer.data(), length)) {
        Log("failed to send near pong to ", mPacket.Addr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif