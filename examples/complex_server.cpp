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
#include <stdarg.h>
#include <string.h>
#include <inttypes.h>
#include <map>

const char * bind_address = "0.0.0.0:50000";
const char * server_address = "127.0.0.1:50000";
const char * server_datacenter = "local";
const char * backend_hostname = "dev.spacecats.net";
const char * customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";

// -------------------------------------------------------------

struct AllocatorEntry
{
    // ...
};

class Allocator
{
    int64_t num_allocations;
    next_mutex_t mutex;
    std::map<void*, AllocatorEntry*> entries;

public:

    Allocator()
    {
        int result = next_mutex_create( &mutex );
		(void) result;
        next_assert( result == NEXT_OK );
        num_allocations = 0;
    }

    ~Allocator()
    {
        next_mutex_destroy( &mutex );
        next_assert( num_allocations == 0 );
        next_assert( entries.size() == 0 );
    }

    void * Alloc( size_t size )
    {
        next_mutex_guard( &mutex );
        void * pointer = malloc( size );
        next_assert( pointer );
        next_assert( entries[pointer] == NULL );
        AllocatorEntry * entry = new AllocatorEntry();
        entries[pointer] = entry;
        num_allocations++;
        return pointer;
    }

    void Free( void * pointer )
    {
        next_mutex_guard( &mutex );
        next_assert( pointer );
        next_assert( num_allocations > 0 );
        std::map<void*, AllocatorEntry*>::iterator itor = entries.find( pointer );
        next_assert( itor != entries.end() );
        entries.erase( itor );
        num_allocations--;
        free( pointer );
    }
};

Allocator global_allocator;

struct Context
{
    Allocator * allocator;
};

struct ServerContext
{
    Allocator * allocator;
    uint32_t server_data;
};

void * malloc_function( void * _context, size_t bytes )
{
    Context * context = (Context*) _context;
    next_assert( context );
    next_assert( context->allocator );
    return context->allocator->Alloc( bytes );
}

void free_function( void * _context, void * p )
{
    Context * context = (Context*) _context;
    next_assert( context );
    next_assert( context->allocator );
    return context->allocator->Free( p );
}

// -------------------------------------------------------------

extern const char * log_level_string( int level )
{
    if ( level == NEXT_LOG_LEVEL_DEBUG )
        return "debug";
    else if ( level == NEXT_LOG_LEVEL_INFO )
        return "info";
    else if ( level == NEXT_LOG_LEVEL_ERROR )
        return "error";
    else if ( level == NEXT_LOG_LEVEL_WARN )
        return "warning";
    else
        return "???";
}

void log_function( int level, const char * format, ... ) 
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    if ( level != NEXT_LOG_LEVEL_NONE )
    {
        const char * level_string = log_level_string( level );
        printf( "%.2f: %s: %s\n", next_time(), level_string, buffer );
    }
    else
    {
        printf( "%s\n", buffer );
    }
    va_end( args );
    fflush( stdout );
}

void assert_function( const char * condition, const char * function, const char * file, int line )
{
    next_printf( NEXT_LOG_LEVEL_NONE, "assert failed: ( %s ), function %s, file %s, line %d\n", condition, function, file, line );
    fflush( stdout );
    #if defined(_MSC_VER)
        __debugbreak();
    #elif defined(__ORBIS__)
        __builtin_trap();
    #elif defined(__clang__)
        __builtin_debugtrap();
    #elif defined(__GNUC__)
        __builtin_trap();
    #elif defined(linux) || defined(__linux) || defined(__linux__) || defined(__APPLE__)
        raise(SIGTRAP);
    #else
        #error "asserts not supported on this platform!"
    #endif
}

// -------------------------------------------------------------

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
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

void server_packet_received( next_server_t * server, void * _context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    ServerContext * context = (ServerContext*) _context;

    (void) context;

    next_assert( context );
    next_assert( context->allocator != NULL );
    next_assert( context->server_data == 0x12345678 );

    verify_packet( packet_data, packet_bytes );

    next_server_send_packet( server, from, packet_data, packet_bytes );
    
    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];

    next_printf( NEXT_LOG_LEVEL_INFO, "server received packet from client %s (%d bytes)", next_address_to_string( from, address_buffer ), packet_bytes );

    if ( !next_server_session_upgraded( server, from ) )
    {
        if ( next_server_upgrade_session( server, from, NULL ) )
        {
            next_server_tag_session( server, from, "pro" );
        }
    }
}

#if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS
#define strncpy_s strncpy
#endif // #if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    Context global_context;
    global_context.allocator = &global_allocator;

    next_log_level( NEXT_LOG_LEVEL_INFO );

    next_log_function( log_function );

    next_assert_function( assert_function );

    next_allocator( malloc_function, free_function );

    next_config_t config;
    next_default_config( &config );
    strncpy_s( config.hostname, backend_hostname, sizeof(config.hostname) - 1 );
    strncpy_s( config.customer_private_key, customer_private_key, sizeof(config.customer_private_key) - 1 );

    if ( next_init( &global_context, &config ) != NEXT_OK )
    {
        printf( "error: could not initialize network next\n" );
        return 1;
    }

    Allocator server_allocator;
    ServerContext server_context;
    server_context.allocator = &server_allocator;
    server_context.server_data = 0x12345678;

    next_server_t * server = next_server_create( &server_context, server_address, bind_address, server_datacenter, server_packet_received, NULL );
    if ( server == NULL )
    {
        printf( "error: failed to create server\n" );
        return 1;
    }
    
    uint16_t server_port = next_server_port( server );

    next_printf( NEXT_LOG_LEVEL_INFO, "server port is %d", server_port );

    while ( !quit )
    {
        next_server_update( server );

        next_sleep( 1.0 / 60.0 );
    }
    
    next_server_destroy( server );
    
    next_term();

    return 0;
}
