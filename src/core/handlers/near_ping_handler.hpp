#pragma once

#include "core/packets/types.hpp"
#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "os/socket.hpp"

using core::packets::Type;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    inline void near_ping_handler(GenericPacket<>& packet, ThroughputRecorder& recorder, const Socket& socket, bool is_signed)
    {
      size_t length = packet.Len;

      if (is_signed) {
        length = packet.Len - crypto::PacketHashLength;
      }

      if (length != 1 + 8 + 8 + 8 + 8) {
        LOG(ERROR, "ignoring near ping packet, length invalid: ", length);
        return;
      }

      length = packet.Len - 16;

      if (is_signed) {
        packet.Buffer[crypto::PacketHashLength] = static_cast<uint8_t>(Type::NearPong);
        crypto::SignNetworkNextPacket(packet.Buffer, length);
      } else {
        packet.Buffer[0] = static_cast<uint8_t>(Type::NearPong);
      }

      recorder.NearPingTx.add(length);

      if (!socket.send(packet.Addr, packet.Buffer.data(), length)) {
        LOG(ERROR, "failed to send near pong to ", packet.Addr);
      }
    }
  }  // namespace handlers
}  // namespace core
