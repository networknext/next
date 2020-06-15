#include "includes.h"
#include "testing/test.hpp"

#include "core/replay_protection.hpp"

namespace
{
  const auto MAX_SEQUENCE = RELAY_REPLAY_PROTECTION_BUFFER_SIZE * 4;
}

Test(ReplayProtection_additional_logic_tests)
{
  core::ReplayProtection rp;

  for (int i = 0; i < 2; i++) {
    rp.reset();

    check(rp.mostRecentSeq() == 0);

    // the first time we receive packets, they should not be already received

    for (uint64_t sequence = 0; sequence < MAX_SEQUENCE; sequence++) {
      check(rp.alreadyReceived(sequence) == false);
      rp.advanceSeq(sequence);
    }

    // old packets outside buffer should be considered already received

    check(rp.alreadyReceived(0) == true);

    // packets received a second time should be flagged already received

    for (uint64_t sequence = MAX_SEQUENCE - 10; sequence < MAX_SEQUENCE; sequence++) {
      check(rp.alreadyReceived(sequence) == true);
    }

    // jumping ahead to a much higher sequence should be considered not already received

    check(rp.alreadyReceived(MAX_SEQUENCE + RELAY_REPLAY_PROTECTION_BUFFER_SIZE) == false);

    // old packets should be considered already received

    for (uint64_t sequence = 0; sequence < MAX_SEQUENCE; sequence++) {
      check(rp.alreadyReceived(sequence) == true);
    }
  }
}

Test(legacy_test_replay_protection)
{
  legacy::relay_replay_protection_t replay_protection;

  int i;
  for (i = 0; i < 2; ++i) {
    legacy::relay_replay_protection_reset(&replay_protection);

    check(replay_protection.most_recent_sequence == 0);

    // the first time we receive packets, they should not be already received

    uint64_t sequence;
    for (sequence = 0; sequence < MAX_SEQUENCE; ++sequence) {
      check(legacy::relay_replay_protection_already_received(&replay_protection, sequence) == 0);
      legacy::relay_replay_protection_advance_sequence(&replay_protection, sequence);
    }

    // old packets outside buffer should be considered already received

    check(legacy::relay_replay_protection_already_received(&replay_protection, 0) == 1);

    // packets received a second time should be flagged already received

    for (sequence = MAX_SEQUENCE - 10; sequence < MAX_SEQUENCE; ++sequence) {
      check(legacy::relay_replay_protection_already_received(&replay_protection, sequence) == 1);
    }

    // jumping ahead to a much higher sequence should be considered not already received

    check(legacy::relay_replay_protection_already_received(
           &replay_protection, MAX_SEQUENCE + RELAY_REPLAY_PROTECTION_BUFFER_SIZE) == 0);

    // old packets should be considered already received

    for (sequence = 0; sequence < MAX_SEQUENCE; ++sequence) {
      check(legacy::relay_replay_protection_already_received(&replay_protection, sequence) == 1);
    }
  }
}
