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
      NearPingHandler(
       GenericPacket<>& packet, const net::Address& from, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket);

     private:
      const net::Address& mFrom;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline NearPingHandler::NearPingHandler(
     GenericPacket<>& packet, const net::Address& from, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mFrom(from), mRecorder(recorder), mStats(stats)
    {}

    template <size_t Size>
    inline void NearPingHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket)
    {
      (void)buff;
      (void)socket;
      if (mPacket.Len != 1 + 8 + 8 + 8 + 8) {
        return;
      }

      mPacket.Buffer[0] = static_cast<uint8_t>(packets::Type::NearPong);
      auto length = mPacket.Len - 16;  // ? why 16
      mRecorder.addToSent(length);
      mStats.BytesPerSecMeasurementTx += length;

#ifdef RELAY_MULTISEND
      buff.push(mFrom, mPacket.Buffer.data(), length);
#else
      if (!socket.send(mFrom, mPacket.Buffer.data(), length)) {
        Log("failed to send near pong to ", mFrom);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif