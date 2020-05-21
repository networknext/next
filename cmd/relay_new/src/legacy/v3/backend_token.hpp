#pragma once

#include "net/address.hpp"

namespace legacy
{
  namespace v3
  {
    struct BackendToken
    {
      net::Address Address;
      std::array<uint8_t, 32> HMAC;
    };
  }  // namespace v3
}  // namespace legacy