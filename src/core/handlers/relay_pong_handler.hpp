#pragma once

#include "core/packet.hpp"
#include "core/packet_types.hpp"
#include "core/relay_manager.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"

using core::Packet;
using core::RELAY_PING_PACKET_SIZE;
using core::RelayManager;
using crypto::PACKET_HASH_LENGTH;

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

      if (packet.length != RELAY_PING_PACKET_SIZE) {
        LOG(ERROR, "ignoring relay pong, invalid packet size");
        return;
      }

      uint64_t sequence_number;
      size_t index = PACKET_HASH_LENGTH + 1;
      if (!encoding::read_uint64(packet.buffer, index, sequence_number)) {
        LOG(ERROR, "could not read sequence number");
        return;
      }
      manager.process_pong(packet.addr, sequence_number);
    }
  }  // namespace handlers
}  // namespace core
