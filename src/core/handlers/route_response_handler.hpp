#pragma once

#include "core/packets/header.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::Packet;
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
    INLINE void route_response_handler(
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

      if (length != Header::ByteSize) {
        LOG(ERROR, "ignoring route response, header byte count invalid: ", length);
        return;
      }

      Header header;

      {
        size_t i = index;
        if (!header.read(packet, i, Direction::ServerToClient)) {
          LOG(ERROR, "ignoring route response, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring route response, could not find session: session = ", header);
        return;
      }

      if (session->expired(router_info)) {
        LOG(ERROR, "ignoring route response, session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (clean_sequence <= session->ServerToClientSeq) {
        LOG(
         ERROR,
         "ignoring route response, packet already received: session = ",
         *session,
         ", ",
         clean_sequence,
         " <= ",
         session->ServerToClientSeq);
        return;
      }

      if (!header.verify(packet, index, Direction::ServerToClient, session->PrivateKey)) {
        LOG(ERROR, "ignoring route response, header is invalid: session = ", *session);
        return;
      }

      session->ServerToClientSeq = clean_sequence;

      recorder.RouteResponseTx.add(packet.Len);

      if (!socket.send(session->PrevAddr, packet.Buffer.data(), packet.Len)) {
        LOG(ERROR, "failed to forward route response to ", session->PrevAddr);
      }
    }
  }  // namespace handlers
}  // namespace core
