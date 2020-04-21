#pragma once

#include "core/packet.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/channel.hpp"
#include "util/json.hpp"

namespace legacy
{
  namespace v3
  {
    // Legacy support for the old v3 backend
    class Backend
    {
     public:
      Backend(const net::Address& addr, os::Socket& socket);
      ~Backend() = default;

      auto init() -> bool;

      auto updateCycle(const volatile bool& handle) -> bool;

     private:
      const net::Address& mAddr;
      os::Socket& mSocket;

      auto update() -> bool;

      auto buildInitJSON(util::JSON& doc) -> bool;
      auto buildUpdateJSON(util::JSON& doc) -> bool;
    };
  }  // namespace v3
}  // namespace legacy
