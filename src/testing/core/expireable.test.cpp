#include "includes.h"
#include "testing/test.hpp"

#include "core/expireable.hpp"
#include "util/clock.hpp"

using namespace std::chrono_literals;

using core::Expireable;
using core::RouterInfo;
using util::Clock;
using util::Second;

TEST(core_Expireable_expired)
{
  Expireable expireable;
  RouterInfo router_info;
  Clock clock;

  router_info.set_timestamp(clock.unix_time<Second>());
  expireable.expire_timestamp = router_info.current_time<uint64_t>() + 10;
  std::this_thread::sleep_for(10s);
  CHECK(!expireable.expired(router_info));
  std::this_thread::sleep_for(1s);
  CHECK(expireable.expired(router_info));
}
