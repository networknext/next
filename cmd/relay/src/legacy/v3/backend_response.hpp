#pragma once

#include "core/packet.hpp"
#include "constants.hpp"
#include "core/packets/types.hpp"

namespace legacy
{
  namespace v3
  {
    struct BackendResponse
    {
      core::packets::Type Type = core::packets::Type::None;
      std::array<uint8_t, 64> Signature = {};
      uint64_t GUID = 0;
      uint8_t FragIndex = 0;
      uint8_t FragCount = 0;
      uint16_t StatusCode = 0;
      core::Packet<std::vector<uint8_t>> Data = {};
      uint64_t At = 0;
    };
  }  // namespace v3
}  // namespace legacy