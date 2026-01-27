/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_VALUE_TRACKER_H
#define NEXT_VALUE_TRACKER_H

struct next_value_tracker_t
{
    NEXT_DECLARE_SENTINEL(0)

    int index;
    int num_samples;
    float minimum;
    float maximum;

    NEXT_DECLARE_SENTINEL(1)

    float values[NEXT_VALUE_TRACKER_HISTORY];

    NEXT_DECLARE_SENTINEL(2)
};

inline void next_value_tracker_initialize_sentinels( next_value_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_INITIALIZE_SENTINEL( tracker, 0 )
    NEXT_INITIALIZE_SENTINEL( tracker, 1 )
    NEXT_INITIALIZE_SENTINEL( tracker, 2 )
}

inline void next_value_tracker_verify_sentinels( next_value_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_VERIFY_SENTINEL( tracker, 0 )
    NEXT_VERIFY_SENTINEL( tracker, 1 )
    NEXT_VERIFY_SENTINEL( tracker, 2 )
}

inline void next_value_tracker_reset( next_value_tracker_t * tracker )
{
    next_assert( tracker );

    memset( tracker, 0, sizeof(next_value_tracker_t) );

    next_value_tracker_initialize_sentinels( tracker );

    tracker->maximum = -FLT_MAX;
    tracker->minimum = +FLT_MAX;

    next_value_tracker_verify_sentinels( tracker );
}

inline void next_value_tracker_add_sample( next_value_tracker_t * tracker, float value )
{
    next_value_tracker_verify_sentinels( tracker );

    if ( value > tracker->maximum )
    {
        tracker->maximum = value;
    }

    if ( value < tracker->minimum )
    {
        tracker->minimum = value;
    }

    tracker->values[tracker->index] = value;
    tracker->index = ( tracker->index + 1 ) % NEXT_VALUE_TRACKER_HISTORY;
    tracker->num_samples++;
}

inline void next_value_tracker_calculate( next_value_tracker_t * tracker, float * min_value, float * max_value, float * avg_value )
{
    next_value_tracker_verify_sentinels( tracker );

    next_assert( min_value );
    next_assert( max_value );
    next_assert( avg_value );

    if ( tracker->num_samples == 0 )
    {
        *min_value = 0.0f;
        *max_value = 0.0f;
        *avg_value = 0.0f;
        return;
    }

    if ( tracker->num_samples > NEXT_VALUE_TRACKER_HISTORY )
    {
        tracker->num_samples = NEXT_VALUE_TRACKER_HISTORY;
    }

    *min_value = tracker->minimum;
    *max_value = tracker->maximum;

    next_assert( tracker->num_samples > 0 );

    double sum = 0.0;
    for ( int i = 0; i < tracker->num_samples; i++ )
    {
        sum += (double) tracker->values[i];
    }
    *avg_value = float( sum / double(tracker->num_samples) );

    next_value_tracker_reset( tracker );
}

#endif // #ifndef NEXT_VALUE_TRACKER_H
