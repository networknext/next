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
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <cinttypes>

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void generate_packet( uint8_t * packet_data, int & packet_bytes, bool high_bandwidth )
{
    if ( high_bandwidth )
    {
        packet_bytes = NEXT_MTU;
    }
    else
    {
        packet_bytes = 1 + ( rand() % NEXT_MTU ) / 10;
    }
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes; ++i )
    {
        packet_data[i] = (uint8_t) ( start + i ) % 256;
    }
}

void verify_packet( const uint8_t * packet_data, int packet_bytes )
{
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes; ++i )
        next_assert( packet_data[i] == (uint8_t) ( ( start + i ) % 256 ) );
}

void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context;
    verify_packet( packet_data, packet_bytes );
}

#define NEXT_CLIENT_COUNTER_MAX 64
extern void next_client_counters( next_client_t * client, uint64_t * counters );
extern float next_fake_direct_packet_loss;
extern float next_fake_direct_rtt;
extern float next_fake_next_packet_loss;
extern float next_fake_next_rtt;
extern bool next_packet_loss;
extern bool next_fake_fallback_to_direct;

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_config_t config;
    next_default_config( &config );

    const char * customer_public_key_env = getenv( "NEXT_CUSTOMER_PUBLIC_KEY" );
    if ( customer_public_key_env )
    {
        strcpy( config.customer_public_key, customer_public_key_env );
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

    double stop_sending_packets_time = -1.0;
    const char * stop_sending_packets_time_env = getenv( "CLIENT_STOP_SENDING_PACKETS_TIME" );
    if ( stop_sending_packets_time_env )
    {
        stop_sending_packets_time = atof( stop_sending_packets_time_env );
    }

    double fallback_to_direct_time = -1.0;
    const char * fallback_to_direct_time_env = getenv( "CLIENT_FALLBACK_TO_DIRECT_TIME" );
    if ( fallback_to_direct_time_env )
    {
        fallback_to_direct_time = atof( fallback_to_direct_time_env );
    }

    next_log_level( NEXT_LOG_LEVEL_DEBUG );

    if ( next_init( NULL, &config ) != NEXT_OK )
        return 1;

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received, NULL );
    if ( client == NULL )
        return 1;

    next_client_open_session( client, "127.0.0.1:32202" );

    const char * client_user_flags_env = getenv( "CLIENT_USER_FLAGS" );
    if ( client_user_flags_env )
    {
        next_client_set_user_flags( client, 0x123 );
    }

    const char * client_packet_loss_env = getenv( "CLIENT_PACKET_LOSS" );
    if ( client_packet_loss_env )
    {
        next_packet_loss = true;
    }

    double stop_time = -1.0f;

    const char * duration_env = getenv( "CLIENT_DURATION" );
    if ( duration_env )
    {
        stop_time = atof( duration_env );
    }

    bool high_bandwidth = false;
    const char * high_bandwidth_env = getenv( "CLIENT_HIGH_BANDWIDTH" );
    if ( high_bandwidth_env )
    {
        high_bandwidth = true;
    }

    double time = 0.0;
    double delta_time = 1.0 / 60.0;

    bool second_connect_completed = false;

    // IMPORTANT: Have to wait a bit here or the first packet will get dropped
    // because of a race condition between the server getting set via OPEN_SESSION_COMMAND
    // and the recvfrom for the response from the server.
    next_client_update( client );
    next_sleep( 0.25 );

    while ( stop_time < 0.0 || time < stop_time )
    {
        if ( quit )
            break;

        if ( connect_time > 0.0f && !second_connect_completed && connect_time < time )
        {
            next_client_open_session( client, connect_address );
            second_connect_completed = true;
        }

        if ( client_user_flags_env )
        {
            next_client_set_user_flags( client, 0x123 );
        }

        next_client_update( client );

        if ( stop_sending_packets_time < 0.0 || time < stop_sending_packets_time )
        {
            uint8_t packet_data[NEXT_MTU];
            memset( packet_data, 0, sizeof( packet_data ) );

            int packet_bytes = 0;
            generate_packet( packet_data, packet_bytes, high_bandwidth );

            next_client_send_packet( client, packet_data, packet_bytes );
        }

        if ( fallback_to_direct_time >= 0.0 && time > fallback_to_direct_time )
        {
            next_fake_fallback_to_direct = true;
        }

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
