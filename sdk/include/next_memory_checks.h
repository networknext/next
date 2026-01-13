/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_MEMORY_CHECKS_H
#define NEXT_MEMORY_CHECKS_H

#include "next.h"

#if NEXT_ENABLE_MEMORY_CHECKS

    #define NEXT_DECLARE_SENTINEL(n) uint32_t next_sentinel_##n[64];

    #define NEXT_INITIALIZE_SENTINEL(pointer,n) for ( int i = 0; i < 64; ++i ) { pointer->next_sentinel_##n[i] = 0xBAADF00D; }

    #define NEXT_VERIFY_SENTINEL(pointer,n) for ( int i = 0; i < 64; ++i ) { next_assert( pointer->next_sentinel_##n[i] == 0xBAADF00D ); }

#else // #if NEXT_ENABLE_MEMORY_CHECKS

    #define NEXT_DECLARE_SENTINEL(n)

    #define NEXT_INITIALIZE_SENTINEL(pointer,n)

    #define NEXT_VERIFY_SENTINEL(pointer,n)

#endif // #if NEXT_ENABLE_MEMORY_CHECKS

#endif // #ifndef NEXT_MEMORY_CHECKS_H
