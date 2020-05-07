#pragma once

#include "backend_request.hpp"
#include "backend_response.hpp"
#include "backend_token.hpp"
#include "core/packet.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "util/channel.hpp"
#include "util/env.hpp"
#include "util/json.hpp"
#include "util/clock.hpp"

namespace legacy
{
  namespace v3
  {
    // Legacy support for the old v3 backend
    class Backend
    {
     public:
      Backend(util::Receiver<core::GenericPacket<>>& receiver, util::Env& env, os::Socket& socket, util::Clock& relayClock);
      ~Backend() = default;

      auto init() -> bool;
      auto config() -> bool;

      auto updateCycle(const volatile bool& handle) -> bool;

     private:
      util::Receiver<core::GenericPacket<>>& mReceiver;
      const util::Env& mEnv;
      os::Socket& mSocket;
      const util::Clock& mClock;
      BackendToken mToken;

      auto tryInit() -> bool;
      auto update() -> bool;

      auto buildConfigJSON(util::JSON& doc) -> bool;
      auto buildUpdateJSON(util::JSON& doc) -> bool;
      auto readResponse(core::GenericPacket<>& packet, BackendRequest& request, BackendResponse& response, std::vector<uint8_t>& completeResponse) -> bool;
      auto buildCompleteResponse(std::vector<uint8_t>& completeBuffer, util::JSON& doc) -> bool;
    };
  }  // namespace v3
}  // namespace legacy
