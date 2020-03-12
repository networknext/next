/*
    Network Next: $(NEXT_VERSION_FULL)
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#include "relay_internal.h"
#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <memory.h>
#include <ctime>

#if defined(linux) || defined(__linux) || defined(__linux__) || defined(__APPLE__)
#include <signal.h> // for asserts
#endif

uint8_t NEXT_KEY_MASTER[] = {0x49, 0x2e, 0x79, 0x74, 0x49, 0x7d, 0x9d, 0x34, 0xa7, 0x55, 0x50, 0xeb, 0xab, 0x03, 0xde, 0xa9, 0x1b, 0xff, 0x61, 0xc6, 0x0e, 0x65, 0x92, 0xd7, 0x09, 0x64, 0xe9, 0x34, 0x12, 0x32, 0x5f, 0x46};

struct next_t
{
    bool initialized;
};

void resolver_update( resolver_t * resolver );

void resolver_init( resolver_t * resolver, const char * address )
{
    memset( resolver, 0, sizeof( *resolver ) );
    next_mutex_init( &resolver->mutex );

    snprintf( resolver->address_string, sizeof(resolver->address_string), "%s", address );
    resolver->resolve_needed = true;
    resolver_update( resolver );
}

const next_address_t * resolver_address( resolver_t * resolver )
{
    return &resolver->address;
}

const char * resolver_address_string( resolver_t * resolver )
{
    return resolver->address_string;
}

static next_thread_return_t NEXT_THREAD_FUNC resolver_thread( void * ctx )
{
    resolver_t * resolver = (resolver_t *)( ctx );
    next_address_t address;
    while ( true )
    {
        if ( next_address_resolve( resolver->address_string, &address ) == NEXT_OK )
        {
            // success!
            next_mutex_acquire( &resolver->mutex );
            resolver->thread_local_address = address;
            next_mutex_release( &resolver->mutex );
            break;
        }
        else
        {
            // failed; try again?
            next_mutex_acquire( &resolver->mutex );
            bool retry = resolver->resolving;
            next_mutex_release( &resolver->mutex );
            if ( retry )
            {
                next_printf( NEXT_LOG_LEVEL_WARN, "failed to resolve master address, retrying" );
                next_sleep( 2.0 );
            }
            else
            {
                break;
            }
        }
    }
    NEXT_THREAD_RETURN();
}

void resolver_update( resolver_t * resolver )
{
    if ( resolver->resolve_needed )
    {
        if ( resolver->resolving )
        {
            next_mutex_acquire( &resolver->mutex );
            if ( resolver->thread_local_address.type == NEXT_ADDRESS_NONE )
            {
                // still resolving
                next_mutex_release( &resolver->mutex );
            }
            else
            {
                // done resolving
                resolver->address = resolver->thread_local_address;
                resolver->resolving = false;
                next_mutex_release( &resolver->mutex );
                next_thread_join( &resolver->thread );

                resolver->resolve_last = next_time();

                char str[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_address_to_string( &resolver->address, str );
                next_printf( NEXT_LOG_LEVEL_DEBUG, "resolved %s to %s", resolver->address_string, str );
            }
        }
        else if ( resolver->address.type == NEXT_ADDRESS_NONE || next_time() - resolver->resolve_last > 600 * NEXT_ONE_SECOND_NS )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "resolving %s...", resolver->address_string );
            resolver->thread_local_address.type = NEXT_ADDRESS_NONE;
            resolver->resolving = true;
            if ( next_thread_create( &resolver->thread, resolver_thread, resolver ) != NEXT_OK )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "failed to start resolve thread" );
                resolver->resolving = false;
            }
        }
    }
}

void resolver_destroy( resolver_t * resolver )
{
    if ( resolver->resolving )
    {
        next_mutex_acquire( &resolver->mutex );
        resolver->resolving = false;
        next_mutex_release( &resolver->mutex );
        next_thread_join( &resolver->thread );
        resolver->resolve_needed = false;
    }

    next_mutex_destroy( &resolver->mutex );
}

static next_t next;

int next_internal_init()
{
    if ( next.initialized )
        return NEXT_OK;

    if ( !next_platform_init() )
        return NEXT_ERROR;

    next.initialized = true;

    return NEXT_OK;
}

int next_resolve_master_address( const char * address )
{
    if ( strlen( address ) == 0 )
        return NEXT_ERROR;

    return NEXT_OK;
}

void next_internal_term()
{
    if ( !next.initialized )
        return;

    next_platform_term();

    next.initialized = false;
}

static void next_default_assert_function( const char * condition, const char * function, const char * file, int line )
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

static int log_level = 0;

int next_log_level()
{
    return log_level;
}

const char * next_log_level_str( int level )
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
        return "log";
}

void next_default_print_function( int level, const char * format, ... ) 
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    const char * level_str = next_log_level_str( level );

    printf( "%s: %s\n", level_str, buffer );
    va_end( args );
    fflush( stdout );
}

static void (*next_print_function)( int level, const char *, ... ) = next_default_print_function;

void (*next_assert_function)( const char *, const char *, const char * file, int line ) = next_default_assert_function;

void next_set_log_level( int level )
{
    log_level = level;
}

void next_set_print_function( void (*function)( int level, const char *, ... ) )
{
    next_assert( function );
    next_print_function = function;
}

void next_set_assert_function( void (*function)( const char *, const char *, const char * file, int line ) )
{
    next_assert_function = function;
}

void next_flow_log( int level, uint64_t flow_id, uint8_t flow_version, const char * format, ... )
{
    if ( level > log_level )
        return;
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    va_end( args );
    next_print_function( level, "[%" PRIx64 ":%hhu] %s", flow_id, flow_version, buffer );
}

void next_printf( int level, const char * format, ... ) 
{
    if ( level > log_level )
        return;
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    next_print_function( level, "%s", buffer );
    va_end( args );
}

// ------------------------------------------------------------------

int64_t next_time()
{
    return next_platform_time();
}

uint16_t next_ntohs( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

uint16_t next_htons( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

// ----------------------------------------------------------------

bool next_address_t::operator==( const next_address_t & other ) const
{
    if ( type != other.type )
        return false;

    if ( port != other.port )
        return false;

    if ( type == NEXT_ADDRESS_IPV4 )
    {
        for ( int i = 0; i < 4; i++ )
        {
            if ( data.ipv4[i] != other.data.ipv4[i] )
                return false;
        }
    }
    else
    {
        for ( int i = 0; i < 8; i++ )
        {
            if ( data.ipv6[i] != other.data.ipv6[i] )
                return false;
        }
    }
    return true;
}

int next_address_parse( next_address_t * address, const char * address_string_in )
{
    next_assert( address );
    next_assert( address_string_in );

    if ( !address )
        return NEXT_ERROR;

    if ( !address_string_in )
        return NEXT_ERROR;

    memset( address, 0, sizeof( next_address_t ) );

    // first try to parse the string as an IPv6 address:
    // 1. if the first character is '[' then it's probably an ipv6 in form "[addr6]:portnum"
    // 2. otherwise try to parse as a raw IPv6 address using inet_pton

    char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH + NEXT_ADDRESS_BUFFER_SAFETY*2];

    char * address_string = buffer + NEXT_ADDRESS_BUFFER_SAFETY;
    strncpy( address_string, address_string_in, NEXT_MAX_ADDRESS_STRING_LENGTH - 1 );
    address_string[NEXT_MAX_ADDRESS_STRING_LENGTH-1] = '\0';

    int address_string_length = (int) strlen( address_string );

    if ( address_string[0] == '[' )
    {
        const int base_index = address_string_length - 1;
        
        // note: no need to search past 6 characters as ":65535" is longest possible port value
        for ( int i = 0; i < 6; ++i )
        {
            const int index = base_index - i;
            if ( index < 3 )
            {
                return NEXT_ERROR;
            }
            if ( address_string[index] == ':' )
            {
                address->port = (uint16_t) ( atoi( &address_string[index + 1] ) );
                address_string[index-1] = '\0';
            }
        }
        address_string += 1;
    }

    uint16_t addr6[8];
    if ( next_inet_pton6( address_string, addr6 ) == NEXT_OK )
    {
        address->type = NEXT_ADDRESS_IPV6;
        for ( int i = 0; i < 8; ++i )
        {
            address->data.ipv6[i] = next_ntohs( addr6[i] );
        }
        return NEXT_OK;
    }

    // otherwise it's probably an IPv4 address:
    // 1. look for ":portnum", if found save the portnum and strip it out
    // 2. parse remaining ipv4 address via inet_pton

    address_string_length = (int) strlen( address_string );
    const int base_index = address_string_length - 1;
    for ( int i = 0; i < 6; ++i )
    {
        const int index = base_index - i;
        if ( index < 0 )
            break;
        if ( address_string[index] == ':' )
        {
            address->port = (uint16_t)( atoi( &address_string[index + 1] ) );
            address_string[index] = '\0';
        }
    }

    uint32_t addr4;
    if ( next_inet_pton4( address_string, &addr4 ) == NEXT_OK )
    {
        address->type = NEXT_ADDRESS_IPV4;
        address->data.ipv4[3] = (uint8_t) ( ( addr4 & 0xFF000000 ) >> 24 );
        address->data.ipv4[2] = (uint8_t) ( ( addr4 & 0x00FF0000 ) >> 16 );
        address->data.ipv4[1] = (uint8_t) ( ( addr4 & 0x0000FF00 ) >> 8  );
        address->data.ipv4[0] = (uint8_t) ( ( addr4 & 0x000000FF )     );
        return NEXT_OK;
    }

    return NEXT_ERROR;
}

char * next_address_to_string( const next_address_t * address, char * buffer )
{
    next_assert( buffer );

    if ( address->type == NEXT_ADDRESS_IPV6 )
    {
#if defined(WINVER) && WINVER <= 0x0502
        // ipv6 not supported
        buffer[0] = '\0';
        return buffer;
#else
        uint16_t ipv6_network_order[8];
        for ( int i = 0; i < 8; ++i )
            ipv6_network_order[i] = next_htons( address->data.ipv6[i] );
        char address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_inet_ntop6( ipv6_network_order, address_string, sizeof( address_string ) );
        if ( address->port == 0 )
        {
            strncpy( buffer, address_string, NEXT_MAX_ADDRESS_STRING_LENGTH );
            return buffer;
        }
        else
        {
            snprintf( buffer, NEXT_MAX_ADDRESS_STRING_LENGTH, "[%s]:%d", address_string, address->port );
            return buffer;
        }
#endif
    }
    else if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        if ( address->port != 0 )
        {
            snprintf( buffer, 
                      NEXT_MAX_ADDRESS_STRING_LENGTH, 
                      "%d.%d.%d.%d:%d", 
                      address->data.ipv4[0], 
                      address->data.ipv4[1], 
                      address->data.ipv4[2], 
                      address->data.ipv4[3], 
                      address->port );
        }
        else
        {
            snprintf( buffer, 
                      NEXT_MAX_ADDRESS_STRING_LENGTH, 
                      "%d.%d.%d.%d", 
                      address->data.ipv4[0], 
                      address->data.ipv4[1], 
                      address->data.ipv4[2], 
                      address->data.ipv4[3] );
        }
        return buffer;
    }
    else
    {
        snprintf( buffer, NEXT_MAX_ADDRESS_STRING_LENGTH, "%s", "NONE" );
        return buffer;
    }
}

int next_address_equal( const next_address_t * a, const next_address_t * b )
{
    next_assert( a );
    next_assert( b );

    if ( a->type != b->type )
        return 0;

    if ( a->port != b->port )
        return 0;

    if ( a->type == NEXT_ADDRESS_IPV4 )
    {
        for ( int i = 0; i < 4; ++i )
        {
            if ( a->data.ipv4[i] != b->data.ipv4[i] ) 
                return 0;
        }
    }
    else if ( a->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int i = 0; i < 8; ++i )
        {
            if ( a->data.ipv6[i] != b->data.ipv6[i] ) 
                return 0;
        }
    }
    else {
        return 0;
    }

    return 1;
}

int next_ip_equal( const next_address_t * a, const next_address_t * b )
{
    next_assert( a );
    next_assert( b );

    if ( a->type != b->type )
        return 0;

    if ( a->type == NEXT_ADDRESS_IPV4 )
    {
        for ( int i = 0; i < 4; ++i )
        {
            if ( a->data.ipv4[i] != b->data.ipv4[i] ) 
                return 0;
        }
    }
    else if ( a->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int i = 0; i < 8; ++i )
        {
            if ( a->data.ipv6[i] != b->data.ipv6[i] ) 
                return 0;
        }
    }
    else {
        return 0;
    }

    return 1;
}

void next_print_bytes( const char * label, const uint8_t * data, int data_bytes )
{
    printf( "%s: ", label );
    for ( int i = 0; i < data_bytes; ++i )
    {
        printf( "0x%02x,", (int) data[i] );
    }
    printf( " (%d bytes)\n", data_bytes );
}

// ----------------------------------------------------------------

void next_write_uint8( uint8_t ** p, uint8_t value )
{
    **p = value;
    ++(*p);
}

void next_write_uint16( uint8_t ** p, uint16_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = value >> 8;
    *p += 2;
}

void next_write_uint32( uint8_t ** p, uint32_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = value >> 24;
    *p += 4;
}

void next_write_float32( uint8_t ** p, float value )
{
    uint32_t value_int;
    char * p_value = (char *)(&value);
    char * p_value_int = (char *)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint32_t));
    next_write_uint32( p, value_int);
}

void next_write_uint64( uint8_t ** p, uint64_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = ( value >> 24 ) & 0xFF;
    (*p)[4] = ( value >> 32 ) & 0xFF;
    (*p)[5] = ( value >> 40 ) & 0xFF;
    (*p)[6] = ( value >> 48 ) & 0xFF;
    (*p)[7] = value >> 56;
    *p += 8;
}

void next_write_float64( uint8_t ** p, double value )
{
    uint64_t value_int;
    char * p_value = (char *)(&value);
    char * p_value_int = (char *)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint64_t));
    next_write_uint64( p, value_int);
}

void next_write_bytes( uint8_t ** p, uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        next_write_uint8( p, byte_array[i] );
    }
}

uint8_t next_read_uint8( uint8_t ** p )
{
    uint8_t value = **p;
    ++(*p);
    return value;
}

uint16_t next_read_uint16( uint8_t ** p )
{
    uint16_t value;
    value = (*p)[0];
    value |= ( ( (uint16_t)( (*p)[1] ) ) << 8 );
    *p += 2;
    return value;
}

uint32_t next_read_uint32( uint8_t ** p )
{
    uint32_t value;
    value  = (*p)[0];
    value |= ( ( (uint32_t)( (*p)[1] ) ) << 8 );
    value |= ( ( (uint32_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint32_t)( (*p)[3] ) ) << 24 );
    *p += 4;
    return value;
}

uint64_t next_read_uint64( uint8_t ** p )
{
    uint64_t value;
    value  = (*p)[0];
    value |= ( ( (uint64_t)( (*p)[1] ) ) << 8  );
    value |= ( ( (uint64_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint64_t)( (*p)[3] ) ) << 24 );
    value |= ( ( (uint64_t)( (*p)[4] ) ) << 32 );
    value |= ( ( (uint64_t)( (*p)[5] ) ) << 40 );
    value |= ( ( (uint64_t)( (*p)[6] ) ) << 48 );
    value |= ( ( (uint64_t)( (*p)[7] ) ) << 56 );
    *p += 8;
    return value;
}

float next_read_float32( uint8_t ** p )
{
    uint32_t value_int = next_read_uint32( p );
    float value_float;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy(pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

double next_read_float64( uint8_t ** p )
{
    uint64_t value_int = next_read_uint64( p );
    double value_float;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy(pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

void next_read_bytes( uint8_t ** p, uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        byte_array[i] = next_read_uint8( p );
    }
}

// ----------------------------------------------------------------

typedef uint64_t next_fnv1a_64_t;

void next_fnv1a_64_init( next_fnv1a_64_t * fnv )
{
    *fnv = 0xCBF29CE484222325;
}

void next_fnv1a_64_write( next_fnv1a_64_t * fnv, const uint8_t * data, size_t size )
{
    for ( size_t i = 0; i < size; i++ )
    {
        (*fnv) ^= data[i];
        (*fnv) *= 0x00000100000001B3;
    }
}

uint64_t next_fnv1a_64_finalize( next_fnv1a_64_t * fnv )
{
    return *fnv;
}

uint64_t next_relay_id( const char * name )
{
    next_fnv1a_64_t fnv;
    next_fnv1a_64_init( &fnv );
    next_fnv1a_64_write( &fnv, (uint8_t*)( name ), strlen( name ) );
    return next_fnv1a_64_finalize( &fnv );
}

// ----------------------------------------------------------------

#if SODIUM_LIBRARY_VERSION_MAJOR > 7 || ( SODIUM_LIBRARY_VERSION_MAJOR && SODIUM_LIBRARY_VERSION_MINOR >= 3 )
#define SODIUM_SUPPORTS_OVERLAPPING_BUFFERS 1
#endif

void next_generate_keypair( uint8_t * public_key, uint8_t * private_key )
{
    next_assert( public_key );
    next_assert( private_key );
    crypto_box_keypair( public_key, private_key );
}

int next_encrypt_aead( uint8_t * message, uint64_t message_length, 
                       uint8_t * additional, uint64_t additional_length,
                       uint8_t * nonce,
                       uint8_t * key )
{
    unsigned long long encrypted_length;

    #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

        int result = crypto_aead_chacha20poly1305_ietf_encrypt( message, &encrypted_length,
                                                                message, (unsigned long long) message_length,
                                                                additional, (unsigned long long) additional_length,
                                                                NULL, nonce, key );
    
        if ( result != 0 )
            return NEXT_ERROR;

    #else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

        uint8_t * temp = (uint8_t *)( alloca( message_length + NEXT_SYMMETRIC_MAC_BYTES ) );

        int result = crypto_aead_chacha20poly1305_ietf_encrypt( temp, &encrypted_length,
                                                                message, (unsigned long long) message_length,
                                                                additional, (unsigned long long) additional_length,
                                                                NULL, nonce, key );
        
        if ( result == 0 )
        {
            memcpy( message, temp, message_length + NEXT_SYMMETRIC_MAC_BYTES );
        }
        else
        {
            return NEXT_ERROR;
        }
    

    #endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    next_assert( encrypted_length == message_length + NEXT_SYMMETRIC_MAC_BYTES );

    return NEXT_OK;
}

int next_decrypt_aead( uint8_t * message, uint64_t message_length, 
                       uint8_t * additional, uint64_t additional_length,
                       uint8_t * nonce,
                       uint8_t * key )
{
    unsigned long long decrypted_length;

    #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

        int result = crypto_aead_chacha20poly1305_ietf_decrypt( message, &decrypted_length,
                                                                NULL,
                                                                message, (unsigned long long) message_length,
                                                                additional, (unsigned long long) additional_length,
                                                                nonce, key );

        if ( result != 0 )
            return NEXT_ERROR;

    #else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

        uint8_t * temp = (uint8_t *)( alloca( message_length ) );

        int result = crypto_aead_chacha20poly1305_ietf_decrypt( temp, &decrypted_length,
                                                                NULL,
                                                                message, (unsigned long long) message_length,
                                                                additional, (unsigned long long) additional_length,
                                                                nonce, key );
        
        if ( result == 0 )
        {
            memcpy( message, temp, decrypted_length );
        }
        else
        {
            return NEXT_ERROR;
        }
    

    #endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    next_assert( decrypted_length == message_length - NEXT_SYMMETRIC_MAC_BYTES );

    return NEXT_OK;
}

// ---------------------------------------------------------------

void next_write_address( uint8_t ** buffer, next_address_t * address )
{
    next_assert( buffer );
    next_assert( *buffer );
    next_assert( address );

    uint8_t * start = *buffer;

    (void) buffer;

    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        next_write_uint8( buffer, NEXT_ADDRESS_IPV4 );
        for ( int i = 0; i < 4; ++i )
        {
            next_write_uint8( buffer, address->data.ipv4[i] );
        }
        next_write_uint16( buffer, address->port );
        for ( int i = 0; i < 12; ++i )
        {
            next_write_uint8( buffer, 0 );
        }
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        next_write_uint8( buffer, NEXT_ADDRESS_IPV6 );
        for ( int i = 0; i < 8; ++i )
        {
            next_write_uint16( buffer, address->data.ipv6[i] );
        }
        next_write_uint16( buffer, address->port );
    }
    else
    {
        for ( int i = 0; i < NEXT_ADDRESS_BYTES; ++i )
        {
            next_write_uint8( buffer, 0 );
        }
    }

    (void) start;

    next_assert( *buffer - start == NEXT_ADDRESS_BYTES );
}

void next_read_address( uint8_t ** buffer, next_address_t * address )
{
    uint8_t * start = *buffer;

    address->type = next_read_uint8( buffer );

    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        for ( int j = 0; j < 4; ++j )
        {
            address->data.ipv4[j] = next_read_uint8( buffer );
        }
        address->port = next_read_uint16( buffer );
        for ( int i = 0; i < 12; ++i )
        {
            uint8_t dummy = next_read_uint8( buffer ); (void) dummy;
        }
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int j = 0; j < 8; ++j )
        {
            address->data.ipv6[j] = next_read_uint16( buffer );
        }
        address->port = next_read_uint16( buffer );
    }
    else
    {
        for ( int i = 0; i < NEXT_ADDRESS_BYTES - 1; ++i )
        {
            uint8_t dummy = next_read_uint8( buffer ); (void) dummy;
        }
    }

    (void) start;

    next_assert( *buffer - start == NEXT_ADDRESS_BYTES );
}

// -----------------------------------------------------------

void next_write_flow_token( next_flow_token_t * token, uint8_t * buffer, int buffer_length )
{
    (void) buffer_length;

    next_assert( token );
    next_assert( buffer );
    next_assert( buffer_length >= NEXT_FLOW_TOKEN_BYTES );

    uint8_t * start = buffer;

    (void) start;

    next_write_uint64( &buffer, token->expire_timestamp );
    next_write_uint64( &buffer, token->flow_id );
    next_write_uint8( &buffer, token->flow_version );
    next_write_uint8( &buffer, token->flow_flags );
    next_write_uint32( &buffer, token->kbps_up );
    next_write_uint32( &buffer, token->kbps_down );
    next_write_address( &buffer, &token->next_address );
    next_write_bytes( &buffer, token->private_key, NEXT_SYMMETRIC_KEY_BYTES );

    next_assert( buffer - start == NEXT_FLOW_TOKEN_BYTES );
}

void next_read_flow_token( next_flow_token_t * token, uint8_t * buffer )
{
    next_assert( token );
    next_assert( buffer );

    uint8_t * start = buffer;

    (void) start;

    token->expire_timestamp = next_read_uint64( &buffer );
    token->flow_id = next_read_uint64( &buffer );
    token->flow_version = next_read_uint8( &buffer );
    token->flow_flags = next_read_uint8( &buffer );
    token->kbps_up = next_read_uint32( &buffer );
    token->kbps_down = next_read_uint32( &buffer );
    next_read_address( &buffer, &token->next_address );
    next_read_bytes( &buffer, token->private_key, NEXT_SYMMETRIC_KEY_BYTES );
    next_assert( buffer - start == NEXT_FLOW_TOKEN_BYTES );
}

int next_encrypt_flow_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length )
{
    next_assert( sender_private_key );
    next_assert( receiver_public_key );
    next_assert( buffer );
    next_assert( buffer_length >= (int) ( NEXT_FLOW_TOKEN_BYTES + crypto_box_MACBYTES ) );

    (void) buffer_length;

#if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    if ( crypto_box_easy( buffer, buffer, NEXT_FLOW_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key ) != 0 )
    {
        return NEXT_ERROR;
    }

#else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    #error this version of sodium does not suppert overlapping buffers. please upgrade your libsodium!

#endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    return NEXT_OK;
}

int next_decrypt_flow_token( uint8_t * sender_public_key, uint8_t * receiver_private_key, uint8_t * nonce, uint8_t * buffer )
{
    next_assert( sender_public_key );
    next_assert( receiver_private_key );
    next_assert( buffer );

#if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    if ( crypto_box_open_easy( buffer, buffer, NEXT_FLOW_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key ) != 0 )
    {
        return NEXT_ERROR;
    }

#else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    #error this version of sodium does not suppert overlapping buffers. please upgrade your libsodium!

#endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    return NEXT_OK;
}

int next_write_encrypted_flow_token( uint8_t ** buffer, next_flow_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( sender_private_key );
    next_assert( receiver_public_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    next_random_bytes( nonce, crypto_box_NONCEBYTES );

    uint8_t * start = *buffer;

    next_write_bytes( buffer, nonce, crypto_box_NONCEBYTES );

    next_write_flow_token( token, *buffer, NEXT_FLOW_TOKEN_BYTES );

    if ( next_encrypt_flow_token( sender_private_key, receiver_public_key, nonce, *buffer, NEXT_FLOW_TOKEN_BYTES + crypto_box_NONCEBYTES ) != NEXT_OK )
        return NEXT_ERROR;

    *buffer += NEXT_FLOW_TOKEN_BYTES + crypto_box_MACBYTES;

    (void) start;

    next_assert( ( *buffer - start ) == NEXT_ENCRYPTED_FLOW_TOKEN_BYTES );

    return NEXT_OK;
}

int next_read_encrypted_flow_token( uint8_t ** buffer, next_flow_token_t * token, uint8_t * sender_public_key, uint8_t * receiver_private_key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( sender_public_key );
    next_assert( receiver_private_key );

    uint8_t * nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if ( next_decrypt_flow_token( sender_public_key, receiver_private_key, nonce, *buffer ) != NEXT_OK )
    {
        return NEXT_ERROR;
    }

    next_read_flow_token( token, *buffer );

    *buffer += NEXT_FLOW_TOKEN_BYTES + crypto_box_MACBYTES;

    return NEXT_OK;
}

// -----------------------------------------------------------

void next_write_continue_token( next_continue_token_t * token, uint8_t * buffer, int buffer_length )
{
    (void) buffer_length;

    next_assert( token );
    next_assert( buffer );
    next_assert( buffer_length >= NEXT_CONTINUE_TOKEN_BYTES );

    uint8_t * start = buffer;

    (void) start;

    next_write_uint64( &buffer, token->expire_timestamp );
    next_write_uint64( &buffer, token->flow_id );
    next_write_uint8( &buffer, token->flow_version );
    next_write_uint8( &buffer, token->flow_flags );

    next_assert( buffer - start == NEXT_CONTINUE_TOKEN_BYTES );
}

void next_read_continue_token( next_continue_token_t * token, uint8_t * buffer )
{
    next_assert( token );
    next_assert( buffer );

    uint8_t * start = buffer;

    (void) start;

    token->expire_timestamp = next_read_uint64( &buffer );
    token->flow_id = next_read_uint64( &buffer );
    token->flow_version = next_read_uint8( &buffer );
    token->flow_flags = next_read_uint8( &buffer );

    next_assert( buffer - start == NEXT_CONTINUE_TOKEN_BYTES );
}

int next_encrypt_continue_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length )
{
    next_assert( sender_private_key );
    next_assert( receiver_public_key );
    next_assert( buffer );
    next_assert( buffer_length >= (int) ( NEXT_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES ) );

    (void) buffer_length;

#if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    if ( crypto_box_easy( buffer, buffer, NEXT_CONTINUE_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key ) != 0 )
    {
        return NEXT_ERROR;
    }

#else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    #error this version of sodium does not suppert overlapping buffers. please upgrade your libsodium!

#endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    return NEXT_OK;
}

int next_decrypt_continue_token( uint8_t * sender_public_key, uint8_t * receiver_private_key, uint8_t * nonce, uint8_t * buffer )
{
    next_assert( sender_public_key );
    next_assert( receiver_private_key );
    next_assert( buffer );

#if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    if ( crypto_box_open_easy( buffer, buffer, NEXT_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key ) != 0 )
    {
        return NEXT_ERROR;
    }

#else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    #error this version of sodium does not suppert overlapping buffers. please upgrade your libsodium!

#endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    return NEXT_OK;
}

int next_write_encrypted_continue_token( uint8_t ** buffer, next_continue_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( sender_private_key );
    next_assert( receiver_public_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    next_random_bytes( nonce, crypto_box_NONCEBYTES );

    uint8_t * start = *buffer;

    next_write_bytes( buffer, nonce, crypto_box_NONCEBYTES );

    next_write_continue_token( token, *buffer, NEXT_CONTINUE_TOKEN_BYTES );

    if ( next_encrypt_continue_token( sender_private_key, receiver_public_key, nonce, *buffer, NEXT_CONTINUE_TOKEN_BYTES + crypto_box_NONCEBYTES ) != NEXT_OK )
        return NEXT_ERROR;

    *buffer += NEXT_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    (void) start;

    next_assert( ( *buffer - start ) == NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES );

    return NEXT_OK;
}

int next_read_encrypted_continue_token( uint8_t ** buffer, next_continue_token_t * token, uint8_t * sender_public_key, uint8_t * receiver_private_key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( sender_public_key );
    next_assert( receiver_private_key );

    uint8_t * nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if ( next_decrypt_continue_token( sender_public_key, receiver_private_key, nonce, *buffer ) != NEXT_OK )
    {
        return NEXT_ERROR;
    }

    next_read_continue_token( token, *buffer );

    *buffer += NEXT_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    return NEXT_OK;
}

// -----------------------------------------------------------

void next_write_server_token( next_server_token_t * token, uint8_t * buffer, int buffer_length )
{
    (void) buffer_length;

    next_assert( token );
    next_assert( buffer );
    next_assert( buffer_length >= (int) NEXT_SERVER_TOKEN_BYTES );

    uint8_t * start = buffer;

    (void) start;

    next_write_uint64( &buffer, token->expire_timestamp );
    next_write_uint64( &buffer, token->flow_id );
    next_write_uint8( &buffer, token->flow_version );
    next_write_uint8( &buffer, token->flow_flags );

    next_assert( buffer - start == NEXT_SERVER_TOKEN_BYTES );
}

void next_read_server_token( next_server_token_t * token, uint8_t * buffer )
{
    next_assert( token );
    next_assert( buffer );

    uint8_t * start = buffer;

    (void) start;

    token->expire_timestamp = next_read_uint64( &buffer );
    token->flow_id = next_read_uint64( &buffer );
    token->flow_version = next_read_uint8( &buffer );
    token->flow_flags = next_read_uint8( &buffer );

    next_assert( buffer - start == NEXT_SERVER_TOKEN_BYTES );
}

int next_encrypt_server_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length )
{
    next_assert( sender_private_key );
    next_assert( receiver_public_key );
    next_assert( buffer );
    next_assert( buffer_length >= (int) ( NEXT_SERVER_TOKEN_BYTES + crypto_box_MACBYTES ) );

    (void) buffer_length;

#if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    if ( crypto_box_easy( buffer, buffer, NEXT_SERVER_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key ) != 0 )
    {
        return NEXT_ERROR;
    }

#else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    #error this version of sodium does not suppert overlapping buffers. please upgrade your libsodium!

#endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    return NEXT_OK;
}

int next_decrypt_server_token( uint8_t * sender_public_key, uint8_t * receiver_private_key, uint8_t * nonce, uint8_t * buffer )
{
    next_assert( sender_public_key );
    next_assert( receiver_private_key );
    next_assert( buffer );

#if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    if ( crypto_box_open_easy( buffer, buffer, NEXT_SERVER_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key ) != 0 )
    {
        return NEXT_ERROR;
    }

#else // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    #error this version of sodium does not suppert overlapping buffers. please upgrade your libsodium!

#endif // #if SODIUM_SUPPORTS_OVERLAPPING_BUFFERS

    return NEXT_OK;
}

int next_write_encrypted_server_token( uint8_t ** buffer, next_server_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( sender_private_key );
    next_assert( receiver_public_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    next_random_bytes( nonce, crypto_box_NONCEBYTES );

    uint8_t * start = *buffer;

    next_write_bytes( buffer, nonce, crypto_box_NONCEBYTES );

    next_write_server_token( token, *buffer, NEXT_SERVER_TOKEN_BYTES );

    if ( next_encrypt_server_token( sender_private_key, receiver_public_key, nonce, *buffer, NEXT_SERVER_TOKEN_BYTES + crypto_box_NONCEBYTES ) != NEXT_OK )
        return NEXT_ERROR;

    *buffer += NEXT_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    (void) start;

    next_assert( ( *buffer - start ) == NEXT_ENCRYPTED_SERVER_TOKEN_BYTES );

    return NEXT_OK;
}

int next_read_encrypted_server_token( uint8_t ** buffer, next_server_token_t * token, uint8_t * sender_public_key, uint8_t * receiver_private_key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( sender_public_key );
    next_assert( receiver_private_key );

    uint8_t * nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if ( next_decrypt_continue_token( sender_public_key, receiver_private_key, nonce, *buffer ) != NEXT_OK )
    {
        return NEXT_ERROR;
    }

    next_read_server_token( token, *buffer );

    *buffer += NEXT_SERVER_TOKEN_BYTES + crypto_box_MACBYTES;

    return NEXT_OK;
}

// -----------------------------------------------------------------------------

static int next_packet_is_server_to_client ( uint8_t type )
{
    return type == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE
        || type == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE
        || type == NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT
        || type == NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT
        || type == NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE
        || type == NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE
        || type == NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE
        || type == NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE
        || type == NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG
        || type == NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG;
}

static int next_packet_is_client_to_server ( uint8_t type )
{
    return type == NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER
        || type == NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER
        || type == NEXT_PACKET_TYPE_V2_MIGRATE
        || type == NEXT_PACKET_TYPE_V3_MIGRATE
        || type == NEXT_PACKET_TYPE_V2_DESTROY
        || type == NEXT_PACKET_TYPE_V3_DESTROY
        || type == NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING
        || type == NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING;
}

static int next_packet_type_has_header( uint8_t type )
{
    return type == NEXT_PACKET_TYPE_V2_MIGRATE
        || type == NEXT_PACKET_TYPE_V3_MIGRATE
        || type == NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE
        || type == NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE
        || type == NEXT_PACKET_TYPE_V2_DESTROY
        || type == NEXT_PACKET_TYPE_V3_DESTROY
        || type == NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE
        || type == NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE
        || type == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE
        || type == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE
        || type == NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER
        || type == NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER
        || type == NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT
        || type == NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT
        || type == NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING
        || type == NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING
        || type == NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG
        || type == NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG;
}

int next_write_header( uint8_t type, 
                       uint64_t sequence, 
                       uint64_t flow_id, 
                       uint8_t flow_version,
                       uint8_t flow_flags,
                       uint8_t * private_key, 
                       uint8_t * buffer, 
                       int buffer_length )
{
    next_assert( private_key );
    next_assert( buffer );
    next_assert( NEXT_HEADER_BYTES <= buffer_length );

    (void) buffer_length;

    uint8_t * start = buffer;

    if ( next_packet_is_server_to_client( type ) )
    {
        // high bit must be set
        sequence |= 1ULL << 63;
    }
    else if ( next_packet_is_client_to_server( type ) )
    {
        // high bit must be clear
        sequence &= ~(1ULL << 63);
    }

    next_write_uint8( &buffer, type );

    const uint8_t * signed_data = buffer;
    const int signed_data_bytes = sizeof(uint64_t) + sizeof(uint64_t) + 1 + 1;

    next_write_uint64( &buffer, sequence );

    uint8_t * additional = buffer;
    const int additional_length = 8 + 2;

    next_write_uint64( &buffer, flow_id );
    next_write_uint8( &buffer, flow_version );
    next_write_uint8( &buffer, flow_flags );

    if ( type >= NEXT_PACKET_TYPE_V3_OFFSET )
    {
        // v3 signing algorithm

        next_assert( NEXT_SYMMETRIC_KEY_BYTES == crypto_onetimeauth_KEYBYTES );
        next_assert( NEXT_SYMMETRIC_MAC_BYTES == crypto_onetimeauth_BYTES );

        crypto_onetimeauth( buffer, signed_data, signed_data_bytes, private_key );
    }
    else
    {
        // v2 signing algorithm

        uint8_t nonce[12];
        {
            uint8_t * p = nonce;
            next_write_uint32( &p, 0 );
            next_write_uint64( &p, sequence );
        }

        if ( next_encrypt_aead( buffer, 0, additional, additional_length, nonce, private_key ) != NEXT_OK )
            return NEXT_ERROR;
    }

    buffer += NEXT_SYMMETRIC_MAC_BYTES;

    int bytes = (int) ( buffer - start );

    next_assert( bytes == NEXT_HEADER_BYTES );

    (void) bytes;

    return NEXT_OK;
}

int next_peek_header( uint8_t * type, 
                      uint64_t * sequence, 
                      uint64_t * flow_id, 
                      uint8_t * flow_version, 
                      uint8_t * flow_flags, 
                      uint8_t * buffer, 
                      int buffer_length )
{
    uint8_t _type;
    uint64_t _sequence;

    next_assert( buffer );

    if ( buffer_length < NEXT_HEADER_BYTES )
        return NEXT_ERROR;

    _type = next_read_uint8( &buffer );

    if ( !next_packet_type_has_header( _type ) )
        return NEXT_ERROR;

    _sequence = next_read_uint64( &buffer );

    if ( next_packet_is_server_to_client( _type ) )
    {
        // high bit must be set
        if ( !( _sequence & ( 1ULL << 63 ) ) )
            return NEXT_ERROR;

        // okay now don't worry about it any more
        _sequence &= ~( 1ULL << 63 );
    }
    else if ( next_packet_is_client_to_server( _type ) )
    {
        // high bit must be clear
        if ( _sequence & ( 1ULL << 63 ) )
            return NEXT_ERROR;
    }

    *type = _type;
    *sequence = _sequence;
    *flow_id = next_read_uint64( &buffer );
    *flow_version = next_read_uint8( &buffer );
    *flow_flags = next_read_uint8( &buffer );

    return NEXT_OK;
}

int next_read_header( uint8_t * type, 
                      uint64_t * sequence, 
                      uint64_t * flow_id, 
                      uint8_t * flow_version, 
                      uint8_t * flow_flags, 
                      uint8_t * private_key, 
                      uint8_t * buffer, 
                      int buffer_length )
{
    next_assert( private_key );
    next_assert( buffer );

    if ( buffer_length < NEXT_HEADER_BYTES )
    {
        // todo
        printf( "next_read_header failed. too small (%d<%d)", buffer_length, NEXT_HEADER_BYTES );
        return NEXT_ERROR;
    }

    uint8_t * start = buffer;

    uint8_t _type;
    uint64_t _sequence;
    uint64_t _flow_id;
    uint8_t _flow_version;
    uint8_t _flow_flags;

    _type = next_read_uint8( &buffer );

    if ( !next_packet_type_has_header( _type ) )
    {
        // todo
        printf( "next packet type should not have header\n" );
        return NEXT_ERROR;
    }

    _sequence = next_read_uint64( &buffer );

    if ( next_packet_is_server_to_client( _type ) )
    {
        // high bit must be set
        if ( !( _sequence & ( 1ULL <<  63) ) )
        {
            // todo
            printf( "high bit not set\n" );
            return NEXT_ERROR;
        }
    }
    else if ( next_packet_is_client_to_server( _type ) )
    {
        // high bit must be clear
        if ( _sequence & ( 1ULL << 63 ) )
        {
            // todo
            printf( "high bit not clear\n" );
            return NEXT_ERROR;
        }
    }

    uint8_t * additional = buffer;
    const int additional_length = 8 + 2;

    _flow_id = next_read_uint64( &buffer );
    _flow_version = next_read_uint8( &buffer );
    _flow_flags = next_read_uint8( &buffer );

    uint8_t nonce[12];
    {
        uint8_t * p = nonce;
        next_write_uint32( &p, 0 );
        next_write_uint64( &p, _sequence );
    }

    if ( next_decrypt_aead( buffer, NEXT_SYMMETRIC_MAC_BYTES, additional, additional_length, nonce, private_key ) != NEXT_OK )
    {
        printf( "next_decrypt_aead failed\n" );
        return NEXT_ERROR;
    }

    buffer += NEXT_SYMMETRIC_MAC_BYTES;

    int bytes = (int) ( buffer - start );

    next_assert( bytes == NEXT_HEADER_BYTES );

    *type = _type;
    *sequence = _sequence & ~( 1ULL << 63 );
    *flow_id = _flow_id;
    *flow_version = _flow_version;
    *flow_flags = _flow_flags;

    (void) bytes;

    return NEXT_OK;
}

// ---------------------------------------------------------------

void next_replay_protection_reset( next_replay_protection_t * replay_protection )
{
    next_assert( replay_protection );
    replay_protection->most_recent_sequence = 0;
    memset( replay_protection->received_packet, 0xFF, sizeof( replay_protection->received_packet ) );
}

int next_replay_protection_already_received( next_replay_protection_t * replay_protection, uint64_t sequence )
{
    next_assert( replay_protection );

    if ( sequence + NEXT_REPLAY_PROTECTION_BUFFER_SIZE <= replay_protection->most_recent_sequence )
        return 1;
    
    int index = (int) ( sequence % NEXT_REPLAY_PROTECTION_BUFFER_SIZE );

    if ( replay_protection->received_packet[index] == 0xFFFFFFFFFFFFFFFFLL )
        return 0;

    if ( replay_protection->received_packet[index] >= sequence )
        return 1;

    return 0;
}

void next_replay_protection_advance_sequence( next_replay_protection_t * replay_protection, uint64_t sequence )
{
    next_assert( replay_protection );

    if ( sequence > replay_protection->most_recent_sequence )
        replay_protection->most_recent_sequence = sequence;

    int index = (int) ( sequence % NEXT_REPLAY_PROTECTION_BUFFER_SIZE );

    replay_protection->received_packet[index] = sequence;
}

void next_flow_bandwidth_reset( next_flow_bandwidth_t * flow_bandwidth )
{
    next_assert( flow_bandwidth );
    flow_bandwidth->last_bandwidth_check_timestamp = NEXT_FLOW_BANDWIDTH_CHECK_INVALID_TIME;
    flow_bandwidth->bits_sent_in_current_interval = NEXT_FLOW_BANDWIDTH_CHECK_INVALID_BITS;
}

bool next_flow_bandwidth_over_budget( next_flow_bandwidth_t * flow_bandwidth, int64_t current_time, int64_t interval, uint32_t kbps_allowed, uint32_t packet_bytes_sent )
{
    next_assert( flow_bandwidth );
    if ( ( flow_bandwidth->last_bandwidth_check_timestamp == NEXT_FLOW_BANDWIDTH_CHECK_INVALID_TIME &&
           flow_bandwidth->bits_sent_in_current_interval == NEXT_FLOW_BANDWIDTH_CHECK_INVALID_BITS ) ||
         ( current_time - flow_bandwidth->last_bandwidth_check_timestamp > interval ) )
    {
        flow_bandwidth->bits_sent_in_current_interval = 0;
        flow_bandwidth->last_bandwidth_check_timestamp = current_time;
    }
    flow_bandwidth->bits_sent_in_current_interval += NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS( packet_bytes_sent );
    if ( flow_bandwidth->bits_sent_in_current_interval > NEXT_FLOW_BANDWIDTH_MAX_BITS_ALLOWED( interval, kbps_allowed ) )
    {
        return 1;
    }
    return 0;
}

int next_base64_encode_string( const char * input, char * output, size_t output_size )
{
    next_assert( input );
    next_assert( output );
    next_assert( output_size > 0 );

    return next_base64_encode_data( (const uint8_t *)( input ), strlen( input ), output, output_size );
}

int next_base64_decode_string( const char * input, char * output, size_t output_size )
{
    next_assert( input );
    next_assert( output );
    next_assert( output_size > 0 );

    int output_length = next_base64_decode_data( input, (uint8_t *)( output ), output_size - 1 );
    if ( output_length < 0 )
    {
        return output_length;
    }

    output[output_length] = '\0';

    return output_length;
}

static const unsigned char base64_table_encode[65] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

int next_base64_encode_data( const uint8_t * input, size_t input_length, char * output, size_t output_size )
{
    next_assert( input );
    next_assert( output );
    next_assert( output_size > 0 );

    char * pos;
    const uint8_t * end;
    const uint8_t * in;

    size_t output_length = 4 * ( ( input_length + 2 ) / 3 ); // 3-byte blocks to 4-byte

    if ( output_length < input_length )
    {
        return -1; // integer overflow
    }

    if ( output_length >= output_size )
    {
        return -1; // not enough room in output buffer
    }

    end = input + input_length;
    in = input;
    pos = output;
    while ( end - in >= 3 )
    {
        *pos++ = base64_table_encode[in[0] >> 2];
        *pos++ = base64_table_encode[( ( in[0] & 0x03 ) << 4 ) | ( in[1] >> 4 )];
        *pos++ = base64_table_encode[( ( in[1] & 0x0f ) << 2 ) | ( in[2] >> 6 )];
        *pos++ = base64_table_encode[in[2] & 0x3f];
        in += 3;
    }

    if (end - in)
    {
        *pos++ = base64_table_encode[in[0] >> 2];
        if (end - in == 1)
        {
            *pos++ = base64_table_encode[(in[0] & 0x03) << 4];
            *pos++ = '=';
        }
        else
        {
            *pos++ = base64_table_encode[((in[0] & 0x03) << 4) | (in[1] >> 4)];
            *pos++ = base64_table_encode[(in[1] & 0x0f) << 2];
        }
        *pos++ = '=';
    }

    output[output_length] = '\0';

    return int( output_length );
}

static const int base64_table_decode[256] =
{
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0, 62, 63, 62, 62, 63, 52, 53, 54, 55,
    56, 57, 58, 59, 60, 61,  0,  0,  0,  0,  0,  0,  0,  0,  1,  2,  3,  4,  5,  6,
    7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,  0,
    0,  0,  0,  63,  0, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
};

int next_base64_decode_data( const char * input, uint8_t * output, size_t output_size )
{
    next_assert( input );
    next_assert( output );
    next_assert( output_size > 0 );

    size_t input_length = strlen( input );
    int pad = input_length > 0 && ( input_length % 4 || input[input_length - 1] == '=' );
    size_t L = ( ( input_length + 3 ) / 4 - pad ) * 4;
    size_t output_length = L / 4 * 3 + pad;

    if ( output_length > output_size )
    {
        return -1;
    }

    for ( size_t i = 0, j = 0; i < L; i += 4 )
    {
        int n = base64_table_decode[int( input[i] )] << 18 | base64_table_decode[int( input[i + 1] )] << 12 | base64_table_decode[int( input[i + 2] )] << 6 | base64_table_decode[int( input[i + 3] )];
        output[j++] = uint8_t( n >> 16 );
        output[j++] = uint8_t( n >> 8 & 0xFF );
        output[j++] = uint8_t( n & 0xFF );
    }

    if (pad)
    {
        int n = base64_table_decode[int( input[L] )] << 18 | base64_table_decode[int( input[L + 1] )] << 12;
        output[output_length - 1] = uint8_t( n >> 16 );

        if (input_length > L + 2 && input[L + 2] != '=')
        {
            n |= base64_table_decode[int( input[L + 2] )] << 6;
            output_length += 1;
            if ( output_length > output_size )
            {
                return -1;
            }
            output[output_length - 1] = uint8_t( n >> 8 & 0xFF );
        }
    }

    return int( output_length );
}

void next_write_uint64( uint8_t * p, uint64_t value )
{
    p[0] = value & 0xFF;
    p[1] = ( value >> 8  ) & 0xFF;
    p[2] = ( value >> 16 ) & 0xFF;
    p[3] = ( value >> 24 ) & 0xFF;
    p[4] = ( value >> 32 ) & 0xFF;
    p[5] = ( value >> 40 ) & 0xFF;
    p[6] = ( value >> 48 ) & 0xFF;
    p[7] = value >> 56;
}

uint64_t next_read_uint64( uint8_t * p )
{
    uint64_t value;
    value  = p[0];
    value |= ( ( (uint64_t)( p[1] ) ) << 8  );
    value |= ( ( (uint64_t)( p[2] ) ) << 16 );
    value |= ( ( (uint64_t)( p[3] ) ) << 24 );
    value |= ( ( (uint64_t)( p[4] ) ) << 32 );
    value |= ( ( (uint64_t)( p[5] ) ) << 40 );
    value |= ( ( (uint64_t)( p[6] ) ) << 48 );
    value |= ( ( (uint64_t)( p[7] ) ) << 56 );
    return value;
}
