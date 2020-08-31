
// Nintendo Switch Testbed

#include <nn/nn_Log.h>
#include <nn/socket.h>
#include <nn/ro.h>
#include <nn/nn_Assert.h>
#include <nn/fs.h>
#include <cstring>
#include <cstdlib>
#include <cstdio>
#include "next.h"

const char * customer_public_key = "pX89DxRQBaQuYebn80myIh+12DPY2mXFhoRhiLRpKFoffKSVE0zotg==";

void packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;
    (void) context;
    (void) packet_data;
    (void) packet_bytes;
    // ...
}

static const char * log_level_str( int level )
{
    if ( level == NEXT_LOG_LEVEL_DEBUG )
        return "debug";
    else if ( level == NEXT_LOG_LEVEL_INFO )
        return "info";
    else if ( level == NEXT_LOG_LEVEL_ERROR )
        return "error";
    else if ( level == NEXT_LOG_LEVEL_WARN )
        return "warning";
    else
        return "???";
}

static void print_function( int level, const char * format, ...)
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    const char * level_str = log_level_str( level );
    va_end( args );
	if ( level != NEXT_LOG_LEVEL_NONE )
	{
	    NN_LOG( "%0.6f %s: %s\n", next_time(), level_str, buffer );
	}
	else
	{
		NN_LOG( "%s\n", buffer );
	}
}

static nn::socket::ConfigDefaultWithMemory socket_config_with_memory;

static volatile int quit = 0;

extern "C" void nnMain()
{
    nn::Result result = nn::socket::Initialize( socket_config_with_memory );
    if ( result.IsFailure() )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to initialize nintendo sockets" );
        exit( 1 );
    }

    next_log_function( print_function );

	next_config_t config;
	next_default_config( &config );
	strncpy( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

	next_init( NULL, &config );

	NN_LOG( "\nRunning tests...\n\n" );

	next_log_level( NEXT_LOG_LEVEL_NONE );

	next_test();
	
	printf( "\nAll tests passed successfully!\n\n" );

	next_log_level( NEXT_LOG_LEVEL_INFO );

	printf( "Starting client...\n\n" );

	next_client_t * client = next_client_create( NULL, "0.0.0.0:0", packet_received );
    if ( !client )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create network next client" );
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

	printf( "Shutting down...\n\n" );

	next_client_destroy( client );

    next_term();

    nn::socket::Finalize();

    NN_LOG( "\n" );
}
