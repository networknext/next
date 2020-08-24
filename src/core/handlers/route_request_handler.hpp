#pragma once

#include "core/packets/types.hpp"
#include "core/router_info.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "os/socket.hpp"

using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    inline void route_request_handler(
     GenericPacket<>& packet,
     const Keychain& keychain,
     SessionMap& session_map,
     ThroughputRecorder& recorder,
     const RouterInfo& router_info,
     const Socket& socket,
     bool is_signed)
    {
      size_t index = 0;
      size_t length = packet.Len;

      if (is_signed) {
        index = crypto::PacketHashLength;
        length = packet.Len - crypto::PacketHashLength;
      }

      if (length < static_cast<size_t>(1 + RouteToken::EncryptedByteSize * 2)) {
        LOG(ERROR, "ignoring route request. bad packet size (", length, ")");
        return;
      }

      core::RouteToken token(router_info);
      {
        size_t i = index + 1;
        if (!token.read_encrypted(packet, index, keychain.RouterPublicKey, keychain.RelayPrivateKey)) {
          LOG(ERROR, "ignoring route request. could not read route token");
          return;
        }
      }

      // don't do anything if the token is expired - probably should log something here
      if (token.expired()) {
        LOG(INFO, "ignoring route request. token expired");
        return;
      }

      // create a new session and add it to the session map
      uint64_t hash = token.hash();

      if (!session_map.get(hash)) {
        // create the session
        auto session = std::make_shared<Session>(router_info);
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
        session->PrevAddr = packet.Addr;
        session->NextAddr = token.NextAddr;
        std::copy(token.PrivateKey.begin(), token.PrivateKey.end(), session->PrivateKey.begin());
        relay_replay_protection_reset(&session->ClientToServerProtection);
        relay_replay_protection_reset(&session->ServerToClientProtection);

        session_map.set(hash, session);

        LOG(INFO, "session created: ", *session);
      } else {
        LOG(DEBUG, "received additional route request for session: ", token);
      }

      // remove this part of the token by offseting it the request packet bytes

      length = packet.Len - RouteToken::EncryptedByteSize;

      if (is_signed) {
        packet.Buffer[RouteToken::EncryptedByteSize + crypto::PacketHashLength] =
         static_cast<uint8_t>(packets::Type::RouteRequest);
        legacy::relay_sign_network_next_packet(&packet.Buffer[RouteToken::EncryptedByteSize], length);
      } else {
        packet.Buffer[RouteToken::EncryptedByteSize] = static_cast<uint8_t>(packets::Type::RouteRequest);
      }

      recorder.RouteRequestTx.add(length);

#ifdef RELAY_MULTISEND
      buff.push(token.NextAddr, &mPacket.Buffer[RouteToken::EncryptedByteSize], length);
#else
      if (!socket.send(token.NextAddr, &packet.Buffer[RouteToken::EncryptedByteSize], length)) {
        LOG(ERROR, "failed to forward route request to ", token.NextAddr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
