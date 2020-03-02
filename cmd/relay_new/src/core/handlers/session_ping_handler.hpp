#ifndef CORE_HANDLERS_SESSION_PING_HANDLER_HPP
#define CORE_HANDLERS_SESSION_PING_HANDLER_HPP

#include "base_handler.hpp"

#include "core/session.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class SessionPingHandler: public BaseHandler
    {
     public:
      SessionPingHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int packetSize,
       core::SessionMap& sessions,
       const os::Socket& socket);

      void handle();

     private:
      core::SessionMap& mSessionMap;
      const os::Socket& mSocket;
    };

    inline SessionPingHandler::SessionPingHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions,
     const os::Socket& socket)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions), mSocket(socket)
    {}

    inline void SessionPingHandler::handle()
    {
      if (mPacketSize > RELAY_HEADER_BYTES + 32) {
        return;
      }

      uint8_t type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (relay::relay_peek_header(
           RELAY_DIRECTION_CLIENT_TO_SERVER, &type, &sequence, &session_id, &session_version, mPacket.Buffer.data(), mPacketSize) !=
          RELAY_OK) {
        return;
      }

      uint64_t hash = session_id ^ session_version;

      core::SessionMap::iterator iter, end;
      {
        std::lock_guard<std::mutex> lk(mSessionMap.Lock);
        iter = mSessionMap.find(hash);
        end = mSessionMap.end();
      }

      if (iter == end) {
        return;
      }

      core::SessionPtr session;
      {
        std::lock_guard<std::mutex> lk(mSessionMap.Lock);
        session = mSessionMap[hash];
      }

      if (sessionIsExpired(session)) {
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
      if (clean_sequence <= session->ClientToServerSeq) {
        return;
      }

      session->ClientToServerSeq = clean_sequence;
      if (relay::relay_verify_header(
           RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      mSocket.send(session->NextAddr, mPacket.Buffer.data(), mPacketSize);
    }
  }  // namespace handlers
}  // namespace core
#endif