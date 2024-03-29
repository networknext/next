/*
    Network Next XDP Relay
*/

#include "relay_ping_history.h"
#include <float.h>

void relay_ping_history_clear( struct relay_ping_history_t * history )
{
    assert( history );
    history->sequence = 0;
    for ( int i = 0; i < RELAY_PING_HISTORY_SIZE; ++i )
    {
        history->entries[i].sequence = 0xFFFFFFFFFFFFFFFFULL;
        history->entries[i].time_ping_sent = -1.0;
        history->entries[i].time_pong_received = -1.0;
    }
}

uint64_t relay_ping_history_ping_sent( struct relay_ping_history_t * history, double time )
{
    assert( history );
    const int index = history->sequence % RELAY_PING_HISTORY_SIZE;
    struct relay_ping_history_entry_t * entry = &history->entries[index];
    entry->sequence = history->sequence;
    entry->time_ping_sent = time;
    entry->time_pong_received = -1.0;
    history->sequence++;
    return entry->sequence;
}

void relay_ping_history_pong_received( struct relay_ping_history_t * history, uint64_t sequence, double time )
{
    const int index = sequence % RELAY_PING_HISTORY_SIZE;
    struct relay_ping_history_entry_t * entry = &history->entries[index];
    if ( entry->sequence == sequence )
    {
        entry->time_pong_received = time;
    }
}

void relay_ping_history_get_stats( const struct relay_ping_history_t * history, double start, double end, struct relay_ping_history_stats_t * stats, double ping_safety )
{
    assert( history );
    assert( stats );
    assert( start < end );

    stats->rtt = 0.0f;
    stats->jitter = 0.0f;
    stats->packet_loss = 100.0f;

    // calculate packet loss

    int num_pings_sent = 0;
    int num_pongs_received = 0;

    for ( int i = 0; i < RELAY_PING_HISTORY_SIZE; i++ )
    {
        const struct relay_ping_history_entry_t * entry = &history->entries[i];

        if ( entry->time_ping_sent >= start && entry->time_ping_sent <= end - ping_safety )
        {
            num_pings_sent++;

            if ( entry->time_pong_received >= entry->time_ping_sent )
                num_pongs_received++;
        }
    }

    if ( num_pings_sent > 0 )
    {
        stats->packet_loss = (float) ( 100.0 * ( 1.0 - ( (double) num_pongs_received ) / (double) num_pings_sent ) );
    }

    // calculate min RTT

    double min_rtt = FLT_MAX;

    for ( int i = 0; i < RELAY_PING_HISTORY_SIZE; i++ )
    {
        const struct relay_ping_history_entry_t * entry = &history->entries[i];

        if ( entry->time_ping_sent >= start && entry->time_ping_sent <= end )
        {
            if ( entry->time_pong_received > entry->time_ping_sent )
            {
                double rtt = ( entry->time_pong_received - entry->time_ping_sent );
                if ( rtt < min_rtt )
                {
                    min_rtt = rtt;
                }
            }
        }
    }

    assert( min_rtt >= 0.0 );

    stats->rtt = 1000.0f * (float) min_rtt;

    // calculate jitter

    int num_jitter_samples = 0;

    double jitter_sum = 0.0;

    for ( int i = 0; i < RELAY_PING_HISTORY_SIZE; i++ )
    {
        const struct relay_ping_history_entry_t * entry = &history->entries[i];

        if ( entry->time_ping_sent >= start && entry->time_ping_sent <= end )
        {
            if ( entry->time_pong_received > entry->time_ping_sent )
            {
                // pong received
                double rtt = ( entry->time_pong_received - entry->time_ping_sent );
                double jitter = rtt - min_rtt;
                jitter_sum += jitter;
                num_jitter_samples++;
            }
        }
    }

    if ( num_jitter_samples > 0 )
    {
        stats->jitter = 1000.0f * (float) ( jitter_sum / num_jitter_samples );
    }
}
