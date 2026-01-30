/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#include "next_config.h"
#include "next_platform.h"
#include "next_address.h"

#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>

static volatile int quit = 0;

extern bool raspberry_fake_latency;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;

    if ( packet_bytes != 256 )
        return;

    next_server_send_packet( server, from, packet_data, packet_bytes );

    if ( !next_server_session_upgraded( server, from ) )
    {
        uint64_t user_id;
        char * user_id_bytes = (char*) &user_id;
        memcpy( user_id_bytes, packet_data, 8 );
        char buffer[256];
        next_server_upgrade_session( server, from, next_user_id_string( user_id, buffer, sizeof(buffer) ) );
    }
}

struct thread_data_t
{
    const char * server_address;
    const char * raspberry_backend_url;
};

void server_update_thread( void * data )
{
    next_assert( data );

    thread_data_t * thread_data = (thread_data_t*) data;

    const char * server_address = thread_data->server_address;
    const char * raspberry_backend_url = thread_data->raspberry_backend_url;
    
    char command_line[1024];
    snprintf( command_line, sizeof(command_line), "curl -s -d \"%s\" -X POST %s/server_update --max-time 10 -v 2>/dev/null", server_address, raspberry_backend_url );

    while ( !quit )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "sending raspberry server update: %s", server_address );

        system( command_line );

        next_platform_sleep( 10.0 );
    }
}

void send_server_updates_to_raspberry_backend( thread_data_t * thread_data )
{
    next_platform_thread_t * thread = next_platform_thread_create( NULL, server_update_thread, thread_data );
    next_assert( thread );
    (void) thread;
}

extern const char * next_platform_getenv( const char * name );

int main()
{
    printf( "\nRaspberry Server\n\n" );

    raspberry_fake_latency = 

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL );

    raspberry_fake_latency = next_platform_getenv( "RASPBERRY_FAKE_LATENCY" ) != NULL;

    if ( raspberry_fake_latency )
    {
        printf( "fake latency mode\n" );
    }

    char raspberry_backend_url[1024];
    next_copy_string( raspberry_backend_url, "http://127.0.0.1:40100", sizeof( raspberry_backend_url ) );
    const char * raspberry_backend_url_override = next_platform_getenv( "RASPBERRY_BACKEND_URL" );
    if ( raspberry_backend_url_override )
    {
        next_copy_string( raspberry_backend_url, raspberry_backend_url_override, sizeof(raspberry_backend_url) );
    }

    next_printf( NEXT_LOG_LEVEL_DEBUG, "raspberry backend url: %s", raspberry_backend_url );

    char server_address[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_copy_string( server_address, "127.0.0.1", sizeof(server_address) );

    // look for a server address override (used in docker env...)

    const char * server_address_override = next_platform_getenv( "NEXT_SERVER_ADDRESS" );
    if ( server_address_override )
    {
        next_copy_string( server_address, server_address_override, sizeof(server_address) );
    }

    // if we are running in google cloud, detect google cloud public IP address

    FILE * file = popen( "curl -s http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H \"Metadata-Flavor: Google\" --max-time 10 -s 2>/dev/null", "r" );

    char buffer[1024];

    while ( file && fgets( buffer, sizeof(buffer), file ) != NULL )
    {
        next_address_t address;
        if ( next_address_parse( &address, buffer ) == NEXT_OK )
        {
            next_address_to_string( &address, server_address );
            break;
        }
    }

    if ( file )
    {
        pclose( file );
    }

    // start server

    next_server_t * server = next_server_create( NULL, server_address, "0.0.0.0", "local", server_packet_received );

    if ( server == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create raspberry server" );
        return 1;
    }

    // now that we know our port number, combine with the address and we have our public address

    int server_port = next_server_port( server );
    char public_address[1024];
    memset( public_address, 0, sizeof(public_address) );
    snprintf( public_address, sizeof(public_address) - 1, "%s:%d", server_address, server_port );
    next_printf( NEXT_LOG_LEVEL_INFO, "raspberry server public address is: %s", public_address );

    // send server updates to the raspberry backend in the background

    thread_data_t thread_data;

    thread_data.server_address = public_address;
    thread_data.raspberry_backend_url = raspberry_backend_url;

    send_server_updates_to_raspberry_backend( &thread_data );

    // main loop for server

    while ( !quit )
    {
        next_server_update( server );

        if ( next_server_direct_only( server ) )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "detected raspberry server is in direct only mode. restarting..." );
            break;
        }

        next_platform_sleep( 1.0 / 60.0 );
    }

    next_server_flush( server );

    next_server_destroy( server );

    next_term();

    return 0;
}
