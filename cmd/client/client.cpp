/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"
#include "next_platform.h"

#include <stdio.h>
#include <signal.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) from; (void) packet_data; (void) packet_bytes;
}

void generate_packet( uint8_t * packet_data, int & packet_bytes )
{
    packet_bytes = 100;
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes; ++i )
        packet_data[i] = (uint8_t) ( start + i ) % 256;
}

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    uint64_t current_session_id = 0;

    next_init( NULL, NULL );

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );

    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    const char * connect_address = "127.0.0.1:30000";

    const char * connect_address_override = getenv( "NEXT_CONNECT_ADDRESS" );
    if ( connect_address_override )
    {
        connect_address = connect_address_override;
    }

    next_client_open_session( client, connect_address );

    while ( !quit )
    {
        next_client_update( client );

        uint64_t session_id = next_client_session_id( client );

        if ( current_session_id == 0 && session_id != 0 )
        {
            printf( "session id is %" PRIx64 "\n", session_id );
            current_session_id = session_id;
        }

        if ( next_client_ready( client ) ) 
        {
            int packet_bytes = 0;
            uint8_t packet_data[NEXT_MTU];
            generate_packet( packet_data, packet_bytes );
            next_client_send_packet( client, packet_data, packet_bytes );
        }

        next_platform_sleep( 1.0 / 60.0 );

        fflush( stdout );
    }

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
