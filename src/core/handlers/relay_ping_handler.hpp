#pragma once

#include "core/packet_types.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"
#include "os/socket.hpp"
#include "util/macros.hpp"

using core::Packet;
using core::RELAY_PING_PACKET_SIZE;
using core::Type;
using crypto::PACKET_HASH_LENGTH;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    INLINE void relay_ping_handler(Packet& packet, ThroughputRecorder& recorder, const Socket& socket, bool should_handle)
    {
      if (!should_handle) {
        LOG(INFO, "relay in process of shutting down, ignoring relay ping packet");
        return;
      }

      if (packet.length != RELAY_PING_PACKET_SIZE) {
        LOG(ERROR, "ignoring relay ping, invalid packet size");
        return;
      }

      packet.buffer[PACKET_HASH_LENGTH] = static_cast<uint8_t>(Type::RelayPong);

      crypto::sign_network_next_packet(packet.buffer, packet.length);

      recorder.inbound_ping_tx.add(packet.length);

      if (!socket.send(packet)) {
        LOG(ERROR, "failed to send new pong to ", packet.addr);
      }
    }
  }  // namespace handlers
}  // namespace core
