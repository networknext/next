#ifndef CORE_HANDLERS_SERVER_TO_CLIENT_HANDLER_HPP
#define CORE_HANDLERS_SERVER_TO_CLIENT_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"
#include "legacy/v3/traffic_stats.hpp"

namespace core
{
  namespace handlers
  {
    class ServerToClientHandler: public BaseHandler
    {
     public:
      ServerToClientHandler(
       GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline ServerToClientHandler::ServerToClientHandler(
     GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSessionMap(sessions), mRecorder(recorder), mStats(stats)
    {}

    template <size_t Size>
    inline void ServerToClientHandler::handle(core::GenericPacketBuffer<Size>& buff)
    {
      if (mPacket.Len <= RELAY_HEADER_BYTES || mPacket.Len > RELAY_HEADER_BYTES + RELAY_MTU) {
        return;
      }

      uint8_t type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT,
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
      if (relay_replay_protection_already_received(&session->ServerToClientProtection, clean_sequence)) {
        return;
      }

      relay_replay_protection_advance_sequence(&session->ServerToClientProtection, clean_sequence);
      if (
       relay::relay_verify_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), mPacket.Buffer.data(), mPacket.Len) != RELAY_OK) {
        return;
      }

      mRecorder.addToSent(mPacket.Len);
      mStats.BytesPerSecPaidTx += mPacket.Len;
      buff.push(session->PrevAddr, mPacket.Buffer.data(), mPacket.Len);
      LogDebug("sent server response to ", session->PrevAddr);
    }
  }  // namespace handlers
}  // namespace core
#endif