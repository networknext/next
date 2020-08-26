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
    auto expired() -> bool;

    /* Returns true if the expire timestamp is less than the argument */
    auto expired(double seconds) -> bool;

    // Time to expire in seconds, unix time
    uint64_t ExpireTimestamp;

   protected:
    Expireable(const RouterInfo& routerInfo); // TODO de-ruby-fy this, pass router info into expired()

   private:
    const RouterInfo& mRouterInfo;

    inline auto timestamp() -> uint64_t;
  };

  inline Expireable::Expireable(const RouterInfo& routerInfo): mRouterInfo(routerInfo) {}

  inline auto Expireable::expired() -> bool
  {
    return this->ExpireTimestamp < timestamp() + 1;
  }

  inline auto Expireable::expired(double seconds) -> bool
  {
    return this->ExpireTimestamp < seconds;
  }

  inline auto Expireable::timestamp() -> uint64_t
  {
    // elapsed time is the amount of seconds since the relay initialized with the backend
    return mRouterInfo.currentTime();
  }
}  // namespace core
#endif