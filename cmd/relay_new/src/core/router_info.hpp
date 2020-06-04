#ifndef CORE_ROUTER_INFO_HPP
#define CORE_ROUTER_INFO_HPP

#include "util/clock.hpp"

namespace core
{
  class RouterInfo
  {
   public:
    RouterInfo(const util::Clock& clock);

    auto currentTime() const -> double;

    int64_t BackendTimestamp = 0;  // in seconds, so the backend and relays have a time sync

   private:
    const util::Clock& mClock;
  };

  inline RouterInfo::RouterInfo(const util::Clock& clock): mClock(clock) {}

  [[gnu::always_inline]] inline auto RouterInfo::currentTime() const -> double
  {
    return this->BackendTimestamp + mClock.elapsed<util::Second>();
  }
}  // namespace core
#endif