/*
    Network Next XDP Relay
*/

#include "relay_manager.h"
#include "relay_platform.h"
#include "relay_ping_history.h"

#include <stdlib.h>

struct relay_manager_t * relay_manager_create()
{
    struct relay_manager_t * manager = (struct relay_manager_t*) malloc( sizeof(struct relay_manager_t) );
    if ( !manager )
        return NULL;
    memset( manager, 0, sizeof(struct relay_manager_t) );
    relay_manager_reset( manager );
    return manager;
}

void relay_manager_reset( struct relay_manager_t * manager )
{
    assert( manager );
    manager->num_relays = 0;
    memset( manager->relay_ids, 0, sizeof(manager->relay_ids) );
    memset( manager->relay_last_ping_time, 0, sizeof(manager->relay_last_ping_time) );
    memset( manager->relay_addresses, 0, sizeof(manager->relay_addresses) );
    memset( manager->relay_ports, 0, sizeof(manager->relay_ports) );
    for ( int i = 0; i < manager->num_relays; i++ )
    {
        free( manager->relay_ping_history[i] );
    }
    memset( manager->relay_ping_history, 0, sizeof(manager->relay_ping_history) );
}

void relay_manager_update( struct relay_manager_t * manager, struct relay_set * new_relays, struct relay_set * delete_relays )
{
    assert( manager );
    assert( new_relays );
    assert( delete_relays );

    // if there are no creates or deletes, there is no work to do

    if ( new_relays->num_relays == 0 && delete_relays->num_relays == 0 )
        return;

    // copy the current set of relays, sans any deletes to a new array

    int num_relays = 0;
    uint64_t relay_ids[MAX_RELAYS];
    uint32_t relay_addresses[MAX_RELAYS];
    uint16_t relay_ports[MAX_RELAYS];
    uint8_t relay_internal[MAX_RELAYS];
    struct relay_ping_history_t * relay_ping_history[MAX_RELAYS];

    for ( int i = 0; i < manager->num_relays; i++ )
    {
        bool found = false;
        for ( int j = 0; j < delete_relays->num_relays; j++ )
        {
            if ( delete_relays->id[j] == manager->relay_ids[i] )
            {
                found = true;
                break;
            }
        }

        if ( !found )
        {
            relay_ids[num_relays] = manager->relay_ids[i];
            relay_addresses[num_relays] = manager->relay_addresses[i];
            relay_ports[num_relays] = manager->relay_ports[i];
            relay_internal[num_relays] = manager->relay_internal[i];
            relay_ping_history[num_relays] = manager->relay_ping_history[i];
            num_relays++;
        }
        else
        {
            free( manager->relay_ping_history[i] );
        }
    }

    // add the new relays to the end

    for ( int i = 0; i < new_relays->num_relays; i++ )
    {
        relay_ids[num_relays] = new_relays->id[i];
        relay_addresses[num_relays] = new_relays->address[i];
        relay_ports[num_relays] = new_relays->port[i];
        relay_internal[num_relays] = new_relays->internal[i];
        relay_ping_history[num_relays] = (struct relay_ping_history_t*) malloc( MAX_RELAYS * sizeof(struct relay_ping_history_t) );
        num_relays++;
    }

    // commit the updated relay array

    manager->num_relays = num_relays;
    memcpy( manager->relay_ids, relay_ids, 8 * num_relays );
    memcpy( manager->relay_addresses, relay_addresses, 4 * num_relays );
    memcpy( manager->relay_ports, relay_ports, 2 * num_relays );
    memcpy( manager->relay_internal, relay_internal, num_relays );
    memcpy( manager->relay_ping_history, relay_ping_history, sizeof(struct relay_ping_history_t*) * num_relays );

    // make sure all ping times are evenly distributed to avoid clusters of ping packets

    double current_time = relay_platform_time();

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        manager->relay_last_ping_time[i] = current_time - RELAY_PING_TIME + i * RELAY_PING_TIME / manager->num_relays;
    }
}

bool relay_manager_process_pong( struct relay_manager_t * manager, uint32_t from_address, uint16_t from_port, uint64_t sequence )
{
    assert( manager );

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        if ( from_address == manager->relay_addresses[i] && from_port == manager->relay_ports[i] )
        {
            relay_ping_history_pong_received( manager->relay_ping_history[i], sequence, relay_platform_time() );
            return true;
        }
    }

    return false;
}

void relay_manager_get_ping_stats( struct relay_manager_t * manager, struct relay_ping_stats_t * ping_stats )
{
    assert( manager );
    assert( ping_stats );

    double current_time = relay_platform_time();

    ping_stats->num_relays = manager->num_relays;

    for ( int i = 0; i < ping_stats->num_relays; ++i )
    {
        struct relay_ping_history_stats_t stats;
        relay_ping_history_get_stats( manager->relay_ping_history[i], current_time - RELAY_PING_STATS_WINDOW, current_time, &stats, RELAY_PING_SAFETY );
        ping_stats->relay_ids[i] = manager->relay_ids[i];
        ping_stats->relay_rtt[i] = stats.rtt;
        ping_stats->relay_jitter[i] = stats.jitter;
        ping_stats->relay_packet_loss[i] = stats.packet_loss;
    }
}

void relay_manager_destroy( struct relay_manager_t * manager )
{
    assert( manager );
    relay_manager_reset( manager );
    free( manager );
}
