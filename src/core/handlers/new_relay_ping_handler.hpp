#pragma once

#include "core/packets/relay_ping_packet.hpp"
#include "core/packets/types.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"
#include "os/socket.hpp"

using core::packets::Type;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  namespace handlers
  {
    inline void relay_ping_handler(GenericPacket<>& packet, util::ThroughputRecorder& recorder, const os::Socket& socket)
    {
      packet.Buffer[crypto::PacketHashLength] = static_cast<uint8_t>(Type::RelayPong);

      crypto::SignNetworkNextPacket(packet.Buffer, packet.Len);

      recorder.InboundPingTx.add(packet.Len);

      if (!socket.send(packet)) {
        LOG(ERROR, "failed to send new pong to ", packet.Addr);
      }
    }
  }  // namespace handlers
}  // namespace core
