/*
    Network Next XDP Relay
*/

#ifndef RELAY_PING_STATS_H
#define RELAY_PING_STATS_H

#include "relay.h"

struct relay_ping_stats_t
{
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    float relay_rtt[MAX_RELAYS];
    float relay_jitter[MAX_RELAYS];
    float relay_packet_loss[MAX_RELAYS];
};

#endif // #ifndef RELAY_PING_STATS_H
