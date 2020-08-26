#pragma once

#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::packets::Direction;
using core::packets::Header;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void session_ping_handler(
     GenericPacket<>& packet, SessionMap& session_map, ThroughputRecorder& recorder, const Socket& socket, bool is_signed)
    {
      size_t index = 0;
      size_t length = packet.Len;

      if (is_signed) {
        index = crypto::PacketHashLength;
        length = packet.Len - crypto::PacketHashLength;
      }

      if (length > Header::ByteSize + 32) {
        LOG(ERROR, "ignoring session ping, packet size too large: ", length);
        return;
      }

      Header header;

      {
        size_t i = 0;
        if (!header.read(packet.Buffer, i, Direction::ClientToServer)) {
          LOG(ERROR, "ignoring session ping packet, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring session ping packet, session does not exist: session = ", hash);
        return;
      }

      if (session->expired()) {
        LOG(ERROR, "ignoring session ping packet, session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (clean_sequence <= session->ClientToServerSeq) {
        LOG(ERROR, "ignoring session ping packet, packet already received: session = ", header);
        return;
      }

      if (!header.verify(packet.Buffer, index, Direction::ClientToServer, session->PrivateKey)) {
        LOG(ERROR, "ignoring session ping packet, could not verify header: session = ", *session);
        return;
      }

      session->ClientToServerSeq = clean_sequence;

      recorder.SessionPingTx.add(packet.Len);

      if (!socket.send(session->NextAddr, packet.Buffer.data(), packet.Len)) {
        LOG(ERROR, "failed to send session pong to ", session->NextAddr);
      }
    }
  }  // namespace handlers
}  // namespace core
