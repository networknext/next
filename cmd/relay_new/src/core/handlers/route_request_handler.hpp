#ifndef CORE_HANDLERS_ROUTE_REQUEST_HANDLER_HPP
#define CORE_HANDLERS_ROUTE_REQUEST_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"
#include "legacy/v3/traffic_stats.hpp"

namespace core
{
  namespace handlers
  {
    class RouteRequestHandler: public BaseHandler
    {
     public:
      RouteRequestHandler(
       const util::Clock& relayClock,
       GenericPacket<>& packet,
       const net::Address& from,
       const crypto::Keychain& keychain,
       core::SessionMap& sessions,
       util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& size);

     private:
      const util::Clock& mRelayClock;
      const net::Address& mFrom;
      const crypto::Keychain& mKeychain;
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
      legacy::v3::TrafficStats& mStats;
    };

    inline RouteRequestHandler::RouteRequestHandler(
     const util::Clock& relayClock,
     GenericPacket<>& packet,
     const net::Address& from,
     const crypto::Keychain& keychain,
     core::SessionMap& sessions,
     util::ThroughputRecorder& recorder, legacy::v3::TrafficStats& stats)
     : BaseHandler(packet),
       mRelayClock(relayClock),
       mFrom(from),
       mKeychain(keychain),
       mSessionMap(sessions),
       mRecorder(recorder),
       mStats(stats)
    {}

    template <size_t Size>
    inline void RouteRequestHandler::handle(core::GenericPacketBuffer<Size>& buff)
    {
      LogDebug("got route request from ", mFrom);

      if (mPacket.Len < int(1 + RouteToken::EncryptedByteSize * 2)) {
        Log("ignoring route request. bad packet size (", mPacket.Len, ")");
        return;
      }

      // ignore the header byte of the packet
      size_t index = 1;
      core::RouteToken token(mRelayClock);

      if (!token.readEncrypted(mPacket, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
        Log("ignoring route request. could not read route token");
        return;
      }

      // don't do anything if the token is expired - probably should log something here
      if (token.expired()) {
        Log("ignoring route request. token expired");
        return;
      }

      // create a new session and add it to the session map
      uint64_t hash = token.key();

      if (!mSessionMap.exists(hash)) {
        // create the session
        auto session = std::make_shared<Session>(mRelayClock);
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

        mSessionMap.set(hash, session);

        Log("session created: ", std::hex, token.SessionID, '.', std::dec, static_cast<unsigned int>(token.SessionVersion));
      }  // TODO else what?

      // remove this part of the token by offseting it the request packet bytes
      mPacket.Buffer[RouteToken::EncryptedByteSize] = RELAY_ROUTE_REQUEST_PACKET;

      LogDebug("sending route request to ", token.NextAddr);

      auto length = mPacket.Len - RouteToken::EncryptedByteSize;
      mRecorder.addToSent(length);
      mStats.BytesPerSecManagementTx += length;
      buff.push(token.NextAddr, &mPacket.Buffer[RouteToken::EncryptedByteSize], length);
    }
  }  // namespace handlers
}  // namespace core
#endif