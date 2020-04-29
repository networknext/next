#ifndef CORE_HANDLERS_CLIENT_TO_SERVER_HANDLER_HPP
#define CORE_HANDLERS_CLIENT_TO_SERVER_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class ClientToServerHandler: public BaseHandler
    {
     public:
      ClientToServerHandler(
       GenericPacket<>& packet, const int packetSize, core::SessionMap& sessions, util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
    };

    inline ClientToServerHandler::ClientToServerHandler(
     GenericPacket<>& packet, const int packetSize, core::SessionMap& sessions, util::ThroughputRecorder& recorder)
     : BaseHandler(packet, packetSize), mSessionMap(sessions), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void ClientToServerHandler::handle(core::GenericPacketBuffer<Size>& buff)
    {
      if (mPacketSize <= RELAY_HEADER_BYTES || mPacketSize > RELAY_HEADER_BYTES + RELAY_MTU) {
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
        mPacketSize) != RELAY_OK) {
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
        return;
      }

      relay_replay_protection_advance_sequence(&session->ClientToServerProtection, clean_sequence);
      if (
       relay::relay_verify_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      LogDebug("sending client packet to ", session->NextAddr);
      mRecorder.addToSent(mPacketSize);
      buff.push(session->NextAddr, mPacket.Buffer.data(), mPacketSize);
    }
  }  // namespace handlers
}  // namespace core
#endif