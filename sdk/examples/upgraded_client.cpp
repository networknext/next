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

const char * bind_address = "0.0.0.0:0";
const char * server_address = "127.0.0.1:50000";
const char * buyer_public_key = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ==";

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) packet_data; (void) packet_bytes; (void) from;
    next_printf( NEXT_LOG_LEVEL_INFO, "client received packet from server (%d bytes)", packet_bytes );
}

#if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS
#define strncpy_s strncpy
#endif // #if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );
    
    next_config_t config;
    next_default_config( &config );
    strncpy_s( config.buyer_public_key, buyer_public_key, sizeof(config.buyer_public_key) - 1 );

    if ( next_init( NULL, &config ) != NEXT_OK )
    {
        printf( "error: could not initialize network next\n" );
        return 1;
    }

    next_client_t * client = next_client_create( NULL, bind_address, client_packet_received );
    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    next_client_open_session( client, server_address );

    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof( packet_data ) );

    while ( !quit )
    {
        next_client_update( client );

        if ( next_client_ready( client ) )
        {
            next_client_send_packet( client, packet_data, sizeof(packet_data) );
        }
        
        next_platform_sleep( 0.25 );
    }

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
