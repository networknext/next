#pragma once

#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "os/socket.hpp"

using core::RouterInfo;
using core::SessionMap;
using core::packets::Direction;
using core::packets::Header;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    inline void session_pong_handler(
     Packet& packet,
     SessionMap& session_map,
     ThroughputRecorder& recorder,
     const RouterInfo& router_info,
     const os::Socket& socket,
     bool is_signed)
    {
      size_t index = 0;
      size_t length = packet.Len;

      if (is_signed) {
        index = crypto::PacketHashLength;
        length = packet.Len - crypto::PacketHashLength;
      }

      if (length > Header::ByteSize + 32) {
        LOG(ERROR, "ignoring session pong, packet size too large: ", length);
        return;
      }

      Header header;

      {
        size_t i = index;
        if (!header.read(packet, i, Direction::ServerToClient)) {
          LOG(ERROR, "ignoring session pong packet, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring session pong packet, session does not exist: session = ", header);
        return;
      }

      if (session->expired(router_info)) {
        LOG(ERROR, "ignoring session pong packet, session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (clean_sequence <= session->ServerToClientSeq) {
        LOG(ERROR, "ignoring session pong packet, packet already received: session = ", *session);
        return;
      }

      if (!header.verify(packet, index, Direction::ServerToClient, session->PrivateKey)) {
        LOG(ERROR, "ignoring session pong packet, could not verify header: session = ", *session);
        return;
      }

      session->ServerToClientSeq = clean_sequence;

      recorder.SessionPongTx.add(packet.Len);

      if (!socket.send(session->PrevAddr, packet.Buffer.data(), packet.Len)) {
        LOG(ERROR, "failed to send session pong to ", session->PrevAddr);
      }
    }
  }  // namespace handlers
}  // namespace core
