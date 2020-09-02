#pragma once

#include "util/clock.hpp"
#include "util/macros.hpp"

using util::Clock;

namespace core
{
  class RouterInfo
  {
   public:
    RouterInfo() = default;

    void set_timestamp(int64_t ts);
    auto current_time() const -> double;

   private:
    mutable std::mutex mutex;
    int64_t backend_timestamp = 0;  // in seconds, so the backend and relays have a time sync
    Clock clock;
  };

  INLINE void RouterInfo::set_timestamp(int64_t ts)
  {
    std::lock_guard<std::mutex> lk(mutex);
    this->backend_timestamp = ts;
    clock.reset();
  }

  INLINE auto RouterInfo::current_time() const -> double
  {
    std::lock_guard<std::mutex> lk(mutex);
    return this->backend_timestamp + clock.elapsed<util::Second>();
  }
}  // namespace core
