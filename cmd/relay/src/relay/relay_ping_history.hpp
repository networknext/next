#ifndef RELAY_RELAY_PING_HISTORY
#define RELAY_RELAY_PING_HISTORY

#include <cinttypes>

#include "config.hpp"

namespace relay
{
    struct relay_ping_history_entry_t
    {
        uint64_t sequence;
        double time_ping_sent;
        double time_pong_received;
    };

    struct relay_ping_history_t
    {
        uint64_t sequence;
        relay_ping_history_entry_t entries[RELAY_PING_HISTORY_ENTRY_COUNT];
    };

    void relay_ping_history_clear(relay_ping_history_t* history);

    uint64_t relay_ping_history_ping_sent(relay_ping_history_t* history, double time);

    void relay_ping_history_pong_received(relay_ping_history_t* history, uint64_t sequence, double time);
}  // namespace relay

#endif