
// Network Next PS4 Testbed

#include "next.h"
#include "next_platform.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <kernel.h>

const char * buyer_public_key = "M/NxwbhSaPjUHES+kePTWD9TFA0bga1kubG+3vg0rTx/3sQoFgMB1w==";

unsigned int sceLibcHeapExtendedAlloc = 1;

size_t sceLibcHeapSize = SCE_LIBC_HEAP_SIZE_EXTENDED_ALLOC_NO_LIMIT;

void packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;
    (void) context;
    (void) from;
    (void) packet_data;
    (void) packet_bytes;

    // ...
}

int32_t main(int argc, const char* const argv[])
{
    next_config_t config;
    next_default_config( &config );
    strncpy_s( config.buyer_public_key, buyer_public_key, sizeof(config.buyer_public_key) - 1 );

    next_init( NULL, &config );

    printf( "Starting client...\n\n" );

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", packet_received );
    if ( !client )
    {
        printf( "error: failed to create network next client" );
        exit( 1 );
    }

    next_client_open_session( client, "173.255.241.176:50000" );

    while ( true )
    {
        next_client_update( client );

        uint8_t packet_data[32];
        memset( packet_data, 0, sizeof(packet_data) );
        next_client_send_packet( client, packet_data, sizeof(packet_data) );

        next_platform_sleep( 1.0f / 60.0f );
    }

    printf( "\nShutting down...\n\n" );

    next_client_destroy( client );

    next_term();

    printf( "\n" );

    return 0;
}