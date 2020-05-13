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
#include <inttypes.h>

const char * bind_address = "0.0.0.0:0";
const char * server_address = "127.0.0.1:32202";
const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) packet_data; (void) packet_bytes;
}

int main()
{
    printf( "\nWelcome to Network Next!\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );
    
    next_config_t config;
    next_default_config( &config );
    strncpy( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

    next_init( NULL, &config ); 

    next_client_t * client = next_client_create( NULL, bind_address, client_packet_received );
    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    next_client_open_session( client, server_address );

    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof( packet_data ) );

    double delta_time = 1.0 / 60.0;

    double accumulator = 0.0;

    while ( !quit )
    {
        next_client_update( client );

        next_client_send_packet( client, packet_data, sizeof( packet_data ) );
        
        next_sleep( delta_time );

        accumulator += delta_time;

        if ( accumulator > 10.0 )
        {
            const next_client_stats_t * stats = next_client_stats( client );
            next_assert( stats );
            printf( "===================================================================\n" );
            /*
            uint64_t packets_sent_client_to_server;
            uint64_t packets_sent_server_to_client;
            uint64_t packets_lost_client_to_server;
            uint64_t packets_lost_server_to_client;
            uint64_t user_flags;
            */
            printf( "flags = %" PRIx64 "\n", stats->flags );
            printf( "platform_id = %" PRIx64 "\n", stats->flags );
            printf( "connection_type = %d\n", stats->connection_type );
            printf( "multipath = %s\n", stats->multipath ? "yes" : "no" );
            printf( "committed = %s\n", stats->committed ? "yes" : "no" );
            printf( "flagged = %s\n", stats->flagged ? "yes" : "no" );
            printf( "direct_min_rtt = %.2f\n", stats->direct_min_rtt );
            printf( "direct_max_rtt = %.2f\n", stats->direct_max_rtt );
            printf( "direct_mean_rtt = %.2f\n", stats->direct_mean_rtt );
            printf( "direct_jitter = %.2f\n", stats->direct_jitter );
            printf( "direct_packet_loss = %.2f\n", stats->direct_packet_loss );
            printf( "next = %s\n", stats->next ? "yes" : "no" );
            printf( "next_min_rtt = %.2f\n", stats->next_min_rtt );
            printf( "next_max_rtt = %.2f\n", stats->next_max_rtt );
            printf( "next_mean_rtt = %.2f\n", stats->next_mean_rtt );
            printf( "next_jitter = %.2f\n", stats->next_jitter );
            printf( "next_packet_loss = %.2f\n", stats->next_packet_loss );
            printf( "next_kbps_up = %.2f\n", stats->next_kbps_up );
            printf( "next_kbps_down = %.2f\n", stats->next_kbps_down );
            printf( "===================================================================\n" );
            accumulator = 0.0;
        }
    }

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
