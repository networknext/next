/*
    Network Next. Copyright © 2017 - 2020 Network Next, Inc.

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
#include <stdint.h>
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

void generate_packet( uint8_t * packet_data, int & packet_bytes )
{
    packet_bytes = 1 + ( rand() % NEXT_MTU );
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes; ++i )
        packet_data[i] = (uint8_t) ( start + i ) % 256;
}

void verify_packet( const uint8_t * packet_data, int packet_bytes )
{
    const int start = packet_bytes % 256;
    for ( int i = 0; i < packet_bytes; ++i )
    {
        if ( packet_data[i] != (uint8_t) ( ( start + i ) % 256 ) )
        {
            printf( "%d: %d != %d (%d)\n", i, packet_data[i], ( start + i ) % 256, packet_bytes );
        }
        next_assert( packet_data[i] == (uint8_t) ( ( start + i ) % 256 ) );
    }
}

int main()
{
    printf( "\nWelcome to Network Next!\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL );

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received, NULL );

    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    next_client_open_session( client, "127.0.0.1:32202" );

    double accumulator = 0.0;

    while ( !quit )
    {
        next_client_update( client );

        int packet_bytes = 0;
        uint8_t packet_data[NEXT_MTU];
        generate_packet( packet_data, packet_bytes );
        next_client_send_packet( client, packet_data, packet_bytes );

        next_sleep( 1.0 / 60.0 );

        accumulator += 1.0 / 60.0;

        if ( accumulator > 10.0 )
        {
            accumulator = 0.0;

            printf( "================================================================\n" );
            
            const next_client_stats_t * stats = next_client_stats( client );

            const char * platform = "unknown";

            switch ( stats->platform_id )
            {
                case NEXT_PLATFORM_WINDOWS:
                    platform = "windows";
                    break;

                case NEXT_PLATFORM_MAC:
                    platform = "mac";
                    break;

                case NEXT_PLATFORM_LINUX:
                    platform = "linux";
                    break;

                case NEXT_PLATFORM_SWITCH:
                    platform = "nintendo switch";
                    break;

                case NEXT_PLATFORM_PS4:
                    platform = "ps4";
                    break;

                case NEXT_PLATFORM_IOS:
                    platform = "ios";
                    break;

                case NEXT_PLATFORM_XBOX_ONE:
                    platform = "xbox one";
                    break;

                default:
                    break;
            }

            const char * state_string = "???";

            const int state = next_client_state( client );
            
            switch ( state )
            {
                case NEXT_CLIENT_STATE_CLOSED:
                    state_string = "closed";
                    break;

                case NEXT_CLIENT_STATE_OPEN:
                    state_string = "open";
                    break;

                case NEXT_CLIENT_STATE_ERROR:
                    state_string = "error";
                    break;

                default:
                    break;
            }

            printf( "state = %s (%d)\n", state_string, state );

            printf( "session_id = %" PRIx64 "\n", next_client_session_id( client ) );

            printf( "platform_id = %s (%d)\n", platform, (int) stats->platform_id );

            const char * connection = "unknown";
            
            switch ( stats->connection_type )
            {
                case NEXT_CONNECTION_TYPE_WIRED:
                    connection = "wired";
                    break;

                case NEXT_CONNECTION_TYPE_WIFI:
                    connection = "wifi";
                    break;

                case NEXT_CONNECTION_TYPE_CELLULAR:
                    connection = "cellular";
                    break;

                default:
                    break;
            }

            printf( "connection_type = %s (%d)\n", connection, stats->connection_type );

            if ( !stats->fallback_to_direct )
            {
                printf( "upgraded = %s\n", stats->upgraded ? "true" : "false" );
                printf( "committed = %s\n", stats->committed ? "true" : "false" );
                printf( "multipath = %s\n", stats->multipath ? "true" : "false" );
                printf( "reported = %s\n", stats->reported ? "true" : "false" );
            }

            printf( "direct_rtt = %.2fms\n", stats->direct_rtt );
            printf( "direct_jitter = %.2fms\n", stats->direct_jitter );
            printf( "direct_packet_loss = %.1f%%\n", stats->direct_packet_loss );

            printf( "fallback_to_direct = %s\n", stats->fallback_to_direct ? "true" : "false" );

            if ( stats->next )
            {
                printf( "next_rtt = %.2fms\n", stats->next_rtt );
                printf( "next_jitter = %.2fms\n", stats->next_jitter );
                printf( "next_packet_loss = %.1f%%\n", stats->next_packet_loss );
                printf( "next_bandwidth_up = %.1fkbps\n", stats->next_kbps_up );
                printf( "next_bandwidth_down = %.1fkbps\n", stats->next_kbps_down );
            }

            if ( stats->upgraded && !stats->fallback_to_direct )
            {
                printf( "packets_sent_client_to_server = %" PRId64 "\n", stats->packets_sent_client_to_server );
                printf( "packets_sent_server_to_client = %" PRId64 "\n", stats->packets_sent_server_to_client );
                printf( "packets_lost_client_to_server = %" PRId64 "\n", stats->packets_lost_client_to_server );
                printf( "packets_lost_server_to_client = %" PRId64 "\n", stats->packets_lost_server_to_client );
                printf( "packets_out_of_order_client_to_server = %" PRId64 "\n", stats->packets_out_of_order_client_to_server );
                printf( "packets_out_of_order_server_to_client = %" PRId64 "\n", stats->packets_out_of_order_server_to_client );
                printf( "jitter_client_to_server = %f\n", stats->jitter_client_to_server );
                printf( "jitter_server_to_client = %f\n", stats->jitter_server_to_client );
            }

            printf( "================================================================\n" );
        }
    }

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
