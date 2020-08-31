#pragma once

#include "core/packet.hpp"
#include "core/relay_manager.hpp"
#include "encoding/read.hpp"

using core::Packet;
using core::RelayManager;
using core::packets::RELAY_PING_PACKET_SIZE;

namespace core
{
  namespace handlers
  {
    inline void relay_pong_handler(Packet& packet, RelayManager& manager, bool should_handle)
    {
      if (!should_handle) {
        LOG(INFO, "relay in process of shutting down, ignoring relay pong packet");
        return;
      }

      if (packet.Len != RELAY_PING_PACKET_SIZE) {
        LOG(ERROR, "ignoring relay pong, invalid packet size");
        return;
      }

      uint64_t sequence_number;
      size_t index = crypto::PacketHashLength + 1;
      if (!encoding::ReadUint64(packet.Buffer, index, sequence_number)) {
        LOG(ERROR, "could not read sequence number");
        return;
      }
      manager.processPong(packet.Addr, sequence_number);
    }
  }  // namespace handlers
}  // namespace core
