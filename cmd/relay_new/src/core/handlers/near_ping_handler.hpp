#ifndef CORE_HANDLERS_NEAR_PING_HANDLER_HPP
#define CORE_HANDLERS_NEAR_PING_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class NearPingHandler: public BaseHandler
    {
     public:
      NearPingHandler(GenericPacket<>& packet, const int packetSize, const net::Address& from, const os::Socket& socket, util::ThroughputRecorder& recorder);

      void handle();

     private:
      const net::Address& mFrom;
      const os::Socket& mSocket;
      util::ThroughputRecorder& mRecorder;
    };

    inline NearPingHandler::NearPingHandler(
     GenericPacket<>& packet, const int packetSize, const net::Address& from, const os::Socket& socket, util::ThroughputRecorder& recorder)
     : BaseHandler(packet, packetSize), mFrom(from), mSocket(socket), mRecorder(recorder)
    {}

    inline void NearPingHandler::handle()
    {
      if (mPacketSize != 1 + 8 + 8 + 8 + 8) {
        return;
      }

      mPacket.Buffer[0] = RELAY_NEAR_PONG_PACKET;
      auto length = mPacketSize - 16; // ? why 16
      mRecorder.addToSent(length);
      mSocket.send(mFrom, mPacket.Buffer.data(), length);  // ? why 16?
    }
  }  // namespace handlers
}  // namespace core
#endif