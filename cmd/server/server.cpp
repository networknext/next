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

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;

    next_server_send_packet( server, from, packet_data, packet_bytes );

    if ( next_server_ready( server ) && !next_server_session_upgraded( server, from ) )
    {
        next_server_upgrade_session( server, from, "12345" );
    }
}

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL ); 

    next_server_t * server = next_server_create( NULL, "127.0.0.1:30000", "0.0.0.0:30000", "local", server_packet_received );

    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

    while ( !quit )
    {
        next_server_update( server );

        if ( next_server_direct_only( server ) )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "detected server is in direct only mode. restarting..." );
            break;
        }

        next_platform_sleep( 0.001 );

        fflush( stdout );
    }

    next_server_flush( server );
    
    next_server_destroy( server );
    
    next_term();

    return 0;
}
