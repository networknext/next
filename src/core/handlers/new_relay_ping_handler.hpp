#pragma once

#include "base_handler.hpp"
#include "core/packets/new_relay_ping_packet.hpp"
#include "core/packets/types.hpp"
#include "crypto/hash.hpp"
#include "encoding/read.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class NewRelayPingHandler: public BaseHandler
    {
     public:
      NewRelayPingHandler(GenericPacket<>& packet, util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket);

     private:
      util::ThroughputRecorder& mRecorder;
    };

    inline NewRelayPingHandler::NewRelayPingHandler(
     GenericPacket<>& packet, util::ThroughputRecorder& recorder)
     : BaseHandler(packet), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void NewRelayPingHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket)
    {
      (void)buff;
      (void)socket;

      mPacket.Buffer[crypto::PacketHashLength] = static_cast<uint8_t>(packets::Type::NewRelayPong);

      crypto::SignNetworkNextPacket(mPacket.Buffer, mPacket.Len);

      mRecorder.InboundPingTx.add(mPacket.Len);

#ifdef RELAY_MULTISEND
      buff.push(mPacket);
#else
      if (!socket.send(mPacket)) {
        Log("failed to send new pong to ", mPacket.Addr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core