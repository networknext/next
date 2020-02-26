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
       GenericPacket& packet,
       const int size,
       const net::Address& from,
       const crypto::Keychain& keychain,
       core::SessionMap& sessions,
       const os::Socket& socket);
      void handle();

     private:
      const net::Address& mFrom;
      const crypto::Keychain& mKeychain;
      core::SessionMap& mSessionMap;
      const os::Socket& mSocket;
    };

    inline RouteRequestHandler::RouteRequestHandler(const util::Clock& relayClock,
     const RouterInfo& routerInfo,
     GenericPacket& packet,
     const int size,
     const net::Address& from,
     const crypto::Keychain& keychain,
     core::SessionMap& sessions,
     const os::Socket& socket)
     : BaseHandler(relayClock, routerInfo, packet, size),
       mFrom(from),
       mKeychain(keychain),
       mSessionMap(sessions),
       mSocket(socket)
    {}

    inline void RouteRequestHandler::handle()
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

      core::SessionMap::iterator iter, end;
      {
        std::lock_guard<std::mutex> lk(mSessionMap.Lock);
        iter = mSessionMap.find(hash);
        end = mSessionMap.end();
      }

      if (iter == end) {
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

        {
          std::lock_guard<std::mutex> lk(mSessionMap.Lock);
          mSessionMap[hash] = session;
        }

        Log("session created: ", std::hex, token.SessionID, '.', std::dec, static_cast<unsigned int>(token.SessionVersion));
      }  // TODO else what?

      // remove this part of the token by offseting it the request packet bytes
      mPacket[RouteToken::EncryptedByteSize] = RELAY_ROUTE_REQUEST_PACKET;

      LogDebug("sending route request to ", token.NextAddr);

      mSocket.send(token.NextAddr, mPacket.data() + RouteToken::EncryptedByteSize, mPacketSize - RouteToken::EncryptedByteSize);

      // net::Message msg(token.NextAddr, mPacket, RouteToken::EncryptedByteSize, mPacketSize - RouteToken::EncryptedByteSize);
      // mSender.queue(msg);  // after this, token & packet are invalid
    }
  }  // namespace handlers
}  // namespace core
#endif