/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.
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

    srand( time(NULL) );

    double restart_time = next_platform_time() + 1.0 * ( double(rand()) / double(RAND_MAX) );

    next_server_t * server = next_server_create( NULL, "127.0.0.1:30000", "0.0.0.0:30000", "local", server_packet_received );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

    while ( !quit )
    {
        next_server_update( server );

        next_platform_sleep( 0.01 );

        if ( next_platform_time() >= restart_time )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "restarting server" );

            next_server_destroy( server );

            server = next_server_create( NULL, "127.0.0.1:30000", "0.0.0.0:30000", "local", server_packet_received );
            if ( server == NULL )
            {
                printf( "error: failed to create server\n" );
                return 1;
            }

            restart_time = next_platform_time() + 1.0 * ( double(rand()) / double(RAND_MAX) );
        }
    }
    
    next_server_destroy( server );

    next_term();

    return 0;
}
