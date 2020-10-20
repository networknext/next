#pragma once

#include "core/packet_header.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::RouterInfo;
using core::SessionMap;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void route_response_handler_sdk4(
     Packet& packet, SessionMap& session_map, ThroughputRecorder& recorder, const RouterInfo& router_info, const Socket& socket)
    {
      size_t index = 0;
      size_t length = packet.length;

      if (length != PacketHeader::SIZE_OF_SIGNED) {
        LOG(ERROR, "ignoring route response, header byte count invalid: ", length);
        return;
      }

      PacketHeader header;

      {
        size_t i = index;
        if (!header.read(packet, i, PacketDirection::ServerToClient)) {
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

      if (clean_sequence <= session->server_to_client_sequence) {
        LOG(
         ERROR,
         "ignoring route response, packet already received: session = ",
         *session,
         ", ",
         clean_sequence,
         " <= ",
         session->server_to_client_sequence);
        return;
      }

      if (!header.verify(packet, index, PacketDirection::ServerToClient, session->private_key)) {
        LOG(ERROR, "ignoring route response, header is invalid: session = ", *session);
        return;
      }

      session->server_to_client_sequence = clean_sequence;

      recorder.route_response_tx.add(packet.length);

      if (!socket.send(session->prev_addr, packet.buffer.data(), packet.length)) {
        LOG(ERROR, "failed to forward route response to ", session->prev_addr);
      }
    }
  }  // namespace handlers
}  // namespace core
