#pragma once

#include "base_handler.hpp"
#include "core/packet.hpp"
#include "util/channel.hpp"

namespace core
{
  namespace handlers
  {
    class V3InitHandler: public BaseHandler
    {
     public:
      V3InitHandler(GenericPacket<>& packet, const int packetSize, util::Channel<GenericPacket<>>& channel);

      void handle();

     private:
      util::Channel<GenericPacket<>>& mChannel;
    };

    V3InitHandler::V3InitHandler(GenericPacket<>& packet, const int packetSize, util::Channel<GenericPacket<>>& channel)
     : BaseHandler(packet, packetSize), mChannel(channel)
    {}

    void V3InitHandler::handle() {
      mChannel.send(mPacket);
    }
  }  // namespace handlers
}  // namespace core
