#pragma once

#include "net/address.hpp"
#include "os/platform.hpp"
#include "packet.hpp"
#include "util/channel.hpp"

namespace core
{
  // Legacy support for the old v3 backend
  class V3Backend
  {
   public:
    V3Backend(const net::Address& addr, os::Socket& socket, util::Channel<GenericPacket<>>& channel);
    ~V3Backend() = default;

    auto init() -> bool;

    auto updateCycle(const volatile bool& handle) -> bool;

   private:
    const net::Address& mAddr;
    os::Socket& mSocket;
    util::Channel<GenericPacket<>>& mChannel;

    auto update() -> bool;
  };
}  // namespace core