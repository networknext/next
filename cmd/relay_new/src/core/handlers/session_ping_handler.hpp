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
       util::ThroughputRecorder& recorder,
       legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline SessionPingHandler::SessionPingHandler(
     GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSessionMap(sessions), mRecorder(recorder), mStats(stats)
    {}

    template <size_t Size>
    inline void SessionPingHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket)
    {
      (void)buff;
      (void)socket;
      if (mPacket.Len > RELAY_HEADER_BYTES + 32) {
        return;
      }

      core::packets::Type type;
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

#ifdef RELAY_MULTISEND
      buff.push(session->NextAddr, mPacket.Buffer.data(), mPacket.Len);
#else
      if (!socket.send(session->NextAddr, mPacket.Buffer.data(), mPacket.Len)) {
        Log("failed to send session pong to ", session->NextAddr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif