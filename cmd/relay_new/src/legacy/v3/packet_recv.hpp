#pragma once

#include "core/packet.hpp"
#include "backend_request.hpp"
#include "backend_response.hpp"

namespace legacy
{
  namespace v3
  {
    auto packet_recv(
     core::GenericPacket<>& packet, BackendRequest& request, BackendResponse& response, std::vector<uint8_t>& completeResponse)
     -> bool;
  }
}  // namespace legacy