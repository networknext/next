/*
    Network Next SDK. Copyright © 2017 - 2022 Network Next, Inc.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following 
    conditions are met:

    1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

    2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions 
       and the following disclaimer in the documentation and/or other materials provided with the distribution.

    3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote 
       products derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. 
    IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR 
    CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; 
    OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING 
    NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include "next.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <string>
#include <map>
#include <unordered_set>

std::map<std::string,uint8_t*> client_map;
std::unordered_set<std::string> match_data_set;

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

bool no_upgrade = false;
int upgrade_count = 0;
int num_upgrades = 0;
bool tags_multi = false;
bool server_events = false;
bool match_data = false;
bool flush = false;

extern bool next_packet_loss;

void verify_packet( const uint8_t * packet_data, int packet_bytes )
{
    next_assert( packet_bytes >= 32 );
    next_assert( packet_bytes <= NEXT_MTU );
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
            if ( server_events && !flush )
            {
                uint64_t event1 = (1<<10);
                uint64_t event2 = (1<<20);
                uint64_t event3 = (1<<30); 
                next_server_event( server, from, event1 | event2 | event3 );
            }

            if ( match_data && !flush && match_data_set.find( address_string ) == match_data_set.end() )
            {
                const double match_values[] = {10.10f, 20.20f, 30.30f};
                int num_match_values = sizeof(match_values) / sizeof(match_values[0]);
                next_server_match( server, from, "test match id", match_values, num_match_values );

                match_data_set.insert( address_string );
            }
        }

        std::map<std::string,uint8_t*>::iterator itor = client_map.find( address_string );

        if ( itor == client_map.end() || memcmp( packet_data, itor->second, 32 ) != 0 )
        {
            next_server_upgrade_session( server, from, 0 );

            if ( tags_multi )
            {
                const char * tags[] = {"pro", "streamer"};
                const int num_tags = 2;
                next_server_tag_session_multiple( server, from, tags, num_tags );
            }
            else
            {
                next_server_tag_session( server, from, "test" );
            }

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

    const char * customer_private_key_env = getenv( "NEXT_CUSTOMER_PRIVATE_KEY" );
    if ( customer_private_key_env )
    {
        strcpy( config.customer_private_key, customer_private_key_env );
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

    server = next_server_create( NULL, "127.0.0.1:32202", "0.0.0.0:32202", "local", server_packet_received, NULL );
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

    const char * server_tags_multi_env = getenv( "SERVER_TAGS_MULTI" );
    if ( server_tags_multi_env )
    {
        tags_multi = true;
    }

    const char * server_events_env = getenv( "SERVER_EVENTS" );
    if ( server_events_env )
    {
        server_events = true;
    }

    const char * match_data_env = getenv( "SERVER_MATCH_DATA" );
    if ( match_data_env )
    {
        match_data = true;
    }

    const char * flush_env = getenv( "SERVER_FLUSH" );
    if ( flush_env )
    {
        flush = true;
    }

    bool restarted = false;

    while ( !quit )
    {
        next_server_update( server );

        if ( restart_time > 0.0 && next_time() > restart_time && !restarted )
        {
            printf( "restarting server\n" );
            next_server_destroy( server );
            server = next_server_create( NULL, "127.0.0.1:32202", "0.0.0.0:32202", "local", server_packet_received, NULL );
            if ( server == NULL )
                return 1;

            restarted = true;
        }

        next_sleep( 1.0 / 60.0 );
    }

    if ( flush )
    {
        uint64_t event1 = (1<<10);
        uint64_t event2 = (1<<20);
        uint64_t event3 = (1<<30);

        const double match_values[] = {10.10f, 20.20f, 30.30f};
        int num_match_values = sizeof(match_values) / sizeof(match_values[0]);

        for ( std::map<std::string,uint8_t*>::iterator itor = client_map.begin(); itor != client_map.end(); ++itor )
        {
            next_address_t client_address;
            if ( next_address_parse( &client_address, itor->first.c_str() ) != NEXT_OK )
                continue;

            if ( server_events )
            {
                next_server_event( server, &client_address, event1 | event2 | event3 );
            }

            if ( match_data )
            {
                next_server_match( server, &client_address, "test match id", match_values, num_match_values );
            }

        }

        next_server_flush ( server );
    }
    
    next_server_destroy( server );
    
    next_term();

    return 0;
}
