#pragma once

#include "core/packet.hpp"
#include "constants.hpp"

namespace legacy
{
  namespace v3
  {
    struct BackendResponse
    {
      PacketType Type;
      std::array<uint8_t, 64> Signature;
      uint64_t GUID;
      uint8_t FragIndex;
      uint8_t FragCount;
      uint16_t StatusCode;
      core::Packet<std::vector<uint8_t>> Data;
      uint64_t At;
    };
  }  // namespace v3
}  // namespace legacy