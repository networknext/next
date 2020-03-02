#ifndef CORE_HANDLERS_CLIENT_TO_SERVER_HANDLER_HPP
#define CORE_HANDLERS_CLIENT_TO_SERVER_HANDLER_HPP

#include "base_handler.hpp"

#include "core/session.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    template <size_t SenderMaxCap, size_t SenderTimeout>
    class ClientToServerHandler: public BaseHandler
    {
     public:
      ClientToServerHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int packetSize,
       core::SessionMap& sessions,
       const os::Socket& socket,
       net::BufferedSender<SenderMaxCap, SenderTimeout>& sender);

      void handle();

     private:
      core::SessionMap& mSessionMap;
      const os::Socket& mSocket;
      net::BufferedSender<SenderMaxCap, SenderTimeout>& mSender;
    };

    template <size_t SenderMaxCap, size_t SenderTimeout>
    inline ClientToServerHandler<SenderMaxCap, SenderTimeout>::ClientToServerHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions,
     const os::Socket& socket,
     net::BufferedSender<SenderMaxCap, SenderTimeout>& sender)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions), mSocket(socket), mSender(sender)
    {}

    template <size_t SenderMaxCap, size_t SenderTimeout>
    inline void ClientToServerHandler<SenderMaxCap, SenderTimeout>::handle()
    {
      if (mPacketSize <= RELAY_HEADER_BYTES || mPacketSize > RELAY_HEADER_BYTES + RELAY_MTU) {
        return;
      }

      uint8_t type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      if (relay::relay_peek_header(RELAY_DIRECTION_CLIENT_TO_SERVER,
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
      {
        std::lock_guard<std::mutex> lk(mSessionMap.Lock);
        if (mSessionMap.find(hash) == mSessionMap.end()) {
          return;
        }
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
      if (relay_replay_protection_already_received(&session->ClientToServerProtection, clean_sequence)) {
        return;
      }

      relay_replay_protection_advance_sequence(&session->ClientToServerProtection, clean_sequence);
      if (relay::relay_verify_header(
           RELAY_DIRECTION_CLIENT_TO_SERVER, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      mSender.queue(session->NextAddr, mPacket.Buffer.data(), mPacketSize);
      LogDebug("sent client packet to ", session->NextAddr);
    }
  }  // namespace handlers
}  // namespace core
#endif