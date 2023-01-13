/*
    Network Next SDK. Copyright © 2017 - 2020 Network Next, Inc.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following
    conditions are met:

    1. Redistributions of must retain the above copyright notice, this list of conditions and the following disclaimer.

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

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;

    if ( packet_bytes != 8 )
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
    else
    {
        next_server_stats_t stats;
        bool session_exists = next_server_stats( server, from, &stats );

        if ( rand() % 2500 == 0 && next_server_session_upgraded( server, from ) && session_exists )
        {
            int num_server_events = rand() % 64;
            uint64_t server_events = 0;

            for ( int i = 0; i < num_server_events; ++i )
            {
                server_events |= 1 << (rand() % 64);
            }

            next_server_event( server, from, server_events );
        }
    }
}

int main()
{
    printf( "\nRaspberry Pi Server\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL );

    next_server_t * server = next_server_create( NULL, "127.0.0.1:50000", "0.0.0.0:50000", "local", server_packet_received );

    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

    while ( !quit )
    {
        next_server_update( server );

        if ( next_server_direct_only( server ) )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "detected server is in direct only mode. restarting..." );
            break;
        }

        next_sleep( 1.0 / 60.0 );
    }

    next_server_flush( server );

    next_server_destroy( server );

    next_term();

    return 0;
}
