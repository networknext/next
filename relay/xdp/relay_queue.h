/*
    Network Next XDP Relay
*/

#ifndef RELAY_QUEUE_H
#define RELAY_QUEUE_H

#include "relay.h"
#include <stdlib.h>

struct relay_queue_t
{
    int size;
    int num_entries;
    int start_index;
    void ** entries;
};

static inline void relay_queue_clear( struct relay_queue_t * queue )
{
    assert( queue ); 

    const int queue_size = queue->size;
    const int start_index = queue->start_index;

    for ( int i = 0; i < queue->num_entries; ++i )
    {
        const int index = (start_index + i ) % queue_size;
        free( queue->entries[index] );
        queue->entries[index] = NULL;
    }

    queue->num_entries = 0;
    queue->start_index = 0;
}

static inline void relay_queue_destroy( struct relay_queue_t * queue )
{
    relay_queue_clear( queue );

    free( queue->entries );

    memset( queue, 0, sizeof(struct relay_queue_t) );
    
    free( queue );
}

static inline struct relay_queue_t * relay_queue_create( int size )
{
    struct relay_queue_t * queue = (struct relay_queue_t*) malloc( sizeof(struct relay_queue_t) );
    assert( queue );
    if ( !queue )
        return NULL;

    queue->size = size;
    queue->num_entries = 0;
    queue->start_index = 0;
    queue->entries = (void**) malloc( size * sizeof(void*) );

    assert( queue->entries );

    if ( !queue->entries )
    {
        relay_queue_destroy( queue );
        return NULL;
    }

    return queue;
}

static inline bool relay_queue_full( struct relay_queue_t * queue )
{
    assert( queue ); 
    return queue->num_entries == queue->size;
}

static inline bool relay_queue_push( struct relay_queue_t * queue, void * entry )
{
    assert( queue ); 
    assert( entry );

    if ( queue->num_entries == queue->size )
    {
        free( entry );
        return false;
    }

    int index = ( queue->start_index + queue->num_entries ) % queue->size;

    queue->entries[index] = entry;
    queue->num_entries++;

    return true;
}

static inline void * relay_queue_pop( struct relay_queue_t * queue )
{
    assert( queue ); 

    if ( queue->num_entries == 0 )
        return NULL;

    void * entry = queue->entries[queue->start_index];

    queue->start_index = ( queue->start_index + 1 ) % queue->size;
    queue->num_entries--;

    return entry;
}

#endif // #ifndef RELAY_QUEUE_H
