#pragma once

#include "core/packet_header.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::PacketDirection;
using core::PacketHeaderV4;
using core::RouterInfo;
using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void continue_response_handler_sdk4(
     Packet& packet, SessionMap& session_map, ThroughputRecorder& recorder, const RouterInfo& router_info, const Socket& socket)
    {
      size_t index = 0;
      size_t length = packet.length;

      if (length != PacketHeaderV4::SIZE_OF_SIGNED) {
        LOG(ERROR, "dropping continue response packet, invalid size: ", length);
        return;
      }

      PacketHeaderV4 header;
      {
        size_t i = index;
        if (!header.read(packet, i, PacketDirection::ServerToClient)) {
          LOG(ERROR, "ignoring continue response, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring continue response, could not find session: session = ", header);
        return;
      }

      if (session->expired(router_info)) {
        LOG(INFO, "ignoring continue response, session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (clean_sequence <= session->server_to_client_sequence) {
        LOG(
         ERROR,
         "ignoring continue response, packet already received: session = ",
         *session,
         ", ",
         clean_sequence,
         " <= ",
         session->server_to_client_sequence);
        return;
      }

      {
        size_t i = index;
        if (!header.verify(packet, i, PacketDirection::ServerToClient, session->private_key)) {
          LOG(ERROR, "ignoring continue response, could not verify header: session = ", *session);
          return;
        }
      }

      session->server_to_client_sequence = clean_sequence;

      recorder.continue_response_tx.add(packet.length);

      if (!socket.send(session->prev_addr, packet.buffer.data(), packet.length)) {
        LOG(ERROR, "failed to forward continue response to ", session->prev_addr);
      }
    }
  }  // namespace handlers
}  // namespace core
