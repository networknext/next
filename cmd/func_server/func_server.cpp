/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#include "next_platform.h"
#include "next_address.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <string>
#include <map>
#include <unordered_set>

std::map<std::string,uint8_t*> client_map;

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

bool no_upgrade = false;
int upgrade_count = 0;
int num_upgrades = 0;
bool session_events = false;
bool flush = false;

extern bool next_packet_loss;

void verify_packet( const uint8_t * packet_data, int packet_bytes )
{
    next_assert( packet_bytes >= 32 );
    next_assert( packet_bytes <= NEXT_MAX_PACKET_BYTES - 1 );
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes - 32; ++i )
    {
        next_assert( packet_data[32+i] == (uint8_t) ( ( start + i ) % 256 ) );
    }
}

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;

    if ( packet_bytes <= 32 )
    {
        printf( "packet too small\n" );
        return;
    }

    verify_packet( packet_data, packet_bytes );

    next_server_send_packet( server, from, packet_data, packet_bytes );

    if ( !next_server_ready( server ) )
        return;

    if ( !no_upgrade && ( upgrade_count == 0 || ( upgrade_count > 0 && num_upgrades < upgrade_count ) ) )
    {
        char address[256];
        next_address_to_string( from, address );
        std::string address_string( address );

        next_server_stats_t stats;
        bool session_exists = next_server_stats( server, from, &stats );

        if ( next_server_session_upgraded( server, from ) && session_exists )
        {
            if ( session_events && !flush )
            {
                uint64_t event1 = (1<<10);
                uint64_t event2 = (1<<20);
                uint64_t event3 = (1<<30); 
                next_server_session_event( server, from, event1 | event2 | event3 );
            }
        }

        std::map<std::string,uint8_t*>::iterator itor = client_map.find( address_string );

        if ( itor == client_map.end() || memcmp( packet_data, itor->second, 32 ) != 0 )
        {
            next_server_upgrade_session( server, from, 0 );

            num_upgrades++;

            uint8_t * client_id = (uint8_t*) malloc( 32 );
            memcpy( client_id, packet_data, 32 );
            if ( itor != client_map.end() )
            {
                client_map.erase( itor );
            }
            client_map.insert( std::make_pair( address_string, client_id ) );
        }
    }
}

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_config_t config;
    next_default_config( &config );

    const char * buyer_private_key_env = getenv( "NEXT_BUYER_PRIVATE_KEY" );
    if ( buyer_private_key_env )
    {
        strcpy( config.buyer_private_key, buyer_private_key_env );
    }
    
    const char * disable_network_next_env = getenv( "SERVER_DISABLE_NETWORK_NEXT" );
    if ( disable_network_next_env )
    {
        config.disable_network_next = true;
    }
    
    next_log_level( NEXT_LOG_LEVEL_DEBUG );

    if ( next_init( NULL, &config ) != NEXT_OK )
        return 1;
    
    const char * server_packet_loss_env = getenv( "SERVER_PACKET_LOSS" );
    if ( server_packet_loss_env )
    {
        next_packet_loss = true;
    }

    next_server_t * server = NULL;

    server = next_server_create( NULL, "127.0.0.1:32202", "127.0.0.1:32202", "local", server_packet_received );
    if ( server == NULL )
        return 1;

    const char * no_upgrade_env = getenv( "SERVER_NO_UPGRADE" );
    if ( no_upgrade_env )
    {
        int value = atoi( no_upgrade_env );
        if ( value != 0 )
        {
            no_upgrade = true;
        }
    }
    
    const char * upgrade_count_env = getenv( "SERVER_UPGRADE_COUNT" );
    if ( upgrade_count_env )
    {
        upgrade_count = atoi( upgrade_count_env );
    }

    double restart_time = 0.0;
    const char * restart_time_env = getenv( "SERVER_RESTART_TIME" );
    if ( restart_time_env )
    {
        restart_time = atof( restart_time_env );
    }

    const char * session_events_env = getenv( "SESSION_EVENTS" );
    if ( session_events_env )
    {
        session_events = true;
    }

    const char * flush_env = getenv( "SERVER_FLUSH" );
    if ( flush_env )
    {
        flush = true;
    }

    bool restarted = false;

    double base_time = next_platform_time();

    while ( !quit )
    {
        next_server_update( server );

        if ( restart_time > 0.0 && ( next_platform_time() - base_time ) > restart_time && !restarted )
        {
            printf( "restarting server\n" );
            next_server_destroy( server );
            server = next_server_create( NULL, "127.0.0.1:32202", "0.0.0.0:32202", "local", server_packet_received );
            if ( server == NULL )
                return 1;

            restarted = true;
        }

        next_platform_sleep( 1.0 / 60.0 );
    }
    
    if ( flush )
    {
        uint64_t event1 = (1<<10);
        uint64_t event2 = (1<<20);
        uint64_t event3 = (1<<30);

        for ( std::map<std::string,uint8_t*>::iterator itor = client_map.begin(); itor != client_map.end(); ++itor )
        {
            next_address_t client_address;
            if ( next_address_parse( &client_address, itor->first.c_str() ) != NEXT_OK )
                continue;

            if ( session_events )
            {
                next_server_session_event( server, &client_address, event1 | event2 | event3 );
            }
        }

        next_server_flush ( server );
    }

    next_server_destroy( server );
    
    next_term();

    return 0;
}
