/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_OUT_OF_ORDER_TRACKER_H
#define NEXT_OUT_OF_ORDER_TRACKER_H

struct next_out_of_order_tracker_t
{
    NEXT_DECLARE_SENTINEL(0)

    uint64_t last_packet_processed;
    uint64_t num_out_of_order_packets;

    NEXT_DECLARE_SENTINEL(1)
};

inline void next_out_of_order_tracker_initialize_sentinels( next_out_of_order_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_INITIALIZE_SENTINEL( tracker, 0 )
    NEXT_INITIALIZE_SENTINEL( tracker, 1 )
}

inline void next_out_of_order_tracker_verify_sentinels( next_out_of_order_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_VERIFY_SENTINEL( tracker, 0 )
    NEXT_VERIFY_SENTINEL( tracker, 1 )
}

inline void next_out_of_order_tracker_reset( next_out_of_order_tracker_t * tracker )
{
    next_assert( tracker );

    next_out_of_order_tracker_initialize_sentinels( tracker );

    tracker->last_packet_processed = 0;
    tracker->num_out_of_order_packets = 0;

    next_out_of_order_tracker_verify_sentinels( tracker );
}

inline void next_out_of_order_tracker_packet_received( next_out_of_order_tracker_t * tracker, uint64_t sequence )
{
    next_out_of_order_tracker_verify_sentinels( tracker );

    if ( sequence < tracker->last_packet_processed )
    {
        tracker->num_out_of_order_packets++;
        return;
    }

    tracker->last_packet_processed = sequence;
}

#endif // #ifndef NEXT_OUT_OF_ORDER_TRACKER_H
