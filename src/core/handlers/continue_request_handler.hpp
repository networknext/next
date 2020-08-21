#pragma once

#include "base_handler.hpp"
#include "core/packets/types.hpp"
#include "core/router_info.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "os/socket.hpp"

using core::packets::Direction;
using core::packets::Header;

namespace core
{
  namespace handlers
  {
    inline void continue_request_handler(
     GenericPacket<>& packet,
     core::SessionMap& session_map,
     const crypto::Keychain& keychain,
     util::ThroughputRecorder& recorder,
     const RouterInfo& router_info,
     const os::Socket& socket,
     bool is_signed)
    {
      size_t index = 0;
      size_t length = packet.Len;

      if (is_signed) {
        index = crypto::PacketHashLength;
        length = packet.Len - crypto::PacketHashLength;
      }

      if (length < int(1 + ContinueToken::EncryptedByteSize * 2)) {
        LOG(ERROR, "ignoring continue request. bad packet size (", length, ")");
        return;
      }

      core::ContinueToken token(router_info);
      {
        size_t i = index + 1;
        if (!token.readEncrypted(packet.Buffer, i, keychain.RouterPublicKey, keychain.RelayPrivateKey)) {
          LOG(ERROR, "ignoring continue request. could not read continue token");
          return;
        }
      }

      if (token.expired()) {
        LOG(INFO, "ignoring continue request. token is expired");
        return;
      }

      auto hash = token.key();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring continue request. session does not exist");
        return;
      }

      if (session->expired()) {
        LOG(INFO, "ignoring continue request. session is expired");
        session_map.erase(hash);
        return;
      }

      if (session->ExpireTimestamp != token.ExpireTimestamp) {
        LOG(INFO, "session continued: ", token);
      }

      session->ExpireTimestamp = token.ExpireTimestamp;

      length = packet.Len - ContinueToken::EncryptedByteSize;

      if (is_signed) {
        mPacket.Buffer[ContinueToken::EncryptedByteSize + crypto::PacketHashLength] =
         static_cast<uint8_t>(packets::Type::ContinueRequest);
        legacy::relay_sign_network_next_packet(&mPacket.Buffer[ContinueToken::EncryptedByteSize], length);
      } else {
        mPacket.Buffer[ContinueToken::EncryptedByteSize] = static_cast<uint8_t>(packets::Type::ContinueRequest);
      }

      mRecorder.ContinueRequestTx.add(length);

      if (!socket.send(session->NextAddr, &mPacket.Buffer[ContinueToken::EncryptedByteSize], length)) {
        LOG("failed to forward continue request");
      }
    }
  }  // namespace handlers
}  // namespace core
#endif