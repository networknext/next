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

    // Returns true if the expire timestamp is less than the current unix time given by the backend
    auto expired(const RouterInfo& router_info) -> bool;

    // Time to expire in seconds, unix time
    uint64_t expire_timestamp;

   protected:
    Expireable() = default;
  };

  INLINE auto Expireable::expired(const RouterInfo& router_info) -> bool
  {
    return this->expire_timestamp < router_info.current_time<uint64_t>();
  }
}  // namespace core
