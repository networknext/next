#pragma once

#include "core/packet_header.hpp"
#include "core/packet_types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeaderV4;
using core::RouterInfo;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void client_to_server_handler_sdk4(
     Packet& packet,
     core::SessionMap& session_map,
     util::ThroughputRecorder& recorder,
     const RouterInfo& router_info,
     const os::Socket& socket)
    {
      size_t index = 0;
      size_t length = packet.length;

      // check if length excluding the hash is right,
      // and then check if the hash + everything else is too large
      if (length <= PacketHeaderV4::SIZE_OF_SIGNED || packet.length > PacketHeaderV4::SIZE_OF_SIGNED + RELAY_MTU) {
        LOG(ERROR, "ignoring client to server packet, invalid size: ", length);
        return;
      }

      PacketHeaderV4 header;
      {
        size_t i = index;
        if (!header.read(packet, i, PacketDirection::ClientToServer)) {
          LOG(ERROR, "ignoring client to server packet, relay header could not be read");
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
        LOG(INFO, "session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (session->client_to_server_protection.is_already_received(clean_sequence)) {
        LOG(ERROR, "ignoring client to server packet, already received packet: session = ", *session);
        return;
      }

      {
        size_t i = index;
        if (!header.verify(packet, i, PacketDirection::ClientToServer, session->private_key)) {
          LOG(ERROR, "ignoring client to server packet, could not verify header: session = ", *session);
          return;
        }
      }

      session->client_to_server_protection.advance_sequence_to(clean_sequence);

      recorder.client_to_server_tx.add(packet.length);

      if (!socket.send(session->next_addr, packet.buffer.data(), packet.length)) {
        LOG(ERROR, "failed to forward client packet to ", session->next_addr, ", session = ", *session);
      }
    }
  }  // namespace handlers
}  // namespace core
