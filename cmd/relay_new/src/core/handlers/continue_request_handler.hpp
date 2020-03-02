#ifndef CORE_HANDLERS_CONTINUE_REQUEST_HANDLER
#define CORE_HANDLERS_CONTINUE_REQUEST_HANDLER

#include "base_handler.hpp"

#include "core/session.hpp"

#include "crypto/keychain.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class ContinueRequestHandler: public BaseHandler
    {
     public:
      ContinueRequestHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int packetSize,
       core::SessionMap& sessions,
       const os::Socket& socket,
       const crypto::Keychain& keychain);

      void handle();

     private:
      core::SessionMap& mSessionMap;
      const os::Socket& mSocket;
      const crypto::Keychain& mKeychain;
    };

    inline ContinueRequestHandler::ContinueRequestHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions,
     const os::Socket& socket,
     const crypto::Keychain& keychain)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions), mSocket(socket), mKeychain(keychain)
    {}

    inline void ContinueRequestHandler::handle()
    {
      if (mPacketSize < int(1 + ContinueToken::EncryptedByteSize * 2)) {
        Log("ignoring continue request. bad packet size (", mPacketSize, ")");
        return;
      }

      size_t index = 1;
      core::ContinueToken token;
      if (!token.readEncrypted(mPacket, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
        Log("ignoring continue request. could not read continue token");
        return;
      }

      if (tokenIsExpired(token)) {
        return;
      }

      auto hash = token.key();

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

      if (session->ExpireTimestamp != token.ExpireTimestamp) {
        Log("session continued: ", std::hex, token.SessionID, '.', std::dec, static_cast<unsigned int>(token.SessionVersion));
      }

      session->ExpireTimestamp = token.ExpireTimestamp;
      mPacket.Buffer[ContinueToken::EncryptedByteSize] = RELAY_CONTINUE_REQUEST_PACKET;

      mSocket.send(
       session->NextAddr, mPacket.Buffer.data() + ContinueToken::EncryptedByteSize, mPacketSize - ContinueToken::EncryptedByteSize);
    }
  }  // namespace handlers
}  // namespace core
#endif