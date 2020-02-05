#include "testing/test.hpp"

#include "core/ping_history.hpp"

Test(PingHistory_clear) {
    core::PingHistory ph;
    for (const auto& entry : ph.Entries) {
        check(entry.Seq == INVALID_SEQUENCE_NUMBER);
        check(entry.TimePingSent == -1.0);
        check(entry.TimePongRecieved == -1.0);
    }
}