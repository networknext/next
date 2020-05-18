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
       GenericPacket<>& packet,
       const net::Address& from,
       util::ThroughputRecorder& recorder,
       legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff);

     private:
      const net::Address& mFrom;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline NearPingHandler::NearPingHandler(
     GenericPacket<>& packet,
     const net::Address& from,
     util::ThroughputRecorder& recorder,
     legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mFrom(from), mRecorder(recorder), mStats(stats)
    {}

    template <size_t Size>
    inline void NearPingHandler::handle(core::GenericPacketBuffer<Size>& buff)
    {
      if (mPacket.Len != 1 + 8 + 8 + 8 + 8) {
        return;
      }

      mPacket.Buffer[0] = static_cast<uint8_t>(packets::Type::NearPong);
      auto length = mPacket.Len - 16;  // ? why 16
      mRecorder.addToSent(length);
      mStats.BytesPerSecMeasurementTx += length;
      buff.push(mFrom, mPacket.Buffer.data(), length);  // ? why 16?
    }
  }  // namespace handlers
}  // namespace core
#endif