/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#include "next_crypto.h"
#include "next_platform.h"

#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <cinttypes>

static volatile int quit = 0;

static uint8_t client_id[32];

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void generate_packet( uint8_t * packet_data, int & packet_bytes, bool high_bandwidth, bool big_packets )
{
    if ( big_packets )
    {
        packet_bytes = NEXT_MAX_PACKET_BYTES - 1;
    }
    else if ( high_bandwidth )
    {
        packet_bytes = NEXT_MTU;
    }
    else
    {
        packet_bytes = 32 + 1 + ( rand() % NEXT_MTU ) / 10;
    }
    memcpy( packet_data, client_id, 32 );
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes - 32; ++i )
    {
        packet_data[32+i] = (uint8_t) ( start + i ) % 256;
    }
}

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

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) from;

    if ( packet_bytes <= 32 )
        return;
   
    if ( memcmp( packet_data, client_id, 32 ) != 0 )
        return;

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

    const char * buyer_public_key_env = getenv( "NEXT_BUYER_PUBLIC_KEY" );
    if ( buyer_public_key_env )
    {
        strcpy( config.buyer_public_key, buyer_public_key_env );
    }

    const char * disable_network_next_env = getenv( "CLIENT_DISABLE_NETWORK_NEXT" );
    if ( disable_network_next_env )
    {
        config.disable_network_next = true;
    }

    bool report_session = false;
    const char * report_session_env = getenv( "CLIENT_REPORT_SESSION" );
    if ( report_session_env )
    {
        report_session = true;
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

    next_client_t * client = next_client_create( NULL, "127.0.0.1:0", client_packet_received );
    if ( client == NULL )
        return 1;

    next_client_open_session( client, "127.0.0.1:32202" );

    next_crypto_random_bytes( client_id, 32 );

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

    bool big_packets = false;
    const char * big_packets_env = getenv( "CLIENT_BIG_PACKETS" );
    if ( big_packets_env )
    {
        big_packets = true;
    }

    double time = 0.0;
    double delta_time = 1.0 / 60.0;

    bool reported = false;
    bool second_connect_completed = false;

    while ( stop_time < 0.0 || time < stop_time )
    {
        if ( quit )
            break;

        if ( connect_time > 0.0f && !second_connect_completed && connect_time < time )
        {
            next_client_open_session( client, connect_address );
            second_connect_completed = true;
            next_crypto_random_bytes( client_id, 32 );
        }

        next_client_update( client );

        if ( next_client_ready( client ) && ( stop_sending_packets_time < 0.0 || time < stop_sending_packets_time ) )
        {
            uint8_t packet_data[NEXT_MTU];
            memset( packet_data, 0, sizeof( packet_data ) );

            int packet_bytes = 0;
            generate_packet( packet_data, packet_bytes, high_bandwidth, big_packets );

            next_client_send_packet( client, packet_data, packet_bytes );
        }

        if ( fallback_to_direct_time >= 0.0 && time > fallback_to_direct_time )
        {
            next_fake_fallback_to_direct = true;
        }

        if ( report_session && time > 30.0 && !reported )
        {
            next_client_report_session( client );
            reported = true;
        }

        next_platform_sleep( delta_time );

        time += delta_time;
    }

    next_client_close_session( client );

    next_platform_sleep( 1.0f );

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
