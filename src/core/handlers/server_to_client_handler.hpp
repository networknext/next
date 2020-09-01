#pragma once

#include "core/packets/header.hpp"
#include "core/packets/types.hpp"
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
      size_t length = packet.Len;

      if (is_signed) {
        index = crypto::PacketHashLength;
        length = packet.Len - crypto::PacketHashLength;
      }

      // check if length excluding the hash is right,
      // and then check if the hash + everything else is too large
      if (length <= Header::ByteSize || packet.Len > Header::ByteSize + RELAY_MTU) {
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

      if (relay_replay_protection_already_received(&session->ServerToClientProtection, clean_sequence)) {
        LOG(ERROR, "ignoring server to client packet, packet already received: session = ", *session);
        return;
      }

      if (!header.verify(packet, index, Direction::ServerToClient, session->PrivateKey)) {
        LOG(ERROR, "ignoring server to client packet, could not verify header: session = ", *session);
        return;
      }

      relay_replay_protection_advance_sequence(&session->ServerToClientProtection, clean_sequence);

      recorder.ServerToClientTx.add(packet.Len);

      if (!socket.send(session->PrevAddr, packet.Buffer.data(), packet.Len)) {
        LOG(ERROR, "failed to forward server packet to ", session->PrevAddr);
      }
    }
  }  // namespace handlers
}  // namespace core