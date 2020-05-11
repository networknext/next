#pragma once

#include "backend_request.hpp"
#include "backend_token.hpp"
#include "constants.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"

namespace legacy
{
  namespace v3
  {
    auto packet_send(
     const os::Socket& socket,
     const BackendToken& master_token,
     core::GenericPacket<>& packet,
     BackendRequest& request) -> bool;
  }
}  // namespace legacy