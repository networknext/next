#pragma once

#include "core/continue_token.hpp"
#include "core/packet.hpp"
#include "core/packet_types.hpp"
#include "core/router_info.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::ContinueToken;
using core::Packet;
using core::RouterInfo;
using core::Type;
using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;
using crypto::PACKET_HASH_LENGTH;

namespace core
{
  namespace handlers
  {
    INLINE void continue_request_handler(
     Packet& packet,
     SessionMap& session_map,
     const Keychain& keychain,
     ThroughputRecorder& recorder,
     const RouterInfo& router_info,
     const Socket& socket,
     bool is_signed)
    {
      size_t index = 0;
      size_t length = packet.length;

      if (is_signed) {
        index = PACKET_HASH_LENGTH;
        length = packet.length - PACKET_HASH_LENGTH;
      }

      if (length < int(1 + ContinueToken::SIZE_OF_ENCRYPTED * 2)) {
        LOG(ERROR, "ignoring continue request. bad packet size (", length, ")");
        return;
      }

      ContinueToken token;
      {
        size_t i = index + 1;
        if (!token.read_encrypted(packet, i, keychain.backend_public_key, keychain.relay_private_key)) {
          LOG(ERROR, "ignoring continue request. could not read continue token");
          return;
        }
      }

      if (token.expired(router_info)) {
        LOG(INFO, "ignoring continue request. token is expired");
        return;
      }

      auto hash = token.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring continue request. session does not exist");
        return;
      }

      if (session->expired(router_info)) {
        LOG(INFO, "ignoring continue request. session is expired");
        session_map.erase(hash);
        return;
      }

      if (session->expire_timestamp != token.expire_timestamp) {
        LOG(INFO, "session continued: ", token);
      }

      session->expire_timestamp = token.expire_timestamp;

      length = packet.length - ContinueToken::SIZE_OF_ENCRYPTED;

      if (is_signed) {
        packet.buffer[ContinueToken::SIZE_OF_ENCRYPTED + PACKET_HASH_LENGTH] =
         static_cast<uint8_t>(Type::ContinueRequest);
        legacy::relay_sign_network_next_packet(&packet.buffer[ContinueToken::SIZE_OF_ENCRYPTED], length);
      } else {
        packet.buffer[ContinueToken::SIZE_OF_ENCRYPTED] = static_cast<uint8_t>(Type::ContinueRequest);
      }

      recorder.continue_request_tx.add(length);

      if (!socket.send(session->next_addr, &packet.buffer[ContinueToken::SIZE_OF_ENCRYPTED], length)) {
        LOG(ERROR, "failed to forward continue request");
      }
    }
  }  // namespace handlers
}  // namespace core
