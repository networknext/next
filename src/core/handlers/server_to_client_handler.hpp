#pragma once

#include "core/packet_header.hpp"
#include "core/packet_types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::RouterInfo;
using core::packets::Direction;
using core::packets::Header;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void server_to_client_handler(
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
        index = crypto::PACKET_HASH_LENGTH;
        length = packet.length - crypto::PACKET_HASH_LENGTH;
      }

      // check if length excluding the hash is right,
      // and then check if the hash + everything else is too large
      if (length <= Header::ByteSize || packet.length > Header::ByteSize + RELAY_MTU) {
        LOG(ERROR, "ignoring server to client packet, invalid size: ", length);
        return;
      }

      Header header;

      {
        size_t i = index;
        if (!header.read(packet, i, Direction::ServerToClient)) {
          LOG(ERROR, "ignoring server to client packet, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "session does not exist: session = ", header);
        return;
      }

      if (session->expired(router_info)) {
        LOG(ERROR, "session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (relay_replay_protection_already_received(&session->server_to_client_protection, clean_sequence)) {
        LOG(ERROR, "ignoring server to client packet, packet already received: session = ", *session);
        return;
      }

      if (!header.verify(packet, index, Direction::ServerToClient, session->private_key)) {
        LOG(ERROR, "ignoring server to client packet, could not verify header: session = ", *session);
        return;
      }

      relay_replay_protection_advance_sequence(&session->server_to_client_protection, clean_sequence);

      recorder.server_to_client_tx.add(packet.length);

      if (!socket.send(session->prev_addr, packet.buffer.data(), packet.length)) {
        LOG(ERROR, "failed to forward server packet to ", session->prev_addr);
      }
    }
  }  // namespace handlers
}  // namespace core