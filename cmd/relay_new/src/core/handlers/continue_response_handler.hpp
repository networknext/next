#ifndef CORE_HANDLERS_CONTINUE_RESPONSE_HANDLER_HPP
#define CORE_HANDLERS_CONTINUE_RESPONSE_HANDLER_HPP

#include "base_handler.hpp"

#include "crypto/keychain.hpp"

#include "core/session.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class ContinueResponseHandler: public BaseHandler
    {
     public:
      ContinueResponseHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int packetSize,
       core::SessionMap& sessions);

      template <typename T, typename F>
      void handle(T& sender, F funcptr);

     private:
      core::SessionMap& mSessionMap;
    };

    inline ContinueResponseHandler::ContinueResponseHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions)
    {}

    template <typename T, typename F>
    inline void ContinueResponseHandler::handle(T& sender, F funcptr)
    {
      if (mPacketSize != RELAY_HEADER_BYTES) {
        return;
      }

      uint8_t type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (relay::relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
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

      if (sessionIsExpired(session)) {
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);

      if (clean_sequence <= session->ServerToClientSeq) {
        return;
      }

      session->ServerToClientSeq = clean_sequence;

      if (relay::relay_verify_header(
           RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      (sender.*funcptr)(session->PrevAddr, mPacket.Buffer.data(), mPacketSize);
    }
  }  // namespace handlers
}  // namespace core

#endif