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

#ifdef _MSC_VER
#define _CRT_SECURE_NO_WARNINGS
#endif

#include "next.h"
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>

const int MaxServers = 8;
const int MaxClients = 100;

static next_server_t * servers[MaxServers];
static next_client_t * clients[MaxClients];

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) packet_data; (void) packet_bytes;
}

void server_packet_received( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) context;
    next_server_send_packet( server, from, packet_data, packet_bytes );
    if ( !next_server_session_upgraded( server, from ) )
    {
        next_server_upgrade_session( server, from, 0, 0, NULL );
    }
}

int main( int argc, char ** argv )
{
    printf( "\nWelcome to Network Next!\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );
    
    next_init( NULL, NULL );
    
    uint8_t packet_data[NEXT_MTU];
    memset( packet_data, 0, sizeof( packet_data ) );

    int iterations = 1000;
    if ( argc == 2 ) 
    {
        iterations = atoi( argv[1] );
    }

    while ( !quit )
    {
        // randomly create clients

        for ( int i = 0; i < MaxClients; ++i )
        {
            if ( clients[i] == NULL && ( rand() % 1000 ) == 0 )
            {
                clients[i] = next_client_create( NULL, client_packet_received );
                next_assert( clients[i] );
                next_printf( NEXT_LOG_LEVEL_INFO, "created client %d", i );
            }
        }

        // randomly destroy clients

        for ( int i = 0; i < MaxClients; ++i )
        {
            if ( clients[i] && ( rand() % 1500 ) == 0 )
            {
                next_client_destroy( clients[i] );
                clients[i] = NULL;
                next_printf( NEXT_LOG_LEVEL_INFO, "destroyed client %d", i );
            }
        }

        // randomly open client sessions

        for ( int i = 0; i < MaxClients; ++i )
        {
            if ( clients[i] && ( rand() % 500 ) == 0 )
            {
                int j = rand() % MaxServers;
                char server_address_string[256]; 
                sprintf( server_address_string, "127.0.0.1:%d", 50000 + j );
                next_client_open_session( clients[i], server_address_string );
            }
        }

        // randomly close client sessions

        for ( int i = 0; i < MaxClients; ++i )
        {
            if ( clients[i] && ( rand() % 750 ) == 0 )
            {
                next_client_close_session( clients[i] );
            }
        }

        // randomly create servers

        for ( int i = 0; i < MaxServers; ++i )
        {
            if ( servers[i] == NULL && ( rand() % 1000 ) == 0 )
            {
                char server_address_string[256]; 
                char bind_address_string[256];
                sprintf( server_address_string, "127.0.0.1:%d", 50000 + i );
                sprintf( bind_address_string, "0.0.0.0:%d", 50000 + i );
                servers[i] = next_server_create( NULL, server_address_string, bind_address_string, "local", server_packet_received );
                if ( servers[i] )
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "created server %d", i );
                }
            }
        }

        // randomly destroy servers

        for ( int i = 0; i < MaxServers; ++i )
        {
            if ( servers[i] && ( rand() % 1500 ) == 0 )
            {
                next_server_destroy( servers[i] );
                servers[i] = NULL;
                next_printf( NEXT_LOG_LEVEL_INFO, "destroyed server %d", i );
            }
        }

        // update clients

        for ( int i = 0; i < MaxClients; ++i )
        {
            if ( clients[i] )
            {
                next_client_update( clients[i] );
                next_client_send_packet( clients[i], packet_data, 1 + ( rand() % NEXT_MTU ) );
            }
        }

        // update servers

        for ( int i = 0; i < MaxServers; ++i )
        {
            if ( servers[i] )
            {
                next_server_update( servers[i] );
            }
        }

        // optionally quit after a number of iterations

        if ( iterations > 0 )
        {
            iterations--;
            if ( iterations == 0 )
                quit = true;
        }

        next_sleep( 0.01 );
    }
    
    // destroy clients

    for ( int i = 0; i < MaxClients; ++i )
    {
        if ( clients[i] )
        {
            next_client_destroy( clients[i] );
            next_printf( NEXT_LOG_LEVEL_INFO, "destroyed client %d", i );
        }
    }

    // destroy servers

    for ( int i = 0; i < MaxServers; ++i )
    {
        if ( servers[i] )
        {
            next_server_destroy( servers[i] );
            next_printf( NEXT_LOG_LEVEL_INFO, "destroyed server %d", i );
        }
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "done." );
    
    next_term();

    return 0;
}
