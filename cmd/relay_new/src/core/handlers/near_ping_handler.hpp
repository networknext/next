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
       const os::Socket& socket,
       util::ThroughputRecorder& recorder,
       legacy::v3::TrafficStats& stats);

      void handle();

     private:
      const net::Address& mFrom;
      const os::Socket& mSocket;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline NearPingHandler::NearPingHandler(
     GenericPacket<>& packet,
     const net::Address& from,
     const os::Socket& socket,
     util::ThroughputRecorder& recorder,
     legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mFrom(from), mSocket(socket), mRecorder(recorder), mStats(stats)
    {}

    inline void NearPingHandler::handle()
    {
      if (mPacket.Len != 1 + 8 + 8 + 8 + 8) {
        return;
      }

      mPacket.Buffer[0] = static_cast<uint8_t>(packets::Type::NearPong);
      auto length = mPacket.Len - 16;  // ? why 16
      mRecorder.addToSent(length);
      mStats.BytesPerSecMeasurementTx += length;
      mSocket.send(mFrom, mPacket.Buffer.data(), length);  // ? why 16?
    }
  }  // namespace handlers
}  // namespace core
#endif