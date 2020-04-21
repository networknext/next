#pragma once

#include "constants.hpp"

namespace legacy
{
  namespace v3
  {
    struct BackendRequestFragment
    {
      std::array<uint8_t, FragmentSize> data;
      uint16_t length;
      bool received;
    };

    struct BackendRequest
    {
      uint64_t id;
      std::array<BackendRequestFragment, FragmentMax> fragments;
      uint8_t fragment_total;
      uint8_t type;
    };
  }  // namespace v3
}  // namespace legacy