#include "replay_protection.test.hpp"

#include "testing/macros.hpp"
#include "core/replay_protection.hpp"

namespace testing
{
	void TestReplayProtection()
	{
	}
}  // namespace testing

namespace legacy_testing
{
	void test_replay_protection()
	{
		legacy::relay_replay_protection_t replay_protection;

		int i;
		for (i = 0; i < 2; ++i) {
			legacy::relay_replay_protection_reset(&replay_protection);

			check(replay_protection.most_recent_sequence == 0);

			// the first time we receive packets, they should not be already received

#define MAX_SEQUENCE (RELAY_REPLAY_PROTECTION_BUFFER_SIZE * 4)

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
}  // namespace legacy_testing
