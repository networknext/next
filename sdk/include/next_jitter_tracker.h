/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_JITTER_TRACKER_H
#define NEXT_JITTER_TRACKER_H

#include "next.h"

#include <math.h>

struct next_jitter_tracker_t
{
    NEXT_DECLARE_SENTINEL(0)

    uint64_t last_packet_processed;
    double last_packet_time;
    double last_packet_delta;
    double jitter;

    NEXT_DECLARE_SENTINEL(1)
};

inline void next_jitter_tracker_initialize_sentinels( next_jitter_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_INITIALIZE_SENTINEL( tracker, 0 )
    NEXT_INITIALIZE_SENTINEL( tracker, 1 )
}

inline void next_jitter_tracker_verify_sentinels( next_jitter_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_VERIFY_SENTINEL( tracker, 0 )
    NEXT_VERIFY_SENTINEL( tracker, 1 )
}

inline void next_jitter_tracker_reset( next_jitter_tracker_t * tracker )
{
    next_assert( tracker );

    next_jitter_tracker_initialize_sentinels( tracker );

    tracker->last_packet_processed = 0;
    tracker->last_packet_time = 0.0;
    tracker->last_packet_delta = 0.0;
    tracker->jitter = 0.0;

    next_jitter_tracker_verify_sentinels( tracker );
}

inline void next_jitter_tracker_packet_received( next_jitter_tracker_t * tracker, uint64_t sequence, double time )
{
    next_jitter_tracker_verify_sentinels( tracker );

    if ( sequence == tracker->last_packet_processed + 1 && tracker->last_packet_time > 0.0 )
    {
        const double delta = time - tracker->last_packet_time;
        const double jitter = fabs( delta - tracker->last_packet_delta );
        tracker->last_packet_delta = delta;

        if ( fabs( jitter - tracker->jitter ) > 0.00001 )
        {
            tracker->jitter += ( jitter - tracker->jitter ) * 0.01;
        }
        else
        {
            tracker->jitter = jitter;
        }
    }

    tracker->last_packet_processed = sequence;
    tracker->last_packet_time = time;
}

#endif // #ifndef NEXT_JITTER_TRACKER_H
