
// Network Next PS4 Testbed

#include "next.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <kernel.h>

const char * customer_public_key = "UoFYERKJnCt18mU53IsWzlEXD2pYD9yd+TiZiq9+cMF9cHG4kMwRtw==";
// const char * linux_test_bed_server_gcp = "34.133.77.37";
// const char * linux_test_bed_server_amazon = "54.161.150.42";
// const char * windows_test_bed_server_gcp = "35.193.180.236";
// const char * windows_test_bed_server_amazon = "3.13.23.117";

unsigned int sceLibcHeapExtendedAlloc = 1;

size_t sceLibcHeapSize = SCE_LIBC_HEAP_SIZE_EXTENDED_ALLOC_NO_LIMIT;

static volatile int quit = 0;

void packet_received( next_client_t * client, void * context, const struct next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;
    (void) context;
    (void) from;
    (void) packet_data;
    (void) packet_bytes;

    // ...
}

int32_t main( int argc, const char * const argv[] )
{
    SceKernelModule next_library = sceKernelLoadStartModule("/app0/next-ps4-4.20.0.prx", 0, NULL, 0, NULL, NULL);
    if ( next_library < 0 )
    {
        printf( "Failed to load next PRX library\n" );
    }

    next_log_level( NEXT_LOG_LEVEL_NONE );

    next_config_t config;
    next_default_config( &config );
    strncpy_s( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

    next_init( NULL, &config );

    printf( "\nRunning tests...\n\n" );

    next_test();

    printf( "\nAll tests passed successfully!\n\n" );

    next_log_level( NEXT_LOG_LEVEL_INFO );

    printf( "Starting client...\n\n" );
    
    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", packet_received, NULL );
    if ( !client )
    {
        printf( "error: failed to create network next client" );
        exit( 1 );
    }

    next_client_open_session( client, "34.133.77.37:50000" );

    while ( !quit )
    {
        next_client_update( client );

        uint8_t packet_data[32];
        memset( packet_data, 0, sizeof(packet_data) );
        next_client_send_packet( client, packet_data, sizeof( packet_data ) );

        next_sleep( 1.0f / 60.0f );
    }

    printf( "\nShutting down...\n\n" );
    
    next_client_destroy( client );

    next_term();

    sceKernelStopUnloadModule( next_library, 0, NULL, 0, NULL, NULL );

    printf( "\n" );

    return 0;
}
