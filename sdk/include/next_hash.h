/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_HASH_H
#define NEXT_HASH_H

#include "next.h"

uint64_t next_hash_string( const char * string );

uint64_t next_relay_id( const char * name );

uint64_t next_datacenter_id( const char * name );

typedef uint64_t next_fnv_t;

void next_fnv_init( next_fnv_t * fnv );

void next_fnv_write( next_fnv_t * fnv, const uint8_t * data, size_t size );

uint64_t next_fnv_finalize( next_fnv_t * fnv );

#endif // #ifndef NEXT_HASH_H
