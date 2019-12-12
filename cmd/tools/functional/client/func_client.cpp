/*
    Network Next SDK $(NEXT_VERSION_FULL)

    Copyright © 2017 - 2020 Network Next, Inc.

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
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) packet_data; (void) packet_bytes;
}

// todo: hack test
#define NEXT_CLIENT_COUNTER_MAX 64
extern void next_client_counters( next_client_t * client, uint64_t * counters );
extern float next_fake_direct_packet_loss;
extern float next_fake_direct_rtt;
extern float next_fake_next_packet_loss;
extern float next_fake_next_rtt;

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_log_level( NEXT_LOG_LEVEL_DEBUG );

    next_config_t config;
    next_default_config( &config );

    const char * customer_public_key_env = getenv( "NEXT_CUSTOMER_PUBLIC_KEY" );
    if ( customer_public_key_env )
    {
        strcpy( config.customer_public_key, customer_public_key_env );
    }
    
    const char * disable_try_before_you_buy_env = getenv( "CLIENT_DISABLE_TRY_BEFORE_YOU_BUY" );
    if ( disable_try_before_you_buy_env )
    {
        config.try_before_you_buy = false;
    }

    const char * disable_network_next_env = getenv( "CLIENT_DISABLE_NETWORK_NEXT" );
    if ( disable_network_next_env )
    {
        config.disable_network_next = true;
    }

    const char * fake_direct_packet_loss_env = getenv( "CLIENT_FAKE_DIRECT_PACKET_LOSS" );
    if ( fake_direct_packet_loss_env )
    {
        next_fake_direct_packet_loss = atof( fake_direct_packet_loss_env );
    }

    const char * fake_direct_rtt_env = getenv( "CLIENT_FAKE_DIRECT_RTT" );
    if ( fake_direct_rtt_env )
    {
        next_fake_direct_rtt = atof( fake_direct_rtt_env );
    }

    const char * fake_next_packet_loss_env = getenv( "CLIENT_FAKE_NEXT_PACKET_LOSS" );
    if ( fake_next_packet_loss_env )
    {
        next_fake_next_packet_loss = atof( fake_next_packet_loss_env );
    }

    const char * fake_next_rtt_env = getenv( "CLIENT_FAKE_NEXT_RTT" );
    if ( fake_next_rtt_env )
    {
        next_fake_next_rtt = atof( fake_next_rtt_env );
    }

    double connect_time = 0.0;
    const char * connect_address = NULL;
    const char * connect_time_env = getenv( "CLIENT_CONNECT_TIME" );
    if ( connect_time_env )
    {
        connect_time = atof( connect_time_env );
        if ( connect_time > 0.0 )
        {
            connect_address = getenv( "CLIENT_CONNECT_ADDRESS" );
            if ( !connect_address )
            {
                connect_time = 0.0;
            }
        }
    }

    if ( next_init( NULL, &config ) != NEXT_OK )
        return 1;

    next_client_t * client = next_client_create( NULL, client_packet_received );
    if ( client == NULL )
        return 1;

    next_client_open_session( client, "127.0.0.1:32202" );

    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof( packet_data ) );

    double stop_time = -1.0f;

    const char * duration_env = getenv( "CLIENT_DURATION" );
    if ( duration_env )
    {
        stop_time = atof( duration_env );
    }

    double time = 0.0;
    double delta_time = 1.0 / 60.0;

    bool second_connect_completed = false;

    while ( stop_time < 0.0 || time < stop_time )
    {
        if ( quit )
            break;

        if ( connect_time > 0.0f && !second_connect_completed && connect_time < time )
        {
            next_client_open_session( client, connect_address );
            second_connect_completed = true;
        }

        next_client_update( client );

        next_client_send_packet( client, packet_data, sizeof( packet_data ) );
        
        next_sleep( delta_time );

        time += delta_time;
    }

    next_client_close_session( client );

    next_sleep( 1.0f );

    uint64_t counters[NEXT_CLIENT_COUNTER_MAX];
    next_client_counters( client, counters );
    for ( int i = 0; i < NEXT_CLIENT_COUNTER_MAX; ++i )
    {
        if ( i != NEXT_CLIENT_COUNTER_MAX - 1 )
        {
            fprintf( stderr, "%" PRIu64 ",", counters[i] );
        }
        else
        {
            fprintf( stderr, "%" PRIu64, counters[i] );
        }
    }
    fprintf( stderr, "\n" );

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
