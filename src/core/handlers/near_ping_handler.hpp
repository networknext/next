#pragma once

#include "core/packet_types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::Packet;
using core::Type;
using crypto::PACKET_HASH_LENGTH;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void near_ping_handler(Packet& packet, ThroughputRecorder& recorder, const Socket& socket, bool is_signed)
    {
      size_t length = packet.length;

      if (is_signed) {
        length = packet.length - PACKET_HASH_LENGTH;
      }

      if (length != 1 + 8 + 8 + 8 + 8) {
        LOG(ERROR, "ignoring near ping packet, length invalid: ", length);
        return;
      }

      length = packet.length - 16;

      if (is_signed) {
        packet.buffer[PACKET_HASH_LENGTH] = static_cast<uint8_t>(Type::NearPong);
        crypto::sign_network_next_packet(packet.buffer, length);
      } else {
        packet.buffer[0] = static_cast<uint8_t>(Type::NearPong);
      }

      recorder.near_ping_tx.add(length);

      if (!socket.send(packet.addr, packet.buffer.data(), length)) {
        LOG(ERROR, "failed to send near pong to ", packet.addr);
      }
    }
  }  // namespace handlers
}  // namespace core
