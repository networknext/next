#ifndef CORE_HANDLERS_SESSION_PING_HANDLER_HPP
#define CORE_HANDLERS_SESSION_PING_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class SessionPingHandler: public BaseHandler
    {
     public:
      SessionPingHandler(GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
    };

    inline SessionPingHandler::SessionPingHandler(
     GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder)
     : BaseHandler(packet), mSessionMap(sessions), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void SessionPingHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
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

      if (length > RELAY_HEADER_BYTES + 32) {
        LOG("ignoring session ping, packet size too large: ", length);
        return;
      }

      core::packets::Type type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_CLIENT_TO_SERVER, &type, &sequence, &session_id, &session_version, data, length) != RELAY_OK) {
        LOG("ignoring session ping packet, relay header could not be read");
        return;
      }

      uint64_t hash = session_id ^ session_version;

      auto session = mSessionMap.get(hash);

      if (!session) {
        LOG(
         "ignoring session ping packet, session does not exist: session = ",
         std::hex,
         session_id,
         '.',
         std::dec,
         static_cast<unsigned int>(session_version));
        return;
      }

      if (session->expired()) {
        LOG("ignoring session ping packet, session expired: session = ", *session);
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);

      if (clean_sequence <= session->ClientToServerSeq) {
        LOG(
         "ignoring session ping packet, packet already received: session = ",
         *session,
         ", ",
         clean_sequence,
         " <= ",
         session->ClientToServerSeq);
        return;
      }

      if (relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), data, length) != RELAY_OK) {
        LOG("ignoring session ping packet, could not verify header: session = ", *session);
        return;
      }

      session->ClientToServerSeq = clean_sequence;

      mRecorder.SessionPingTx.add(mPacket.Len);

#ifdef RELAY_MULTISEND
      buff.push(session->NextAddr, mPacket.Buffer.data(), mPacket.Len);
#else
      if (!socket.send(session->NextAddr, mPacket.Buffer.data(), mPacket.Len)) {
        LOG("failed to send session pong to ", session->NextAddr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif