#ifndef CORE_HANDLERS_SERVER_TO_CLIENT_HANDLER_HPP
#define CORE_HANDLERS_SERVER_TO_CLIENT_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class ServerToClientHandler: public BaseHandler
    {
     public:
      ServerToClientHandler(GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
    };

    inline ServerToClientHandler::ServerToClientHandler(
     GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder)
     : BaseHandler(packet), mSessionMap(sessions), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void ServerToClientHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
    {
      (void)buff;
      (void)socket;

      uint8_t* data;
      size_t length;

      if (isSigned) {
        data = &mPacket.Buffer[crypto::PacketHashLength];
        length = mPacket.Len - crypto::PacketHashLength;
      } else {
        data = &mPacket.Buffer[0];
        length = mPacket.Len;
      }

      // check if length excluding the hash is right,
      // and then check if the hash + everything else is too large
      if (length <= RELAY_HEADER_BYTES || mPacket.Len > RELAY_HEADER_BYTES + RELAY_MTU) {
        LOG("ignoring server to client packet, invalid size: ", length);
        return;
      }

      packets::Type type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, data, length) != RELAY_OK) {
        LOG("ignoring server to client packet, relay header could not be read");
        return;
      }

      uint64_t hash = session_id ^ session_version;

      auto session = mSessionMap.get(hash);

      if (!session) {
        LOG(
         "session does not exist: session = ", std::hex, session_id, '.', std::dec, static_cast<unsigned int>(session_version));
        return;
      }

      if (session->expired()) {
        LOG("session expired: session = ", *session);
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
      if (relay_replay_protection_already_received(&session->ServerToClientProtection, clean_sequence)) {
        LOG("ignoring server to client packet, packet already received: session = ", *session);
        return;
      }

      if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), data, length) != RELAY_OK) {
        LOG("ignoring server to client packet, could not verify header: session = ", *session);
        return;
      }

      relay_replay_protection_advance_sequence(&session->ServerToClientProtection, clean_sequence);

      mRecorder.ServerToClientTx.add(mPacket.Len);

#ifdef RELAY_MULTISEND
      buff.push(session->PrevAddr, mPacket.Buffer.data(), mPacket.Len);
#else
      if (!socket.send(session->PrevAddr, mPacket.Buffer.data(), mPacket.Len)) {
        LOG("failed to forward server packet to ", session->PrevAddr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif