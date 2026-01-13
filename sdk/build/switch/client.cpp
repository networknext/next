/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include <nn/nn_Log.h>
#include <nn/socket.h>
#include <nn/ro.h>
#include <nn/nn_Assert.h>
#include <nn/fs.h>
#include <cstring>
#include <cstdlib>
#include <cstdio>
#include "next.h"
#include "next_tests.h"
#include "next_platform.h"

const char * server_address = "35.232.190.226:30000";

const char * buyer_public_key = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ==";

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void)client; (void)context; (void)from; (void)packet_data; (void)packet_bytes;
}

extern "C" void nnMain()
{
    next_config_t config;
    next_default_config( &config );

    next_init( NULL, &config );

    printf( "\nRunning tests...\n\n" );

    next_log_level( NEXT_LOG_LEVEL_NONE );

    next_run_tests();
    
    printf( "\nAll tests passed successfully!\n\n" );

    next_term();

    printf( "Starting client...\n\n" );

    next_log_level( NEXT_LOG_LEVEL_INFO );

    strncpy( config.buyer_public_key, buyer_public_key, sizeof(config.buyer_public_key) - 1 );

    next_init( NULL, &config );

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
    if ( !client )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create network next client" );
        exit( 1 );
    }

    next_client_open_session( client, server_address );

    while ( true )
    {
        next_client_update( client );

        uint8_t packet_data[32];
        memset( packet_data, 0, sizeof(packet_data) );
        next_client_send_packet( client, packet_data, sizeof( packet_data ) );

        next_platform_sleep( 1.0f / 60.0f );
    }

    printf( "Shutting down...\n\n" );

    next_client_destroy( client );

    next_term();
}
