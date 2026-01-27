/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_QUEUE_H
#define NEXT_QUEUE_H

#include "next.h"
#include "next_memory_checks.h"

struct next_queue_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int size;
    int num_entries;
    int start_index;
    void ** entries;

    NEXT_DECLARE_SENTINEL(1)
};

inline void next_queue_initialize_sentinels( next_queue_t * queue )
{
    (void) queue;
    next_assert( queue );
    NEXT_INITIALIZE_SENTINEL( queue, 0 )
    NEXT_INITIALIZE_SENTINEL( queue, 1 )
}

inline void next_queue_verify_sentinels( next_queue_t * queue )
{
    (void) queue;
    next_assert( queue );
    NEXT_VERIFY_SENTINEL( queue, 0 )
    NEXT_VERIFY_SENTINEL( queue, 1 )
}

inline void next_queue_destroy( next_queue_t * queue );

inline next_queue_t * next_queue_create( void * context, int size )
{
    next_queue_t * queue = (next_queue_t*) next_malloc( context, sizeof(next_queue_t) );
    next_assert( queue );
    if ( !queue )
        return NULL;

    next_queue_initialize_sentinels( queue );

    queue->context = context;
    queue->size = size;
    queue->num_entries = 0;
    queue->start_index = 0;
    queue->entries = (void**) next_malloc( context, size * sizeof(void*) );

    next_assert( queue->entries );

    if ( !queue->entries )
    {
        next_queue_destroy( queue );
        return NULL;
    }

    next_queue_verify_sentinels( queue );

    return queue;
}

inline void next_queue_clear( next_queue_t * queue )
{
    next_queue_verify_sentinels( queue );

    const int queue_size = queue->size;
    const int start_index = queue->start_index;

    for ( int i = 0; i < queue->num_entries; ++i )
    {
        const int index = (start_index + i ) % queue_size;
        next_free( queue->context, queue->entries[index] );
        queue->entries[index] = NULL;
    }

    queue->num_entries = 0;
    queue->start_index = 0;
}

inline void next_queue_destroy( next_queue_t * queue )
{
    next_queue_verify_sentinels( queue );

    next_queue_clear( queue );

    next_free( queue->context, queue->entries );

    next_clear_and_free( queue->context, queue, sizeof( queue ) );
}

inline int next_queue_push( next_queue_t * queue, void * entry )
{
    next_queue_verify_sentinels( queue );

    next_assert( entry );

    if ( queue->num_entries == queue->size )
    {
        next_free( queue->context, entry );
        return NEXT_ERROR;
    }

    int index = ( queue->start_index + queue->num_entries ) % queue->size;

    queue->entries[index] = entry;
    queue->num_entries++;

    return NEXT_OK;
}

inline void * next_queue_pop( next_queue_t * queue )
{
    next_queue_verify_sentinels( queue );

    if ( queue->num_entries == 0 )
        return NULL;

    void * entry = queue->entries[queue->start_index];

    queue->start_index = ( queue->start_index + 1 ) % queue->size;
    queue->num_entries--;

    return entry;
}

#endif // #ifndef NEXT_QUEUE_H
