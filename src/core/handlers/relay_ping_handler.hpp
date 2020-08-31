#pragma once

#include "core/packets/types.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"
#include "os/socket.hpp"

using core::Packet;
using core::packets::RELAY_PING_PACKET_SIZE;
using core::packets::Type;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    inline void relay_ping_handler(Packet& packet, ThroughputRecorder& recorder, const Socket& socket, bool should_handle)
    {
      if (!should_handle) {
        LOG(INFO, "relay in process of shutting down, ignoring relay ping packet");
        return;
      }

      if (packet.Len != RELAY_PING_PACKET_SIZE) {
        LOG(ERROR, "ignoring relay ping, invalid packet size");
        return;
      }

      packet.Buffer[crypto::PacketHashLength] = static_cast<uint8_t>(Type::RelayPong);

      crypto::SignNetworkNextPacket(packet.Buffer, packet.Len);

      recorder.InboundPingTx.add(packet.Len);

      if (!socket.send(packet)) {
        LOG(ERROR, "failed to send new pong to ", packet.Addr);
      }
    }
  }  // namespace handlers
}  // namespace core
