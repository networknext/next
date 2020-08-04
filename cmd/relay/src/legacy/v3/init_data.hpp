#pragma once

#include "backend_token.hpp"

namespace legacy
{
  namespace v3
  {
    struct InitData
    {
      uint64_t Timestamp;  // in nanosecond resolution, converted by (Timestamp = backend_timestamp / 1000000 - (received - requested) / 2)
      int64_t Requested;
      int64_t Received;
      BackendToken Token;
    }
  }  // namespace v3
}  // namespace legacy