/*
    Network Next XDP Relay
*/

#ifndef RELAY_HASH_H
#define RELAY_HASH_H

#include <memory.h>

struct relay_hash
{
    uint64_t values[RELAY_HASH_SIZE];
};

inline void relay_hash_initialize( struct relay_hash * hash, uint64_t * relay_ids, int num_relays )
{
    assert( hash );
    assert( num_relays >= 0 );
    assert( num_relays < MAX_RELAYS );

    memset( hash->values, 0, sizeof(hash->values) );

    for ( int i = 0; i < num_relays; i++ )
    {
        char data[8];
        memcpy( data, &relay_ids[i], 8 );
        uint64_t hash_value = 0xCBF29CE484222325;
        for ( int j = 0; j < 8; j++ )
        {
            hash_value ^= data[j];
            hash_value *= 0x00000100000001B3;
        }

        int index = (int) ( hash_value % RELAY_HASH_SIZE );
        while ( 1 )
        {
            if ( hash->values[index] == 0 )
            {
                hash->values[index] = relay_ids[i];
                break;
            }
            index = ( index + 1 ) % RELAY_HASH_SIZE;
        }
    }
}

inline bool relay_hash_exists( struct relay_hash * hash, uint64_t relay_id )
{
    char data[8];
    memcpy( data, &relay_id, 8 );
    uint64_t hash_value = 0xCBF29CE484222325;
    for ( int i = 0; i < 8; i++ )
    {
        hash_value ^= data[i];
        hash_value *= 0x00000100000001B3;
    }

    int index = (int) ( hash_value % RELAY_HASH_SIZE );
    while ( 1 )
    {
        if ( hash->values[index] == relay_id )
            return true;
        if ( hash->values[index] == 0 )
            return false;
    }
}

#endif // #ifndef RELAY_HASH_H
