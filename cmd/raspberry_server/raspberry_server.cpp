/*
    Network Next SDK. Copyright © 2017 - 2023 Network Next, Inc.

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

struct thread_data_t
{
    const char * server_address;
    const char * raspberry_backend_address;
};

next_platform_thread_return_t NEXT_PLATFORM_THREAD_FUNC server_update_thread( void * data )
{
    next_assert( data );

    thread_data_t * thread_data = (thread_data_t*) data;

    const char * server_address = thread_data->server_address;
    const char * raspberry_backend_address = thread_data->raspberry_backend_address;
    
    char command_line[1024];
    snprintf( command_line, sizeof(command_line), "curl -d \"%s\" -X POST http://%s/server_update --max-time 2 2>/dev/null", server_address, raspberry_backend_address );

    while ( !quit )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "sending server update: %s", server_address );

        FILE * file = popen( command_line, "r" );
        if ( file )
            pclose( file );

        next_sleep( 10.0 );
    }

    NEXT_PLATFORM_THREAD_RETURN();
}

void send_server_updates_to_raspberry_backend( thread_data_t * thread_data )
{
    next_platform_thread_t * thread = next_platform_thread_create( NULL, server_update_thread, thread_data );
    next_assert( thread );
    (void) thread;
}

extern const char * next_platform_getenv( const char * name );

int main()
{
    printf( "\nRaspberry Server\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_init( NULL, NULL );

    char raspberry_backend_address[1024];
    next_copy_string( raspberry_backend_address, "127.0.0.1:40100", sizeof( raspberry_backend_address ) );
    const char * raspberry_backend_address_override = next_platform_getenv( "RASPBERRY_BACKEND_ADDRESS" );
    if ( raspberry_backend_address_override )
    {
    	next_copy_string( raspberry_backend_address, raspberry_backend_address_override, sizeof(raspberry_backend_address) );
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "raspberry backend address: %s", raspberry_backend_address );

    char server_address[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_copy_string( server_address, "127.0.0.1", sizeof(server_address) );

    // if we are running in google cloud, detect google cloud public IP address

    FILE * file = popen( "curl http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H \"Metadata-Flavor: Google\" --max-time 10 -vs 2>/dev/null", "r" );

    char buffer[1024];

    while ( file && fgets( buffer, sizeof(buffer), file ) != NULL )
    {
		next_address_t address;
		if ( next_address_parse( &address, buffer ) == NEXT_OK )
		{
			next_address_to_string( &address, server_address );
			break;
		}
    }

    if ( file )
    {
    	pclose( file );
    }

    // start server

    next_server_t * server = next_server_create( NULL, server_address, "0.0.0.0", "local", server_packet_received );

    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }

    // now that we know our port number, combine with the address and we have our public address

    int server_port = next_server_port( server );
    char public_address[256];
    snprintf( public_address, sizeof(public_address), "%s:%d", server_address, server_port );
    next_printf( NEXT_LOG_LEVEL_INFO, "public address is: %s", public_address );

    // send server updates to the raspberry backend in the background

    thread_data_t thread_data;

    thread_data.server_address = public_address;
    thread_data.raspberry_backend_address = raspberry_backend_address;

    send_server_updates_to_raspberry_backend( &thread_data );

    // main loop for server

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
