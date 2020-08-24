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
#include <stdlib.h>
#include <string.h>

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;

    next_server_send_packet( server, from, packet_data, packet_bytes );

    // todo: disabled temporarily
    /*
    if ( !next_server_session_upgraded( server, from ) )
    {
        next_server_upgrade_session( server, from, 0 );
    }
    */
}

int main()
{
    printf( "\nWelcome to Network Next!\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL ); 
    
    next_server_t * server = next_server_create( NULL, "127.0.0.1:32202", "0.0.0.0:32202", "local", server_packet_received, NULL );

    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
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
