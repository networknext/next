#ifndef RELAY_RELAY__REPLAY_PROTECTION_HPP
#define RELAY_RELAY__REPLAY_PROTECTION_HPP

#include <cinttypes>

#include "config.hpp"

namespace relay
{
    struct relay_replay_protection_t
    {
        uint64_t most_recent_sequence;
        uint64_t received_packet[RELAY_REPLAY_PROTECTION_BUFFER_SIZE];
    };

    void relay_replay_protection_reset(relay_replay_protection_t* replay_protection);

    int relay_replay_protection_already_received(relay_replay_protection_t* replay_protection, uint64_t sequence);

    void relay_replay_protection_advance_sequence(relay_replay_protection_t* replay_protection, uint64_t sequence);

}  // namespace relay
#endif
