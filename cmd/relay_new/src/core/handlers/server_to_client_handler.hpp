#ifndef CORE_HANDLERS_SERVER_TO_CLIENT_HANDLER_HPP
#define CORE_HANDLERS_SERVER_TO_CLIENT_HANDLER_HPP
#include "base_handler.hpp"

#include "core/session.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class ServerToClientHandler: public BaseHandler
    {
     public:
      ServerToClientHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int packetSize,
       core::SessionMap& sessions);

      template <typename T, typename F>
      void handle(T& sender, F funcptr);

     private:
      core::SessionMap& mSessionMap;
    };

    inline ServerToClientHandler::ServerToClientHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions)
    {}

    template <typename T, typename F>
    inline void ServerToClientHandler::handle(T& sender, F funcptr)
    {
      if (mPacketSize <= RELAY_HEADER_BYTES || mPacketSize > RELAY_HEADER_BYTES + RELAY_MTU) {
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
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
      if (relay_replay_protection_already_received(&session->ServerToClientProtection, clean_sequence)) {
        return;
      }

      relay_replay_protection_advance_sequence(&session->ServerToClientProtection, clean_sequence);
      if (relay::relay_verify_header(
           RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      (sender.*funcptr)(session->PrevAddr, mPacket.Buffer.data(), mPacketSize);
      LogDebug("sent server packet to ", session->PrevAddr);
    }
  }  // namespace handlers
}  // namespace core
#endif