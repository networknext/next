#include "includes.h"
#include "testing/test.hpp"

#include "core/session.hpp"

using core::SessionHasher;

TEST(core_SessionHasher_hash)
{
  SessionHasher hasher;
  hasher.session_id = 0x12345600;
  hasher.session_version = 0xFF;

  CHECK(hasher.hash() == 0x123456FF);
}
