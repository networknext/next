/*
    Network Next SDK. Copyright © 2017 - 2020 Network Next, Inc.

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

std::map<std::string,uint8_t*> client_map;

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

bool no_upgrade = false;
int upgrade_count = 0;
int num_upgrades = 0;

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

    if ( !no_upgrade && ( upgrade_count == 0 || ( upgrade_count > 0 && num_upgrades < upgrade_count ) ) )
    {
        char address[256];
        next_address_to_string( from, address );
        std::string address_string( address );

        std::map<std::string,uint8_t*>::iterator itor = client_map.find( address_string );

        if ( itor == client_map.end() || memcmp( packet_data, itor->second, 32 ) != 0 )
        {
            next_server_upgrade_session( server, from, 0 );

            next_server_tag_session( server, from, "test" );

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

    while ( !quit )
    {
        next_server_update( server );

        next_sleep( 1.0 / 60.0 );
    }
    
    next_server_destroy( server );
    
    next_term();

    return 0;
}
