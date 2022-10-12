#include "includes.h"
#include "testing/test.hpp"

#include "core/session_map.hpp"

using core::RouterInfo;
using core::Session;
using core::SessionMap;

TEST(core_SessionMap_set_and_get)
{
  SessionMap map;

  uint64_t session_hash = 123456;
  size_t env_up = 123;
  size_t env_down = 456;
  auto session = std::make_shared<Session>();
  session->kbps_up = env_up;
  session->kbps_down = env_down;

  map.set(session_hash, session);

  CHECK(map.get(session_hash) == session);
  CHECK(map.get(~session_hash) == nullptr);
  CHECK(map.envelope_up_total() == env_up);
  CHECK(map.envelope_down_total() == env_down);
}

TEST(core_SessionMap_erase)
{
  SessionMap map;

  uint64_t session_hash = 123456;
  size_t env_up = 123;
  size_t env_down = 456;
  auto session = std::make_shared<Session>();
  session->kbps_up = env_up;
  session->kbps_down = env_down;

  map.set(session_hash, session);

  CHECK(map.get(session_hash) == session);

  CHECK(map.erase(session_hash));
  // erase decrements the envelope counts
  CHECK(map.envelope_up_total() == 0);
  CHECK(map.envelope_down_total() == 0);

  CHECK(!map.erase(session_hash));
  CHECK(map.envelope_up_total() == 0);
  CHECK(map.envelope_down_total() == 0);

  CHECK(map.get(session_hash) == nullptr);
}

TEST(core_SessionMap_size_and_purge)
{
  SessionMap map;
  RouterInfo info;

  info.set_timestamp(100);

  size_t total_up = 0;
  size_t total_down = 0;
  size_t expected_decrement_up = 0;
  size_t expected_decrement_down = 0;
  for (int i = 0; i < 100; i++) {
    auto session = std::make_shared<Session>();
    session->kbps_up = i;
    session->kbps_down = i * 2;

    total_up += i;
    total_down += i * 2;

    if (i & 1) {
      session->expire_timestamp = 50;
      expected_decrement_up += i;
      expected_decrement_down += i * 2;
    } else {
      session->expire_timestamp = 150;
    }

    map.set(i, session);
  }

  CHECK(map.size() == 100);
  CHECK(map.envelope_up_total() == total_up);
  CHECK(map.envelope_down_total() == total_down);

  map.purge(info.current_time<uint64_t>());

  CHECK(map.size() == 50);
  CHECK(map.envelope_up_total() == total_up - expected_decrement_up);
  CHECK(map.envelope_down_total() == total_down - expected_decrement_down);
}
