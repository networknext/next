/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next_hash.h"

#include <string.h>

typedef uint64_t next_fnv_t;

void next_fnv_init( next_fnv_t * fnv )
{
    *fnv = 0xCBF29CE484222325;
}

void next_fnv_write( next_fnv_t * fnv, const uint8_t * data, size_t size )
{
    for ( size_t i = 0; i < size; i++ )
    {
        (*fnv) ^= data[i];
        (*fnv) *= 0x00000100000001B3;
    }
}

uint64_t next_fnv_finalize( next_fnv_t * fnv )
{
    return *fnv;
}

uint64_t next_hash_string( const char * string )
{
    next_fnv_t fnv;
    next_fnv_init( &fnv );
    next_fnv_write( &fnv, (uint8_t *)( string ), strlen( string ) );
    return next_fnv_finalize( &fnv );
}

uint64_t next_relay_id( const char * name )
{
    return next_hash_string( name );
}

uint64_t next_datacenter_id( const char * name )
{
    return next_hash_string( name );
}
