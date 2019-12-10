/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "relay.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
#include <stdarg.h>
#include <sodium.h>
#include <math.h>
#include <curl/curl.h>

#define RELAY_MTU                                               1300

#define RELAY_HEADER_BYTES                                        35

#define RELAY_ADDRESS_BYTES                                       19
#define RELAY_ADDRESS_BUFFER_SAFETY                               32

#define RELAY_REPLAY_PROTECTION_BUFFER_SIZE                      256

#define RELAY_BANDWIDTH_LIMITER_INTERVAL                         1.0

#define RELAY_ROUTE_TOKEN_BYTES                                   77
#define RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES                        117
#define RELAY_CONTINUE_TOKEN_BYTES                                18
#define RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES                      58

#define RELAY_DIRECTION_CLIENT_TO_SERVER                           0
#define RELAY_DIRECTION_SERVER_TO_CLIENT                           1

#define RELAY_ROUTE_REQUEST_PACKET                                 1
#define RELAY_ROUTE_RESPONSE_PACKET                                2
#define RELAY_CLIENT_TO_SERVER_PACKET                              3
#define RELAY_SERVER_TO_CLIENT_PACKET                              4
#define RELAY_SESSION_PING_PACKET                                 11
#define RELAY_SESSION_PONG_PACKET                                 12
#define RELAY_CONTINUE_REQUEST_PACKET                             13
#define RELAY_CONTINUE_RESPONSE_PACKET                            14
#define RELAY_NEAR_PING_PACKET                                    73
#define RELAY_NEAR_PONG_PACKET                                    74

// -------------------------------------------------------------------------------------

extern int relay_platform_init();

extern void relay_platform_term();

extern const char * relay_platform_getenv( const char * );

extern uint16_t relay_platform_ntohs( uint16_t in );

extern uint16_t relay_platform_htons( uint16_t in );

extern int relay_platform_inet_pton4( const char * address_string, uint32_t * address_out );

extern int relay_platform_inet_pton6( const char * address_string, uint16_t * address_out );

extern int relay_platform_inet_ntop6( const uint16_t * address, char * address_string, size_t address_string_size );

extern double relay_platform_time();

extern void relay_platform_sleep( double time );

extern relay_platform_socket_t * relay_platform_socket_create( void * context, struct relay_address_t * address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size );

extern void relay_platform_socket_destroy( relay_platform_socket_t * socket );

extern void relay_platform_socket_send_packet( relay_platform_socket_t * socket, const relay_address_t * to, const void * packet_data, int packet_bytes );

extern int relay_platform_socket_receive_packet( relay_platform_socket_t * socket, relay_address_t * from, void * packet_data, int max_packet_size );

extern relay_platform_thread_t * relay_platform_thread_create( void * context, relay_platform_thread_func_t * func, void * arg );

extern void relay_platform_thread_join( relay_platform_thread_t * thread );

extern void relay_platform_thread_destroy( relay_platform_thread_t * thread );

extern void relay_platform_thread_set_sched_max( relay_platform_thread_t * thread );

extern relay_platform_mutex_t * relay_platform_mutex_create( void * context );

extern void relay_platform_mutex_acquire( relay_platform_mutex_t * mutex );

extern void relay_platform_mutex_release( relay_platform_mutex_t * mutex );

extern void relay_platform_mutex_destroy( relay_platform_mutex_t * mutex );

struct relay_mutex_helper_t
{
    relay_platform_mutex_t * mutex;
    relay_mutex_helper_t( relay_platform_mutex_t * mutex );
    ~relay_mutex_helper_t();
};

#define relay_mutex_guard( _mutex ) relay_mutex_helper_t __mutex_helper( _mutex )

relay_mutex_helper_t::relay_mutex_helper_t( relay_platform_mutex_t * mutex ) : mutex( mutex )
{
    assert( mutex );
    relay_platform_mutex_acquire( mutex );
}

relay_mutex_helper_t::~relay_mutex_helper_t()
{
    assert( mutex );
    relay_platform_mutex_release( mutex );
    mutex = NULL;
}
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------

static int log_level = RELAY_LOG_LEVEL_INFO;

void relay_log_level( int level )
{
    log_level = level;
}

const char * relay_log_level_string( int level )
{
    if ( level == RELAY_LOG_LEVEL_DEBUG )
        return "debug";
    else if ( level == RELAY_LOG_LEVEL_INFO )
        return "info";
    else if ( level == RELAY_LOG_LEVEL_ERROR )
        return "error";
    else if ( level == RELAY_LOG_LEVEL_WARN )
        return "warning";
    else
        return "???";
}

static void default_log_function( int level, const char * format, ... ) 
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    const char * level_string = relay_log_level_string( level );
    printf( "%.6f: %s: %s\n", relay_platform_time(), level_string, buffer );
    va_end( args );
    fflush( stdout );
}

static void (*log_function)( int level, const char * format, ... ) = default_log_function;

void relay_log_function( void (*function)( int level, const char * format, ... ) )
{
    log_function = function;
}

void relay_printf( int level, const char * format, ... ) 
{
    if ( level > log_level )
        return;
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    log_function( level, "%s", buffer );
    va_end( args );
}

// -----------------------------------------------------------------------------

int relay_init()
{
    if ( relay_platform_init() != RELAY_OK )
    {
        relay_printf( RELAY_LOG_LEVEL_ERROR, "failed to initialize platform" );
        return RELAY_ERROR;
    }

    if ( sodium_init() == -1 )
    {
        relay_printf( RELAY_LOG_LEVEL_ERROR, "failed to initialize sodium" );
        return RELAY_ERROR;
    }

    const char * log_level_override = relay_platform_getenv( "RELAY_LOG_LEVEL" );
    if ( log_level_override )
    {
        log_level = atoi( log_level_override );
    }

    return RELAY_OK;
}

void relay_term()
{
    relay_platform_term();
}

// -----------------------------------------------------------------------------

void relay_random_bytes( uint8_t * buffer, int bytes )
{
    randombytes_buf( buffer, bytes );
}

uint16_t relay_ntohs( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

uint16_t relay_htons( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

// -----------------------------------------------------------------------------

void relay_write_uint8( uint8_t ** p, uint8_t value )
{
    **p = value;
    ++(*p);
}

void relay_write_uint16( uint8_t ** p, uint16_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = value >> 8;
    *p += 2;
}

void relay_write_uint32( uint8_t ** p, uint32_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = value >> 24;
    *p += 4;
}

void relay_write_uint64( uint8_t ** p, uint64_t value )
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

void relay_write_float32( uint8_t ** p, float value )
{
    uint32_t value_int = 0;
    char * p_value = (char*)(&value);
    char * p_value_int = (char*)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint32_t));
    relay_write_uint32( p, value_int);
}

void relay_write_float64( uint8_t ** p, double value )
{
    uint64_t value_int = 0;
    char * p_value = (char *)(&value);
    char * p_value_int = (char *)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint64_t));
    relay_write_uint64( p, value_int);
}

void relay_write_bytes( uint8_t ** p, const uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        relay_write_uint8( p, byte_array[i] );
    }
}

void relay_write_string( uint8_t ** p, const char * string_data, uint32_t max_length )
{
    uint32_t length = strlen( string_data );
    assert( length <= max_length );
    if ( length > max_length )
        length = max_length;
    relay_write_uint32( p, length );
    for ( uint32_t i = 0; i < length; ++i )
    {
        relay_write_uint8( p, string_data[i] );
    }
}

uint8_t relay_read_uint8( const uint8_t ** p )
{
    uint8_t value = **p;
    ++(*p);
    return value;
}

uint16_t relay_read_uint16( const uint8_t ** p )
{
    uint16_t value;
    value = (*p)[0];
    value |= ( ( (uint16_t)( (*p)[1] ) ) << 8 );
    *p += 2;
    return value;
}

uint32_t relay_read_uint32( const uint8_t ** p )
{
    uint32_t value;
    value  = (*p)[0];
    value |= ( ( (uint32_t)( (*p)[1] ) ) << 8 );
    value |= ( ( (uint32_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint32_t)( (*p)[3] ) ) << 24 );
    *p += 4;
    return value;
}

uint64_t relay_read_uint64( const uint8_t ** p )
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

