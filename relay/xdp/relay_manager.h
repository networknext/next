/*
    Network Next XDP Relay
*/

#ifndef RELAY_MANAGER_H
#define RELAY_MANAGER_H

#include "relay.h"
#include "relay_set.h"
#include "relay_ping_stats.h"

struct relay_manager_t
{
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    double relay_last_ping_time[MAX_RELAYS];
    uint32_t relay_addresses[MAX_RELAYS];
    uint16_t relay_ports[MAX_RELAYS];
    uint8_t relay_internal[MAX_RELAYS];
    struct relay_ping_history_t * relay_ping_history[MAX_RELAYS];
};

struct relay_manager_t * relay_manager_create();

void relay_manager_reset( struct relay_manager_t * manager );

void relay_manager_update( struct relay_manager_t * manager, struct relay_set * new_relays, struct relay_set * delete_relays );

bool relay_manager_process_pong( struct relay_manager_t * manager, uint32_t from_address, uint16_t from_port, uint64_t sequence );

void relay_manager_get_ping_stats( struct relay_manager_t * manager, struct relay_ping_stats_t * ping_stats );

void relay_manager_destroy( struct relay_manager_t * manager );

#endif // #ifndef RELAY_MANAGER_H
