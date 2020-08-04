#pragma once

#include "constants.hpp"
#include "core/packets/types.hpp"

namespace legacy
{
  namespace v3
  {
    struct BackendRequestFragment
    {
      std::array<uint8_t, FragmentSize> Data = {};
      uint16_t Length = 0;
      bool Received = false;
    };

    struct BackendRequest
    {
      uint64_t ID = 0;
      std::array<BackendRequestFragment, FragmentMax> Fragments = {};
      uint8_t FragmentTotal = 0;
      core::packets::Type Type = core::packets::Type::None;
      uint64_t At = 0;  // when the request was made
    };
  }  // namespace v3
}  // namespace legacy