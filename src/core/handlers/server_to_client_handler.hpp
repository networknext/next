#pragma once

#include "core/packet_header.hpp"
#include "core/packet_types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::PacketDirection;
using core::PacketHeader;
using core::RouterInfo;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void server_to_client_handler(
     Packet& packet, SessionMap& session_map, ThroughputRecorder& recorder, const RouterInfo& router_info, const Socket& socket)
    {
      size_t index = 0;
      size_t length = packet.length;

      // check if length excluding the hash is right,
      // and then check if the hash + everything else is too large
      if (length <= PacketHeader::SIZE_OF_SIGNED || packet.length > PacketHeader::SIZE_OF_SIGNED + RELAY_MTU) {
        LOG(ERROR, "ignoring server to client packet, invalid size: ", length);
        return;
      }

      PacketHeader header;

      {
        size_t i = index;
        if (!header.read(packet, i, PacketDirection::ServerToClient)) {
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

      if (session->expired(router_info.current_time<uint64_t>())) {
        LOG(ERROR, "session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (session->server_to_client_protection.is_already_received(clean_sequence)) {
        LOG(ERROR, "ignoring server to client packet, packet already received: session = ", *session);
        return;
      }

      if (!header.verify(packet, index, PacketDirection::ServerToClient, session->private_key)) {
        LOG(ERROR, "ignoring server to client packet, could not verify header: session = ", *session);
        return;
      }

      session->server_to_client_protection.advance_sequence_to(clean_sequence);

      recorder.server_to_client_tx.add(packet.length);

      if (!socket.send(session->prev_addr, packet.buffer.data(), packet.length)) {
        LOG(ERROR, "failed to forward server packet to ", session->prev_addr);
      }
    }
  }  // namespace handlers
}  // namespace core