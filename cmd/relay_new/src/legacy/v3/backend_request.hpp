#pragma once

#include "constants.hpp"

namespace legacy
{
  namespace v3
  {
    struct BackendRequestFragment
    {
      std::array<uint8_t, FragmentSize> Data;
      uint16_t Length;
      bool Received;
    };

    struct BackendRequest
    {
      uint64_t ID;
      std::array<BackendRequestFragment, FragmentMax> Fragments;
      uint8_t FragmentTotal;
      PacketType Type;
      uint64_t At; // when the request was made
    };
  }  // namespace v3
}  // namespace legacy