#ifndef CORE_HANDLERS_CLIENT_TO_SERVER_HANDLER_HPP
#define CORE_HANDLERS_CLIENT_TO_SERVER_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "crypto/hash.hpp"
#include "os/platform.hpp"
#include "relay/relay.hpp"
#include "util/throughput_recorder.hpp"
#include "legacy/v3/traffic_stats.hpp"

namespace core
{
  namespace handlers
  {
    class ClientToServerHandler: public BaseHandler
    {
     public:
      ClientToServerHandler(
       GenericPacket<>& packet,
       core::SessionMap& sessions,
       util::ThroughputRecorder& recorder,
       legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline ClientToServerHandler::ClientToServerHandler(
     GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet), mSessionMap(sessions), mRecorder(recorder), mStats(stats)
    {}

    template <size_t Size>
    inline void ClientToServerHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
    {
      (void)buff;
      (void)socket;
      if (mPacket.Len <= RELAY_HEADER_BYTES || mPacket.Len > RELAY_HEADER_BYTES + RELAY_MTU) {
        return;
      }

      packets::Type type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      uint8_t* data;
      size_t length;

      if (isSigned) {
        data = &mPacket.Buffer[crypto::PacketHashLength];
        length = mPacket.Len - crypto::PacketHashLength;
      } else {
        data = &mPacket.Buffer[0];
        length = mPacket.Len;
      }

      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER, &type, &sequence, &session_id, &session_version, data, length) != RELAY_OK) {
        return;
      }

      uint64_t hash = session_id ^ session_version;

      // check if the session is registered
      if (!mSessionMap.exists(hash)) {
        return;
      }

      auto session = mSessionMap.get(hash);

      if (session->expired()) {
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
      if (relay_replay_protection_already_received(&session->ClientToServerProtection, clean_sequence)) {
        LogDebug("client to server packet already received");
        return;
      }

      relay_replay_protection_advance_sequence(&session->ClientToServerProtection, clean_sequence);
      if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), data, length) != RELAY_OK) {
        Log("failed to verify client to server packet header");
        return;
      }

      LogDebug("sending client packet to ", session->NextAddr);
      mRecorder.addToSent(mPacket.Len);
      mStats.BytesPerSecPaidTx += mPacket.Len;

// no signing required, packet doesn't change
#ifdef RELAY_MULTISEND
      buff.push(session->NextAddr, mPacket.Buffer.data(), mPacket.Len);
#else
      if (!socket.send(session->NextAddr, mPacket.Buffer.data(), mPacket.Len)) {
        Log("failed to forward client packet");
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif