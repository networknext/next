#ifndef CORE_EXPIREABLE_HPP
#define CORE_EXPIREABLE_HPP

#include "core/router_info.hpp"
#include "util/clock.hpp"
#include "util/logger.hpp"

namespace core
{
  class Expireable
  {
   public:
    virtual ~Expireable() = default;

    /* Returns true if the expire timestamp is less than the current unix time */
    auto expired(const RouterInfo& router_info) -> bool;

    /* Returns true if the expire timestamp is less than the number of specified seconds */
    auto expired(double seconds) -> bool;

    // Time to expire in seconds, unix time
    uint64_t ExpireTimestamp;

   protected:
    Expireable() = default;

   private:
  };

  inline auto Expireable::expired(const RouterInfo& router_info) -> bool
  {
    return this->ExpireTimestamp < router_info.currentTime() + 1;
  }

  inline auto Expireable::expired(double seconds) -> bool
  {
    return this->ExpireTimestamp < seconds;
  }
}  // namespace core
#endif