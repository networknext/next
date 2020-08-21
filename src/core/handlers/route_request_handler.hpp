#pragma once

#include "base_handler.hpp"
#include "core/packets/types.hpp"
#include "core/router_info.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "os/socket.hpp"

namespace core
{
  namespace handlers
  {
    class RouteRequestHandler: public BaseHandler
    {
     public:
      RouteRequestHandler(
       GenericPacket<>& packet,
       const net::Address& from,
       const crypto::Keychain& keychain,
       core::SessionMap& sessions,
       util::ThroughputRecorder& recorder,
       const RouterInfo& routerInfo);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& size, const os::Socket& socket, bool isSigned);

     private:
      const net::Address& mFrom;
      const crypto::Keychain& mKeychain;
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
      const RouterInfo& mRouterInfo;
    };

    inline RouteRequestHandler::RouteRequestHandler(
     GenericPacket<>& packet,
     const net::Address& from,
     const crypto::Keychain& keychain,
     core::SessionMap& sessions,
     util::ThroughputRecorder& recorder,
     const RouterInfo& routerInfo)
     : BaseHandler(packet),
       mFrom(from),
       mKeychain(keychain),
       mSessionMap(sessions),
       mRecorder(recorder),
       mRouterInfo(routerInfo)
    {}

    template <size_t Size>
    inline void RouteRequestHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
    {
      (void)buff;
      (void)socket;

      uint8_t* data;
      size_t length;

      if (isSigned) {
        data = &mPacket.Buffer[crypto::PacketHashLength];
        length = mPacket.Len - crypto::PacketHashLength;
      } else {
        data = &mPacket.Buffer[0];
        length = mPacket.Len;
      }

      if (mPacket.Len < int(1 + RouteToken::EncryptedByteSize * 2)) {
        LOG("ignoring route request. bad packet size (", mPacket.Len, ")");
        return;
      }

      // ignore the header byte of the packet
      size_t index = 1;
      core::RouteToken token(mRouterInfo);

      if (!token.readEncrypted(data, length, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
        LOG("ignoring route request. could not read route token");
        return;
      }

      // don't do anything if the token is expired - probably should log something here
      if (token.expired()) {
        LOG("ignoring route request. token expired");
        return;
      }

      // create a new session and add it to the session map
      uint64_t hash = token.key();

      if (!mSessionMap.get(hash)) {
        // create the session
        auto session = std::make_shared<Session>(mRouterInfo);
        assert(session);

        // fill it with data in the token
        session->ExpireTimestamp = token.ExpireTimestamp;
        session->SessionID = token.SessionID;
        session->SessionVersion = token.SessionVersion;

        // initialize the rest of the fields
        session->ClientToServerSeq = 0;
        session->ServerToClientSeq = 0;
        session->KbpsUp = token.KbpsUp;
        session->KbpsDown = token.KbpsDown;
        session->PrevAddr = mFrom;
        session->NextAddr = token.NextAddr;
        std::copy(token.PrivateKey.begin(), token.PrivateKey.end(), session->PrivateKey.begin());
        relay_replay_protection_reset(&session->ClientToServerProtection);
        relay_replay_protection_reset(&session->ServerToClientProtection);

        mSessionMap.set(hash, session);

        LOG("session created: ", *session);
      } else {
        LogDebug("received additional route request for session: ", token);
      }

      // remove this part of the token by offseting it the request packet bytes

      length = mPacket.Len - RouteToken::EncryptedByteSize;

      if (isSigned) {
        mPacket.Buffer[RouteToken::EncryptedByteSize + crypto::PacketHashLength] =
         static_cast<uint8_t>(packets::Type::RouteRequest);
        legacy::relay_sign_network_next_packet(&mPacket.Buffer[RouteToken::EncryptedByteSize], length);
      } else {
        mPacket.Buffer[RouteToken::EncryptedByteSize] = static_cast<uint8_t>(packets::Type::RouteRequest);
      }

      mRecorder.RouteRequestTx.add(length);

#ifdef RELAY_MULTISEND
      buff.push(token.NextAddr, &mPacket.Buffer[RouteToken::EncryptedByteSize], length);
#else
      if (!socket.send(token.NextAddr, &mPacket.Buffer[RouteToken::EncryptedByteSize], length)) {
        LOG("failed to forward route request to ", token.NextAddr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