float relay_read_float32( const uint8_t ** p )
{
    uint32_t value_int = relay_read_uint32( p );
    float value_float = 0.0f;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

double relay_read_float64( const uint8_t ** p )
{
    uint64_t value_int = relay_read_uint64( p );
    double value_float = 0.0;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

void relay_read_bytes( const uint8_t ** p, uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        byte_array[i] = relay_read_uint8( p );
    }
}

void relay_read_string( const uint8_t ** p, char * string_data, uint32_t max_length )
{
    uint32_t length = relay_read_uint32( p );
    if ( length > max_length )
    {
        length = 0;
        return;
    }
    uint32_t i = 0;
    for ( ; i < length; ++i )
    {
        string_data[i] = relay_read_uint8( p );
    }
    string_data[i] = 0;
}

// -----------------------------------------------------------------------------

int relay_address_parse( relay_address_t * address, const char * address_string_in )
{
    assert( address );
    assert( address_string_in );

    if ( !address )
        return RELAY_ERROR;

    if ( !address_string_in )
        return RELAY_ERROR;

    memset( address, 0, sizeof( relay_address_t ) );

    // first try to parse the string as an IPv6 address:
    // 1. if the first character is '[' then it's probably an ipv6 in form "[addr6]:portnum"
    // 2. otherwise try to parse as a raw IPv6 address using inet_pton

    char buffer[RELAY_MAX_ADDRESS_STRING_LENGTH + RELAY_ADDRESS_BUFFER_SAFETY*2];

    char * address_string = buffer + RELAY_ADDRESS_BUFFER_SAFETY;
    strncpy( address_string, address_string_in, RELAY_MAX_ADDRESS_STRING_LENGTH - 1 );
    address_string[RELAY_MAX_ADDRESS_STRING_LENGTH-1] = '\0';

    int address_string_length = (int) strlen( address_string );

    if ( address_string[0] == '[' )
    {
        const int base_index = address_string_length - 1;

        // note: no need to search past 6 characters as ":65535" is longest possible port value
        for ( int i = 0; i < 6; ++i )
        {
            const int index = base_index - i;
            if ( index < 0 )
            {
                return RELAY_ERROR;
            }
            if ( address_string[index] == ':' )
            {
                address->port = (uint16_t) ( atoi( &address_string[index + 1] ) );
                address_string[index-1] = '\0';
                break;
            }
            else if ( address_string[index] == ']' )
            {
                // no port number
                address->port = 0;
                address_string[index] = '\0';
                break;
            }
        }
        address_string += 1;
    }
    uint16_t addr6[8];
    if ( relay_platform_inet_pton6( address_string, addr6 ) == RELAY_OK )
    {
        address->type = RELAY_ADDRESS_IPV6;
        for ( int i = 0; i < 8; ++i )
        {
            address->data.ipv6[i] = relay_platform_ntohs( addr6[i] );
        }
        return RELAY_OK;
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
    if ( relay_platform_inet_pton4( address_string, &addr4 ) == RELAY_OK )
    {
        address->type = RELAY_ADDRESS_IPV4;
        address->data.ipv4[3] = (uint8_t) ( ( addr4 & 0xFF000000 ) >> 24 );
        address->data.ipv4[2] = (uint8_t) ( ( addr4 & 0x00FF0000 ) >> 16 );
        address->data.ipv4[1] = (uint8_t) ( ( addr4 & 0x0000FF00 ) >> 8  );
        address->data.ipv4[0] = (uint8_t) ( ( addr4 & 0x000000FF )     );
        return RELAY_OK;
    }

    return RELAY_ERROR;
}

void relay_read_address( const uint8_t ** buffer, relay_address_t * address )
{
    const uint8_t * start = *buffer;

    address->type = relay_read_uint8( buffer );

    if ( address->type == RELAY_ADDRESS_IPV4 )
    {
        for ( int j = 0; j < 4; ++j )
        {
            address->data.ipv4[j] = relay_read_uint8( buffer );
        }
        address->port = relay_read_uint16( buffer );
        for ( int i = 0; i < 12; ++i )
        {
            uint8_t dummy = relay_read_uint8( buffer ); (void) dummy;
        }
    }
    else if ( address->type == RELAY_ADDRESS_IPV6 )
    {
        for ( int j = 0; j < 8; ++j )
        {
            address->data.ipv6[j] = relay_read_uint16( buffer );
        }
        address->port = relay_read_uint16( buffer );
    }
    else
    {
        for ( int i = 0; i < RELAY_ADDRESS_BYTES - 1; ++i )
        {
            uint8_t dummy = relay_read_uint8( buffer ); (void) dummy;
        }
    }

    (void) start;

    assert( *buffer - start == RELAY_ADDRESS_BYTES );
}

void relay_write_address( uint8_t ** buffer, const relay_address_t * address )
{
    assert( buffer );
    assert( *buffer );
    assert( address );

    uint8_t * start = *buffer;

    (void) buffer;

    if ( address->type == RELAY_ADDRESS_IPV4 )
    {
        relay_write_uint8( buffer, RELAY_ADDRESS_IPV4 );
        for ( int i = 0; i < 4; ++i )
        {
            relay_write_uint8( buffer, address->data.ipv4[i] );
        }
        relay_write_uint16( buffer, address->port );
        for ( int i = 0; i < 12; ++i )
        {
            relay_write_uint8( buffer, 0 );
        }
    }
    else if ( address->type == RELAY_ADDRESS_IPV6 )
    {
        relay_write_uint8( buffer, RELAY_ADDRESS_IPV6 );
        for ( int i = 0; i < 8; ++i )
        {
            relay_write_uint16( buffer, address->data.ipv6[i] );
        }
        relay_write_uint16( buffer, address->port );
    }
    else
    {
        for ( int i = 0; i < RELAY_ADDRESS_BYTES; ++i )
        {
            relay_write_uint8( buffer, 0 );
        }
    }

    (void) start;

    assert( *buffer - start == RELAY_ADDRESS_BYTES );
}

const char * relay_address_to_string( const relay_address_t * address, char * buffer )
{
    assert( buffer );

    if ( address->type == RELAY_ADDRESS_IPV6 )
    {
#if defined(WINVER) && WINVER <= 0x0502
        // ipv6 not supported
        buffer[0] = '\0';
        return buffer;
#else
        uint16_t ipv6_network_order[8];
        for ( int i = 0; i < 8; ++i )
            ipv6_network_order[i] = relay_htons( address->data.ipv6[i] );
        char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
        relay_platform_inet_ntop6( ipv6_network_order, address_string, sizeof( address_string ) );
        if ( address->port == 0 )
        {
            strncpy( buffer, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH );
            return buffer;
        }
        else
        {
            if ( snprintf( buffer, RELAY_MAX_ADDRESS_STRING_LENGTH, "[%s]:%hu", address_string, address->port ) < 0 )
            {
                relay_printf( RELAY_LOG_LEVEL_ERROR, "address string truncated: [%s]:%hu", address_string, address->port );
            }
            return buffer;
        }
#endif
    }
    else if ( address->type == RELAY_ADDRESS_IPV4 )
    {
        if ( address->port != 0 )
        {
            snprintf( buffer, 
                      RELAY_MAX_ADDRESS_STRING_LENGTH, 
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
                      RELAY_MAX_ADDRESS_STRING_LENGTH, 
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
        snprintf( buffer, RELAY_MAX_ADDRESS_STRING_LENGTH, "%s", "NONE" );
        return buffer;
    }
}

int relay_address_equal( const relay_address_t * a, const relay_address_t * b )
{
    assert( a );
    assert( b );

    if ( a->type != b->type )
        return 0;

    if ( a->type == RELAY_ADDRESS_IPV4 )
    {
        if ( a->port != b->port )
            return 0;

        for ( int i = 0; i < 4; ++i )
        {
            if ( a->data.ipv4[i] != b->data.ipv4[i] )
                return 0;
        }
    }
    else if ( a->type == RELAY_ADDRESS_IPV6 )
    {
        if ( a->port != b->port )
            return 0;

        for ( int i = 0; i < 8; ++i )
        {
            if ( a->data.ipv6[i] != b->data.ipv6[i] )
                return 0;
        }
    }

    return 1;
}

// -----------------------------------------------------------------------------

struct relay_replay_protection_t
{
    uint64_t most_recent_sequence;
    uint64_t received_packet[RELAY_REPLAY_PROTECTION_BUFFER_SIZE];
};

void relay_replay_protection_reset( relay_replay_protection_t * replay_protection )
{
    assert( replay_protection );
    replay_protection->most_recent_sequence = 0;
    memset( replay_protection->received_packet, 0xFF, sizeof( replay_protection->received_packet ) );
}

int relay_replay_protection_already_received( relay_replay_protection_t * replay_protection, uint64_t sequence )
{
    assert( replay_protection );

    if ( sequence + RELAY_REPLAY_PROTECTION_BUFFER_SIZE <= replay_protection->most_recent_sequence )
        return 1;
    
    int index = (int) ( sequence % RELAY_REPLAY_PROTECTION_BUFFER_SIZE );

    if ( replay_protection->received_packet[index] == 0xFFFFFFFFFFFFFFFFLL )
        return 0;

    if ( replay_protection->received_packet[index] >= sequence )
        return 1;

    return 0;
}

void relay_replay_protection_advance_sequence( relay_replay_protection_t * replay_protection, uint64_t sequence )
{
    assert( replay_protection );

    if ( sequence > replay_protection->most_recent_sequence )
    {
        replay_protection->most_recent_sequence = sequence;
    }

    int index = (int) ( sequence % RELAY_REPLAY_PROTECTION_BUFFER_SIZE );

    replay_protection->received_packet[index] = sequence;
}

// -----------------------------------------------------------------------------

namespace relay
{
    /**
        Calculates the population count of an unsigned 32 bit integer at compile time.
        See "Hacker's Delight" and http://www.hackersdelight.org/hdcodetxt/popArrayHS.c.txt
     */

    template <uint32_t x> struct PopCount
    {
        enum {   a = x - ( ( x >> 1 )       & 0x55555555 ),
                 b =   ( ( ( a >> 2 )       & 0x33333333 ) + ( a & 0x33333333 ) ),
                 c =   ( ( ( b >> 4 ) + b ) & 0x0f0f0f0f ),
                 d =   c + ( c >> 8 ),
                 e =   d + ( d >> 16 ),

            result = e & 0x0000003f 
        };
    };

    /**
        Calculates the log 2 of an unsigned 32 bit integer at compile time.
     */

    template <uint32_t x> struct Log2
    {
        enum {   a = x | ( x >> 1 ),
                 b = a | ( a >> 2 ),
                 c = b | ( b >> 4 ),
                 d = c | ( c >> 8 ),
                 e = d | ( d >> 16 ),
                 f = e >> 1,

            result = PopCount<f>::result
        };
    };

    /**
        Calculates the number of bits required to serialize an integer value in [min,max] at compile time.
     */

    template <int64_t min, int64_t max> struct BitsRequired
    {
        static const uint32_t result = ( min == max ) ? 0 : ( Log2<uint32_t(max-min)>::result + 1 );
    };

    /**
        Calculates the population count of an unsigned 32 bit integer.
        The population count is the number of bits in the integer set to 1.
        @param x The input integer value.
        @returns The number of bits set to 1 in the input value.
     */

    inline uint32_t popcount( uint32_t x )
    {
    #ifdef __GNUC__
        return __builtin_popcount( x );
    #else // #ifdef __GNUC__
        const uint32_t a = x - ( ( x >> 1 )       & 0x55555555 );
        const uint32_t b =   ( ( ( a >> 2 )       & 0x33333333 ) + ( a & 0x33333333 ) );
        const uint32_t c =   ( ( ( b >> 4 ) + b ) & 0x0f0f0f0f );
        const uint32_t d =   c + ( c >> 8 );
        const uint32_t e =   d + ( d >> 16 );
        const uint32_t result = e & 0x0000003f;
        return result;
    #endif // #ifdef __GNUC__
    }

    /**
        Calculates the log base 2 of an unsigned 32 bit integer.
        @param x The input integer value.
        @returns The log base 2 of the input.
     */

    inline uint32_t log2( uint32_t x )
    {
        const uint32_t a = x | ( x >> 1 );
        const uint32_t b = a | ( a >> 2 );
        const uint32_t c = b | ( b >> 4 );
        const uint32_t d = c | ( c >> 8 );
        const uint32_t e = d | ( d >> 16 );
        const uint32_t f = e >> 1;
        return popcount( f );
    }

    /**
        Calculates the number of bits required to serialize an integer in range [min,max].
        @param min The minimum value.
        @param max The maximum value.
        @returns The number of bits required to serialize the integer.
     */

    inline int bits_required( uint32_t min, uint32_t max )
    {
    #ifdef __GNUC__
        return ( min == max ) ? 0 : 32 - __builtin_clz( max - min );
    #else // #ifdef __GNUC__
        return ( min == max ) ? 0 : log2( max - min ) + 1;
    #endif // #ifdef __GNUC__
    }

    /**
        Reverse the order of bytes in a 64 bit integer.
        @param value The input value.
        @returns The input value with the byte order reversed.
     */

    inline uint64_t bswap( uint64_t value )
    {
    #ifdef __GNUC__
        return __builtin_bswap64( value );
    #else // #ifdef __GNUC__
        value = ( value & 0x00000000FFFFFFFF ) << 32 | ( value & 0xFFFFFFFF00000000 ) >> 32;
        value = ( value & 0x0000FFFF0000FFFF ) << 16 | ( value & 0xFFFF0000FFFF0000 ) >> 16;
        value = ( value & 0x00FF00FF00FF00FF ) << 8  | ( value & 0xFF00FF00FF00FF00 ) >> 8;
        return value;
    #endif // #ifdef __GNUC__
    }

    /**
        Reverse the order of bytes in a 32 bit integer.
        @param value The input value.
        @returns The input value with the byte order reversed.
     */

    inline uint32_t bswap( uint32_t value )
    {
    #ifdef __GNUC__
        return __builtin_bswap32( value );
    #else // #ifdef __GNUC__
        return ( value & 0x000000ff ) << 24 | ( value & 0x0000ff00 ) << 8 | ( value & 0x00ff0000 ) >> 8 | ( value & 0xff000000 ) >> 24;
    #endif // #ifdef __GNUC__
    }

    /**
        Reverse the order of bytes in a 16 bit integer.
        @param value The input value.
        @returns The input value with the byte order reversed.
     */

    inline uint16_t bswap( uint16_t value )
    {
        return ( value & 0x00ff ) << 8 | ( value & 0xff00 ) >> 8;
    }

    /**
        Template to convert an integer value from local byte order to network byte order.
        IMPORTANT: Because most machines running relay are little endian, relay defines network byte order to be little endian.
        @param value The input value in local byte order. Supported integer types: uint64_t, uint32_t, uint16_t.
        @returns The input value converted to network byte order. If this processor is little endian the output is the same as the input. If the processor is big endian, the output is the input byte swapped.
     */

    template <typename T> T host_to_network( T value )
    {
    #if RELAY_BIG_ENDIAN
        return bswap( value );
    #else // #if RELAY_BIG_ENDIAN
        return value;
    #endif // #if RELAY_BIG_ENDIAN
    }

    /**
        Template to convert an integer value from network byte order to local byte order.
        IMPORTANT: Because most machines running relay are little endian, relay defines network byte order to be little endian.
        @param value The input value in network byte order. Supported integer types: uint64_t, uint32_t, uint16_t.
        @returns The input value converted to local byte order. If this processor is little endian the output is the same as the input. If the processor is big endian, the output is the input byte swapped.
     */

    template <typename T> T network_to_host( T value )
    {
    #if RELAY_BIG_ENDIAN
        return bswap( value );
    #else // #if RELAY_BIG_ENDIAN
        return value;
    #endif // #if RELAY_BIG_ENDIAN
    }

    /** 
        Compares two 16 bit sequence numbers and returns true if the first one is greater than the second (considering wrapping).
        IMPORTANT: This is not the same as s1 > s2!
        Greater than is defined specially to handle wrapping sequence numbers. 
        If the two sequence numbers are close together, it is as normal, but they are far apart, it is assumed that they have wrapped around.
        Thus, sequence_greater_than( 1, 0 ) returns true, and so does sequence_greater_than( 0, 65535 )!
        @param s1 The first sequence number.
        @param s2 The second sequence number.
        @returns True if the s1 is greater than s2, with sequence number wrapping considered.
     */

    inline bool sequence_greater_than( uint16_t s1, uint16_t s2 )
    {
        return ( ( s1 > s2 ) && ( s1 - s2 <= 32768 ) ) || 
               ( ( s1 < s2 ) && ( s2 - s1  > 32768 ) );
    }

    /** 
        Compares two 16 bit sequence numbers and returns true if the first one is less than the second (considering wrapping).
        IMPORTANT: This is not the same as s1 < s2!
        Greater than is defined specially to handle wrapping sequence numbers. 
        If the two sequence numbers are close together, it is as normal, but they are far apart, it is assumed that they have wrapped around.
        Thus, sequence_less_than( 0, 1 ) returns true, and so does sequence_greater_than( 65535, 0 )!
        @param s1 The first sequence number.
        @param s2 The second sequence number.
        @returns True if the s1 is less than s2, with sequence number wrapping considered.
     */

    inline bool sequence_less_than( uint16_t s1, uint16_t s2 )
    {
        return sequence_greater_than( s2, s1 );
    }

    /**
        Bitpacks unsigned integer values to a buffer.
        Integer bit values are written to a 64 bit scratch value from right to left.
        Once the low 32 bits of the scratch is filled with bits it is flushed to memory as a dword and the scratch value is shifted right by 32.
        The bit stream is written to memory in little endian order, which is considered network byte order for this library.
     */

    class BitWriter
    {
    public:

        /**
            Bit writer constructor.
            Creates a bit writer object to write to the specified buffer. 
            @param data The pointer to the buffer to fill with bitpacked data.
            @param bytes The size of the buffer in bytes. Must be a multiple of 4, because the bitpacker reads and writes memory as dwords, not bytes.
         */

        BitWriter( void * data, int bytes ) : m_data( (uint32_t*) data ), m_numWords( bytes / 4 )
        {
            assert( data );
            assert( ( bytes % 4 ) == 0 );
            m_numBits = m_numWords * 32;
            m_bitsWritten = 0;
            m_wordIndex = 0;
            m_scratch = 0;
            m_scratchBits = 0;
        }

        /**
            Write bits to the buffer.
            Bits are written to the buffer as-is, without padding to nearest byte. Will assert if you try to write past the end of the buffer.
            A boolean value writes just 1 bit to the buffer, a value in range [0,31] can be written with just 5 bits and so on.
            IMPORTANT: When you have finished writing to your buffer, take care to call BitWrite::FlushBits, otherwise the last dword of data will not get flushed to memory!
            @param value The integer value to write to the buffer. Must be in [0,(1<<bits)-1].
            @param bits The number of bits to encode in [1,32].
         */

        void WriteBits( uint32_t value, int bits )
        {
            assert( bits > 0 );
            assert( bits <= 32 );
            assert( m_bitsWritten + bits <= m_numBits );
            assert( uint64_t( value ) <= ( ( 1ULL << bits ) - 1 ) );

            m_scratch |= uint64_t( value ) << m_scratchBits;

            m_scratchBits += bits;

            if ( m_scratchBits >= 32 )
            {
                assert( m_wordIndex < m_numWords );
                m_data[m_wordIndex] = host_to_network( uint32_t( m_scratch & 0xFFFFFFFF ) );
                m_scratch >>= 32;
                m_scratchBits -= 32;
                m_wordIndex++;
            }

            m_bitsWritten += bits;
        }

        /**
            Write an alignment to the bit stream, padding zeros so the bit index becomes is a multiple of 8.
            This is useful if you want to write some data to a packet that should be byte aligned. For example, an array of bytes, or a string.
            IMPORTANT: If the current bit index is already a multiple of 8, nothing is written.
         */

        void WriteAlign()
        {
            const int remainderBits = m_bitsWritten % 8;

            if ( remainderBits != 0 )
            {
                uint32_t zero = 0;
                WriteBits( zero, 8 - remainderBits );
                assert( ( m_bitsWritten % 8 ) == 0 );
            }
        }

        /**
            Write an array of bytes to the bit stream.
            Use this when you have to copy a large block of data into your bitstream.
            Faster than just writing each byte to the bit stream via BitWriter::WriteBits( value, 8 ), because it aligns to byte index and copies into the buffer without bitpacking.
            @param data The byte array data to write to the bit stream.
            @param bytes The number of bytes to write.
         */

        void WriteBytes( const uint8_t * data, int bytes )
        {
            assert( GetAlignBits() == 0 );
            assert( m_bitsWritten + bytes * 8 <= m_numBits );
            assert( ( m_bitsWritten % 32 ) == 0 || ( m_bitsWritten % 32 ) == 8 || ( m_bitsWritten % 32 ) == 16 || ( m_bitsWritten % 32 ) == 24 );

            int headBytes = ( 4 - ( m_bitsWritten % 32 ) / 8 ) % 4;
            if ( headBytes > bytes )
                headBytes = bytes;
            for ( int i = 0; i < headBytes; ++i )
                WriteBits( data[i], 8 );
            if ( headBytes == bytes )
                return;

            FlushBits();

            assert( GetAlignBits() == 0 );

            int numWords = ( bytes - headBytes ) / 4;
            if ( numWords > 0 )
            {
                assert( ( m_bitsWritten % 32 ) == 0 );
                memcpy( &m_data[m_wordIndex], data + headBytes, numWords * 4 );
                m_bitsWritten += numWords * 32;
                m_wordIndex += numWords;
                m_scratch = 0;
            }

            assert( GetAlignBits() == 0 );

            int tailStart = headBytes + numWords * 4;
            int tailBytes = bytes - tailStart;
            assert( tailBytes >= 0 && tailBytes < 4 );
            for ( int i = 0; i < tailBytes; ++i )
                WriteBits( data[tailStart+i], 8 );

            assert( GetAlignBits() == 0 );

            assert( headBytes + numWords * 4 + tailBytes == bytes );
        }

        /**
            Flush any remaining bits to memory.
            Call this once after you've finished writing bits to flush the last dword of scratch to memory!
         */

        void FlushBits()
        {
            if ( m_scratchBits != 0 )
            {
                assert( m_scratchBits <= 32 );
                assert( m_wordIndex < m_numWords );
                m_data[m_wordIndex] = host_to_network( uint32_t( m_scratch & 0xFFFFFFFF ) );
                m_scratch >>= 32;
                m_scratchBits = 0;
                m_wordIndex++;                
            }
        }

        /**
            How many align bits would be written, if we were to write an align right now?
            @returns Result in [0,7], where 0 is zero bits required to align (already aligned) and 7 is worst case.
         */

        int GetAlignBits() const
        {
            return ( 8 - ( m_bitsWritten % 8 ) ) % 8;
        }

        /** 
            How many bits have we written so far?
            @returns The number of bits written to the bit buffer.
         */

        int GetBitsWritten() const
        {
            return m_bitsWritten;
        }

        /**
            How many bits are still available to write?
            For example, if the buffer size is 4, we have 32 bits available to write, if we have already written 10 bytes then 22 are still available to write.
            @returns The number of bits available to write.
         */

        int GetBitsAvailable() const
        {
            return m_numBits - m_bitsWritten;
        }
        
        /**
            Get a pointer to the data written by the bit writer.
            Corresponds to the data block passed in to the constructor.
            @returns Pointer to the data written by the bit writer.
         */

        const uint8_t * GetData() const
        {
            return (uint8_t*) m_data;
        }

        /**
            The number of bytes flushed to memory.
            This is effectively the size of the packet that you should send after you have finished bitpacking values with this class.
            The returned value is not always a multiple of 4, even though we flush dwords to memory. You won't miss any data in this case because the order of bits written is designed to work with the little endian memory layout.
            IMPORTANT: Make sure you call BitWriter::FlushBits before calling this method, otherwise you risk missing the last dword of data.
         */

        int GetBytesWritten() const
        {
            return ( m_bitsWritten + 7 ) / 8;
        }

    private:

        uint32_t * m_data;              ///< The buffer we are writing to, as a uint32_t * because we're writing dwords at a time.
        uint64_t m_scratch;             ///< The scratch value where we write bits to (right to left). 64 bit for overflow. Once # of bits in scratch is >= 32, the low 32 bits are flushed to memory.
        int m_numBits;                  ///< The number of bits in the buffer. This is equivalent to the size of the buffer in bytes multiplied by 8. Note that the buffer size must always be a multiple of 4.
        int m_numWords;                 ///< The number of words in the buffer. This is equivalent to the size of the buffer in bytes divided by 4. Note that the buffer size must always be a multiple of 4.
        int m_bitsWritten;              ///< The number of bits written so far.
        int m_wordIndex;                ///< The current word index. The next word flushed to memory will be at this index in m_data.
        int m_scratchBits;              ///< The number of bits in scratch. When this is >= 32, the low 32 bits of scratch is flushed to memory as a dword and scratch is shifted right by 32.
    };

    /**
        Reads bit packed integer values from a buffer.
        Relies on the user reconstructing the exact same set of bit reads as bit writes when the buffer was written. This is an unattributed bitpacked binary stream!
        Implementation: 32 bit dwords are read in from memory to the high bits of a scratch value as required. The user reads off bit values from the scratch value from the right, after which the scratch value is shifted by the same number of bits.
     */

    class BitReader
    {
    public:

        /**
            Bit reader constructor.
            Non-multiples of four buffer sizes are supported, as this naturally tends to occur when packets are read from the network.
            However, actual buffer allocated for the packet data must round up at least to the next 4 bytes in memory, because the bit reader reads dwords from memory not bytes.
            @param data Pointer to the bitpacked data to read.
            @param bytes The number of bytes of bitpacked data to read.
         */

    #ifndef NDEBUG
        BitReader( const void * data, int bytes ) : m_data( (const uint32_t*) data ), m_numBytes( bytes ), m_numWords( ( bytes + 3 ) / 4)
    #else // #ifndef NDEBUG
        BitReader( const void * data, int bytes ) : m_data( (const uint32_t*) data ), m_numBytes( bytes )
    #endif // #ifndef NDEBUG
        {
            assert( data );
            m_numBits = m_numBytes * 8;
            m_bitsRead = 0;
            m_scratch = 0;
            m_scratchBits = 0;
            m_wordIndex = 0;
        }

        /**
            Would the bit reader would read past the end of the buffer if it read this many bits?
            @param bits The number of bits that would be read.
            @returns True if reading the number of bits would read past the end of the buffer.
         */

        bool WouldReadPastEnd( int bits ) const
        {
            return m_bitsRead + bits > m_numBits;
        }

        /**
            Read bits from the bit buffer.
            This function will assert in debug builds if this read would read past the end of the buffer.
            In production situations, the higher level ReadStream takes care of checking all packet data and never calling this function if it would read past the end of the buffer.
            @param bits The number of bits to read in [1,32].
            @returns The integer value read in range [0,(1<<bits)-1].
         */

        uint32_t ReadBits( int bits )
        {
            assert( bits > 0 );
            assert( bits <= 32 );
            assert( m_bitsRead + bits <= m_numBits );

            m_bitsRead += bits;

            assert( m_scratchBits >= 0 && m_scratchBits <= 64 );

            if ( m_scratchBits < bits )
            {
                assert( m_wordIndex < m_numWords );
                m_scratch |= uint64_t( network_to_host( m_data[m_wordIndex] ) ) << m_scratchBits;
                m_scratchBits += 32;
                m_wordIndex++;
            }

            assert( m_scratchBits >= bits );

            const uint32_t output = m_scratch & ( (uint64_t(1)<<bits) - 1 );

            m_scratch >>= bits;
            m_scratchBits -= bits;

            return output;
        }

        /**
            Read an align.
            Call this on read to correspond to a WriteAlign call when the bitpacked buffer was written. 
            This makes sure we skip ahead to the next aligned byte index. As a safety check, we verify that the padding to next byte is zero bits and return false if that's not the case. 
            This will typically abort packet read. Just another safety measure...
            @returns True if we successfully read an align and skipped ahead past zero pad, false otherwise (probably means, no align was written to the stream).
         */

        bool ReadAlign()
        {
            const int remainderBits = m_bitsRead % 8;
            if ( remainderBits != 0 )
            {
                uint32_t value = ReadBits( 8 - remainderBits );
                assert( m_bitsRead % 8 == 0 );
                if ( value != 0 )
                    return false;
            }
            return true;
        }

        /**
            Read bytes from the bitpacked data.
         */

        void ReadBytes( uint8_t * data, int bytes )
        {
            assert( GetAlignBits() == 0 );
            assert( m_bitsRead + bytes * 8 <= m_numBits );
            assert( ( m_bitsRead % 32 ) == 0 || ( m_bitsRead % 32 ) == 8 || ( m_bitsRead % 32 ) == 16 || ( m_bitsRead % 32 ) == 24 );

            int headBytes = ( 4 - ( m_bitsRead % 32 ) / 8 ) % 4;
            if ( headBytes > bytes )
                headBytes = bytes;
            for ( int i = 0; i < headBytes; ++i )
                data[i] = (uint8_t) ReadBits( 8 );
            if ( headBytes == bytes )
                return;

            assert( GetAlignBits() == 0 );

            int numWords = ( bytes - headBytes ) / 4;
            if ( numWords > 0 )
            {
                assert( ( m_bitsRead % 32 ) == 0 );
                memcpy( data + headBytes, &m_data[m_wordIndex], numWords * 4 );
                m_bitsRead += numWords * 32;
                m_wordIndex += numWords;
                m_scratchBits = 0;
            }

            assert( GetAlignBits() == 0 );

            int tailStart = headBytes + numWords * 4;
            int tailBytes = bytes - tailStart;
            assert( tailBytes >= 0 && tailBytes < 4 );
            for ( int i = 0; i < tailBytes; ++i )
                data[tailStart+i] = (uint8_t) ReadBits( 8 );

            assert( GetAlignBits() == 0 );

            assert( headBytes + numWords * 4 + tailBytes == bytes );
        }

        /**
            How many align bits would be read, if we were to read an align right now?
            @returns Result in [0,7], where 0 is zero bits required to align (already aligned) and 7 is worst case.
         */

        int GetAlignBits() const
        {
            return ( 8 - m_bitsRead % 8 ) % 8;
        }

        /** 
            How many bits have we read so far?
            @returns The number of bits read from the bit buffer so far.
         */

        int GetBitsRead() const
        {
            return m_bitsRead;
        }

        /**
            How many bits are still available to read?
            For example, if the buffer size is 4, we have 32 bits available to read, if we have already written 10 bytes then 22 are still available.
            @returns The number of bits available to read.
         */

        int GetBitsRemaining() const
        {
            return m_numBits - m_bitsRead;
        }

    private:

        const uint32_t * m_data;            ///< The bitpacked data we're reading as a dword array.
        uint64_t m_scratch;                 ///< The scratch value. New data is read in 32 bits at a top to the left of this buffer, and data is read off to the right.
        int m_numBits;                      ///< Number of bits to read in the buffer. Of course, we can't *really* know this so it's actually m_numBytes * 8.
        int m_numBytes;                     ///< Number of bytes to read in the buffer. We know this, and this is the non-rounded up version.
    #ifndef NDEBUG
        int m_numWords;                     ///< Number of words to read in the buffer. This is rounded up to the next word if necessary.
    #endif // #ifndef NDEBUG
        int m_bitsRead;                     ///< Number of bits read from the buffer so far.
        int m_scratchBits;                  ///< Number of bits currently in the scratch value. If the user wants to read more bits than this, we have to go fetch another dword from memory.
        int m_wordIndex;                    ///< Index of the next word to read from memory.
    };

    /** 
        Functionality common to all stream classes.
     */

    class BaseStream
    {
    public:

        /**
            Base stream constructor.
         */

        explicit BaseStream() : m_context( NULL ) {}

        /**
            Set a context on the stream.
         */

        void SetContext( void * context )
        {
            m_context = context;
        }

        /**
            Get the context pointer set on the stream.
            @returns The context pointer. May be NULL.
         */

        void * GetContext() const
        {
            return m_context;
        }

    private:

        void * m_context;                           ///< The context pointer set on the stream. May be NULL.
    };

    /**
        Stream class for writing bitpacked data.
        This class is a wrapper around the bit writer class. Its purpose is to provide unified interface for reading and writing.
        You can determine if you are writing to a stream by calling Stream::IsWriting inside your templated serialize method.
        This is evaluated at compile time, letting the compiler generate optimized serialize functions without the hassle of maintaining separate read and write functions.
        IMPORTANT: Generally, you don't call methods on this class directly. Use the serialize_* macros instead. See test/shared.h for some examples.
     */

    class WriteStream : public BaseStream
    {
    public:

        enum { IsWriting = 1 };
        enum { IsReading = 0 };

        /**
            Write stream constructor.
            @param buffer The buffer to write to.
            @param bytes The number of bytes in the buffer. Must be a multiple of four.
            @param allocator The allocator to use for stream allocations. This lets you dynamically allocate memory as you read and write packets.
         */

        WriteStream( uint8_t * buffer, int bytes ) : m_writer( buffer, bytes ) {}

        /**
            Serialize an integer (write).
            @param value The integer value in [min,max].
            @param min The minimum value.
            @param max The maximum value.
            @returns Always returns true. All checking is performed by debug asserts only on write.
         */

        bool SerializeInteger( int32_t value, int32_t min, int32_t max )
        {
            assert( min < max );
            assert( value >= min );
            assert( value <= max );
            const int bits = bits_required( min, max );
            uint32_t unsigned_value = value - min;
            m_writer.WriteBits( unsigned_value, bits );
            return true;
        }

        /**
            Serialize a number of bits (write).
            @param value The unsigned integer value to serialize. Must be in range [0,(1<<bits)-1].
            @param bits The number of bits to write in [1,32].
            @returns Always returns true. All checking is performed by debug asserts on write.
         */

        bool SerializeBits( uint32_t value, int bits )
        {
            assert( bits > 0 );
            assert( bits <= 32 );
            m_writer.WriteBits( value, bits );
            return true;
        }

        /**
            Serialize an array of bytes (write).
            @param data Array of bytes to be written.
            @param bytes The number of bytes to write.
            @returns Always returns true. All checking is performed by debug asserts on write.
         */

        bool SerializeBytes( const uint8_t * data, int bytes )
        {
            assert( data );
            assert( bytes >= 0 );
            SerializeAlign();
            m_writer.WriteBytes( data, bytes );
            return true;
        }

        /**
            Serialize an align (write).
            @returns Always returns true. All checking is performed by debug asserts on write.
         */

        bool SerializeAlign()
        {
            m_writer.WriteAlign();
            return true;
        }

        /** 
            If we were to write an align right now, how many bits would be required?
            @returns The number of zero pad bits required to achieve byte alignment in [0,7].
         */

        int GetAlignBits() const
        {
            return m_writer.GetAlignBits();
        }

        /**
            Flush the stream to memory after you finish writing.
            Always call this after you finish writing and before you call WriteStream::GetData, or you'll potentially truncate the last dword of data you wrote.
         */

        void Flush()
        {
            m_writer.FlushBits();
        }

        /**
            Get a pointer to the data written by the stream.
            IMPORTANT: Call WriteStream::Flush before you call this function!
            @returns A pointer to the data written by the stream
         */

        const uint8_t * GetData() const
        {
            return m_writer.GetData();
        }

        /**
            How many bytes have been written so far?
            @returns Number of bytes written. This is effectively the packet size.
         */

        int GetBytesProcessed() const
        {
            return m_writer.GetBytesWritten();
        }

        /**
            Get number of bits written so far.
            @returns Number of bits written.
         */

        int GetBitsProcessed() const
        {
            return m_writer.GetBitsWritten();
        }

    private:

        BitWriter m_writer;                 ///< The bit writer used for all bitpacked write operations.
    };

    /**
        Stream class for reading bitpacked data.
        This class is a wrapper around the bit reader class. Its purpose is to provide unified interface for reading and writing.
        You can determine if you are reading from a stream by calling Stream::IsReading inside your templated serialize method.
        This is evaluated at compile time, letting the compiler generate optimized serialize functions without the hassle of maintaining separate read and write functions.
        IMPORTANT: Generally, you don't call methods on this class directly. Use the serialize_* macros instead. See test/shared.h for some examples.
     */

    class ReadStream : public BaseStream
    {
    public:

        enum { IsWriting = 0 };
        enum { IsReading = 1 };

        /**
            Read stream constructor.
            @param buffer The buffer to read from.
            @param bytes The number of bytes in the buffer. May be a non-multiple of four, however if it is, the underlying buffer allocated should be large enough to read the any remainder bytes as a dword.
            @param allocator The allocator to use for stream allocations. This lets you dynamically allocate memory as you read and write packets.
         */

        ReadStream( const uint8_t * buffer, int bytes ) : BaseStream(), m_reader( buffer, bytes ) {}

        /**
            Serialize an integer (read).
            @param value The integer value read is stored here. It is guaranteed to be in [min,max] if this function succeeds.
            @param min The minimum allowed value.
            @param max The maximum allowed value.
            @returns Returns true if the serialize succeeded and the value is in the correct range. False otherwise.
         */

        bool SerializeInteger( int32_t & value, int32_t min, int32_t max )
        {
            assert( min < max );
            const int bits = bits_required( min, max );
            if ( m_reader.WouldReadPastEnd( bits ) )
                return false;
            uint32_t unsigned_value = m_reader.ReadBits( bits );
            value = (int32_t) unsigned_value + min;
            return true;
        }

        /**
            Serialize a number of bits (read).
            @param value The integer value read is stored here. Will be in range [0,(1<<bits)-1].
            @param bits The number of bits to read in [1,32].
            @returns Returns true if the serialize read succeeded, false otherwise.
         */

        bool SerializeBits( uint32_t & value, int bits )
        {
            assert( bits > 0 );
            assert( bits <= 32 );
            if ( m_reader.WouldReadPastEnd( bits ) )
                return false;
            uint32_t read_value = m_reader.ReadBits( bits );
            value = read_value;
            return true;
        }

        /**
            Serialize an array of bytes (read).
            @param data Array of bytes to read.
            @param bytes The number of bytes to read.
            @returns Returns true if the serialize read succeeded. False otherwise.
         */

        bool SerializeBytes( uint8_t * data, int bytes )
        {
            if ( !SerializeAlign() )
                return false;
            if ( m_reader.WouldReadPastEnd( bytes * 8 ) )
                return false;
            m_reader.ReadBytes( data, bytes );
            return true;
        }

        /**
            Serialize an align (read).
            @returns Returns true if the serialize read succeeded. False otherwise.
         */

        bool SerializeAlign()
        {
            const int alignBits = m_reader.GetAlignBits();
            if ( m_reader.WouldReadPastEnd( alignBits ) )
                return false;
            if ( !m_reader.ReadAlign() )
                return false;
            return true;
        }

        /** 
            If we were to read an align right now, how many bits would we need to read?
            @returns The number of zero pad bits required to achieve byte alignment in [0,7].
         */

        int GetAlignBits() const
        {
            return m_reader.GetAlignBits();
        }

        /**
            Get number of bits read so far.
            @returns Number of bits read.
         */

        int GetBitsProcessed() const
        {
            return m_reader.GetBitsRead();
        }

        /**
            How many bytes have been read so far?
            @returns Number of bytes read. Effectively this is the number of bits read, rounded up to the next byte where necessary.
         */

        int GetBytesProcessed() const
        {
            return ( m_reader.GetBitsRead() + 7 ) / 8;
        }

    private:

        BitReader m_reader;             ///< The bit reader used for all bitpacked read operations.
    };

    /**
        Serialize integer value (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The integer value to serialize in [min,max].
        @param min The minimum value.
        @param max The maximum value.
     */

    #define serialize_int( stream, value, min, max )                    \
        do                                                              \
        {                                                               \
            assert( min < max );                                  \
            int32_t int32_value = 0;                                    \
            if ( Stream::IsWriting )                                    \
            {                                                           \
                assert( int64_t(value) >= int64_t(min) );         \
                assert( int64_t(value) <= int64_t(max) );         \
                int32_value = (int32_t) value;                          \
            }                                                           \
            if ( !stream.SerializeInteger( int32_value, min, max ) )    \
            {                                                           \
                return false;                                           \
            }                                                           \
            if ( Stream::IsReading )                                    \
            {                                                           \
                value = int32_value;                                    \
                if ( int64_t(value) < int64_t(min) ||                   \
                     int64_t(value) > int64_t(max) )                    \
                {                                                       \
                    return false;                                       \
                }                                                       \
            }                                                           \
        } while (0)

    /**
        Serialize bits to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The unsigned integer value to serialize.
        @param bits The number of bits to serialize in [1,32].
     */

    #define serialize_bits( stream, value, bits )                       \
        do                                                              \
        {                                                               \
            assert( bits > 0 );                                   \
            assert( bits <= 32 );                                 \
            uint32_t uint32_value = 0;                                  \
            if ( Stream::IsWriting )                                    \
            {                                                           \
                uint32_value = (uint32_t) value;                        \
            }                                                           \
            if ( !stream.SerializeBits( uint32_value, bits ) )          \
            {                                                           \
                return false;                                           \
            }                                                           \
            if ( Stream::IsReading )                                    \
            {                                                           \
                value = uint32_value;                                   \
            }                                                           \
        } while (0)

    /**
        Serialize a boolean value to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The boolean value to serialize.
     */

    #define serialize_bool( stream, value )                             \
        do                                                              \
        {                                                               \
            uint32_t uint32_bool_value = 0;                             \
            if ( Stream::IsWriting )                                    \
            {                                                           \
                uint32_bool_value = value ? 1 : 0;                      \
            }                                                           \
            serialize_bits( stream, uint32_bool_value, 1 );             \
            if ( Stream::IsReading )                                    \
            {                                                           \
                value = uint32_bool_value ? true : false;               \
            }                                                           \
        } while (0)

    template <typename Stream> bool serialize_float_internal( Stream & stream, float & value )
    {
        uint32_t int_value;
        if ( Stream::IsWriting )
        {
            memcpy( &int_value, &value, 4 );
        }
        bool result = stream.SerializeBits( int_value, 32 );
        if ( Stream::IsReading && result )
        {
            memcpy( &value, &int_value, 4 );
        }
        return result;
    }

    /**
        Serialize floating point value (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The float value to serialize.
     */

    #define serialize_float( stream, value )                                        \
        do                                                                          \
        {                                                                           \
            if ( !relay::serialize_float_internal( stream, value ) )                \
            {                                                                       \
                return false;                                                       \
            }                                                                       \
        } while (0)

    /**
        Serialize a 32 bit unsigned integer to the stream (read/write).
        This is a helper macro to make unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The unsigned 32 bit integer value to serialize.
     */

    #define serialize_uint32( stream, value ) serialize_bits( stream, value, 32 );

    template <typename Stream> bool serialize_uint64_internal( Stream & stream, uint64_t & value )
    {
        uint32_t hi = 0, lo = 0;
        if ( Stream::IsWriting )
        {
            lo = value & 0xFFFFFFFF;
            hi = value >> 32;
        }
        serialize_bits( stream, lo, 32 );
        serialize_bits( stream, hi, 32 );
        if ( Stream::IsReading )
        {
            value = ( uint64_t(hi) << 32 ) | lo;
        }
        return true;
    }

    /**
        Serialize a 64 bit unsigned integer to the stream (read/write).
        This is a helper macro to make unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The unsigned 64 bit integer value to serialize.
     */

    #define serialize_uint64( stream, value )                                       \
        do                                                                          \
        {                                                                           \
            if ( !relay::serialize_uint64_internal( stream, value ) )               \
                return false;                                                       \
        } while (0)

    template <typename Stream> bool serialize_double_internal( Stream & stream, double & value )
    {
        union DoubleInt
        {
            double double_value;
            uint64_t int_value;
        };
        DoubleInt tmp = { 0 };
        if ( Stream::IsWriting )
        {
            tmp.double_value = value;
        }
        serialize_uint64( stream, tmp.int_value );
        if ( Stream::IsReading )
        {
            value = tmp.double_value;
        }
        return true;
    }

    /**
        Serialize double precision floating point value to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param value The double precision floating point value to serialize.
     */

    #define serialize_double( stream, value )                                       \
        do                                                                          \
        {                                                                           \
            if ( !relay::serialize_double_internal( stream, value ) )               \
            {                                                                       \
                return false;                                                       \
            }                                                                       \
        } while (0)

    template <typename Stream> bool serialize_bytes_internal( Stream & stream, uint8_t * data, int bytes )
    {
        return stream.SerializeBytes( data, bytes );
    }

    /**
        Serialize an array of bytes to the stream (read/write).
        This is a helper macro to make unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param data Pointer to the data to be serialized.
        @param bytes The number of bytes to serialize.
     */

    #define serialize_bytes( stream, data, bytes )                                  \
        do                                                                          \
        {                                                                           \
            if ( !relay::serialize_bytes_internal( stream, data, bytes ) )          \
            {                                                                       \
                return false;                                                       \
            }                                                                       \
        } while (0)

    template <typename Stream> bool serialize_string_internal( Stream & stream, char * string, int buffer_size )
    {
        int length = 0;
        if ( Stream::IsWriting )
        {
            length = (int) strlen( string );
            assert( length < buffer_size );
        }
        serialize_int( stream, length, 0, buffer_size - 1 );
        serialize_bytes( stream, (uint8_t*)string, length );
        if ( Stream::IsReading )
        {
            string[length] = '\0';
        }
        return true;
    }

    /**
        Serialize a string to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param string The string to serialize write. Pointer to buffer to be filled on read.
        @param buffer_size The size of the string buffer. String with terminating null character must fit into this buffer.
     */

    #define serialize_string( stream, string, buffer_size )                                 \
        do                                                                                  \
        {                                                                                   \
            if ( !relay::serialize_string_internal( stream, string, buffer_size ) )         \
            {                                                                               \
                return false;                                                               \
            }                                                                               \
        } while (0)

    /**
        Serialize an alignment to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
     */

    #define serialize_align( stream )                                                       \
        do                                                                                  \
        {                                                                                   \
            if ( !stream.SerializeAlign() )                                                 \
            {                                                                               \
                return false;                                                               \
            }                                                                               \
        } while (0)

    /**
        Serialize an object to the stream (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param object The object to serialize. Must have a serialize method on it.
     */

    #define serialize_object( stream, object )                                              \
        do                                                                                  \
        {                                                                                   \
            if ( !object.Serialize( stream ) )                                              \
            {                                                                               \
                return false;                                                               \
            }                                                                               \
        }                                                                                   \
        while(0)

    template <typename Stream, typename T> bool serialize_int_relative_internal( Stream & stream, T previous, T & current )
    {
        uint32_t difference = 0;
        if ( Stream::IsWriting )
        {
            assert( previous < current );
            difference = current - previous;
        }

        bool oneBit = false;
        if ( Stream::IsWriting )
        {
            oneBit = difference == 1;
        }
        serialize_bool( stream, oneBit );
        if ( oneBit )
        {
            if ( Stream::IsReading )
            {
                current = previous + 1;
            }
            return true;
        }
        
        bool twoBits = false;
        if ( Stream::IsWriting )
        {
            twoBits = difference <= 6;
        }
        serialize_bool( stream, twoBits );
        if ( twoBits )
        {
            serialize_int( stream, difference, 2, 6 );
            if ( Stream::IsReading )
            {
                current = previous + difference;
            }
            return true;
        }
        
        bool fourBits = false;
        if ( Stream::IsWriting )
        {
            fourBits = difference <= 23;
        }
        serialize_bool( stream, fourBits );
        if ( fourBits )
        {
            serialize_int( stream, difference, 7, 23 );
            if ( Stream::IsReading )
            {
                current = previous + difference;
            }
            return true;
        }

        bool eightBits = false;
        if ( Stream::IsWriting )
        {
            eightBits = difference <= 280;
        }
        serialize_bool( stream, eightBits );
        if ( eightBits )
        {
            serialize_int( stream, difference, 24, 280 );
            if ( Stream::IsReading )
            {
                current = previous + difference;
            }
            return true;
        }

        bool twelveBits = false;
        if ( Stream::IsWriting )
        {
            twelveBits = difference <= 4377;
        }
        serialize_bool( stream, twelveBits );
        if ( twelveBits )
        {
            serialize_int( stream, difference, 281, 4377 );
            if ( Stream::IsReading )
            {
                current = previous + difference;
            }
            return true;
        }

        bool sixteenBits = false;
        if ( Stream::IsWriting )
        {
            sixteenBits = difference <= 69914;
        }
        serialize_bool( stream, sixteenBits );
        if ( sixteenBits )
        {
            serialize_int( stream, difference, 4378, 69914 );
            if ( Stream::IsReading )
            {
                current = previous + difference;
            }
            return true;
        }

        uint32_t value = current;
        serialize_uint32( stream, value );
        if ( Stream::IsReading )
        {
            current = value;
        }

        return true;
    }

    /**
        Serialize an integer value relative to another (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param previous The previous integer value.
        @param current The current integer value.
     */

    #define serialize_int_relative( stream, previous, current )                             \
        do                                                                                  \
        {                                                                                   \
            if ( !relay::serialize_int_relative_internal( stream, previous, current ) )     \
            {                                                                               \
                return false;                                                               \
            }                                                                               \
        } while (0)

    template <typename Stream> bool serialize_ack_relative_internal( Stream & stream, uint16_t sequence, uint16_t & ack )
    {
        int ack_delta = 0;
        bool ack_in_range = false;
        if ( Stream::IsWriting )
        {
            if ( ack < sequence )
            {
                ack_delta = sequence - ack;
            }
            else
            {
                ack_delta = (int)sequence + 65536 - ack;
            }
            assert( ack_delta > 0 );
            assert( uint16_t( sequence - ack_delta ) == ack );
            ack_in_range = ack_delta <= 64;
        }
        serialize_bool( stream, ack_in_range );
        if ( ack_in_range )
        {
            serialize_int( stream, ack_delta, 1, 64 );
            if ( Stream::IsReading )
            {
                ack = sequence - ack_delta;
            }
        }
        else
        {
            serialize_bits( stream, ack, 16 );
        }
        return true;
    }

    /**
        Serialize an ack relative to the current sequence number (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param sequence The current sequence number.
        @param ack The ack sequence number, which is typically near the current sequence number.
     */

    #define serialize_ack_relative( stream, sequence, ack  )                                        \
        do                                                                                          \
        {                                                                                           \
            if ( !relay::serialize_ack_relative_internal( stream, sequence, ack ) )                 \
            {                                                                                       \
                return false;                                                                       \
            }                                                                                       \
        } while (0)

    template <typename Stream> bool serialize_sequence_relative_internal( Stream & stream, uint16_t sequence1, uint16_t & sequence2 )
    {
        if ( Stream::IsWriting )
        {
            uint32_t a = sequence1;
            uint32_t b = sequence2 + ( ( sequence1 > sequence2 ) ? 65536 : 0 );
            serialize_int_relative( stream, a, b );
        }
        else
        {
            uint32_t a = sequence1;
            uint32_t b = 0;
            serialize_int_relative( stream, a, b );
            if ( b >= 65536 )
            {
                b -= 65536;
            }
            sequence2 = uint16_t( b );
        }

        return true;
    }

    /**
        Serialize a sequence number relative to another (read/write).
        This is a helper macro to make writing unified serialize functions easier.
        Serialize macros returns false on error so we don't need to use exceptions for error handling on read. This is an important safety measure because packet data comes from the network and may be malicious.
        IMPORTANT: This macro must be called inside a templated serialize function with template \<typename Stream\>. The serialize method must have a bool return value.
        @param stream The stream object. May be a read or write stream.
        @param sequence1 The first sequence number to serialize relative to.
        @param sequence2 The second sequence number to be encoded relative to the first.
     */

    #define serialize_sequence_relative( stream, sequence1, sequence2 )                             \
        do                                                                                          \
        {                                                                                           \
            if ( !relay::serialize_sequence_relative_internal( stream, sequence1, sequence2 ) )     \
            {                                                                                       \
                return false;                                                                       \
            }                                                                                       \
        } while (0)

    template <typename Stream> bool serialize_address_internal( Stream & stream, relay_address_t & address )
    {
        serialize_bits( stream, address.type, 2 );
        if ( address.type == RELAY_ADDRESS_IPV4 )
        {
            serialize_bytes( stream, address.data.ipv4, 4 );
            serialize_bits( stream, address.port, 16 );
        }
        else if ( address.type == RELAY_ADDRESS_IPV6 )
        {
            for ( int i = 0; i < 8; ++i )
            {
                serialize_bits( stream, address.data.ipv6[i], 16 );
            }
            serialize_bits( stream, address.port, 16 );
        }
        else 
        {
            if ( Stream::IsReading )
            {
                memset( &address, 0, sizeof(relay_address_t) );
            }
        }
        return true;
    }

    #define serialize_address( stream, address )                                                    \
        do                                                                                          \
        {                                                                                           \
            if ( !relay::serialize_address_internal( stream, address ) )                            \
            {                                                                                       \
                return false;                                                                       \
            }                                                                                       \
        } while (0)
}

// --------------------------------------------------------------------------

int relay_wire_packet_bits( int packet_bytes )
{
    return ( 14 + 20 + 8 + packet_bytes + 4 ) * 8;
}

struct relay_bandwidth_limiter_t
{
    uint64_t bits_sent;
    double last_check_time;
    double average_kbps;
};

void relay_bandwidth_limiter_reset( relay_bandwidth_limiter_t * bandwidth_limiter )
{
    assert( bandwidth_limiter );
    bandwidth_limiter->last_check_time = -100.0;
    bandwidth_limiter->bits_sent = 0;
    bandwidth_limiter->average_kbps = 0.0;
}

bool relay_bandwidth_limiter_add_packet( relay_bandwidth_limiter_t * bandwidth_limiter, double current_time, uint32_t kbps_allowed, uint32_t packet_bits )
{
    assert( bandwidth_limiter );
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if ( invalid || current_time - bandwidth_limiter->last_check_time >= RELAY_BANDWIDTH_LIMITER_INTERVAL - 0.001f )
    {
        bandwidth_limiter->bits_sent = 0;
        bandwidth_limiter->last_check_time = current_time;
    }
    bandwidth_limiter->bits_sent += packet_bits;
    return bandwidth_limiter->bits_sent > (uint64_t) ( kbps_allowed * 1000 * RELAY_BANDWIDTH_LIMITER_INTERVAL );
}

void relay_bandwidth_limiter_add_sample( relay_bandwidth_limiter_t * bandwidth_limiter, double kbps )
{
    if ( bandwidth_limiter->average_kbps == 0.0 && kbps != 0.0 )
    {
        bandwidth_limiter->average_kbps = kbps;
        return;
    }

    if ( bandwidth_limiter->average_kbps != 0.0 && kbps == 0.0 )
    {
        bandwidth_limiter->average_kbps = 0.0;
        return;
    }

    const double delta = kbps - bandwidth_limiter->average_kbps;

    if ( delta < 0.000001f )
    {
        bandwidth_limiter->average_kbps = kbps;
        return;
    }

    bandwidth_limiter->average_kbps += delta * 0.1f;
}

double relay_bandwidth_limiter_usage_kbps( relay_bandwidth_limiter_t * bandwidth_limiter, double current_time )
{
    assert( bandwidth_limiter );
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if ( !invalid )
    {
        const double delta_time = current_time - bandwidth_limiter->last_check_time;
        if ( delta_time > 0.1f )
        {
            const double kbps = bandwidth_limiter->bits_sent / delta_time / 1000.0;
            relay_bandwidth_limiter_add_sample( bandwidth_limiter, kbps );
        }
    }
    return bandwidth_limiter->average_kbps;
}

// --------------------------------------------------------------------------

struct relay_route_token_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;
    int kbps_up;
    int kbps_down;
    relay_address_t next_address;
    uint8_t private_key[crypto_box_SECRETKEYBYTES];
};

void relay_write_route_token( relay_route_token_t * token, uint8_t * buffer, int buffer_length )
{
    (void) buffer_length;

    assert( token );
    assert( buffer );
    assert( buffer_length >= RELAY_ROUTE_TOKEN_BYTES );

    uint8_t * start = buffer;

    (void) start;

    relay_write_uint64( &buffer, token->expire_timestamp );
    relay_write_uint64( &buffer, token->session_id );
    relay_write_uint8( &buffer, token->session_version );
    relay_write_uint8( &buffer, token->session_flags );
    relay_write_uint32( &buffer, token->kbps_up );
    relay_write_uint32( &buffer, token->kbps_down );
    relay_write_address( &buffer, &token->next_address );
    relay_write_bytes( &buffer, token->private_key, crypto_box_SECRETKEYBYTES );

    assert( buffer - start == RELAY_ROUTE_TOKEN_BYTES );
}

void relay_read_route_token( relay_route_token_t * token, const uint8_t * buffer )
{
    assert( token );
    assert( buffer );

    const uint8_t * start = buffer;

    (void) start;   

    token->expire_timestamp = relay_read_uint64( &buffer );
    token->session_id = relay_read_uint64( &buffer );
    token->session_version = relay_read_uint8( &buffer );
    token->session_flags = relay_read_uint8( &buffer );
    token->kbps_up = relay_read_uint32( &buffer );
    token->kbps_down = relay_read_uint32( &buffer );
    relay_read_address( &buffer, &token->next_address );
    relay_read_bytes( &buffer, token->private_key, crypto_box_SECRETKEYBYTES );
    assert( buffer - start == RELAY_ROUTE_TOKEN_BYTES );
}

int relay_encrypt_route_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length )
{
    assert( sender_private_key );
    assert( receiver_public_key );
    assert( buffer );
    assert( buffer_length >= (int) ( RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES ) );

    (void) buffer_length;

    if ( crypto_box_easy( buffer, buffer, RELAY_ROUTE_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key ) != 0 )
    {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_decrypt_route_token( const uint8_t * sender_public_key, const uint8_t * receiver_private_key, const uint8_t * nonce, uint8_t * buffer )
{
    assert( sender_public_key );
    assert( receiver_private_key );
    assert( buffer );

    if ( crypto_box_open_easy( buffer, buffer, RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key ) != 0 )
    {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_write_encrypted_route_token( uint8_t ** buffer, relay_route_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key )
{
    assert( buffer );
    assert( token );
    assert( sender_private_key );
    assert( receiver_public_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes( nonce, crypto_box_NONCEBYTES );

    uint8_t * start = *buffer;

    (void) start;

    relay_write_bytes( buffer, nonce, crypto_box_NONCEBYTES );

    relay_write_route_token( token, *buffer, RELAY_ROUTE_TOKEN_BYTES );

    if ( relay_encrypt_route_token( sender_private_key, receiver_public_key, nonce, *buffer, RELAY_ROUTE_TOKEN_BYTES + crypto_box_NONCEBYTES ) != RELAY_OK )
        return RELAY_ERROR;

    *buffer += RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES;

    assert( ( *buffer - start ) == RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES );

    return RELAY_OK;
}

int relay_read_encrypted_route_token( uint8_t ** buffer, relay_route_token_t * token, const uint8_t * sender_public_key, const uint8_t * receiver_private_key )
{
    assert( buffer );
    assert( token );
    assert( sender_public_key );
    assert( receiver_private_key );

    const uint8_t * nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if ( relay_decrypt_route_token( sender_public_key, receiver_private_key, nonce, *buffer ) != RELAY_OK )
    {
        return RELAY_ERROR;
    }

    relay_read_route_token( token, *buffer );

    *buffer += RELAY_ROUTE_TOKEN_BYTES + crypto_box_MACBYTES;

    return RELAY_OK;
}

// --------------------------------------------------------------------------

struct relay_continue_token_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;
};

void relay_write_continue_token( relay_continue_token_t * token, uint8_t * buffer, int buffer_length )
{
    (void) buffer_length;

    assert( token );
    assert( buffer );
    assert( buffer_length >= RELAY_CONTINUE_TOKEN_BYTES );

    uint8_t * start = buffer;

    (void) start;

    relay_write_uint64( &buffer, token->expire_timestamp );
    relay_write_uint64( &buffer, token->session_id );
    relay_write_uint8( &buffer, token->session_version );
    relay_write_uint8( &buffer, token->session_flags );

    assert( buffer - start == RELAY_CONTINUE_TOKEN_BYTES );
}

void relay_read_continue_token( relay_continue_token_t * token, const uint8_t * buffer )
{
    assert( token );
    assert( buffer );

    const uint8_t * start = buffer;

    (void) start;

    token->expire_timestamp = relay_read_uint64( &buffer );
    token->session_id = relay_read_uint64( &buffer );
    token->session_version = relay_read_uint8( &buffer );
    token->session_flags = relay_read_uint8( &buffer );

    assert( buffer - start == RELAY_CONTINUE_TOKEN_BYTES );
}

int relay_encrypt_continue_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length )
{
    assert( sender_private_key );
    assert( receiver_public_key );
    assert( buffer );
    assert( buffer_length >= (int) ( RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES ) );

    (void) buffer_length;

    if ( crypto_box_easy( buffer, buffer, RELAY_CONTINUE_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key ) != 0 )
    {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_decrypt_continue_token( const uint8_t * sender_public_key, const uint8_t * receiver_private_key, const uint8_t * nonce, uint8_t * buffer )
{
    assert( sender_public_key );
    assert( receiver_private_key );
    assert( buffer );

    if ( crypto_box_open_easy( buffer, buffer, RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key ) != 0 )
    {
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

int relay_write_encrypted_continue_token( uint8_t ** buffer, relay_continue_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key )
{
    assert( buffer );
    assert( token );
    assert( sender_private_key );
    assert( receiver_public_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes( nonce, crypto_box_NONCEBYTES );

    uint8_t * start = *buffer;

    relay_write_bytes( buffer, nonce, crypto_box_NONCEBYTES );

    relay_write_continue_token( token, *buffer, RELAY_CONTINUE_TOKEN_BYTES );

    if ( relay_encrypt_continue_token( sender_private_key, receiver_public_key, nonce, *buffer, RELAY_CONTINUE_TOKEN_BYTES + crypto_box_NONCEBYTES ) != RELAY_OK )
        return RELAY_ERROR;

    *buffer += RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    (void) start;

    assert( ( *buffer - start ) == RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES );

    return RELAY_OK;
}

int relay_read_encrypted_continue_token( uint8_t ** buffer, relay_continue_token_t * token, const uint8_t * sender_public_key, const uint8_t * receiver_private_key )
{
    assert( buffer );
    assert( token );
    assert( sender_public_key );
    assert( receiver_private_key );

    const uint8_t * nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if ( relay_decrypt_continue_token( sender_public_key, receiver_private_key, nonce, *buffer ) != RELAY_OK )
    {
        return RELAY_ERROR;
    }

    relay_read_continue_token( token, *buffer );

    *buffer += RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    return RELAY_OK;
}

// --------------------------------------------------------------------------

int relay_write_header( int direction, uint8_t type, uint64_t sequence, uint64_t session_id, uint8_t session_version, uint8_t session_flags, const uint8_t * private_key, uint8_t * buffer, int buffer_length )
{
    assert( private_key );
    assert( buffer );
    assert( RELAY_HEADER_BYTES <= buffer_length );

    (void) buffer_length;

    uint8_t * start = buffer;

    (void) start;

    if ( direction == RELAY_DIRECTION_SERVER_TO_CLIENT )
    {
        // high bit must be set
        assert( sequence & ( 1ULL << 63 ) );
    }
    else
    {
        // high bit must be clear
        assert( ( sequence & ( 1ULL << 63 ) ) == 0 );
    }

    if ( type == RELAY_SESSION_PING_PACKET || type == RELAY_SESSION_PONG_PACKET || type == RELAY_ROUTE_RESPONSE_PACKET || type == RELAY_CONTINUE_RESPONSE_PACKET )
    {
        // second highest bit must be set
        assert( sequence & ( 1ULL << 62 ) );
    }
    else
    {
        // second highest bit must be clear
        assert( ( sequence & ( 1ULL << 62 ) ) == 0 );
    }

    relay_write_uint8( &buffer, type );

    relay_write_uint64( &buffer, sequence );

    uint8_t * additional = buffer;
    const int additional_length = 8 + 2;

    relay_write_uint64( &buffer, session_id );
    relay_write_uint8( &buffer, session_version );
    relay_write_uint8( &buffer, session_flags );

    uint8_t nonce[12];
    {
        uint8_t * p = nonce;
        relay_write_uint32( &p, 0 );
        relay_write_uint64( &p, sequence );
    }

    unsigned long long encrypted_length = 0;

    int result = crypto_aead_chacha20poly1305_ietf_encrypt( buffer, &encrypted_length,
                                                            buffer, 0,
                                                            additional, (unsigned long long) additional_length,
                                                            NULL, nonce, private_key );

    if ( result != 0 )
        return RELAY_ERROR;

    buffer += encrypted_length;

    assert( int( buffer - start ) == RELAY_HEADER_BYTES );

    return RELAY_OK;
}

int relay_peek_header( int direction, uint8_t * type, uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, uint8_t * session_flags, const uint8_t * buffer, int buffer_length )
{
    uint8_t packet_type;
    uint64_t packet_sequence;

    assert( buffer );

    if ( buffer_length < RELAY_HEADER_BYTES )
        return RELAY_ERROR;

    packet_type = relay_read_uint8( &buffer );

    packet_sequence = relay_read_uint64( &buffer );

    if ( direction == RELAY_DIRECTION_SERVER_TO_CLIENT )
    {
        // high bit must be set
        if ( !( packet_sequence & ( 1ULL << 63 ) ) )
            return RELAY_ERROR;
    }
    else
    {
        // high bit must be clear
        if ( packet_sequence & ( 1ULL << 63 ) )
            return RELAY_ERROR;
    }

    *type = packet_type;

    if ( *type == RELAY_SESSION_PING_PACKET || *type == RELAY_SESSION_PONG_PACKET || *type == RELAY_ROUTE_RESPONSE_PACKET || *type == RELAY_CONTINUE_RESPONSE_PACKET )
    {
        // second highest bit must be set
        assert( packet_sequence & ( 1ULL << 62 ) );
    }
    else
    {
        // second highest bit must be clear
        assert( ( packet_sequence & ( 1ULL << 62 ) ) == 0 );
    }


    *sequence = packet_sequence;
    *session_id = relay_read_uint64( &buffer );
    *session_version = relay_read_uint8( &buffer );
    *session_flags = relay_read_uint8( &buffer );

    return RELAY_OK;
}

int relay_read_header( int direction, uint8_t * type, uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, uint8_t * session_flags, const uint8_t * private_key, uint8_t * buffer, int buffer_length )
{
    assert( private_key );
    assert( buffer );

    if ( buffer_length < RELAY_HEADER_BYTES )
    {
        return RELAY_ERROR;
    }

    const uint8_t * p = buffer;

    uint8_t packet_type = relay_read_uint8( &p );

    uint64_t packet_sequence = relay_read_uint64( &p );

    if ( direction == RELAY_DIRECTION_SERVER_TO_CLIENT )
    {
        // high bit must be set
        if ( !( packet_sequence & ( 1ULL <<  63) ) )
        {
            return RELAY_ERROR;
        }
    }
    else
    {
        // high bit must be clear
        if ( packet_sequence & ( 1ULL << 63 ) )
        {
            return RELAY_ERROR;
        }
    }

    if ( packet_type == RELAY_SESSION_PING_PACKET || packet_type == RELAY_SESSION_PONG_PACKET || packet_type == RELAY_ROUTE_RESPONSE_PACKET || packet_type == RELAY_CONTINUE_RESPONSE_PACKET )
    {
        // second highest bit must be set
        assert( packet_sequence & ( 1ULL << 62 ) );
    }
    else
    {
        // second highest bit must be clear
        assert( ( packet_sequence & ( 1ULL << 62 ) ) == 0 );
    }

    const uint8_t * additional = p;

    const int additional_length = 8 + 2;

    uint64_t packet_session_id = relay_read_uint64( &p );
    uint8_t packet_session_version = relay_read_uint8( &p );
    uint8_t packet_session_flags = relay_read_uint8( &p );

    uint8_t nonce[12];
    {
        uint8_t * q = nonce;
        relay_write_uint32( &q, 0 );
        relay_write_uint64( &q, packet_sequence );
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt( buffer + 19, &decrypted_length, NULL,
                                                            buffer + 19, (unsigned long long) crypto_aead_chacha20poly1305_IETF_ABYTES,
                                                            additional, (unsigned long long) additional_length,
                                                            nonce, private_key );

    if ( result != 0 )
    {
        return RELAY_ERROR;
    }

    *type = packet_type;
    *sequence = packet_sequence;
    *session_id = packet_session_id;
    *session_version = packet_session_version;
    *session_flags = packet_session_flags;

    return RELAY_OK;
}

// -------------------------------------------------------------

static const unsigned char base64_table_encode[65] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

int relay_base64_encode_data( const uint8_t * input, size_t input_length, char * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

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

    if ( end - in )
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

int relay_base64_decode_data( const char * input, uint8_t * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

    size_t input_length = strlen( input );
    int pad = input_length > 0 && ( input_length % 4 || input[input_length - 1] == '=' );
    size_t L = ( ( input_length + 3 ) / 4 - pad ) * 4;
    size_t output_length = L / 4 * 3 + pad;

    if ( output_length > output_size )
    {
        return 0;
    }

    for ( size_t i = 0, j = 0; i < L; i += 4 )
    {
        int n = base64_table_decode[int( input[i] )] << 18 | base64_table_decode[int( input[i + 1] )] << 12 | base64_table_decode[int( input[i + 2] )] << 6 | base64_table_decode[int( input[i + 3] )];
        output[j++] = uint8_t( n >> 16 );
        output[j++] = uint8_t( n >> 8 & 0xFF );
        output[j++] = uint8_t( n & 0xFF );
    }

    if ( pad )
    {
        int n = base64_table_decode[int( input[L] )] << 18 | base64_table_decode[int( input[L + 1] )] << 12;
        output[output_length - 1] = uint8_t( n >> 16 );

        if (input_length > L + 2 && input[L + 2] != '=')
        {
            n |= base64_table_decode[int( input[L + 2] )] << 6;
            output_length += 1;
            if ( output_length > output_size )
            {
                return 0;
            }
            output[output_length - 1] = uint8_t( n >> 8 & 0xFF );
        }
    }

    return int( output_length );
}

int relay_base64_encode_string( const char * input, char * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

    return relay_base64_encode_data( (const uint8_t *)( input ), strlen( input ), output, output_size );
}

int relay_base64_decode_string( const char * input, char * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

    int output_length = relay_base64_decode_data( input, (uint8_t *)( output ), output_size - 1 );
    if ( output_length < 0 )
    {
        return 0;
    }

    output[output_length] = '\0';

    return output_length;
}

// --------------------------------------------------------------------------

#define RUN_TEST( test_function )                                           \
    do                                                                      \
    {                                                                       \
        printf( "    " #test_function "\n" );                               \
        fflush( stdout );                                                   \
        test_function();                                                    \
    }                                                                       \
    while (0)

static void check_handler( const char * condition,
                           const char * function,
                           const char * file,
                           int line )
{
    printf( "check failed: ( %s ), function %s, file %s, line %d\n", condition, function, file, line );
    fflush( stdout );
#ifndef NDEBUG
    #if defined( __GNUC__ )
        __builtin_trap();
    #elif defined( _MSC_VER )
        __debugbreak();
    #endif
#endif
    exit( 1 );
}

#define check( condition )                                                                                  \
do                                                                                                          \
{                                                                                                           \
    if ( !(condition) )                                                                                     \
    {                                                                                                       \
        check_handler( #condition, (const char*) __FUNCTION__, (const char*) __FILE__, __LINE__ );          \
    }                                                                                                       \
} while(0)

const int MaxItems = 11;

struct TestData
{
    TestData()
    {
        memset( this, 0, sizeof( TestData ) );
    }

    int a,b,c;
    uint32_t d : 8;
    uint32_t e : 8;
    uint32_t f : 8;
    bool g;
    int numItems;
    int items[MaxItems];
    float float_value;
    double double_value;
    uint64_t uint64_value;
    uint8_t bytes[17];
    char string[256];
    relay_address_t address_a, address_b, address_c;
};

struct TestContext
{
    int min;
    int max;
};

struct TestObject
{
    TestData data;

    void Init()
    {
        data.a = 1;
        data.b = -2;
        data.c = 150;
        data.d = 55;
        data.e = 255;
        data.f = 127;
        data.g = true;

        data.numItems = MaxItems / 2;
        for ( int i = 0; i < data.numItems; ++i )
            data.items[i] = i + 10;

        data.float_value = 3.1415926f;
        data.double_value = 1 / 3.0;
        data.uint64_value = 0x1234567898765432L;

        for ( int i = 0; i < (int) sizeof( data.bytes ); ++i )
            data.bytes[i] = ( i * 37 ) % 255;

        strcpy( data.string, "hello world!" );

        memset( &data.address_a, 0, sizeof(relay_address_t) );

        relay_address_parse( &data.address_b, "127.0.0.1:50000" );

        relay_address_parse( &data.address_c, "[::1]:50000" );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        const TestContext & context = *(const TestContext*) stream.GetContext();

        serialize_int( stream, data.a, context.min, context.max );
        serialize_int( stream, data.b, context.min, context.max );

        serialize_int( stream, data.c, -100, 10000 );

        serialize_bits( stream, data.d, 6 );
        serialize_bits( stream, data.e, 8 );
        serialize_bits( stream, data.f, 7 );

        serialize_align( stream );

        serialize_bool( stream, data.g );

	    serialize_int( stream, data.numItems, 0, MaxItems - 1 );
        for ( int i = 0; i < data.numItems; ++i )
            serialize_bits( stream, data.items[i], 8 );

        serialize_float( stream, data.float_value );

        serialize_double( stream, data.double_value );

        serialize_uint64( stream, data.uint64_value );

        serialize_bytes( stream, data.bytes, sizeof( data.bytes ) );

        serialize_string( stream, data.string, sizeof( data.string ) );

        serialize_address( stream, data.address_a );
        serialize_address( stream, data.address_b );
        serialize_address( stream, data.address_c );

        return true;
    }

    bool operator == ( const TestObject & other ) const
    {
        return memcmp( &data, &other.data, sizeof( TestData ) ) == 0;
    }

    bool operator != ( const TestObject & other ) const
    {
        return ! ( *this == other );
    }
};

static void test_endian()
{
    uint32_t value = 0x11223344;

    const char * bytes = (const char*) &value;

#if RELAY_LITTLE_ENDIAN

    check( bytes[0] == 0x44 );
    check( bytes[1] == 0x33 );
    check( bytes[2] == 0x22 );
    check( bytes[3] == 0x11 );

#else // #if RELAY_LITTLE_ENDIAN

    check( bytes[3] == 0x44 );
    check( bytes[2] == 0x33 );
    check( bytes[1] == 0x22 );
    check( bytes[0] == 0x11 );

#endif // #if RELAY_LITTLE_ENDIAN
}

static void test_bitpacker()
{
    const int BufferSize = 256;

    uint8_t buffer[BufferSize];

    relay::BitWriter writer( buffer, BufferSize );

    check( writer.GetData() == buffer );
    check( writer.GetBitsWritten() == 0 );
    check( writer.GetBytesWritten() == 0 );
    check( writer.GetBitsAvailable() == BufferSize * 8 );

    writer.WriteBits( 0, 1 );
    writer.WriteBits( 1, 1 );
    writer.WriteBits( 10, 8 );
    writer.WriteBits( 255, 8 );
    writer.WriteBits( 1000, 10 );
    writer.WriteBits( 50000, 16 );
    writer.WriteBits( 9999999, 32 );
    writer.FlushBits();

    const int bitsWritten = 1 + 1 + 8 + 8 + 10 + 16 + 32;

    check( writer.GetBytesWritten() == 10 );
    check( writer.GetBitsWritten() == bitsWritten );
    check( writer.GetBitsAvailable() == BufferSize * 8 - bitsWritten );

    const int bytesWritten = writer.GetBytesWritten();

    check( bytesWritten == 10 );

    memset( buffer + bytesWritten, 0, BufferSize - bytesWritten );

    relay::BitReader reader( buffer, bytesWritten );

    check( reader.GetBitsRead() == 0 );
    check( reader.GetBitsRemaining() == bytesWritten * 8 );
  
    uint32_t a = reader.ReadBits( 1 );
    uint32_t b = reader.ReadBits( 1 );
    uint32_t c = reader.ReadBits( 8 );
    uint32_t d = reader.ReadBits( 8 );
    uint32_t e = reader.ReadBits( 10 );
    uint32_t f = reader.ReadBits( 16 );
    uint32_t g = reader.ReadBits( 32 );

    check( a == 0 );
    check( b == 1 );
    check( c == 10 );
    check( d == 255 );
    check( e == 1000 );
    check( f == 50000 );
    check( g == 9999999 );

    check( reader.GetBitsRead() == bitsWritten );
    check( reader.GetBitsRemaining() == bytesWritten * 8 - bitsWritten );
}

static void test_stream()
{
    const int BufferSize = 1024;

    uint8_t buffer[BufferSize];

    TestContext context;
    context.min = -10;
    context.max = +10;

    relay::WriteStream writeStream( buffer, BufferSize );

    struct TestObject writeObject;
    writeObject.Init();
    writeStream.SetContext( &context );
    writeObject.Serialize( writeStream );
    writeStream.Flush();

    const int bytesWritten = writeStream.GetBytesProcessed();

    memset( buffer + bytesWritten, 0, BufferSize - bytesWritten );

    struct TestObject readObject;
    relay::ReadStream readStream( buffer, bytesWritten );
    readStream.SetContext( &context );
    readObject.Serialize( readStream );

    check( readObject == writeObject );
}

static void test_address()
{
    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "[" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "[]" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "[]:" ) == RELAY_ERROR );
        check( relay_address_parse( &address, ":" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "1" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "12" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "123" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "1234" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "1234.0.12313.0000" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "1234.0.12313.0000.0.0.0.0.0" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131" ) == RELAY_ERROR );
        check( relay_address_parse( &address, "." ) == RELAY_ERROR );
        check( relay_address_parse( &address, ".." ) == RELAY_ERROR );
        check( relay_address_parse( &address, "..." ) == RELAY_ERROR );
        check( relay_address_parse( &address, "...." ) == RELAY_ERROR );
        check( relay_address_parse( &address, "....." ) == RELAY_ERROR );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "107.77.207.77" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV4 );
        check( address.port == 0 );
        check( address.data.ipv4[0] == 107 );
        check( address.data.ipv4[1] == 77 );
        check( address.data.ipv4[2] == 207 );
        check( address.data.ipv4[3] == 77 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "127.0.0.1" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV4 );
        check( address.port == 0 );
        check( address.data.ipv4[0] == 127 );
        check( address.data.ipv4[1] == 0 );
        check( address.data.ipv4[2] == 0 );
        check( address.data.ipv4[3] == 1 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "107.77.207.77:40000" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV4 );
        check( address.port == 40000 );
        check( address.data.ipv4[0] == 107 );
        check( address.data.ipv4[1] == 77 );
        check( address.data.ipv4[2] == 207 );
        check( address.data.ipv4[3] == 77 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "127.0.0.1:40000" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV4 );
        check( address.port == 40000 );
        check( address.data.ipv4[0] == 127 );
        check( address.data.ipv4[1] == 0 );
        check( address.data.ipv4[2] == 0 );
        check( address.data.ipv4[3] == 1 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "fe80::202:b3ff:fe1e:8329" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV6 );
        check( address.port == 0 );
        check( address.data.ipv6[0] == 0xfe80 );
        check( address.data.ipv6[1] == 0x0000 );
        check( address.data.ipv6[2] == 0x0000 );
        check( address.data.ipv6[3] == 0x0000 );
        check( address.data.ipv6[4] == 0x0202 );
        check( address.data.ipv6[5] == 0xb3ff );
        check( address.data.ipv6[6] == 0xfe1e );
        check( address.data.ipv6[7] == 0x8329 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "::" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV6 );
        check( address.port == 0 );
        check( address.data.ipv6[0] == 0x0000 );
        check( address.data.ipv6[1] == 0x0000 );
        check( address.data.ipv6[2] == 0x0000 );
        check( address.data.ipv6[3] == 0x0000 );
        check( address.data.ipv6[4] == 0x0000 );
        check( address.data.ipv6[5] == 0x0000 );
        check( address.data.ipv6[6] == 0x0000 );
        check( address.data.ipv6[7] == 0x0000 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "::1" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV6 );
        check( address.port == 0 );
        check( address.data.ipv6[0] == 0x0000 );
        check( address.data.ipv6[1] == 0x0000 );
        check( address.data.ipv6[2] == 0x0000 );
        check( address.data.ipv6[3] == 0x0000 );
        check( address.data.ipv6[4] == 0x0000 );
        check( address.data.ipv6[5] == 0x0000 );
        check( address.data.ipv6[6] == 0x0000 );
        check( address.data.ipv6[7] == 0x0001 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "[fe80::202:b3ff:fe1e:8329]:40000" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV6 );
        check( address.port == 40000 );
        check( address.data.ipv6[0] == 0xfe80 );
        check( address.data.ipv6[1] == 0x0000 );
        check( address.data.ipv6[2] == 0x0000 );
        check( address.data.ipv6[3] == 0x0000 );
        check( address.data.ipv6[4] == 0x0202 );
        check( address.data.ipv6[5] == 0xb3ff );
        check( address.data.ipv6[6] == 0xfe1e );
        check( address.data.ipv6[7] == 0x8329 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "[::]:40000" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV6 );
        check( address.port == 40000 );
        check( address.data.ipv6[0] == 0x0000 );
        check( address.data.ipv6[1] == 0x0000 );
        check( address.data.ipv6[2] == 0x0000 );
        check( address.data.ipv6[3] == 0x0000 );
        check( address.data.ipv6[4] == 0x0000 );
        check( address.data.ipv6[5] == 0x0000 );
        check( address.data.ipv6[6] == 0x0000 );
        check( address.data.ipv6[7] == 0x0000 );
    }

    {
        struct relay_address_t address;
        check( relay_address_parse( &address, "[::1]:40000" ) == RELAY_OK );
        check( address.type == RELAY_ADDRESS_IPV6 );
        check( address.port == 40000 );
        check( address.data.ipv6[0] == 0x0000 );
        check( address.data.ipv6[1] == 0x0000 );
        check( address.data.ipv6[2] == 0x0000 );
        check( address.data.ipv6[3] == 0x0000 );
        check( address.data.ipv6[4] == 0x0000 );
        check( address.data.ipv6[5] == 0x0000 );
        check( address.data.ipv6[6] == 0x0000 );
        check( address.data.ipv6[7] == 0x0001 );
    }
}

static void test_replay_protection()
{
    relay_replay_protection_t replay_protection;

    int i;
    for ( i = 0; i < 2; ++i )
    {
        relay_replay_protection_reset( &replay_protection );

        check( replay_protection.most_recent_sequence == 0 );

        // the first time we receive packets, they should not be already received

        #define MAX_SEQUENCE ( RELAY_REPLAY_PROTECTION_BUFFER_SIZE * 4 )

        uint64_t sequence;
        for ( sequence = 0; sequence < MAX_SEQUENCE; ++sequence )
        {
            check( relay_replay_protection_already_received( &replay_protection, sequence ) == 0 );
            relay_replay_protection_advance_sequence( &replay_protection, sequence );
        }

        // old packets outside buffer should be considered already received

        check( relay_replay_protection_already_received( &replay_protection, 0 ) == 1 );

        // packets received a second time should be flagged already received

        for ( sequence = MAX_SEQUENCE - 10; sequence < MAX_SEQUENCE; ++sequence )
        {
            check( relay_replay_protection_already_received( &replay_protection, sequence ) == 1 );
        }

        // jumping ahead to a much higher sequence should be considered not already received

        check( relay_replay_protection_already_received( &replay_protection, MAX_SEQUENCE + RELAY_REPLAY_PROTECTION_BUFFER_SIZE ) == 0 );

        // old packets should be considered already received

        for ( sequence = 0; sequence < MAX_SEQUENCE; ++sequence )
        {
            check( relay_replay_protection_already_received( &replay_protection, sequence ) == 1 );
        }
    }
}

static void test_random_bytes()
{
    const int BufferSize = 64;
    uint8_t buffer[BufferSize];
    relay_random_bytes( buffer, BufferSize );
    for ( int i = 0; i < 100; ++i )
    {
        uint8_t next_buffer[BufferSize];
        relay_random_bytes( next_buffer, BufferSize );
        check( memcmp( buffer, next_buffer, BufferSize ) != 0 );
        memcpy( buffer, next_buffer, BufferSize );
    }
}

static void test_crypto_box()
{
    #define CRYPTO_BOX_MESSAGE (const unsigned char *) "test"
    #define CRYPTO_BOX_MESSAGE_LEN 4
    #define CRYPTO_BOX_CIPHERTEXT_LEN ( crypto_box_MACBYTES + CRYPTO_BOX_MESSAGE_LEN )

    unsigned char sender_publickey[crypto_box_PUBLICKEYBYTES];
    unsigned char sender_secretkey[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair( sender_publickey, sender_secretkey );

    unsigned char receiver_publickey[crypto_box_PUBLICKEYBYTES];
    unsigned char receiver_secretkey[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair( receiver_publickey, receiver_secretkey );

    unsigned char nonce[crypto_box_NONCEBYTES];
    unsigned char ciphertext[CRYPTO_BOX_CIPHERTEXT_LEN];
    relay_random_bytes( nonce, sizeof nonce );
    check( crypto_box_easy( ciphertext, CRYPTO_BOX_MESSAGE, CRYPTO_BOX_MESSAGE_LEN, nonce, receiver_publickey, sender_secretkey ) == 0 );

    unsigned char decrypted[CRYPTO_BOX_MESSAGE_LEN];
    check( crypto_box_open_easy( decrypted, ciphertext, CRYPTO_BOX_CIPHERTEXT_LEN, nonce, sender_publickey, receiver_secretkey ) == 0 );

    check( memcmp( decrypted, CRYPTO_BOX_MESSAGE, CRYPTO_BOX_MESSAGE_LEN ) == 0 );
}

static void test_crypto_secret_box()
{
    #define CRYPTO_SECRET_BOX_MESSAGE ((const unsigned char *) "test")
    #define CRYPTO_SECRET_BOX_MESSAGE_LEN 4
    #define CRYPTO_SECRET_BOX_CIPHERTEXT_LEN (crypto_secretbox_MACBYTES + CRYPTO_SECRET_BOX_MESSAGE_LEN)

    unsigned char key[crypto_secretbox_KEYBYTES];
    unsigned char nonce[crypto_secretbox_NONCEBYTES];
    unsigned char ciphertext[CRYPTO_SECRET_BOX_CIPHERTEXT_LEN];

    crypto_secretbox_keygen( key );
    randombytes_buf( nonce, crypto_secretbox_NONCEBYTES );
    crypto_secretbox_easy( ciphertext, CRYPTO_SECRET_BOX_MESSAGE, CRYPTO_SECRET_BOX_MESSAGE_LEN, nonce, key );

    unsigned char decrypted[CRYPTO_SECRET_BOX_MESSAGE_LEN];
    check( crypto_secretbox_open_easy( decrypted, ciphertext, CRYPTO_SECRET_BOX_CIPHERTEXT_LEN, nonce, key ) == 0 );
}

static void test_crypto_aead()
{
    #define CRYPTO_AEAD_MESSAGE (const unsigned char *) "test"
    #define CRYPTO_AEAD_MESSAGE_LEN 4
    #define CRYPTO_AEAD_ADDITIONAL_DATA (const unsigned char *) "123456"
    #define CRYPTO_AEAD_ADDITIONAL_DATA_LEN 6

    unsigned char nonce[crypto_aead_chacha20poly1305_NPUBBYTES];
    unsigned char key[crypto_aead_chacha20poly1305_KEYBYTES];
    unsigned char ciphertext[CRYPTO_AEAD_MESSAGE_LEN + crypto_aead_chacha20poly1305_ABYTES];
    unsigned long long ciphertext_len;

    crypto_aead_chacha20poly1305_keygen( key );
    randombytes_buf( nonce, sizeof(nonce) );

    crypto_aead_chacha20poly1305_encrypt( ciphertext, &ciphertext_len,
                                          CRYPTO_AEAD_MESSAGE, CRYPTO_AEAD_MESSAGE_LEN,
                                          CRYPTO_AEAD_ADDITIONAL_DATA, CRYPTO_AEAD_ADDITIONAL_DATA_LEN,
                                          NULL, nonce, key );

    unsigned char decrypted[CRYPTO_AEAD_MESSAGE_LEN];
    unsigned long long decrypted_len;
    check( crypto_aead_chacha20poly1305_decrypt( decrypted, &decrypted_len,
                                                 NULL,
                                                 ciphertext, ciphertext_len,
                                                 CRYPTO_AEAD_ADDITIONAL_DATA,
                                                 CRYPTO_AEAD_ADDITIONAL_DATA_LEN,
                                                 nonce, key) == 0 );
}

static void test_crypto_aead_ietf()
{
    #define CRYPTO_AEAD_IETF_MESSAGE (const unsigned char *) "test"
    #define CRYPTO_AEAD_IETF_MESSAGE_LEN 4
    #define CRYPTO_AEAD_IETF_ADDITIONAL_DATA (const unsigned char *) "123456"
    #define CRYPTO_AEAD_IETF_ADDITIONAL_DATA_LEN 6

    unsigned char nonce[crypto_aead_xchacha20poly1305_ietf_NPUBBYTES];
    unsigned char key[crypto_aead_xchacha20poly1305_ietf_KEYBYTES];
    unsigned char ciphertext[CRYPTO_AEAD_IETF_MESSAGE_LEN + crypto_aead_xchacha20poly1305_ietf_ABYTES];
    unsigned long long ciphertext_len;

    crypto_aead_xchacha20poly1305_ietf_keygen( key );
    randombytes_buf( nonce, sizeof(nonce) );

    crypto_aead_xchacha20poly1305_ietf_encrypt( ciphertext, &ciphertext_len, CRYPTO_AEAD_IETF_MESSAGE, CRYPTO_AEAD_IETF_MESSAGE_LEN, CRYPTO_AEAD_IETF_ADDITIONAL_DATA, CRYPTO_AEAD_IETF_ADDITIONAL_DATA_LEN, NULL, nonce, key);

    unsigned char decrypted[CRYPTO_AEAD_IETF_MESSAGE_LEN];
    unsigned long long decrypted_len;
    check(crypto_aead_xchacha20poly1305_ietf_decrypt( decrypted, &decrypted_len, NULL, ciphertext, ciphertext_len, CRYPTO_AEAD_IETF_ADDITIONAL_DATA, CRYPTO_AEAD_IETF_ADDITIONAL_DATA_LEN, nonce, key ) == 0 );
}

static void test_crypto_sign()
{
    #define CRYPTO_SIGN_MESSAGE (const unsigned char *) "test"
    #define CRYPTO_SIGN_MESSAGE_LEN 4

    unsigned char public_key[crypto_sign_PUBLICKEYBYTES];
    unsigned char private_key[crypto_sign_SECRETKEYBYTES];
    crypto_sign_keypair( public_key, private_key );

    unsigned char signed_message[crypto_sign_BYTES + CRYPTO_SIGN_MESSAGE_LEN];
    unsigned long long signed_message_len;

    crypto_sign( signed_message, &signed_message_len, CRYPTO_SIGN_MESSAGE, CRYPTO_SIGN_MESSAGE_LEN, private_key );

    unsigned char unsigned_message[CRYPTO_SIGN_MESSAGE_LEN];
    unsigned long long unsigned_message_len;
    check( crypto_sign_open( unsigned_message, &unsigned_message_len, signed_message, signed_message_len, public_key ) == 0 );
}

static void test_crypto_sign_detached()
{
    #define MESSAGE_PART1 ((const unsigned char *) "Arbitrary data to hash")
    #define MESSAGE_PART1_LEN 22

    #define MESSAGE_PART2 ((const unsigned char *) "is longer than expected")
    #define MESSAGE_PART2_LEN 23

    unsigned char public_key[crypto_sign_PUBLICKEYBYTES];
    unsigned char private_key[crypto_sign_SECRETKEYBYTES];
    crypto_sign_keypair( public_key, private_key );

    crypto_sign_state state;

    unsigned char signature[crypto_sign_BYTES];

    crypto_sign_init( &state );
    crypto_sign_update( &state, MESSAGE_PART1, MESSAGE_PART1_LEN );
    crypto_sign_update( &state, MESSAGE_PART2, MESSAGE_PART2_LEN );
    crypto_sign_final_create( &state, signature, NULL, private_key );

    crypto_sign_init( &state );
    crypto_sign_update( &state, MESSAGE_PART1, MESSAGE_PART1_LEN );
    crypto_sign_update( &state, MESSAGE_PART2, MESSAGE_PART2_LEN );
    check( crypto_sign_final_verify( &state, signature, public_key ) == 0 );
}

static void test_crypto_key_exchange()
{
    uint8_t client_public_key[crypto_kx_PUBLICKEYBYTES];
    uint8_t client_private_key[crypto_kx_SECRETKEYBYTES];
    crypto_kx_keypair( client_public_key, client_private_key );

    uint8_t server_public_key[crypto_kx_PUBLICKEYBYTES];
    uint8_t server_private_key[crypto_kx_SECRETKEYBYTES];
    crypto_kx_keypair( server_public_key, server_private_key );

    uint8_t client_send_key[crypto_kx_SESSIONKEYBYTES];
    uint8_t client_receive_key[crypto_kx_SESSIONKEYBYTES];
    check( crypto_kx_client_session_keys( client_receive_key, client_send_key, client_public_key, client_private_key, server_public_key ) == 0 );

    uint8_t server_send_key[crypto_kx_SESSIONKEYBYTES];
    uint8_t server_receive_key[crypto_kx_SESSIONKEYBYTES];
    check( crypto_kx_server_session_keys( server_receive_key, server_send_key, server_public_key, server_private_key, client_public_key ) == 0 );

    check( memcmp( client_send_key, server_receive_key, crypto_kx_SESSIONKEYBYTES ) == 0 );
    check( memcmp( server_send_key, client_receive_key, crypto_kx_SESSIONKEYBYTES ) == 0 );
}

static void test_basic_read_and_write()
{
    uint8_t buffer[1024];

    uint8_t * p = buffer;
    relay_write_uint8( &p, 105 );
    relay_write_uint16( &p, 10512 );
    relay_write_uint32( &p, 105120000 );
    relay_write_uint64( &p, 105120000000000000LL );
    relay_write_float32( &p, 100.0f );
    relay_write_float64( &p, 100000000000000.0 );
    relay_write_bytes( &p, (uint8_t*)"hello", 6 );
    relay_write_string( &p, "hey ho, let's go!", 32 );

    const uint8_t * q = buffer;

    uint8_t a = relay_read_uint8( &q );
    uint16_t b = relay_read_uint16( &q );
    uint32_t c = relay_read_uint32( &q );
    uint64_t d = relay_read_uint64( &q );
    float e = relay_read_float32( &q );
    double f = relay_read_float64( &q );
    uint8_t g[6];
    relay_read_bytes( &q, g, 6 );
    char string_buffer[32+1];
    memset( string_buffer, 0xFF, sizeof(string_buffer) );
    relay_read_string( &q, string_buffer, 32 );
    check( strcmp( string_buffer, "hey ho, let's go!" ) == 0 );

    check( a == 105 );
    check( b == 10512 );
    check( c == 105120000 );
    check( d == 105120000000000000LL );
    check( e == 100.0f );
    check( f == 100000000000000.0 );
    check( memcmp( g, "hello", 6 ) == 0 );

}

static void test_address_read_and_write()
{
    struct relay_address_t a, b, c;

    memset( &a, 0, sizeof(a) );

    relay_address_parse( &b, "127.0.0.1:50000" );

    relay_address_parse( &c, "[::1]:50000" );

    uint8_t buffer[1024];

    uint8_t * p = buffer;

    relay_write_address( &p, &a );
    relay_write_address( &p, &b );
    relay_write_address( &p, &c );

    struct relay_address_t read_a, read_b, read_c;

    const uint8_t * q = buffer;

    relay_read_address( &q, &read_a );
    relay_read_address( &q, &read_b );
    relay_read_address( &q, &read_c );

    check( relay_address_equal( &a, &read_a ) );
    check( relay_address_equal( &b, &read_b ) );
    check( relay_address_equal( &c, &read_c ) );
}

static void test_platform_socket()
{
    // non-blocking socket (ipv4)
    {
        relay_address_t bind_address;
        relay_address_t local_address;
        relay_address_parse( &bind_address, "0.0.0.0" );
        relay_address_parse( &local_address, "127.0.0.1" );
        relay_platform_socket_t * socket = relay_platform_socket_create( NULL, &bind_address, RELAY_PLATFORM_SOCKET_NON_BLOCKING, 0, 64*1024, 64*1024 );
        local_address.port = bind_address.port;
        check( socket );
        uint8_t packet[256];
        memset( packet, 0, sizeof(packet) );
        relay_platform_socket_send_packet( socket, &local_address, packet, sizeof(packet) );
        relay_address_t from;
        while ( relay_platform_socket_receive_packet( socket, &from, packet, sizeof(packet) ) )
        {
            check( relay_address_equal( &from, &local_address ) );
        }
        relay_platform_socket_destroy( socket );
    }

    // blocking socket with timeout (ipv4)
    {
        relay_address_t bind_address;
        relay_address_t local_address;
        relay_address_parse( &bind_address, "0.0.0.0" );
        relay_address_parse( &local_address, "127.0.0.1" );
        relay_platform_socket_t * socket = relay_platform_socket_create( NULL, &bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.01f, 64*1024, 64*1024 );
        local_address.port = bind_address.port;
        check( socket );
        uint8_t packet[256];
        memset( packet, 0, sizeof(packet) );
        relay_platform_socket_send_packet( socket, &local_address, packet, sizeof(packet) );
        relay_address_t from;
        while ( relay_platform_socket_receive_packet( socket, &from, packet, sizeof(packet) ) )
        {
            check( relay_address_equal( &from, &local_address ) );
        }
        relay_platform_socket_destroy( socket );
    }

    // blocking socket with no timeout (ipv4)
    {
        relay_address_t bind_address;
        relay_address_t local_address;
        relay_address_parse( &bind_address, "0.0.0.0" );
        relay_address_parse( &local_address, "127.0.0.1" );
        relay_platform_socket_t * socket = relay_platform_socket_create( NULL, &bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, -1.0f, 64*1024, 64*1024 );
        local_address.port = bind_address.port;
        check( socket );
        uint8_t packet[256];
        memset( packet, 0, sizeof(packet) );
        relay_platform_socket_send_packet( socket, &local_address, packet, sizeof(packet) );
        relay_address_t from;
        relay_platform_socket_receive_packet( socket, &from, packet, sizeof(packet) );
        check( relay_address_equal( &from, &local_address ) );
        relay_platform_socket_destroy( socket );
    }

    // non-blocking socket (ipv6)
#if RELAY_PLATFORM_HAS_IPV6
    {
        relay_address_t bind_address;
        relay_address_t local_address;
        relay_address_parse( &bind_address, "[::]" );
        relay_address_parse( &local_address, "[::1]" );
        relay_platform_socket_t * socket = relay_platform_socket_create( NULL, &bind_address, RELAY_PLATFORM_SOCKET_NON_BLOCKING, 0, 64*1024, 64*1024 );
        local_address.port = bind_address.port;
        check( socket );
        uint8_t packet[256];
        memset( packet, 0, sizeof(packet) );
        relay_platform_socket_send_packet( socket, &local_address, packet, sizeof(packet) );
        relay_address_t from;
        while ( relay_platform_socket_receive_packet( socket, &from, packet, sizeof(packet) ) )
        {
            check( relay_address_equal( &from, &local_address ) );
        }
        relay_platform_socket_destroy( socket );
    }

    // blocking socket with timeout (ipv6)
    {
        relay_address_t bind_address;
        relay_address_t local_address;
        relay_address_parse( &bind_address, "[::]" );
        relay_address_parse( &local_address, "[::1]" );
        relay_platform_socket_t * socket = relay_platform_socket_create( NULL, &bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.01f, 64*1024, 64*1024 );
        local_address.port = bind_address.port;
        check( socket );
        uint8_t packet[256];
        memset( packet, 0, sizeof(packet) );
        relay_platform_socket_send_packet( socket, &local_address, packet, sizeof(packet) );
        relay_address_t from;
        while ( relay_platform_socket_receive_packet( socket, &from, packet, sizeof(packet) ) )
        {
            check( relay_address_equal( &from, &local_address ) );
        }
        relay_platform_socket_destroy( socket );
    }

    // blocking socket with no timeout (ipv6)
    {
        relay_address_t bind_address;
        relay_address_t local_address;
        relay_address_parse( &bind_address, "[::]" );
        relay_address_parse( &local_address, "[::1]" );
        relay_platform_socket_t * socket = relay_platform_socket_create( NULL, &bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, -1.0f, 64*1024, 64*1024 );
        local_address.port = bind_address.port;
        check( socket );
        uint8_t packet[256];
        memset( packet, 0, sizeof(packet) );
        relay_platform_socket_send_packet( socket, &local_address, packet, sizeof(packet) );
        relay_address_t from;
        relay_platform_socket_receive_packet( socket, &from, packet, sizeof(packet) );
        check( relay_address_equal( &from, &local_address ) );
        relay_platform_socket_destroy( socket );
    }
#endif
}

static bool threads_work = false;

static relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC test_thread_function(void*)
{
    threads_work = true;
    RELAY_PLATFORM_THREAD_RETURN();
}

static void test_platform_thread()
{
    relay_platform_thread_t * thread = relay_platform_thread_create( NULL, test_thread_function, NULL );
    check( thread );
    relay_platform_thread_join( thread );
    relay_platform_thread_destroy( thread );
    check( threads_work );
}

static void test_platform_mutex()
{
    relay_platform_mutex_t * mutex = relay_platform_mutex_create( NULL );
    check( mutex );
    relay_platform_mutex_acquire( mutex );
    relay_platform_mutex_release( mutex );
    {
        relay_mutex_guard( mutex );
        // ...
    }
    relay_platform_mutex_destroy( mutex );
}

static void test_bandwidth_limiter()
{
    relay_bandwidth_limiter_t bandwidth_limiter;

    relay_bandwidth_limiter_reset( &bandwidth_limiter );

    check( relay_bandwidth_limiter_usage_kbps( &bandwidth_limiter, 0.0 ) == 0.0 );

    // come in way under
    {
        const int kbps_allowed = 1000;
        const int packet_bits = 50;

        for ( int i = 0; i < 10; ++i )
        {
            check( !relay_bandwidth_limiter_add_packet( &bandwidth_limiter, i * ( RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0 ), kbps_allowed, packet_bits ) );
        }
    }

    // get really close
    {
        relay_bandwidth_limiter_reset( &bandwidth_limiter );        

        const int kbps_allowed = 1000;
        const int packet_bits = kbps_allowed / 10 * 1000;

        for ( int i = 0; i < 10; ++i )
        {
            check( !relay_bandwidth_limiter_add_packet( &bandwidth_limiter, i * ( RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0 ), kbps_allowed, packet_bits ) );
        }
    }    

    // really close for several intervals
    {
        relay_bandwidth_limiter_reset( &bandwidth_limiter );        

        const int kbps_allowed = 1000;
        const int packet_bits = kbps_allowed / 10 * 1000;

        for ( int i = 0; i < 30; ++i )
        {
            check( !relay_bandwidth_limiter_add_packet( &bandwidth_limiter, i * ( RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0 ), kbps_allowed, packet_bits ) );
        }
    } 

    // go over budget
    {
        relay_bandwidth_limiter_reset( &bandwidth_limiter );        

        const int kbps_allowed = 1000;
        const int packet_bits = kbps_allowed / 10 * 1000 * 1.01f;

        bool over_budget = false;

        for ( int i = 0; i < 30; ++i )
        {
            over_budget |= relay_bandwidth_limiter_add_packet( &bandwidth_limiter, i * ( RELAY_BANDWIDTH_LIMITER_INTERVAL / 10.0 ), kbps_allowed, packet_bits );
        }

        check( over_budget );
    }
}

static void test_route_token()
{
    uint8_t buffer[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES];

    relay_route_token_t input_token;
    memset( &input_token, 0, sizeof( input_token ) );

    input_token.expire_timestamp = 1234241431241LL;
    input_token.session_id = 1234241431241LL;
    input_token.session_version = 5;
    input_token.session_flags = 1;
    input_token.next_address.type = RELAY_ADDRESS_IPV4;
    input_token.next_address.data.ipv4[0] = 127;
    input_token.next_address.data.ipv4[1] = 0;
    input_token.next_address.data.ipv4[2] = 0;
    input_token.next_address.data.ipv4[3] = 1;
    input_token.next_address.port = 40000;

    relay_write_route_token( &input_token, buffer, RELAY_ROUTE_TOKEN_BYTES );

    unsigned char sender_public_key[crypto_box_PUBLICKEYBYTES];
    unsigned char sender_private_key[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair( sender_public_key, sender_private_key );

    unsigned char receiver_public_key[crypto_box_PUBLICKEYBYTES];
    unsigned char receiver_private_key[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair( receiver_public_key, receiver_private_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes( nonce, crypto_box_NONCEBYTES );

    check( relay_encrypt_route_token( sender_private_key, receiver_public_key, nonce, buffer, sizeof( buffer ) ) == RELAY_OK );

    check( relay_decrypt_route_token( sender_public_key, receiver_private_key, nonce, buffer ) == RELAY_OK );

    relay_route_token_t output_token;

    relay_read_route_token( &output_token, buffer );

    check( input_token.expire_timestamp == output_token.expire_timestamp );
    check( input_token.session_id == output_token.session_id );
    check( input_token.session_version == output_token.session_version );
    check( input_token.session_flags == output_token.session_flags );
    check( input_token.kbps_up == output_token.kbps_up );
    check( input_token.kbps_down == output_token.kbps_down );
    check( memcmp( input_token.private_key, output_token.private_key, crypto_box_SECRETKEYBYTES ) == 0 );
    check( relay_address_equal( &input_token.next_address, &output_token.next_address ) == 1 );

    uint8_t * p = buffer;

    check( relay_write_encrypted_route_token( &p, &input_token, sender_private_key, receiver_public_key ) == RELAY_OK );

    p = buffer;

    check( relay_read_encrypted_route_token( &p, &output_token, sender_public_key, receiver_private_key ) == RELAY_OK );

    check( input_token.expire_timestamp == output_token.expire_timestamp );
    check( input_token.session_id == output_token.session_id );
    check( input_token.session_version == output_token.session_version );
    check( input_token.session_flags == output_token.session_flags );
    check( input_token.kbps_up == output_token.kbps_up );
    check( input_token.kbps_down == output_token.kbps_down );
    check( memcmp( input_token.private_key, output_token.private_key, crypto_box_SECRETKEYBYTES ) == 0 );
    check( relay_address_equal( &input_token.next_address, &output_token.next_address ) == 1 );
}

static void test_continue_token()
{
    uint8_t buffer[RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES];

    relay_continue_token_t input_token;
    memset( &input_token, 0, sizeof( input_token ) );

    input_token.expire_timestamp = 1234241431241LL;
    input_token.session_id = 1234241431241LL;
    input_token.session_version = 5;
    input_token.session_flags = 1;

    relay_write_continue_token( &input_token, buffer, RELAY_CONTINUE_TOKEN_BYTES );

    unsigned char sender_public_key[crypto_box_PUBLICKEYBYTES];
    unsigned char sender_private_key[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair( sender_public_key, sender_private_key );

    unsigned char receiver_public_key[crypto_box_PUBLICKEYBYTES];
    unsigned char receiver_private_key[crypto_box_SECRETKEYBYTES];
    crypto_box_keypair( receiver_public_key, receiver_private_key );

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes( nonce, crypto_box_NONCEBYTES );

    check( relay_encrypt_continue_token( sender_private_key, receiver_public_key, nonce, buffer, sizeof( buffer ) ) == RELAY_OK );

    check( relay_decrypt_continue_token( sender_public_key, receiver_private_key, nonce, buffer ) == RELAY_OK );

    relay_continue_token_t output_token;

    relay_read_continue_token( &output_token, buffer );

    check( input_token.expire_timestamp == output_token.expire_timestamp );
    check( input_token.session_id == output_token.session_id );
    check( input_token.session_version == output_token.session_version );
    check( input_token.session_flags == output_token.session_flags );

    uint8_t * p = buffer;

    check( relay_write_encrypted_continue_token( &p, &input_token, sender_private_key, receiver_public_key ) == RELAY_OK );

    p = buffer;

    memset( &output_token, 0, sizeof(output_token) );

    check( relay_read_encrypted_continue_token( &p, &output_token, sender_public_key, receiver_private_key ) == RELAY_OK );

    check( input_token.expire_timestamp == output_token.expire_timestamp );
    check( input_token.session_id == output_token.session_id );
    check( input_token.session_flags == output_token.session_flags );
}

static void test_header()
{
    uint8_t private_key[crypto_box_SECRETKEYBYTES];

    relay_random_bytes( private_key, crypto_box_SECRETKEYBYTES );

    uint8_t buffer[RELAY_MTU];

    // client -> server
    {
        uint64_t sequence = 123123130131LL;
        uint64_t session_id = 0x12313131;
        uint8_t session_version = 0x12;
        uint8_t session_flags = 0x1;

        check( relay_write_header( RELAY_DIRECTION_CLIENT_TO_SERVER, RELAY_CLIENT_TO_SERVER_PACKET, sequence, session_id, session_version, session_flags, private_key, buffer, sizeof( buffer ) ) == RELAY_OK );

        uint8_t read_type = 0;
        uint64_t read_sequence = 0;
        uint64_t read_session_id = 0;
        uint8_t read_session_version = 0;
        uint8_t read_session_flags = 0;

        check( relay_peek_header( RELAY_DIRECTION_CLIENT_TO_SERVER, &read_type, &read_sequence, &read_session_id, &read_session_version, &read_session_flags, buffer, sizeof( buffer ) ) == RELAY_OK );

        check( read_type == RELAY_CLIENT_TO_SERVER_PACKET );
        check( read_sequence == sequence );
        check( read_session_id == session_id );
        check( read_session_version == session_version );
        check( read_session_flags == session_flags );

        read_type = 0;
        read_sequence = 0;
        read_session_id = 0;
        read_session_version = 0;
        read_session_flags = 0;

        check( relay_read_header( RELAY_DIRECTION_CLIENT_TO_SERVER, &read_type, &read_sequence, &read_session_id, &read_session_version, &read_session_flags, private_key, buffer, sizeof( buffer ) ) == RELAY_OK );

        check( read_type == RELAY_CLIENT_TO_SERVER_PACKET );
        check( read_sequence == sequence );
        check( read_session_id == session_id );
        check( read_session_version == session_version );
        check( read_session_flags == session_flags );
    }

    // server -> client
    {
        uint64_t sequence = 123123130131LL | ( 1ULL << 63 );
        uint64_t session_id = 0x12313131;
        uint8_t session_version = 0x12;
        uint8_t session_flags = 0x1;

        check( relay_write_header( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_SERVER_TO_CLIENT_PACKET, sequence, session_id, session_version, session_flags, private_key, buffer, sizeof( buffer ) ) == RELAY_OK );

        uint8_t read_type = 0;
        uint64_t read_sequence = 0;
        uint64_t read_session_id = 0;
        uint8_t read_session_version = 0;
        uint8_t read_session_flags = 0;

        check( relay_peek_header( RELAY_DIRECTION_SERVER_TO_CLIENT, &read_type, &read_sequence, &read_session_id, &read_session_version, &read_session_flags, buffer, sizeof( buffer ) ) == RELAY_OK );

        check( read_type == RELAY_SERVER_TO_CLIENT_PACKET );
        check( read_sequence == sequence );
        check( read_session_id == session_id );
        check( read_session_version == session_version );
        check( read_session_flags == session_flags );

        read_type = 0;
        read_sequence = 0;
        read_session_id = 0;
        read_session_version = 0;
        read_session_flags = 0;

        check( relay_read_header( RELAY_DIRECTION_SERVER_TO_CLIENT, &read_type, &read_sequence, &read_session_id, &read_session_version, &read_session_flags, private_key, buffer, sizeof( buffer ) ) == RELAY_OK );

        check( read_type == RELAY_SERVER_TO_CLIENT_PACKET );
        check( read_sequence == sequence );
        check( read_session_id == session_id );
        check( read_session_version == session_version );
        check( read_session_flags == session_flags );
    }
}

static void test_base64()
{
    const char * input = "a test string. let's see if it works properly";
    char encoded[1024];
    char decoded[1024];
    check( relay_base64_encode_string( input, encoded, sizeof(encoded) ) > 0 );
    check( relay_base64_decode_string( encoded, decoded, sizeof(decoded) ) > 0 );
    check( strcmp( decoded, input ) == 0 );
    check( relay_base64_decode_string( encoded, decoded, 10 ) == 0 );
}

void relay_test()
{
    printf( "\nRunning relay tests:\n\n" );

    check( relay_init() == RELAY_OK );

    RUN_TEST( test_endian );
    RUN_TEST( test_bitpacker );
    RUN_TEST( test_stream );
    RUN_TEST( test_address );
    RUN_TEST( test_replay_protection );
    RUN_TEST( test_random_bytes );
    RUN_TEST( test_crypto_box );
    RUN_TEST( test_crypto_secret_box );
    RUN_TEST( test_crypto_aead );
    RUN_TEST( test_crypto_aead_ietf );
    RUN_TEST( test_crypto_sign );
    RUN_TEST( test_crypto_sign_detached );
    RUN_TEST( test_crypto_key_exchange );
    RUN_TEST( test_basic_read_and_write );
    RUN_TEST( test_address_read_and_write );
    RUN_TEST( test_platform_socket );
    RUN_TEST( test_platform_thread );
    RUN_TEST( test_platform_mutex );
    RUN_TEST( test_bandwidth_limiter );
    RUN_TEST( test_route_token );
    RUN_TEST( test_continue_token );
    RUN_TEST( test_header );
    RUN_TEST( test_base64 );
    
    printf( "\n" );

    fflush( stdout );

    relay_term();
}

// -----------------------------------------------------------------------------

template <typename T> struct relay_vector_t
{
    T * data;
    int length;
    int reserved;

    inline relay_vector_t( int reserve_count = 0 )
    {
        assert( reserve_count >= 0 );
        data = 0;
        length = 0;
        reserved = 0;
        if ( reserve_count > 0 )
        {
            reserve( reserve_count );
        }
    }

    inline ~relay_vector_t()
    {
        clear();
    }

    inline void clear() 
    {
        if ( data )
        {
            free( data );
        }
        data = NULL;
        length = 0;
        reserved = 0;
    }

    inline const T & operator [] ( int i ) const
    {
        assert( data );
        assert( i >= 0 && i < length );
        return *( data + i );
    }

    inline T & operator [] ( int i )
    {
        assert( data );
        assert( i >= 0 && i < length );
        return *( data + i );
    }

    void reserve( int size )
    {
        assert( size >= 0 );
        if ( size > reserved )
        {
            const double VECTOR_GROWTH_FACTOR = 1.5;
            const int VECTOR_INITIAL_RESERVATION = 1;
            unsigned int next_size = (unsigned int)( pow( VECTOR_GROWTH_FACTOR, int( log( double( size ) ) / log( double( VECTOR_GROWTH_FACTOR ) ) ) + 1 ) );
            if ( !reserved )
            {
                next_size = next_size > VECTOR_INITIAL_RESERVATION ? next_size : VECTOR_INITIAL_RESERVATION;
                data = (T*)( malloc( next_size * sizeof(T) ) );
            }
            else
            {
                data = (T*)( realloc( data, next_size * sizeof(T) ) );
            }
            assert( data );
            memset( (void*)( &data[reserved] ), 0, ( next_size - reserved ) * sizeof(T) );
            reserved = next_size;
        }
    }

    void resize( int i )
    {
        reserve( i );
        length = i;
    }

    void remove( int i )
    {
        assert( data );
        assert( i >= 0 && i < length );
        if ( i != length - 1 )
        {
            data[i] = data[length - 1];
        }
        length--;
    }

    void remove_ordered( int i )
    {
        assert( data );
        assert( i >= 0 && i < length );
        memmove( &data[i], &data[i + 1], sizeof( T ) * ( length - ( i + 1 ) ) );
        length--;
    }

    T * insert( int i )
    {
        assert( i >= 0 && i <= length );
        resize( length + 1 );
        memmove( &data[i + 1], &data[i], sizeof( T ) * ( length - 1 - i ) );
        return &data[i];
    }

    T * insert( int i, const T & t )
    {
        T * p = insert( i );
        *p = t;
        return p;
    }

    T * add()
    {
        reserve( ++length );
        return &data[length - 1];
    }

    T * add( const T & t )
    {
        T * p = add();
        *p = t;
        return p;
    }
};

// -----------------------------------------------------------------------------

typedef void CURL;
typedef void CURLM;
struct curl_slist;

typedef bool relay_http_callback_t( int status, const char * response, void * user_data ); // return false if the callback deleted the http context

struct relay_http_request_t
{
    relay_http_callback_t * callback;
    CURL * easy;
    void * user_data;
    curl_slist * request_headers;
    relay_vector_t<char> data;
    char error[256];
};

struct relay_http_t
{
    CURLM * multi;
    CURL * easy;
    relay_vector_t<relay_http_request_t*> active;
    char url_base[1024];
};

bool relay_curl_init()
{
    CURLcode result;
    if ( ( result = curl_global_init_mem( CURL_GLOBAL_ALL, malloc, free, realloc, strdup, calloc ) ) != CURLE_OK )
    {
        relay_printf( RELAY_LOG_LEVEL_ERROR, "failed to initialize curl: %s", curl_easy_strerror( result ) );
        return false;
    }

    return true;
}

void relay_http_term()
{
    curl_global_cleanup();
}

void relay_http_create( relay_http_t * context, const char * url )
{
    assert( url );
    assert( context );

    memset( context, 0, sizeof( *context ) );

    strncpy( context->url_base, url, 1023 );
}

static void http_request_cleanup( relay_http_request_t * request )
{
    curl_easy_cleanup( request->easy );
    request->data.clear();
    if ( request->request_headers )
    {
        curl_slist_free_all( request->request_headers );
    }
    free( request );
}

void relay_http_cancel_all( relay_http_t * context )
{
    for ( int i = 0; i < context->active.length; i++ )
    {
        relay_http_request_t * request = context->active[i];
        curl_multi_remove_handle( context->multi, request->easy );
        http_request_cleanup( request );
    }
    context->active.length = 0;
}

void relay_http_destroy( relay_http_t * context )
{
    if ( context->easy )
    {
        curl_easy_cleanup( context->easy );
    }

    relay_http_cancel_all( context );
    context->active.clear();

    if ( context->multi )
    {
        curl_multi_cleanup( context->multi );
    }
}

struct http_response_buffer
{
    char * pointer;
    size_t available;
};

static size_t relay_http_response_callback( void * contents, size_t size, size_t nmemb, void * userp )
{
    http_response_buffer * buffer = (http_response_buffer *)( userp );

    if ( buffer->available > 0 )
    {
        size_t available = size_t( buffer->available - 1 );
        size *= nmemb;
        size_t to_copy = ( size > available ) ? available : size;
        memcpy( buffer->pointer, contents, to_copy );
        buffer->available -= to_copy;
        buffer->pointer += to_copy;
        buffer->pointer[0] = '\0';
        return to_copy;
    }

    return 0;
}

static size_t relay_http_response_callback_null( void * contents, size_t size, size_t nmemb, void * userp )
{
    (void) contents;
    (void) userp;
    return nmemb * size;
}

void relay_platform_curl_easy_init( CURL * curl )
{
    (void) curl;
    // ...
}

static CURL * http_easy_init()
{
    CURL * curl = curl_easy_init();
    if ( curl )
    {
        curl_easy_setopt( curl, CURLOPT_VERBOSE, ( log_level >= RELAY_LOG_LEVEL_DEBUG ) ? 1L : 0L );
        curl_easy_setopt( curl, CURLOPT_SSL_VERIFYPEER, 1L );
        curl_easy_setopt( curl, CURLOPT_SSL_VERIFYHOST, 2L );
        curl_easy_setopt( curl, CURLOPT_USERAGENT, "next/1.0");
        relay_platform_curl_easy_init( curl );
    }
    return curl;
}

int relay_http_get( relay_http_t * context, const char * path, char * response, int * response_bytes, int timeout_ms )
{
    assert( path );
    assert( response );
    assert( response_bytes );
    assert( *response_bytes > 0 );

    if ( !context->easy )
    {
        context->easy = http_easy_init();
        if ( !context->easy )
            return -1;
    }

    curl_easy_setopt( context->easy, CURLOPT_HTTPGET, 1L ); 

    char url[1024];
    snprintf( url, sizeof(url), "%s%s", context->url_base, path );

    curl_easy_setopt( context->easy, CURLOPT_URL, url );

    curl_easy_setopt( context->easy, CURLOPT_HTTPHEADER, NULL );

    curl_easy_setopt( context->easy, CURLOPT_TIMEOUT_MS, long( timeout_ms ) );

    http_response_buffer buffer;
    buffer.pointer = response;
    buffer.available = *response_bytes;

    curl_easy_setopt( context->easy, CURLOPT_WRITEDATA, (void *)( &buffer ) );
    curl_easy_setopt( context->easy, CURLOPT_WRITEFUNCTION, relay_http_response_callback );

    int response_code;

    CURLcode result;
    if ( ( result = curl_easy_perform( context->easy ) ) == CURLE_OK )
    {
        long code;
        curl_easy_getinfo( context->easy, CURLINFO_RESPONSE_CODE, &code );
        response_code = int( code );
        *response_bytes = *response_bytes - int( buffer.available );
    }
    else
    {
        relay_printf( RELAY_LOG_LEVEL_WARN, "http request failed: %s", curl_easy_strerror( result ) );
        response[0] = '\0';
        *response_bytes = 0;
        response_code = -1;
    }

    return response_code;
}

int relay_http_post_json( relay_http_t * context, const char * path, const char * body, char * response, int * response_bytes, int timeout_ms )
{
    assert( path );
    assert( response );
    assert( response_bytes );
    assert( *response_bytes > 0 );

    if ( !context->easy )
    {
        context->easy = http_easy_init();
        if ( !context->easy )
            return -1;
    }

    curl_easy_setopt( context->easy, CURLOPT_POST, 1L );
    curl_easy_setopt( context->easy, CURLOPT_POSTFIELDS, body );

    char url[1024];
    snprintf( url, 1024, "%s%s", context->url_base, path );
    curl_easy_setopt( context->easy, CURLOPT_URL, url );

    struct curl_slist * headers = NULL;

    headers = curl_slist_append( headers, "Content-Type: application/json" );
    headers = curl_slist_append( headers, "Expect:" );

    curl_easy_setopt( context->easy, CURLOPT_HTTPHEADER, headers );

    curl_easy_setopt( context->easy, CURLOPT_TIMEOUT_MS, long( timeout_ms ) );

    http_response_buffer buffer;
    if ( response )
    {
        buffer.pointer = response;
        buffer.available = *response_bytes;

        curl_easy_setopt( context->easy, CURLOPT_WRITEDATA, (void *)( &buffer ) );
        curl_easy_setopt( context->easy, CURLOPT_WRITEFUNCTION, relay_http_response_callback );
    }
    else
    {
        curl_easy_setopt( context->easy, CURLOPT_WRITEDATA, NULL );
        curl_easy_setopt( context->easy, CURLOPT_WRITEFUNCTION, relay_http_response_callback_null );
        memset( &buffer, 0, sizeof( buffer ) );
    }

    int response_code;

    CURLcode result;
    if ( ( result = curl_easy_perform( context->easy ) ) == CURLE_OK )
    {
        long code;
        curl_easy_getinfo( context->easy, CURLINFO_RESPONSE_CODE, &code );
        response_code = int( code );
    }
    else
    {
        response_code = -1;
    }

    curl_easy_setopt( context->easy, CURLOPT_HTTPHEADER, NULL );

    curl_slist_free_all( headers );

    if ( response_bytes )
        *response_bytes = *response_bytes - int( buffer.available );

    return response_code;
}

static size_t relay_http_nonblock_response_callback( void * contents, size_t size, size_t nmemb, void * userp )
{
    relay_http_request_t * request = (relay_http_request_t *)( userp );
    size_t total = size * nmemb;
    int offset = request->data.length == 0 ? 0 : request->data.length - 1;
    request->data.resize( offset + int( total ) + 1 );
    memcpy( &request->data[offset], contents, total );
    request->data[request->data.length - 1] = '\0';
    return total;
}

void relay_http_nonblock_get( relay_http_t * context, const char * path, relay_http_callback_t * callback, void * user_data, int timeout_ms )
{
    assert( path );

    if ( !context->multi )
    {
        context->multi = curl_multi_init();
        assert( context->multi );
    }

    relay_http_request_t * request = (relay_http_request_t *)( malloc( sizeof( relay_http_request_t ) ) );
    assert( request );

    request->easy = http_easy_init();
    assert( request->easy );

    request->callback = callback;
    request->user_data = user_data;
    memset( &request->data, 0, sizeof( request->data ) );
    request->error[0] = '\0';
    request->request_headers = NULL;

    context->active.add( request );

    char url[1024];
    snprintf( url, sizeof(url), "%s%s", context->url_base, path );
    curl_easy_setopt( request->easy, CURLOPT_URL, url );

    curl_easy_setopt( request->easy, CURLOPT_HTTPGET, 1L ); 
    curl_easy_setopt( request->easy, CURLOPT_TIMEOUT_MS, long( timeout_ms ) );
    curl_easy_setopt( request->easy, CURLOPT_WRITEDATA, (void *)( request ) );

    if ( callback )
    {
        curl_easy_setopt( request->easy, CURLOPT_WRITEFUNCTION, relay_http_nonblock_response_callback );
    }
    else
    {
        curl_easy_setopt( request->easy, CURLOPT_WRITEFUNCTION, relay_http_response_callback_null );
    }

    curl_multi_add_handle( context->multi, request->easy );
}

void relay_http_nonblock_post_json( relay_http_t * context, const char * path, const char * body, relay_http_callback_t * callback, void * user_data, int timeout_ms )
{
    assert( path );

    if ( !context->multi )
    {
        context->multi = curl_multi_init();
        assert( context->multi );
    }

    relay_http_request_t * request = (relay_http_request_t *)( malloc( sizeof( relay_http_request_t ) ) );
    assert( request );

    request->easy = http_easy_init();
    assert( request->easy );

    request->callback = callback;
    request->user_data = user_data;
    memset( &request->data, 0, sizeof( request->data ) );
    request->error[0] = '\0';
    request->request_headers = NULL;
    request->request_headers = curl_slist_append( request->request_headers, "Content-Type: application/json" );
    request->request_headers = curl_slist_append( request->request_headers, "Expect:" );

    context->active.add( request );

    char url[1024];
    snprintf( url, sizeof(url), "%s%s", context->url_base, path );
    curl_easy_setopt( request->easy, CURLOPT_URL, url );

    curl_easy_setopt( request->easy, CURLOPT_POST, 1L );
    curl_easy_setopt( request->easy, CURLOPT_COPYPOSTFIELDS, body );
    curl_easy_setopt( request->easy, CURLOPT_HTTPHEADER, request->request_headers );
    curl_easy_setopt( request->easy, CURLOPT_TIMEOUT_MS, long( timeout_ms ) );
    curl_easy_setopt( request->easy, CURLOPT_WRITEDATA, (void *)( request ) );
    if ( callback )
    {
        curl_easy_setopt( request->easy, CURLOPT_WRITEFUNCTION, relay_http_nonblock_response_callback );
    }
    else
    {
        curl_easy_setopt( request->easy, CURLOPT_WRITEFUNCTION, relay_http_response_callback_null );
    }

    curl_multi_add_handle( context->multi, request->easy );
}

void relay_http_nonblock_update( relay_http_t * context )
{
    if ( !context->multi )
        return;

    int _; // never used
    curl_multi_perform(context->multi, &_);

    while ( CURLMsg * msg = curl_multi_info_read(context->multi, &_) )
    {
        if ( msg->msg == CURLMSG_DONE )
        {
            CURL * easy = msg->easy_handle;
            curl_multi_remove_handle( context->multi, easy );
            relay_http_request_t * request = NULL;
            for (int i = 0; i < context->active.length; i++ )
            {
                relay_http_request_t * r = context->active[i];
                if (r->easy == easy)
                {
                    request = r;
                    context->active.remove( i );
                    break;
                }
            }
            assert( request );

            long response_code;
            curl_easy_getinfo(request->easy, CURLINFO_RESPONSE_CODE, &response_code);
            const char* response = request->data.length > 0 ? &request->data[0] : NULL;
#if RELAY_DEBUG_HTTP
            {
                char * url = NULL;
                curl_easy_getinfo(request->easy, CURLINFO_EFFECTIVE_URL, &url);
                relay_printf( RELAY_LOG_LEVEL_DEBUG, "HTTP response - %s - code %d: %s", url, response_code, response );
                if ( request->error[0] )
                    relay_printf( RELAY_LOG_LEVEL_DEBUG, "HTTP error: %s", request->error );
            }
#endif // #if RELAY_DEBUG_HTTP
            if (request->callback)
            {
                if ( !request->callback(response_code, response, request->user_data) )
                {
                    // if the callback returns false, that means it deleted the HTTP context. exit immediately.
                    break;
                }
            }
            http_request_cleanup( request );
        }
    }
}

// -----------------------------------------------------------------------------

struct relay_t
{
    // ...
};

int main( int argc, const char ** argv )
{
    if ( argc == 2 && strcmp( argv[1], "test" ) == 0 )
    {
        relay_test();
        return 0;
    }

    const char * relay_id = relay_platform_getenv( "RELAY_ID" );
    if ( !relay_id || relay_id[0] == '\0' )
    {
        printf( "\nerror: RELAY_ID not set\n\n" );
        return 1;
    }

    const char * relay_address_env = relay_platform_getenv( "RELAY_ADDRESS" );
    if ( !relay_address_env )
    {
        printf( "\nerror: RELAY_ADDRESS not set\n\n" );
        return 1;
    }

    relay_address_t relay_address;
    if ( relay_address_parse( &relay_address, relay_address_env ) != RELAY_OK )
    {
        printf( "\nerror: invalid relay address '%s'\n\n", relay_address_env );
        return 1;
    }

    uint16_t relay_port = relay_address.port;

    const char * relay_private_key_env = relay_platform_getenv( "RELAY_PRIVATE_KEY" );
    if ( !relay_private_key_env )
    {
        printf( "\nerror: RELAY_PRIVATE_KEY not set\n\n" );
        return 1;
    }

    uint8_t relay_private_key[crypto_sign_SECRETKEYBYTES];
    if ( relay_base64_decode_data( relay_private_key_env, relay_private_key, crypto_sign_SECRETKEYBYTES ) != crypto_sign_SECRETKEYBYTES )
    {
        printf( "\nerror: invalid relay private key\n\n" );
        return 1;
    }

    const char * relay_public_key_env = relay_platform_getenv( "RELAY_PUBLIC_KEY" );
    if ( !relay_public_key_env )
    {
        printf( "\nerror: RELAY_PUBLIC_KEY not set\n\n" );
        return 1;
    }

    uint8_t relay_public_key[crypto_sign_PUBLICKEYBYTES];
    if ( relay_base64_decode_data( relay_public_key_env, relay_public_key, crypto_sign_PUBLICKEYBYTES ) != crypto_sign_PUBLICKEYBYTES )
    {
        printf( "\nerror: invalid relay public key\n\n" );
        return 1;
    }

    const char * router_public_key_env = relay_platform_getenv( "RELAY_ROUTER_PUBLIC_KEY" );
    if ( !router_public_key_env )
    {
        printf( "\nerror: RELAY_ROUTER_PUBLIC_KEY not set\n\n" );
        return 1;
    }

    uint8_t router_public_key[crypto_sign_PUBLICKEYBYTES];
    if ( relay_base64_decode_data( router_public_key_env, router_public_key, crypto_sign_PUBLICKEYBYTES ) != crypto_sign_PUBLICKEYBYTES )
    {
        printf( "\nerror: invalid router public key\n\n" );
        return 1;
    }

    const char * backend_hostname = relay_platform_getenv( "RELAY_BACKEND_HOSTNAME" );
    if ( !backend_hostname )
    {
        printf( "\nerror: RELAY_BACKEND_HOSTNAME not set\n\n" );
        return 1;
    }

    (void) relay_id;
    (void) relay_address;
    (void) relay_port;
    (void) relay_private_key;
    (void) relay_public_key;
    (void) router_public_key;

    if ( relay_init() != RELAY_OK )
    {
        printf( "\nerror: failed to initialize relay\n\n" );
        return 1;
    }

    printf( "\nHello, relay world!\n\n" );

    relay_t relay;

    (void) relay;

    // ---------------------------------------------------------------------

    const uint32_t hello_magic = 0x9083708f;

    uint32_t hello_version = 0;

    uint8_t hello_data[1024];
    memset( hello_data, 0, sizeof(hello_data) );

    uint8_t hello_nonce[32];
    relay_random_bytes( hello_nonce, 32 );

    uint32_t hello_length = 0;

    uint8_t * p = hello_data;
    relay_write_uint32( &p, hello_magic );
    relay_write_uint32( &p, hello_version );
    uint8_t * q = p;
    relay_write_uint32( &p, hello_length );
    relay_write_string( &p, relay_id, 256 );
    relay_write_bytes( &p, hello_nonce, 32 );

    hello_length = (uint32_t) ( p - hello_data );

    relay_write_uint32( &q, hello_length );

    uint8_t * sign_data = hello_data;
    int sign_length = hello_length;

    uint8_t signed_hello_data[crypto_sign_BYTES + 1024];
    unsigned long long signed_hello_length;
    if ( crypto_sign( signed_hello_data, &signed_hello_length, sign_data, sign_length, relay_private_key ) != 0 )
    {
        printf( "\nerror: failed to sign relay hello data\n\n" );
    }

    printf( "signed hello data is %d bytes\n\n", (int) signed_hello_length );

    // ---------------------------------------------------------------------

    CURL * curl = curl_easy_init();
    if ( !curl )
    {
        printf( "\nerror: could not initialize curl\n\n" );
        return 1;
    }

    struct curl_slist * slist = curl_slist_append( NULL, "Content-Type:application/octet-stream" );

    curl_easy_setopt( curl, CURLOPT_BUFFERSIZE, 102400L );
    curl_easy_setopt( curl, CURLOPT_URL, "http://localhost:30000/relay_hello" );
    curl_easy_setopt( curl, CURLOPT_NOPROGRESS, 1L );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDS, signed_hello_data );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)signed_hello_length );
    curl_easy_setopt( curl, CURLOPT_HTTPHEADER, slist );
    curl_easy_setopt( curl, CURLOPT_USERAGENT, "curl/7.64.1" );
    curl_easy_setopt( curl, CURLOPT_MAXREDIRS, 50L );
    curl_easy_setopt( curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS );
    curl_easy_setopt( curl, CURLOPT_TCP_KEEPALIVE, 1L );
    curl_easy_setopt( curl, CURLOPT_TIMEOUT_MS, long( 10*1000 ) );

    CURLcode ret = curl_easy_perform( curl );

    curl_slist_free_all( slist );
    slist = NULL;

    if ( ret != 0 )
    {
        printf( "\nerror: could not post relay hello\n\n" );
        curl_easy_cleanup( curl );
        relay_term();
        return 1;
    }

    // ---------------------------------------------------------------------

    // ...

    // ---------------------------------------------------------------------

    curl_easy_cleanup( curl );

    relay_term();

    return 0;
}
