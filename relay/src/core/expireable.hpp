#pragma once

#include "core/router_info.hpp"
#include "util/logger.hpp"

namespace testing
{
  class _test_core_Expireable_expired_;
}

namespace core
{
  class Expireable
  {
    friend testing::_test_core_Expireable_expired_;

   public:
    virtual ~Expireable() = default;

    static const size_t SIZE_OF = 8;

    // Returns true if the expire timestamp is less than the current unix time given by the backend + time since last sync
    auto expired(const uint64_t backend_timestamp) -> bool;

    // Time to expire in seconds, unix time
    uint64_t expire_timestamp;

   protected:
    Expireable() = default;
  };

  INLINE auto Expireable::expired(const uint64_t backend_timestamp) -> bool
  {
    return this->expire_timestamp < backend_timestamp;
  }
}  // namespace core
