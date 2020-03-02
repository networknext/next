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
    template <size_t SenderMaxCap, size_t SenderTimeout>
    class ContinueResponseHandler: public BaseHandler
    {
     public:
      ContinueResponseHandler(const util::Clock& relayClock,
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
    inline ContinueResponseHandler<SenderMaxCap, SenderTimeout>::ContinueResponseHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions,
     const os::Socket& socket,
     net::BufferedSender<SenderMaxCap, SenderTimeout>& sender)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions), mSocket(socket), mSender(sender)
    {}

    template <size_t SenderMaxCap, size_t SenderTimeout>
    inline void ContinueResponseHandler<SenderMaxCap, SenderTimeout>::handle()
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

      if (clean_sequence <= session->ServerToClientSeq) {
        return;
      }

      session->ServerToClientSeq = clean_sequence;

      if (relay::relay_verify_header(
           RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        return;
      }

      mSender.queue(session->PrevAddr, mPacket.Buffer.data(), mPacketSize);
    }
  }  // namespace handlers
}  // namespace core

#endif