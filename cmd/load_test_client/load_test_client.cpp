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
#include <sstream>

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

    next_init( NULL, NULL );

    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received, NULL );

    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    const char* server_ip = std::getenv("SERVER_IP");
    if (!server_ip) {
        printf( "error: server ip env var not defined\n" );
        return 1;
    }

    char * cores_str = std::getenv("CORES");
    if (!cores_str) {
        printf( "error: cores env var not defined\n" );
        return 1;
    }

    int cores = std::atoi(cores_str);

    std::stringstream server_addrs_ss[cores];

    for (int i = 0; i < cores; i++) {
      server_addrs_ss[i] << server_ip << ":" << std::to_string(50000 + i);
    }

    next_client_open_session( client, server_addrs_ss[rand() % cores].str().c_str() );

    uint8_t packet_data[32];
    memset( packet_data, 0, sizeof( packet_data ) );

    while ( !quit )
    {
        next_client_update( client );

        next_client_send_packet( client, packet_data, sizeof( packet_data ) );

        next_sleep( 6.0f / 60.0f );
    }

    next_client_destroy( client );

    next_term();

    return 0;
}
