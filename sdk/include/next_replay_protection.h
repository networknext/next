/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_REPLAY_PROTECTION_H
#define NEXT_REPLAY_PROTECTION_H

#include "next.h"
#include "next_config.h"
#include "next_memory_checks.h"

struct next_replay_protection_t
{
    NEXT_DECLARE_SENTINEL(0)

    uint64_t most_recent_sequence;
    uint64_t received_packet[NEXT_REPLAY_PROTECTION_BUFFER_SIZE];

    NEXT_DECLARE_SENTINEL(1)
};

inline void next_replay_protection_initialize_sentinels( next_replay_protection_t * replay_protection )
{
    (void) replay_protection;
    next_assert( replay_protection );
    NEXT_INITIALIZE_SENTINEL( replay_protection, 0 )
    NEXT_INITIALIZE_SENTINEL( replay_protection, 1 )
}

inline void next_replay_protection_verify_sentinels( next_replay_protection_t * replay_protection )
{
    (void) replay_protection;
    next_assert( replay_protection );
    NEXT_VERIFY_SENTINEL( replay_protection, 0 )
    NEXT_VERIFY_SENTINEL( replay_protection, 1 )
}

inline void next_replay_protection_reset( next_replay_protection_t * replay_protection )
{
    next_replay_protection_initialize_sentinels( replay_protection );

    replay_protection->most_recent_sequence = 0;

    memset( replay_protection->received_packet, 0xFF, sizeof( replay_protection->received_packet ) );

    next_replay_protection_verify_sentinels( replay_protection );
}

inline bool next_replay_protection_already_received( next_replay_protection_t * replay_protection, uint64_t sequence )
{
    next_replay_protection_verify_sentinels( replay_protection );

    if ( sequence + NEXT_REPLAY_PROTECTION_BUFFER_SIZE <= replay_protection->most_recent_sequence )
    {
        return true;
    }

    int index = (int) ( sequence % NEXT_REPLAY_PROTECTION_BUFFER_SIZE );

    if ( replay_protection->received_packet[index] == 0xFFFFFFFFFFFFFFFFLL )
    {
        replay_protection->received_packet[index] = sequence;
        return false;
    }

    if ( replay_protection->received_packet[index] >= sequence )
    {
        return true;
    }

    return false;
}

inline void next_replay_protection_advance_sequence( next_replay_protection_t * replay_protection, uint64_t sequence )
{
    next_replay_protection_verify_sentinels( replay_protection );

    if ( sequence > replay_protection->most_recent_sequence )
    {
        replay_protection->most_recent_sequence = sequence;
    }

    int index = (int) ( sequence % NEXT_REPLAY_PROTECTION_BUFFER_SIZE );

    replay_protection->received_packet[index] = sequence;
}

#endif // #ifndef NEXT_REPLAY_PROTECTION_H
