#pragma once

#include "core/packets/header.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "os/socket.hpp"

using core::packets::Direction;
using core::packets::Header;

namespace core
{
  namespace handlers
  {
    inline void client_to_server_handler(
     GenericPacket<>& packet,
     core::SessionMap& session_map,
     util::ThroughputRecorder& recorder,
     const os::Socket& socket,
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
        LOG(ERROR, "ignoring client to server packet, invalid size: ", length);
        return;
      }

      Header header;

      {
        size_t i = index;
        if (!header.read(packet.Buffer, i, Direction::ClientToServer)) {
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

      if (session->expired()) {
        LOG(INFO, "session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (relay_replay_protection_already_received(&session->ClientToServerProtection, clean_sequence)) {
        LOG(ERROR, "ignoring client to server packet, already received packet: session = ", *session);
        return;
      }

      {
        size_t i = index;
        if (!header.verify(packet.Buffer, i, Direction::ClientToServer, session->PrivateKey)) {
          LOG(ERROR, "ignoring client to server packet, could not verify header: session = ", *session);
          return;
        }
      }

      relay_replay_protection_advance_sequence(&session->ClientToServerProtection, clean_sequence);

      recorder.ClientToServerTx.add(packet.Len);

      if (!socket.send(session->NextAddr, packet.Buffer.data(), packet.Len)) {
        LOG(ERROR, "failed to forward client packet to ", session->NextAddr);
      }
    }
  }  // namespace handlers
}  // namespace core
