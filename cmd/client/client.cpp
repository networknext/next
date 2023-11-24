/*
    Network Next Accelerate. Copyright © 2017 - 2023 Network Next, Inc.

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
#include "next_platform.h"

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

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) from; (void) packet_data; (void) packet_bytes;
}

void generate_packet( uint8_t * packet_data, int & packet_bytes )
{
    packet_bytes = 100;
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
    if ( getenv( "NEXT_DELAY" ) )
    {
        next_platform_sleep( 10.0 );
    }

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL );

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );

    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    const char * connect_address = "34.30.158.110:40000"; // "127.0.0.1:30000";

    const char * connect_address_override = getenv( "NEXT_CONNECT_ADDRESS" );
    if ( connect_address_override )
    {
        connect_address = connect_address_override;
    }

    next_client_open_session( client, connect_address );

    while ( !quit )
    {
        next_client_update( client );

        if ( next_client_ready( client ) ) 
        {
            int packet_bytes = 0;
            uint8_t packet_data[NEXT_MTU];
            generate_packet( packet_data, packet_bytes );
            next_client_send_packet( client, packet_data, packet_bytes );
        }

        next_platform_sleep( 1.0 / 60.0 );

        fflush( stdout );
    }

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
