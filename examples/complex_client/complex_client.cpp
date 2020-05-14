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
#include <map>

const char * bind_address = "0.0.0.0:0";
const char * server_address = "127.0.0.1:50000";
const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";

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

struct ClientContext
{
    Allocator * allocator;
    uint32_t client_data;
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
    printf( "assert failed: ( %s ), function %s, file %s, line %d\n", condition, function, file, line );
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

void client_packet_received( next_client_t * client, void * _context, const uint8_t * packet_data, int packet_bytes )
{
    (void) client;

    ClientContext * context = (ClientContext*) _context;

    next_assert( context );
    next_assert( context->allocator != NULL );
    next_assert( context->client_data == 0x12345 );

    next_printf( NEXT_LOG_LEVEL_INFO, "client received packet from server (%d bytes)", packet_bytes );

    verify_packet( packet_data, packet_bytes );
}

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
    strncpy( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

    if ( next_init( &global_context, &config ) != NEXT_OK )
    {
        printf( "error: could not initialize network next\n" );
        return 1;
    }

    Allocator client_allocator;
    ClientContext client_context;
    client_context.allocator = &client_allocator;
    client_context.client_data = 0x12345;

    next_client_t * client = next_client_create( &client_context, bind_address, client_packet_received );
    if ( client == NULL )
    {
        printf( "error: failed to create client\n" );
        return 1;
    }

    next_client_open_session( client, server_address );

    while ( !quit )
    {
        next_client_update( client );

        const int state = next_client_state( client );

        if ( state == NEXT_CLIENT_STATE_ERROR )
        {
            printf( "error: client is in an error state\n" );
            break;
        }

        int packet_bytes = 0;
        uint8_t packet_data[NEXT_MTU];
        generate_packet( packet_data, packet_bytes );
        next_client_send_packet( client, packet_data, packet_bytes );
        
        next_sleep( 0.25 );
    }

    next_client_destroy( client );
    
    next_term();
    
    return 0;
}
