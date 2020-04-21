#pragma once

#include "net/address.hpp"
#include "core/packet.hpp"
#include "backend_token.hpp"

namespace legacy
{
  namespace v3
  {
    auto build_udp_fragment(
     uint8_t packet_type,
     const BackendToken& master_token,
     uint64_t id,
     uint8_t fragment_index,
     uint8_t fragment_total,
     const core::GenericPacket<>& packet,
     core::Packet<std::vector<uint8_t>>& out) -> bool;
  }  // namespace v3
}  // namespace legacy