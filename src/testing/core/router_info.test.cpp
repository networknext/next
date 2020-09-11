#include "includes.h"
#include "testing/test.hpp"

#include "core/router_info.hpp"

using namespace std::chrono_literals;

using core::RouterInfo;
using util::Second;

TEST(core_RouterInfo_set_timestamp) {
  RouterInfo router_info;

  CHECK(router_info.clock.elapsed<Second>() < 1.0);
  CHECK(router_info.backend_timestamp == 0);

  std::this_thread::sleep_for(1s);

  CHECK(router_info.clock.elapsed<Second>() >= 1.0);
  CHECK(router_info.backend_timestamp == 0);

  router_info.set_timestamp(100);

  CHECK(router_info.clock.elapsed<Second>() < 1.0);
  CHECK(router_info.backend_timestamp == 100);
}

TEST(core_RouterInfo_current_time) {
  RouterInfo router_info;

  router_info.backend_timestamp = 100;
  std::this_thread::sleep_for(1s);

  CHECK(router_info.current_time<double>() >= 101.0);
}
