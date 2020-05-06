#pragma once

#include "backend_response.hpp"
#include "core/packet.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/channel.hpp"
#include "util/env.hpp"
#include "util/json.hpp"

namespace legacy
{
  namespace v3
  {
    // Legacy support for the old v3 backend
    class Backend
    {
     public:
      Backend(util::Receiver<core::GenericPacket<>>& receiver, util::Env& env, os::Socket& socket);
      ~Backend() = default;

      auto init() -> bool;
      auto config() -> bool;

      auto updateCycle(const volatile bool& handle) -> bool;

     private:
      util::Receiver<core::GenericPacket<>>& mReceiver;
      const util::Env& mEnv;
      os::Socket& mSocket;

      auto tryInit() -> bool;
      auto update() -> bool;

      auto buildInitJSON(util::JSON& doc) -> bool;
      auto buildConfigJSON(util::JSON& doc) -> bool;
      auto buildUpdateJSON(util::JSON& doc) -> bool;
      auto readResponse(core::GenericPacket<>& packet, BackendRequest& request, BackendResponse& response) -> bool;
    };
  }  // namespace v3
}  // namespace legacy
