#ifndef CORE_ROUTER_INFO_HPP
#define CORE_ROUTER_INFO_HPP

#include "util/clock.hpp"

namespace core
{
  class RouterInfo
  {
   public:
    RouterInfo() = default;

    void setTimestamp(int64_t ts);
    auto currentTime() const -> double;

   private:
    mutable std::mutex mLock;
    int64_t mBackendTimestamp = 0;  // in seconds, so the backend and relays have a time sync
    util::Clock mClock;
  };

  [[gnu::always_inline]] inline void RouterInfo::setTimestamp(int64_t ts)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mBackendTimestamp = ts;
    mClock.reset();
  }

  [[gnu::always_inline]] inline auto RouterInfo::currentTime() const -> double
  {
    std::lock_guard<std::mutex> lk(mLock);
    return mBackendTimestamp + mClock.elapsed<util::Second>();
  }
}  // namespace core
#endif