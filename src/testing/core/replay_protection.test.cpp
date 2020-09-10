#include "includes.h"
#include "testing/test.hpp"

#include "core/replay_protection.hpp"

namespace
{
  const auto MAX_SEQUENCE = RELAY_REPLAY_PROTECTION_BUFFER_SIZE * 4;
}

TEST(core_ReplayProtection_additional_logic_tests)
{
  core::ReplayProtection rp;

  for (int i = 0; i < 2; i++) {
    rp.reset();

    CHECK(rp.most_recent_sequence == 0);

    // the first time we receive packets, they should not be already received

    for (uint64_t sequence = 0; sequence < MAX_SEQUENCE; sequence++) {
      CHECK(rp.is_already_received(sequence) == false);
      rp.advance_sequence_to(sequence);
    }

    // old packets outside buffer should be considered already received

    CHECK(rp.is_already_received(0) == true);

    // packets received a second time should be flagged already received

    for (uint64_t sequence = MAX_SEQUENCE - 10; sequence < MAX_SEQUENCE; sequence++) {
      CHECK(rp.is_already_received(sequence) == true);
    }

    // jumping ahead to a much higher sequence should be considered not already received

    CHECK(rp.is_already_received(MAX_SEQUENCE + RELAY_REPLAY_PROTECTION_BUFFER_SIZE) == false);

    // old packets should be considered already received

    for (uint64_t sequence = 0; sequence < MAX_SEQUENCE; sequence++) {
      CHECK(rp.is_already_received(sequence) == true);
    }
  }
}
