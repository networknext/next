#pragma once

#include "core/packet.hpp"

namespace core {
  namespace packets {
    class OldRelayPongPacket
    {
     public:
      // type | relay id | sequence
      static const size_t ByteSize = 1 + 8 + 8;

      OldRelayPongPacket(GenericPacket<>& packet);

      GenericPacket<>& Internal;

      auto getID() -> uint64_t;
      auto getSequence() -> uint64_t;
    };

    inline OldRelayPongPacket::OldRelayPongPacket(GenericPacket<>& packet): Internal(packet) {}

    inline auto OldRelayPongPacket::getID() -> uint64_t
    {
      size_t index = 1;
      return encoding::ReadUint64(Internal.Buffer, index);
    }

    inline auto OldRelayPongPacket::getSequence() -> uint64_t
    {
      size_t index = 9;
      return encoding::ReadUint64(Internal.Buffer, index);
    }
  }
}