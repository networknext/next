/*
    Network Next XDP Relay
*/

#include "relay_debug.h"
#include "relay_platform.h"
#include "relay_base64.h"
#include "relay_config.h"
#include "relay_bpf.h"
#include "relay_shared.h"

#include <stdio.h>
#include <stdlib.h>
#include <memory.h>
#include <errno.h>
#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <xdp/libxdp.h>
#include <arpa/inet.h>

#if RELAY_DEBUG

int debug_init( struct debug_t * debug, struct config_t * config, struct bpf_t * bpf )
{
    // copy across bpf map file descriptors

    debug->config_fd = bpf->config_fd;
    debug->state_fd = bpf->state_fd;
    debug->stats_fd = bpf->stats_fd;
    debug->relay_map_fd = bpf->relay_map_fd;
    debug->session_map_fd = bpf->session_map_fd;

    // get relay ping key from environment, since in debug mode we don't talk with the relay backend

    const char * relay_ping_key_env = getenv( "RELAY_PING_KEY" );
    if ( !relay_ping_key_env )
    {
        printf( "\nerror: RELAY_PING_KEY not set\n\n" );
        return RELAY_ERROR;
    }

    if ( relay_base64_decode_data( relay_ping_key_env, debug->relay_ping_key, RELAY_PING_KEY_BYTES ) != RELAY_PING_KEY_BYTES )
    {
        printf( "\nerror: invalid relay ping key\n\n" );
        return RELAY_ERROR;
    }

    printf( "Relay ping key is %s\n", relay_ping_key_env );

    // hard code the relay secret key to match client.go

    uint8_t relay_secret_key[32] = { 0x22, 0x3c, 0x0c, 0xc6, 0x70, 0x7b, 0x99, 0xc4, 0xdd, 0x44, 0xb9, 0xe8, 0x3c, 0x78, 0x1c, 0xd7, 0xd3, 0x2f, 0x9b, 0xad, 0x70, 0xbf, 0x8d, 0x9f, 0xe3, 0xa6, 0xd4, 0xc7, 0xe3, 0xb2, 0x98, 0x90 };
    memcpy( debug->relay_secret_key, relay_secret_key, RELAY_SECRET_KEY_BYTES );

    // load the relay config to bpf
    {
        struct relay_config relay_config;

        memset( &relay_config, 0, sizeof(struct relay_config) );

        relay_config.relay_port = htons( config->relay_port );
        relay_config.relay_public_address = htonl( config->relay_public_address );
        relay_config.relay_internal_address = htonl( config->relay_internal_address );
        memcpy( relay_config.relay_secret_key, relay_secret_key, RELAY_SECRET_KEY_BYTES );
        memcpy( relay_config.relay_backend_public_key, config->relay_backend_public_key, RELAY_BACKEND_PUBLIC_KEY_BYTES );

        __u32 key = 0;
        int err = bpf_map_update_elem( debug->config_fd, &key, &relay_config, BPF_ANY );
        if ( err != 0 )
        {
            printf( "\nerror: failed to set relay config: %s\n\n", strerror(errno) );
            return RELAY_ERROR;
        }
    }

    // for testing add some IP addresses as known relays in the relay map

    uint64_t dummy_value = 1;

    printf( "adding known relay 192.168.1.20:30000\n" );
    {
        __u64 key = ( ( (__u64)0x1401a8c0 ) << 32 ) | htons(30000);

        int err = bpf_map_update_elem( debug->relay_map_fd, &key, &dummy_value, BPF_ANY );
        if ( err != 0 )
        {
            printf( "\nerror: failed to add relay to map: %s\n\n", strerror(errno) );
            return RELAY_ERROR;
        }

        uint64_t value;
        int result = bpf_map_lookup_elem( debug->relay_map_fd, &key, &value );
        if ( result != 0 )
        {
            printf( "\nerror: relay map is broken\n\n" );
            return RELAY_ERROR;
        }
    }

    return RELAY_OK;
}

extern bool quit;

int debug_run( struct debug_t * debug )
{
    printf( "Starting debug thread\n" );

    // loop until the user exits with CTRL-C

    struct relay_state state;

    while ( !quit ) 
    {
        state.current_timestamp = time( NULL );

        for ( int i = 0; i < 8; i++ )
        {
            state.current_magic[i] = (uint8_t) i;
        }
        memset( state.previous_magic, 1, sizeof( state.previous_magic ) );
        memset( state.next_magic,     2, sizeof( state.next_magic     ) );
        memcpy( state.ping_key, debug->relay_ping_key, RELAY_PING_KEY_BYTES );     
        {
            __u32 key = 0;
            int err = bpf_map_update_elem( debug->state_fd, &key, &state, BPF_ANY );
            if ( err != 0 )
            {
                printf( "\nerror: failed to set relay state: %s\n\n", strerror(errno) );
                return 1;
            }
        }

        relay_platform_sleep( 1.0 );
    }

    return 0;
}

void debug_shutdown( struct debug_t * debug )
{
    // ...
}

#else // #if RELAY_DEBUG

int debug_init( struct debug_t * debug, struct config_t * config, struct bpf_t * bpf )
{
    return RELAY_OK;
}

void debug_run( struct debug_t * debug )
{
    // ...
}

void debug_shutdown( struct debug_t * debuc )
{
    // ...
}

#endif // #if RELAY_DEBUG
