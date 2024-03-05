/*
    Network Next XDP Relay
*/

#ifndef RELAY_MESSAGES_H
#define RELAY_MESSAGES_H

#include "relay.h"
#include "relay_set.h"
#include "relay_ping_stats.h"

struct relay_control_message
{
    uint64_t current_timestamp;
    uint8_t current_magic[8];
    uint8_t ping_key[RELAY_PING_KEY_BYTES];
    struct relay_set new_relays;
    struct relay_set delete_relays;
};

struct relay_stats_message
{
    uint64_t pings_sent;
    uint64_t bytes_sent;
    struct relay_ping_stats_t ping_stats;
};

#endif // #ifndef RELAY_MESSAGES_H
