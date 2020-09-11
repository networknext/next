#pragma once

#include "util/clock.hpp"
#include "util/macros.hpp"

using util::Clock;
using util::Second;

namespace core
{
  class RouterInfo
  {
   public:
    RouterInfo() = default;

    void set_timestamp(int64_t ts);

    template <typename R>
    auto current_time() const -> R;

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

  template <typename R>
  INLINE auto RouterInfo::current_time() const -> R
  {
    std::lock_guard<std::mutex> lk(mutex);
    return static_cast<R>(this->backend_timestamp + clock.elapsed<Second>());
  }
}  // namespace core
