#pragma once

#include "core/packet_types.hpp"
#include "core/route_token.hpp"
#include "core/router_info.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::PacketType;
using core::RouterInfo;
using core::RouteToken;
using core::RouteTokenV4;
using core::SessionMap;
using crypto::Keychain;
using crypto::PACKET_HASH_LENGTH;
using os::Socket;
using util::ThroughputRecorder;
namespace core
{
  namespace handlers
  {
    INLINE void route_request_handler_sdk4(
     Packet& packet,
     const Keychain& keychain,
     SessionMap& session_map,
     ThroughputRecorder& recorder,
     const RouterInfo& router_info,
     const Socket& socket)
    {
      size_t index = 0;
      size_t length = packet.length;

      if (length < static_cast<size_t>(1 + RouteTokenV4::SIZE_OF_SIGNED * 2)) {
        LOG(ERROR, "ignoring route request. bad packet size (", length, ")");
        return;
      }

      RouteTokenV4 token;
      {
        size_t i = index + 1;
        if (!token.read_encrypted(packet, i, keychain.backend_public_key, keychain.relay_private_key)) {
          LOG(ERROR, "ignoring route request. could not read route token");
          return;
        }
      }

      if (token.expired(router_info.current_time<uint64_t>())) {
        LOG(INFO, "ignoring route request, token expired, session = ", token);
        return;
      }

      uint64_t hash = token.hash();

      // create a new session and add it to the session map
      if (!session_map.get(hash)) {
        // create the session
        auto session = std::make_shared<Session>();
        assert(session);

        // fill it with data in the token
        session->expire_timestamp = token.expire_timestamp;
        session->session_id = token.session_id;
        session->session_version = token.session_version;

        // initialize the rest of the fields
        session->client_to_server_sequence = 0;
        session->server_to_client_sequence = 0;
        session->kbps_up = token.kbps_up;
        session->kbps_down = token.kbps_down;
        session->prev_addr = packet.addr;
        session->next_addr = token.next_addr;
        std::copy(token.private_key.begin(), token.private_key.end(), session->private_key.begin());
        session->client_to_server_protection.reset();
        session->server_to_client_protection.reset();

        session_map.set(hash, session);

        LOG(INFO, "session created: ", *session);
      } else {
        LOG(DEBUG, "received additional route request for session: ", token);
      }

      // remove this part of the token by offseting it the request packet bytes

      length = packet.length - RouteTokenV4::SIZE_OF_SIGNED;

      packet.buffer[RouteTokenV4::SIZE_OF_SIGNED] = static_cast<uint8_t>(PacketType::RouteRequest4);

      recorder.route_request_tx.add(length);

      if (!socket.send(token.next_addr, &packet.buffer[RouteTokenV4::SIZE_OF_SIGNED], length)) {
        LOG(ERROR, "failed to forward route request to ", token.next_addr);
      }
    }

    INLINE void route_request_handler(
     Packet& packet,
     const Keychain& keychain,
     SessionMap& session_map,
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

      if (length < static_cast<size_t>(1 + RouteToken::SIZE_OF_ENCRYPTED * 2)) {
        LOG(ERROR, "ignoring route request. bad packet size (", length, ")");
        return;
      }

      RouteToken token;
      {
        size_t i = index + 1;
        if (!token.read_encrypted(packet, i, keychain.backend_public_key, keychain.relay_private_key)) {
          LOG(ERROR, "ignoring route request. could not read route token");
          return;
        }
      }

      if (token.expired(router_info.current_time<uint64_t>())) {
        LOG(INFO, "ignoring route request, token expired, session = ", token);
        return;
      }

      // create a new session and add it to the session map
      uint64_t hash = token.hash();

      if (!session_map.get(hash)) {
        // create the session
        auto session = std::make_shared<Session>();
        assert(session);

        // fill it with data in the token
        session->expire_timestamp = token.expire_timestamp;
        session->session_id = token.session_id;
        session->session_version = token.session_version;

        // initialize the rest of the fields
        session->client_to_server_sequence = 0;
        session->server_to_client_sequence = 0;
        session->kbps_up = token.kbps_up;
        session->kbps_down = token.kbps_down;
        session->prev_addr = packet.addr;
        session->next_addr = token.next_addr;
        std::copy(token.private_key.begin(), token.private_key.end(), session->private_key.begin());
        session->client_to_server_protection.reset();
        session->server_to_client_protection.reset();

        session_map.set(hash, session);

        LOG(INFO, "session created: ", *session);
      } else {
        LOG(DEBUG, "received additional route request for session: ", token);
      }

      // remove this part of the token by offseting it the request packet bytes

      length = packet.length - RouteToken::SIZE_OF_ENCRYPTED;

      if (is_signed) {
        size_t index = RouteToken::SIZE_OF_ENCRYPTED;
        packet.buffer[index + PACKET_HASH_LENGTH] = static_cast<uint8_t>(PacketType::RouteRequest);
        if (!crypto::sign_network_next_packet(packet.buffer, index, length)) {
          LOG(ERROR, "unable to sign route request packet for session ", token);
        }
      } else {
        packet.buffer[RouteToken::SIZE_OF_ENCRYPTED] = static_cast<uint8_t>(PacketType::RouteRequest);
      }

      recorder.route_request_tx.add(length);

      if (!socket.send(token.next_addr, &packet.buffer[RouteToken::SIZE_OF_ENCRYPTED], length)) {
        LOG(ERROR, "failed to forward route request to ", token.next_addr);
      }
    }
  }  // namespace handlers
}  // namespace core
