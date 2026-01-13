/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_BANDWIDTH_LIMITER_H
#define NEXT_BANDWIDTH_LIMITER_H

#include <stdio.h>

inline int next_wire_packet_bits( int payload_bytes )
{
    return ( NEXT_IPV4_HEADER_BYTES + NEXT_UDP_HEADER_BYTES + 18 + NEXT_HEADER_BYTES + payload_bytes + 2 ) * 8;
}

struct next_bandwidth_limiter_t
{
    uint64_t bits_sent;
    double last_check_time;
    double average_kbps;
};

inline void next_bandwidth_limiter_reset( next_bandwidth_limiter_t * bandwidth_limiter )
{
    next_assert( bandwidth_limiter );
    bandwidth_limiter->last_check_time = -100.0;
    bandwidth_limiter->bits_sent = 0;
    bandwidth_limiter->average_kbps = 0.0;
}

inline void next_bandwidth_limiter_add_sample( next_bandwidth_limiter_t * bandwidth_limiter, double kbps )
{
    if ( bandwidth_limiter->average_kbps == 0.0 && kbps != 0.0 )
    {
        bandwidth_limiter->average_kbps = kbps;
        return;
    }

    if ( bandwidth_limiter->average_kbps != 0.0 && kbps == 0.0 )
    {
        bandwidth_limiter->average_kbps = 0.0;
        return;
    }

    const double delta = kbps - bandwidth_limiter->average_kbps;

    if ( delta < 0.000001f )
    {
        bandwidth_limiter->average_kbps = kbps;
        return;
    }

    bandwidth_limiter->average_kbps += delta * 0.1f;
}

inline bool next_bandwidth_limiter_add_packet( next_bandwidth_limiter_t * bandwidth_limiter, double current_time, uint32_t kbps_allowed, uint32_t packet_bits )
{
    next_assert( bandwidth_limiter );

    const bool invalid = bandwidth_limiter->last_check_time < 0.0;

    const bool new_period = ( current_time - bandwidth_limiter->last_check_time ) >= NEXT_BANDWIDTH_LIMITER_INTERVAL - 0.00001;

    if ( invalid || new_period )
    {
        if ( new_period )
        {
            const double kbps = bandwidth_limiter->bits_sent / ( current_time - bandwidth_limiter->last_check_time) / 1000.0;
            next_bandwidth_limiter_add_sample( bandwidth_limiter, kbps );
        }
        bandwidth_limiter->bits_sent = 0;
        bandwidth_limiter->last_check_time = current_time;
    }

    bandwidth_limiter->bits_sent += packet_bits;

    return bandwidth_limiter->bits_sent > uint64_t(kbps_allowed) * 1000 * NEXT_BANDWIDTH_LIMITER_INTERVAL;
}

inline double next_bandwidth_limiter_usage_kbps( next_bandwidth_limiter_t * bandwidth_limiter )
{
    next_assert( bandwidth_limiter );
    return bandwidth_limiter->average_kbps;
}

#endif // #ifndef NEXT_BANDWIDTH_LIMITER_H
