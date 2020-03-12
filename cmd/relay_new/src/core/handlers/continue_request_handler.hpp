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
       const crypto::Keychain& keychain);

      template <typename T, typename F>
      void handle(T& sender, F funcptr);

     private:
      core::SessionMap& mSessionMap;
      const crypto::Keychain& mKeychain;
    };

    inline ContinueRequestHandler::ContinueRequestHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions,
     const crypto::Keychain& keychain)
     : BaseHandler(relayClock, routerInfo, packet, packetSize), mSessionMap(sessions), mKeychain(keychain)
    {}

    template <typename T, typename F>
    inline void ContinueRequestHandler::handle(T& sender, F funcptr)
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
        Log("ignoring continue request. token is expired");
        return;
      }

      auto hash = token.key();

      if (!mSessionMap.exists(hash)) {
        Log("ignoring continue request. session does not exist");
        return;
      }

      auto session = mSessionMap.get(hash);

      if (sessionIsExpired(session)) {
        Log("ignoring continue request. session is expired");
        return;
      }

      if (session->ExpireTimestamp != token.ExpireTimestamp) {
        Log("session continued: ", std::hex, token.SessionID, '.', std::dec, static_cast<unsigned int>(token.SessionVersion));
      }

      session->ExpireTimestamp = token.ExpireTimestamp;
      mPacket.Buffer[ContinueToken::EncryptedByteSize] = RELAY_CONTINUE_REQUEST_PACKET;

      (sender.*funcptr)(
       session->NextAddr, &mPacket.Buffer[ContinueToken::EncryptedByteSize], mPacketSize - ContinueToken::EncryptedByteSize);
    }
  }  // namespace handlers
}  // namespace core
#endif