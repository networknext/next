/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"
#include "next_platform.h"

#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>

const char * bind_address = "0.0.0.0:50000";
const char * server_address = "127.0.0.1:50000";
const char * server_datacenter = "local";
const char * server_backend_hostname = "server-dev.virtualgo.net";

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;

    next_server_send_packet( server, from, packet_data, packet_bytes );

    next_printf( NEXT_LOG_LEVEL_INFO, "server received packet from client (%d bytes)", packet_bytes );

    if ( next_server_ready( server ) && !next_server_session_upgraded( server, from ) )
    {
        const char * user_id_string = "12345";
        next_server_upgrade_session( server, from, user_id_string );
    }
}

#if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS
#define strncpy_s strncpy
#endif // #if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_config_t config;
    next_default_config( &config );
    strncpy_s( config.server_backend_hostname, server_backend_hostname, sizeof(config.server_backend_hostname) - 1 );

    if ( next_init( NULL, &config ) != NEXT_OK )
    {
        printf( "error: could not initialize network next\n" );
        return 1;
    }

    next_server_t * server = next_server_create( NULL, server_address, bind_address, server_datacenter, server_packet_received );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }
    
    while ( !quit )
    {
        next_server_update( server );

        next_platform_sleep( 1.0 / 60.0 );
    }

    printf( "\nshutting down\n" );

    next_server_flush( server );
    
    next_server_destroy( server );
    
    next_term();

    return 0;
}
