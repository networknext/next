#include "includes.h"
#include "testing/test.hpp"

#include "core/session_map.hpp"

using core::RouterInfo;
using core::Session;
using core::SessionMap;

TEST(core_SessionMap_set_and_get)
{
  SessionMap map;

  auto session_hash = 123456;
  auto session = std::make_shared<Session>();

  map.set(session_hash, session);

  CHECK(map.get(session_hash) == session);
  CHECK(map.get(~session_hash) == nullptr);
}

TEST(core_SessionMap_erase)
{
  SessionMap map;

  auto session_hash = 123456;
  auto session = std::make_shared<Session>();

  map.set(session_hash, session);

  CHECK(map.get(session_hash) == session);
  CHECK(map.erase(session_hash));
  CHECK(!map.erase(session_hash));
  CHECK(map.get(session_hash) == nullptr);
}

TEST(core_SessionMap_size_and_purge)
{
  SessionMap map;
  RouterInfo info;

  info.set_timestamp(100);

  for (int i = 0; i < 100; i++) {
    auto session = std::make_shared<Session>();
    session->expire_timestamp = (i & 1) ? 50 : 150;
    map.set(i, session);
  }

  CHECK(map.size() == 100);

  map.purge(info);

  CHECK(map.size() == 50);
}
