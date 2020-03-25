#ifndef CORE_HANDLERS_SESSION_PONG_HANDLER_HPP
#define CORE_HANDLERS_SESSION_PONG_HANDLER_HPP

#include "base_handler.hpp"

#include "crypto/keychain.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class SessionPongHandler: public BaseHandler
    {
     public:
      SessionPongHandler(GenericPacket<>& packet, const int packetSize, core::SessionMap& sessions, const os::Socket& socket);

      void handle();

     private:
      core::SessionMap& mSessionMap;
      const os::Socket& mSocket;
    };

    inline SessionPongHandler::SessionPongHandler(
     GenericPacket<>& packet, const int packetSize, core::SessionMap& sessions, const os::Socket& socket)
     : BaseHandler(packet, packetSize), mSessionMap(sessions), mSocket(socket)
    {}

    inline void SessionPongHandler::handle()
    {
      if (mPacketSize > RELAY_HEADER_BYTES + 32) {
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
        mPacketSize) != RELAY_OK) {
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
      if (clean_sequence <= session->ServerToClientSeq) {
        return;
      }

      session->ServerToClientSeq = clean_sequence;
      if (
       relay::relay_verify_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      mSocket.send(session->PrevAddr, mPacket.Buffer.data(), mPacketSize);
    }
  }  // namespace handlers
}  // namespace core
#endif