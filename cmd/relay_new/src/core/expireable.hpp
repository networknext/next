#ifndef CORE_EXPIREABLE_HPP
#define CORE_EXPIREABLE_HPP

#include "core/router_info.hpp"
#include "util/clock.hpp"

namespace core
{
  class Expireable
  {
   public:
    virtual ~Expireable() = default;
    auto expired() -> bool;

    uint64_t ExpireTimestamp;

   protected:
    Expireable(const util::Clock& relayClock, const core::RouterInfo& routerInfo);

   private:
    const util::Clock& mRelayClock;
    const core::RouterInfo& mRouterInfo;

    inline auto timestamp() -> uint64_t;
  };

  inline Expireable::Expireable(const util::Clock& relayClock, const core::RouterInfo& routerInfo)
   : mRelayClock(relayClock), mRouterInfo(routerInfo)
  {}

  inline auto Expireable::expired() -> bool
  {
    return this->ExpireTimestamp < timestamp();
  }

  inline auto Expireable::timestamp() -> uint64_t
  {
    auto seconds_since_initialize = mRelayClock.elapsed<util::Second>();
    return mRouterInfo.InitalizeTimeInSeconds + seconds_since_initialize;
  }
}  // namespace core
#endif