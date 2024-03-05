/*
    Network Next XDP Relay
*/

#ifndef RELAY_DEBUG_H
#define RELAY_DEBUG_H

#include "relay.h"

struct debug_t
{
    int config_fd;
    int state_fd;
    int stats_fd;
    int relay_map_fd;
    int session_map_fd;
    uint8_t relay_ping_key[RELAY_PING_KEY_BYTES];
    uint8_t relay_secret_key[RELAY_SECRET_KEY_BYTES];
};

struct config_t;
struct bpf_t;

int debug_init( struct debug_t * debug, struct config_t * config, struct bpf_t * bpf );

int debug_run( struct debug_t * debug );

void debug_shutdown( struct debug_t * debug );

#endif // #ifndef RELAY_DEBUG_H
