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
using core::PacketType;
using core::RouterInfo;
using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;

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
     const Socket& socket)
    {
      LOG(DEBUG, "got continue request from ", packet.addr);

      size_t index = 0;
      size_t length = packet.length;

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
        LOG(ERROR, "ignoring continue request. token is expired");
        return;
      }

      auto hash = token.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring continue request. session does not exist");
        return;
      }

      if (session->expired(router_info)) {
        LOG(ERROR, "ignoring continue request. session is expired");
        session_map.erase(hash);
        return;
      }

      if (session->expire_timestamp != token.expire_timestamp) {
        LOG(INFO, "session continued: ", token);
      }

      session->expire_timestamp = token.expire_timestamp;

      length = packet.length - ContinueToken::SIZE_OF_ENCRYPTED;

      packet.buffer[ContinueToken::SIZE_OF_ENCRYPTED] = static_cast<uint8_t>(PacketType::ContinueRequest);

      recorder.continue_request_tx.add(length);

      LOG(DEBUG, "forwarding continue request to ", session->next_addr);
      if (!socket.send(session->next_addr, &packet.buffer[ContinueToken::SIZE_OF_ENCRYPTED], length)) {
        LOG(ERROR, "failed to forward continue request");
      }
    }
  }  // namespace handlers
}  // namespace core
