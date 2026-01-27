/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#include "next_platform.h"
#include "next_address.h"

#include <time.h>
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>
#ifndef _WIN32
#include <unistd.h>
#include <net/if.h>
#include <netinet/in.h>
#include <sys/ioctl.h>
#include <sys/syscall.h>
#endif

#define MAX_CLIENTS 1000

static volatile int quit = 0;

static bool raspberry_low_bandwidth;

char raspberry_backend_url[1024];

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) packet_data; (void) packet_bytes; (void) from;
}

uint64_t raspberry_user_id()
{
    srand( time( NULL ) );

    uint64_t user_id = 0;

    char data[8];
    for ( int i = 0; i < 8; i++ )
    {
        data[i] = rand() % 256;
    }
    memcpy((char*)&user_id, data, 8);

    return user_id;
}

struct client_data_t
{
    uint64_t user_id;
    int thread_index;
};

void client_thread_function( void * data )
{
    client_data_t * client_data = (client_data_t*) data;

    const int GameLength = 360;

    const int MaxServers = 256;

    next_platform_sleep( rand() % GameLength );

    while ( !quit )
    {
        // update list of server addresses

        int num_servers = 0;
        next_address_t server_addresses[MaxServers];
        memset( server_addresses, 0, sizeof( server_addresses ) );

        char cmd[2048];
        snprintf( cmd, sizeof(cmd), "curl -s %s/servers --max-time 10 2>/dev/null", raspberry_backend_url );
        FILE * file = popen( cmd, "r" );
        if ( !file )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "could not get list of servers" );
            exit(1);
        }

        char buffer[10*1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL )
        {
            if ( num_servers >= MaxServers )
                break;
            int i = 0;
            while ( true )
            {
                if ( buffer[i] == '\0' )
                    break;
                if ( buffer[i] == '\n' || buffer[i] == '\r' )
                {
                    buffer[i] = '\0';
                    break;
                }
                i++;
            }
            next_address_t address;
            if ( next_address_parse( &address, buffer ) == NEXT_OK )
            {
                server_addresses[num_servers] = address;
                num_servers++;
            }
            else
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "could not parse '%s'", buffer );
            }
        }

        pclose( file );

        // if we don't have any servers to connect to, just wait 10 seconds and try again

        if ( num_servers == 0 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "no servers found" );
            next_platform_sleep( 1.0 );
            continue;
        }

        // create a client

        next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
        if ( client == NULL )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create client" );
            exit(1);
        }

        // connect to random server and send packets for game length of time

        char connect_address[NEXT_MAX_ADDRESS_STRING_LENGTH];
        
        next_address_to_string( &server_addresses[rand() % num_servers], connect_address );

        next_client_open_session( client, connect_address );

        uint8_t packet_data[256];
        memcpy( packet_data, &client_data->user_id, 8 );

        double connect_time = next_platform_time();

        uint64_t current_session_id = 0;

        while ( !quit )
        {
            next_client_game_stats( client, 25.0f, 10.0f, 0.0f );

            next_client_send_packet( client, packet_data, sizeof( packet_data ) );

            next_client_update( client );

            uint64_t session_id = next_client_session_id( client );

            if ( session_id != 0 && current_session_id == 0 )
            {
                next_printf( NEXT_LOG_LEVEL_INFO, "session id is %016" PRIx64, session_id );
                current_session_id = session_id;
            }

            if ( next_platform_time() > connect_time + GameLength )
            {
                next_printf( NEXT_LOG_LEVEL_INFO, "game has finished. reconnecting..." );
                break;
            }

            if ( !raspberry_low_bandwidth )
            {
                next_platform_sleep( 1.0f / 100.0f );
            }
            else
            {
                next_platform_sleep( 1.0f );
            }
        }

        next_client_destroy( client );
    }
}

void run_clients( int num_clients )
{
    if ( num_clients > MAX_CLIENTS )
    {
        num_clients = MAX_CLIENTS;
    }

    uint64_t user_id = raspberry_user_id();

    if ( user_id != 0 )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "user id is: %" PRIu64, user_id );
    }

    client_data_t client_data[MAX_CLIENTS];

    for ( int i = 0; i < num_clients; i++ )
    {
        client_data[i].thread_index = i;
        client_data[i].user_id = user_id;

        next_platform_thread_t * thread = next_platform_thread_create( NULL, client_thread_function, &client_data[i] );
        next_assert( thread );
        (void) thread;
    }
}

extern const char * next_platform_getenv( const char * name );

int main()
{
    printf( "\nRaspberry Client\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_config_t config;
    next_default_config( &config );
#ifdef _WIN32
    strncpy_s( config.buyer_public_key, "gtdzp3hCfJ9Y+6OOpsWoMChMXhXGDRnY7vkFdHwNqVW0bdp6jjTx6Q==", 256 );
#else
    strncpy( config.buyer_public_key, "gtdzp3hCfJ9Y+6OOpsWoMChMXhXGDRnY7vkFdHwNqVW0bdp6jjTx6Q==", 256 );
#endif
    config.buyer_public_key[255] = 0;

    next_init( NULL, &config );

    next_copy_string( raspberry_backend_url, "http://127.0.0.1:40100", sizeof(raspberry_backend_url) );
    const char * raspberry_backend_url_override = next_platform_getenv( "RASPBERRY_BACKEND_URL" );
    if ( raspberry_backend_url_override )
    {
        next_copy_string( raspberry_backend_url, raspberry_backend_url_override, sizeof(raspberry_backend_url) );
    }

    int num_clients = 25;
    const char * num_clients_override = next_platform_getenv( "RASPBERRY_NUM_CLIENTS" );
    if ( num_clients_override )
    {
        num_clients = atoi( num_clients_override );
    }

    next_printf( NEXT_LOG_LEVEL_DEBUG, "raspberry backend url: %s", raspberry_backend_url );

    raspberry_low_bandwidth = next_platform_getenv( "RASPBERRY_LOW_BANDWIDTH" ) != NULL;

    if ( raspberry_low_bandwidth )
    {
        printf( "low bandwidth mode\n" );
    }

    printf( "simulating %d clients\n", num_clients );

    run_clients( num_clients );

    while ( !quit )
    {
        next_platform_sleep( 1.0 );
    }

    next_term();

    printf( "\n" );

    return 0;
}
