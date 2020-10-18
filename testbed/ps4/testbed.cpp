
// Network Next PS4 Testbed

#include "next.h"
#include <stdlib.h>
#include <libsecure.h>
#include <string.h>
#include <kernel.h>

const char * customer_public_key = "pX89DxRQBaQuYebn80myIh+12DPY2mXFhoRhiLRpKFoffKSVE0zotg==";

unsigned int sceLibcHeapExtendedAlloc = 1;

size_t sceLibcHeapSize = SCE_LIBC_HEAP_SIZE_EXTENDED_ALLOC_NO_LIMIT;

static volatile int quit = 0;

void packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;
    (void) context;
    (void) packet_data;
    (void) packet_bytes;
    // ...
}

static uint8_t random_seed[4160];

int32_t main( int argc, const char * const argv[] )
{
    SceKernelModule next_library = sceKernelLoadStartModule("/app0/next-ps4-3.4.5.prx", 0, NULL, 0, NULL, NULL);
    if ( next_library < 0 )
    {
        printf( "Failed to load next PRX library\n" );
    }

    SceLibSecureBlock random_seed_block = { random_seed, sizeof( random_seed ) };
    if ( sceLibSecureInit( SCE_LIBSECURE_FLAGS_RANDOM_GENERATOR, &random_seed_block ) != SCE_LIBSECURE_OK )
        exit( 1 );

	next_log_level(NEXT_LOG_LEVEL_NONE);

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

	next_client_open_session( client, "173.255.241.176:32202" );

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

    sceLibSecureDestroy();

    printf( "\n" );

    return 0;
}
