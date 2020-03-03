#ifndef CORE_HANDLERS_ROUTE_REQUEST_HANDLER_HPP
#define CORE_HANDLERS_ROUTE_REQUEST_HANDLER_HPP

#include "base_handler.hpp"

#include "core/session.hpp"

#include "crypto/keychain.hpp"

#include "net/address.hpp"

#include "os/platform.hpp"

namespace core
{
  namespace handlers
  {
    class RouteRequestHandler: public BaseHandler
    {
     public:
      RouteRequestHandler(const util::Clock& relayClock,
       const RouterInfo& routerInfo,
       GenericPacket<>& packet,
       const int size,
       const net::Address& from,
       const crypto::Keychain& keychain,
       core::SessionMap& sessions);

      template <typename T, typename F>
      void handle(T& sender, F funcptr);

     private:
      const net::Address& mFrom;
      const crypto::Keychain& mKeychain;
      core::SessionMap& mSessionMap;
    };

    inline RouteRequestHandler::RouteRequestHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket<>& packet,
     const int size,
     const net::Address& from,
     const crypto::Keychain& keychain,
     core::SessionMap& sessions)
     : BaseHandler(relayClock, routerInfo, packet, size), mFrom(from), mKeychain(keychain), mSessionMap(sessions)
    {}

    template <typename T, typename F>
    inline void RouteRequestHandler::handle(T& sender, F funcptr)
    {
      LogDebug("got route request from ", mFrom);

      if (mPacketSize < int(1 + RouteToken::EncryptedByteSize * 2)) {
        Log("ignoring route request. bad packet size (", mPacketSize, ")");
        return;
      }

      // ignore the header byte of the packet
      size_t index = 1;
      core::RouteToken token;

      if (!token.readEncrypted(mPacket, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
        Log("ignoring route request. could not read route token");
        return;
      }

      // don't do anything if the token is expired - probably should log something here
      if (tokenIsExpired(token)) {
        Log("ignoring route request. token expired");
        return;
      }

      // create a new session and add it to the session map
      uint64_t hash = token.key();

      if (!mSessionMap.exists(hash)) {
        // create the session
        auto session = std::make_shared<Session>();
        assert(session);

        // fill it with data in the token
        session->ExpireTimestamp = token.ExpireTimestamp;
        session->SessionID = token.SessionID;
        session->SessionVersion = token.SessionVersion;
        session->ClientToServerSeq = 0;
        session->ServerToClientSeq = 0;
        session->KbpsUp = token.KbpsUp;
        session->KbpsDown = token.KbpsDown;
        session->PrevAddr = mFrom;
        session->NextAddr = token.NextAddr;

        // store it
        std::copy(token.PrivateKey.begin(), token.PrivateKey.end(), session->PrivateKey.begin());
        relay_replay_protection_reset(&session->ClientToServerProtection);
        relay_replay_protection_reset(&session->ServerToClientProtection);

        mSessionMap[hash] = session;

        Log("session created: ", std::hex, token.SessionID, '.', std::dec, static_cast<unsigned int>(token.SessionVersion));
      }  // TODO else what?

      // remove this part of the token by offseting it the request packet bytes
      mPacket.Buffer[RouteToken::EncryptedByteSize] = RELAY_ROUTE_REQUEST_PACKET;

      LogDebug("sending route request to ", token.NextAddr);

      (sender.*funcptr)(
       token.NextAddr, &mPacket.Buffer[RouteToken::EncryptedByteSize], mPacketSize - RouteToken::EncryptedByteSize);
    }
  }  // namespace handlers
}  // namespace core
#endif