#pragma once

#include "os/platform.hpp"
#include "net/address.hpp"
#include "backend_token.hpp"
#include "backend_request.hpp"

namespace legacy
{
  namespace v3
  {
    auto packet_send(
     const os::Socket& socket,
     const net::Address& master_address,
     const BackendToken& master_token,
     uint8_t packet_type,
     const BackendRequest& request,
     core::GenericPacket<>& packet) -> bool;
  }
}  // namespace legacy