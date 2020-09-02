#pragma once

#include "core/packet.hpp"
#include "core/packet_header.hpp"
#include "core/packet_types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::PacketDirection;
using core::PacketHeader;
using core::RouterInfo;
using core::SessionMap;
using crypto::PACKET_HASH_LENGTH;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void session_ping_handler(
     Packet& packet,
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

      if (length > PacketHeader::SIZE_OF + 32) {
        LOG(ERROR, "ignoring session ping, packet size too large: ", length);
        return;
      }

      PacketHeader header;

      {
        size_t i = index;
        if (!header.read(packet, i, PacketDirection::ClientToServer)) {
          LOG(ERROR, "ignoring session ping packet, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring session ping packet, session does not exist: session = ", header);
        return;
      }

      if (session->expired(router_info)) {
        LOG(ERROR, "ignoring session ping packet, session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (clean_sequence <= session->client_to_server_sequence) {
        LOG(ERROR, "ignoring session ping packet, packet already received: session = ", *session);
        return;
      }

      if (!header.verify(packet, index, PacketDirection::ClientToServer, session->private_key)) {
        LOG(ERROR, "ignoring session ping packet, could not verify header: session = ", *session);
        return;
      }

      session->client_to_server_sequence = clean_sequence;

      recorder.session_ping_tx.add(packet.length);

      if (!socket.send(session->next_addr, packet.buffer.data(), packet.length)) {
        LOG(ERROR, "failed to send session pong to ", session->next_addr);
      }
    }
  }  // namespace handlers
}  // namespace core
