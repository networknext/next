#pragma once

#include "core/packet_types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::Packet;
using core::PacketType;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void near_ping_handler_sdk4(Packet& packet, ThroughputRecorder& recorder, const Socket& socket)
    {
      size_t length = packet.length;

      if (length != 1 + 8 + 8 + 8 + 8) {
        LOG(ERROR, "ignoring near ping packet, length invalid: ", length);
        return;
      }

      length = packet.length - 16;

      packet.buffer[0] = static_cast<uint8_t>(PacketType::NearPong4);

      recorder.near_ping_tx.add(length);

      if (!socket.send(packet.addr, packet.buffer.data(), length)) {
        LOG(ERROR, "failed to send near pong to ", packet.addr);
      }
    }
  }  // namespace handlers
}  // namespace core
