/*
    Network Next XDP Relay
*/

#ifndef RELAY_PING_HISTORY_H
#define RELAY_PING_HISTORY_H

#include "relay.h"

struct relay_ping_history_stats_t
{
    float rtt;
    float jitter;
    float packet_loss;
};

struct relay_ping_history_entry_t
{
    uint64_t sequence;
    double time_ping_sent;
    double time_pong_received;
};

struct relay_ping_history_t
{
    uint64_t sequence;
    struct relay_ping_history_entry_t entries[RELAY_PING_HISTORY_SIZE];
};

void relay_ping_history_clear( struct relay_ping_history_t * history );

uint64_t relay_ping_history_ping_sent( struct relay_ping_history_t * history, double time );

void relay_ping_history_pong_received( struct relay_ping_history_t * history, uint64_t sequence, double time );

void relay_ping_history_get_stats( const struct relay_ping_history_t * history, double start, double end, struct relay_ping_history_stats_t * stats, double ping_safety );

#endif // #ifndef RELAY_PING_HISTORY_H
