#ifndef CORE_HANDLERS_SESSION_PING_HANDLER_HPP
#define CORE_HANDLERS_SESSION_PING_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"
#include "legacy/v3/traffic_stats.hpp"

namespace core
{
  namespace handlers
  {
    class SessionPingHandler: public BaseHandler
    {
     public:
      SessionPingHandler(
       GenericPacket<>& packet,
       core::SessionMap& sessions,
       const os::Socket& socket,
       util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats);

      void handle();

     private:
      core::SessionMap& mSessionMap;
      const os::Socket& mSocket;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline SessionPingHandler::SessionPingHandler(
     GenericPacket<>& packet,
     core::SessionMap& sessions,
     const os::Socket& socket,
     util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSessionMap(sessions), mSocket(socket), mRecorder(recorder), mStats(stats)
    {}

    inline void SessionPingHandler::handle()
    {
      if (mPacket.Len > RELAY_HEADER_BYTES + 32) {
        return;
      }

      uint8_t type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER,
        &type,
        &sequence,
        &session_id,
        &session_version,
        mPacket.Buffer.data(),
        mPacket.Len) != RELAY_OK) {
        return;
      }

      uint64_t hash = session_id ^ session_version;

      if (!mSessionMap.exists(hash)) {
        return;
      }

      auto session = mSessionMap.get(hash);

      if (session->expired()) {
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
      if (clean_sequence <= session->ClientToServerSeq) {
        return;
      }

      session->ClientToServerSeq = clean_sequence;
      if (
       relay::relay_verify_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), mPacket.Buffer.data(), mPacket.Len) != RELAY_OK) {
        return;
      }

      mRecorder.addToSent(mPacket.Len);
      mStats.BytesPerSecMeasurementTx += mPacket.Len;
      mSocket.send(session->NextAddr, mPacket.Buffer.data(), mPacket.Len);
    }
  }  // namespace handlers
}  // namespace core
#endif