/*
 * Network Next Relay.
 * Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
 */

#include "relay.h"
#include "relay_version.h"
#include <assert.h>
#include <string.h>
#include <stdio.h>
#include <inttypes.h>
#include <stdarg.h>
#include <sodium.h>
#include <math.h>
#include <map>
#include <float.h>
#include <signal.h>
#include <atomic>

#include "curl/curl.h"

#define RELAY_DEVELOPMENT                                          1

#define INTENSIVE_RELAY_DEBUGGING                                  0

#define RELAY_MTU                                               1300

#define RELAY_HEADER_BYTES_SDK5                                   33

#define RELAY_ADDRESS_BYTES                                       19
#define RELAY_ADDRESS_BYTES_SHORT                                  7
#define RELAY_ADDRESS_BUFFER_SAFETY                               32

#define RELAY_REPLAY_PROTECTION_BUFFER_SIZE                      256

#define RELAY_BANDWIDTH_LIMITER_INTERVAL                         1.0

#define RELAY_ROUTE_TOKEN_BYTES                                   76
#define RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES                        116
#define RELAY_CONTINUE_TOKEN_BYTES                                17
#define RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES                      57

#define RELAY_PING_TOKEN_BYTES_SDK5                               46
#define RELAY_ENCRYPTED_PING_TOKEN_BYTES_SDK5                     86

#define RELAY_DIRECTION_CLIENT_TO_SERVER                           0
#define RELAY_DIRECTION_SERVER_TO_CLIENT                           1

#define RELAY_ROUTE_REQUEST_PACKET_SDK5                            9
#define RELAY_ROUTE_RESPONSE_PACKET_SDK5                          10
#define RELAY_CLIENT_TO_SERVER_PACKET_SDK5                        11
#define RELAY_SERVER_TO_CLIENT_PACKET_SDK5                        12
#define RELAY_SESSION_PING_PACKET_SDK5                            13
#define RELAY_SESSION_PONG_PACKET_SDK5                            14
#define RELAY_CONTINUE_REQUEST_PACKET_SDK5                        15
#define RELAY_CONTINUE_RESPONSE_PACKET_SDK5                       16
#define RELAY_NEAR_PING_PACKET_SDK5                               20
#define RELAY_NEAR_PONG_PACKET_SDK5                               21

#define RELAY_PING_HISTORY_ENTRY_COUNT                           256

#define RELAY_PING_TIME                                          0.1

#define RELAY_STATS_WINDOW                                      10.0
#define RELAY_PING_SAFETY                                        1.0

#define RELAY_MAX_PACKET_BYTES                                  1500

#define RELAY_PUBLIC_KEY_BYTES                                    32
#define RELAY_PRIVATE_KEY_BYTES                                   32

#define RELAY_MAX_UPDATE_ATTEMPTS                                 30

#define RELAY_COUNTER_PACKETS_SENT                                                               0
#define RELAY_COUNTER_PACKETS_RECEIVED                                                           1
#define RELAY_COUNTER_BYTES_SENT                                                                 2
#define RELAY_COUNTER_BYTES_RECEIVED                                                             3
#define RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET                                         4
#define RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET                                      5
#define RELAY_COUNTER_SESSION_CREATED                                                            6
#define RELAY_COUNTER_SESSION_CONTINUED                                                          7
#define RELAY_COUNTER_SESSION_DESTROYED                                                          8

#define RELAY_COUNTER_RELAY_PING_PACKET_SENT                                                    10
#define RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED                                                11
#define RELAY_COUNTER_RELAY_PONG_PACKET_SENT                            		 		        12
#define RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED                         			 		    13

#define RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED                                                 20
#define RELAY_COUNTER_NEAR_PING_PACKET_BAD_SIZE                                                 21
#define RELAY_COUNTER_NEAR_PING_PACKET_RESPONDED_WITH_PONG                                      22

#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED                					 		    30
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_BAD_SIZE                					 		    31
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_READ_TOKEN              		 		 	32
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED                  					 	33
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS				 	34
#define RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS   	   			35

#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED                       	   	    		    40
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_BAD_SIZE                                			41
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER                          		42
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION                         		43
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_SESSION_EXPIRED                           	    	44
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED                          			45
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY                     			46
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS      		47
#define RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS    		48

#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED                              			50
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_BAD_SIZE                              			51
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_READ_TOKEN                         		52
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED                                     53
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_SESSION_EXPIRED                                   54
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS           		55
#define RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS         		56

#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED											60
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_BAD_SIZE										    61
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER							62
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED                     			63
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION							64
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED 									65
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY                       		66
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS      		67
#define RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS    		68

#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED 											70
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL 										71
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG 											72
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_PEEK_HEADER 					 		73
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION 							74
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_SESSION_EXPIRED 									75
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED 									76
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_VERIFY_HEADER 							77
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS 				78
#define RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS 				79

#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED 									        80
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL 										81
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG 											82
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_PEEK_HEADER 							83
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION                			84
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED									85
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED 									86
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_VERIFY_HEADER 							87
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS 	    	88
#define RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS      	89

#define RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED                                              90
#define RELAY_COUNTER_SESSION_PING_PACKET_BAD_PACKET_SIZE                                       91
#define RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_PEEK_HEADER                                 92
#define RELAY_COUNTER_SESSION_PING_PACKET_SESSION_DOES_NOT_EXIST                                93
#define RELAY_COUNTER_SESSION_PING_PACKET_SESSION_EXPIRED                                       94
#define RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED                                      95
#define RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_VERIFY_HEADER                               96
#define RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS                    97
#define RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS                  98

#define RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED                                             100
#define RELAY_COUNTER_SESSION_PONG_PACKET_BAD_SIZE                                             101
#define RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_PEEK_HEADER                                102
#define RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_DOES_NOT_EXIST                               103
#define RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_EXPIRED                                      104
#define RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED                                     105
#define RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_VERIFY_HEADER                              106
#define RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS               107
#define RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS             108

#define NUM_RELAY_COUNTERS                                                                     128

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

extern relay_platform_socket_t * relay_platform_socket_create( struct relay_address_t * address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size );

extern void relay_platform_socket_destroy( relay_platform_socket_t * socket );

extern void relay_platform_socket_send_packet( relay_platform_socket_t * socket, const relay_address_t * to, const void * packet_data, int packet_bytes );

extern int relay_platform_socket_receive_packet( relay_platform_socket_t * socket, relay_address_t * from, void * packet_data, int max_packet_size );

extern relay_platform_thread_t * relay_platform_thread_create( relay_platform_thread_func_t * func, void * arg );

extern void relay_platform_thread_join( relay_platform_thread_t * thread );

extern void relay_platform_thread_destroy( relay_platform_thread_t * thread );

extern void relay_platform_thread_set_sched_max( relay_platform_thread_t * thread );

extern relay_platform_mutex_t * relay_platform_mutex_create();

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

int relay_initialize()
{
    if ( relay_platform_init() != RELAY_OK )
    {
        printf( "error: failed to initialize platform" );
        return RELAY_ERROR;
    }

    if ( sodium_init() == -1 )
    {
        printf( "error: failed to initialize sodium" );
        return RELAY_ERROR;
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
    if ( length > max_length - 1 )
        length = max_length - 1;
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

void relay_read_address_short( const uint8_t ** buffer, relay_address_t * address )
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
    }
    else
    {
        for ( int i = 0; i < RELAY_ADDRESS_BYTES_SHORT - 1; ++i )
        {
            uint8_t dummy = relay_read_uint8( buffer ); (void) dummy;
        }
    }

    (void) start;

    assert( *buffer - start == RELAY_ADDRESS_BYTES_SHORT );
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

void relay_write_address_short( uint8_t ** buffer, const relay_address_t * address )
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
    }
    else
    {
        for ( int i = 0; i < RELAY_ADDRESS_BYTES_SHORT; ++i )
        {
            relay_write_uint8( buffer, 0 );
        }
    }

    (void) start;

    assert( *buffer - start == RELAY_ADDRESS_BYTES_SHORT );
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
                printf( "address string truncated: [%s]:%hu\n", address_string, address->port );
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
    int kbps_up;
    int kbps_down;
    relay_address_t next_address;
    uint8_t next_internal;
    uint8_t prev_internal;
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
    relay_write_uint32( &buffer, token->kbps_up );
    relay_write_uint32( &buffer, token->kbps_down );
    relay_write_address_short( &buffer, &token->next_address );
    relay_write_uint8( &buffer, token->next_internal );
    relay_write_uint8( &buffer, token->prev_internal );
    for ( int i = 0; i < RELAY_ADDRESS_BYTES - (RELAY_ADDRESS_BYTES_SHORT + 2); ++i )
    {
        uint8_t dummy = 0;
        relay_write_uint8( &buffer, dummy );
    }

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
    token->kbps_up = relay_read_uint32( &buffer );
    token->kbps_down = relay_read_uint32( &buffer );
    relay_read_address_short( &buffer, &token->next_address );
    token->next_internal = relay_read_uint8( &buffer );
    token->prev_internal = relay_read_uint8( &buffer );
    for ( int i = 0; i < RELAY_ADDRESS_BYTES - (RELAY_ADDRESS_BYTES_SHORT + 2); ++i )
    {
        uint8_t dummy = relay_read_uint8( &buffer );
        (void) dummy;
    }

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

// -------------------------------------------------------------

int relay_write_header_sdk5( int direction, uint8_t type, uint64_t sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint8_t * buffer )
{
    assert( private_key );
    assert( buffer );

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

    if ( type == RELAY_SESSION_PING_PACKET_SDK5 || type == RELAY_SESSION_PONG_PACKET_SDK5 || type == RELAY_ROUTE_RESPONSE_PACKET_SDK5 || type == RELAY_CONTINUE_RESPONSE_PACKET_SDK5 )
    {
        // second highest bit must be set
        assert( sequence & ( 1ULL << 62 ) );
    }
    else
    {
        // second highest bit must be clear
        assert( ( sequence & ( 1ULL << 62 ) ) == 0 );
    }

    relay_write_uint64( &buffer, sequence );

    uint8_t * additional = buffer;
    const int additional_length = 8 + 1;

    relay_write_uint64( &buffer, session_id );
    relay_write_uint8( &buffer, session_version );

    uint8_t nonce[12];
    {
        uint8_t * p = nonce;
        relay_write_uint32( &p, type );
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

    assert( int( buffer - start ) == RELAY_HEADER_BYTES_SDK5 );

    return RELAY_OK;
}

int relay_peek_header_sdk5( int direction, int packet_type, uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, const uint8_t * buffer, int buffer_length )
{
    uint64_t packet_sequence;

    assert( buffer );

    if ( buffer_length < RELAY_HEADER_BYTES_SDK5 )
        return RELAY_ERROR;

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

    if ( packet_type == RELAY_SESSION_PING_PACKET_SDK5 || packet_type == RELAY_SESSION_PONG_PACKET_SDK5 || packet_type == RELAY_ROUTE_RESPONSE_PACKET_SDK5 || packet_type == RELAY_CONTINUE_RESPONSE_PACKET_SDK5 )
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

    return RELAY_OK;
}

int relay_verify_header_sdk5( int direction, int packet_type, const uint8_t * private_key, uint8_t * buffer, int buffer_length )
{
    assert( private_key );
    assert( buffer );

    if ( buffer_length < RELAY_HEADER_BYTES_SDK5 )
    {
        return RELAY_ERROR;
    }

    const uint8_t * p = buffer;

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

    if ( packet_type == RELAY_SESSION_PING_PACKET_SDK5 || packet_type == RELAY_SESSION_PONG_PACKET_SDK5 || packet_type == RELAY_ROUTE_RESPONSE_PACKET_SDK5 || packet_type == RELAY_CONTINUE_RESPONSE_PACKET_SDK5 )
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

    const int additional_length = 8 + 1;

    uint64_t packet_session_id = relay_read_uint64( &p );
    uint8_t packet_session_version = relay_read_uint8( &p );

    (void) packet_session_id;
    (void) packet_session_version;

    uint8_t nonce[12];
    {
        uint8_t * q = nonce;
        relay_write_uint32( &q, packet_type );
        relay_write_uint64( &q, packet_sequence );
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt( buffer + 17, &decrypted_length, NULL,
                                                            buffer + 17, (unsigned long long) crypto_aead_chacha20poly1305_IETF_ABYTES,
                                                            additional, (unsigned long long) additional_length,
                                                            nonce, private_key );

    if ( result != 0 )
    {
        return RELAY_ERROR;
    }

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

// ---------------------------------------------------------------

typedef uint64_t relay_fnv_t;

void relay_fnv_init( relay_fnv_t * fnv )
{
    *fnv = 0xCBF29CE484222325;
}

void relay_fnv_write( relay_fnv_t * fnv, const uint8_t * data, size_t size )
{
    for ( size_t i = 0; i < size; i++ )
    {
        (*fnv) ^= data[i];
        (*fnv) *= 0x00000100000001B3;
    }
}

uint64_t relay_fnv_finalize( relay_fnv_t * fnv )
{
    return *fnv;
}

uint64_t relay_hash_string( const char * string )
{
    relay_fnv_t fnv;
    relay_fnv_init( &fnv );
    relay_fnv_write( &fnv, (uint8_t *)( string ), strlen( string ) );
    return relay_fnv_finalize( &fnv );
}

// ---------------------------------------------------------------

static void relay_generate_pittle_sdk5( uint8_t * output, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    assert( output );
    assert( from_address );
    assert( from_address_bytes > 0 );
    assert( to_address );
    assert( to_address_bytes >= 0 );
    assert( packet_length > 0 );
#if RELAY_BIG_ENDIAN
    relay_bswap( from_port );
    relay_bswap( to_port );
    relay_bswap( packet_length );
#endif // #if RELAY_BIG_ENDIAN
    uint16_t sum = 0;
    for ( int i = 0; i < from_address_bytes; ++i ) { sum += uint8_t(from_address[i]); }
    const char * from_port_data = (const char*) &from_port;
    sum += uint8_t(from_port_data[0]);
    sum += uint8_t(from_port_data[1]);
    for ( int i = 0; i < to_address_bytes; ++i ) { sum += uint8_t(to_address[i]); }
    const char * to_port_data = (const char*) &to_port;
    sum += uint8_t(to_port_data[0]);
    sum += uint8_t(to_port_data[1]);
    const char * packet_length_data = (const char*) &packet_length;
    sum += uint8_t(packet_length_data[0]);
    sum += uint8_t(packet_length_data[1]);
    sum += uint8_t(packet_length_data[2]);
    sum += uint8_t(packet_length_data[3]);
#if RELAY_BIG_ENDIAN
    relay_bswap( sum );
#endif // #if RELAY_BIG_ENDIAN
    const char * sum_data = (const char*) &sum;
    output[0] = 1 | ( uint8_t(sum_data[0]) ^ uint8_t(sum_data[1]) ^ 193 );
    output[1] = 1 | ( ( 255 - output[0] ) ^ 113 );
}

static void relay_generate_chonkle_sdk5( uint8_t * output, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    assert( output );
    assert( magic );
    assert( from_address );
    assert( from_address_bytes >= 0 );
    assert( to_address );
    assert( to_address_bytes >= 0 );
    assert( packet_length > 0 );
#if RELAY_BIG_ENDIAN
    relay_bswap( from_port );
    relay_bswap( to_port );
    relay_bswap( packet_length );
#endif // #if RELAY_BIG_ENDIAN
    relay_fnv_t fnv;
    relay_fnv_init( &fnv );
    relay_fnv_write( &fnv, magic, 8 );
    relay_fnv_write( &fnv, from_address, from_address_bytes );
    relay_fnv_write( &fnv, (const uint8_t*) &from_port, 2 );
    relay_fnv_write( &fnv, to_address, to_address_bytes );
    relay_fnv_write( &fnv, (const uint8_t*) &to_port, 2 );
    relay_fnv_write( &fnv, (const uint8_t*) &packet_length, 4 );
    uint64_t hash = relay_fnv_finalize( &fnv );
#if RELAY_BIG_ENDIAN
    relay_bswap( hash );
#endif // #if RELAY_BIG_ENDIAN
    const char * data = (const char*) &hash;
    output[0] = ( ( data[6] & 0xC0 ) >> 6 ) + 42;
    output[1] = ( data[3] & 0x1F ) + 200;
    output[2] = ( ( data[2] & 0xFC ) >> 2 ) + 5;
    output[3] = data[0];
    output[4] = ( data[2] & 0x03 ) + 78;
    output[5] = ( data[4] & 0x7F ) + 96;
    output[6] = ( ( data[1] & 0xFC ) >> 2 ) + 100;
    if ( ( data[7] & 1 ) == 0 ) { output[7] = 79; } else { output[7] = 7; }
    if ( ( data[4] & 0x80 ) == 0 ) { output[8] = 37; } else { output[8] = 83; }
    output[9] = ( data[5] & 0x07 ) + 124;
    output[10] = ( ( data[1] & 0xE0 ) >> 5 ) + 175;
    output[11] = ( data[6] & 0x3F ) + 33;
    const int value = ( data[1] & 0x03 );
    if ( value == 0 ) { output[12] = 97; } else if ( value == 1 ) { output[12] = 5; } else if ( value == 2 ) { output[12] = 43; } else { output[12] = 13; }
    output[13] = ( ( data[5] & 0xF8 ) >> 3 ) + 210;
    output[14] = ( ( data[7] & 0xFE ) >> 1 ) + 17;
}

bool relay_basic_packet_filter_sdk5( const uint8_t * data, int packet_length )
{
    if ( packet_length == 0 )
        return false;

    if ( data[0] == 0 )
        return true;

    if ( packet_length < 18 )
        return false;

    if ( data[0] < 0x01 || data[0] > 0x63 )
        return false;

    if ( data[1] < 0x2A || data[1] > 0x2D )
        return false;

    if ( data[2] < 0xC8 || data[2] > 0xE7 )
        return false;

    if ( data[3] < 0x05 || data[3] > 0x44 )
        return false;

    if ( data[5] < 0x4E || data[5] > 0x51 )
        return false;

    if ( data[6] < 0x60 || data[6] > 0xDF )
        return false;

    if ( data[7] < 0x64 || data[7] > 0xE3 )
        return false;

    if ( data[8] != 0x07 && data[8] != 0x4F )
        return false;

    if ( data[9] != 0x25 && data[9] != 0x53 )
        return false;

    if ( data[10] < 0x7C || data[10] > 0x83 )
        return false;

    if ( data[11] < 0xAF || data[11] > 0xB6 )
        return false;

    if ( data[12] < 0x21 || data[12] > 0x60 )
        return false;

    if ( data[13] != 0x61 && data[13] != 0x05 && data[13] != 0x2B && data[13] != 0x0D )
        return false;

    if ( data[14] < 0xD2 || data[14] > 0xF1 )
        return false;

    if ( data[15] < 0x11 || data[15] > 0x90 )
        return false;

    return true;
}

void relay_address_data_sdk5( const relay_address_t * address, uint8_t * address_data, int * address_bytes, uint16_t * address_port )
{
    assert( address );
    if ( address->type == RELAY_ADDRESS_IPV4 )
    {
        address_data[0] = address->data.ipv4[0];
        address_data[1] = address->data.ipv4[1];
        address_data[2] = address->data.ipv4[2];
        address_data[3] = address->data.ipv4[3];
        *address_bytes = 4;
    }
    else if ( address->type == RELAY_ADDRESS_IPV6 )
    {
        for ( int i = 0; i < 8; ++i )
        {
            address_data[i*2]   = address->data.ipv6[i] >> 8;
            address_data[i*2+1] = address->data.ipv6[i] & 0xFF;
        }
        *address_bytes = 16;
    }
    else
    {
        *address_bytes = 0;
    }
    *address_port = address->port;
}

bool relay_advanced_packet_filter_sdk5( const uint8_t * data, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    if ( data[0] == 0 )
        return true;
    if ( packet_length < 18 )
        return false;
    uint8_t a[15];
    uint8_t b[2];
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    if ( memcmp( a, data + 1, 15 ) != 0 )
        return false;
    if ( memcmp( b, data + packet_length - 2, 2 ) != 0 )
        return false;
    return true;
}

// ---------------------------------------------------------------

struct relay_route_stats_t
{
    float rtt;
    float jitter;
    float packet_loss;
};

struct relay_ping_history_entry_t
{
    uint64_t sequence;
    double time_ping_sent;
    double time_pong_received;
};

struct relay_ping_history_t
{
    uint64_t sequence;
    relay_ping_history_entry_t entries[RELAY_PING_HISTORY_ENTRY_COUNT];
};

void relay_ping_history_clear( relay_ping_history_t * history )
{
    assert( history );
    history->sequence = 0;
    for ( int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; ++i )
    {
        history->entries[i].sequence = 0xFFFFFFFFFFFFFFFFULL;
        history->entries[i].time_ping_sent = -1.0;
        history->entries[i].time_pong_received = -1.0;
    }
}

uint64_t relay_ping_history_ping_sent( relay_ping_history_t * history, double time )
{
    assert( history );
    const int index = history->sequence % RELAY_PING_HISTORY_ENTRY_COUNT;
    relay_ping_history_entry_t * entry = &history->entries[index];
    entry->sequence = history->sequence;
    entry->time_ping_sent = time;
    entry->time_pong_received = -1.0;
    history->sequence++;
    return entry->sequence;
}

void relay_ping_history_pong_received( relay_ping_history_t * history, uint64_t sequence, double time )
{
    const int index = sequence % RELAY_PING_HISTORY_ENTRY_COUNT;
    relay_ping_history_entry_t * entry = &history->entries[index];
    if ( entry->sequence == sequence )
    {
        entry->time_pong_received = time;
    }
}

void relay_route_stats_from_ping_history( const relay_ping_history_t * history, double start, double end, relay_route_stats_t * stats, double ping_safety )
{
    assert( history );
    assert( stats );
    assert( start < end );

    stats->rtt = 0.0f;
    stats->jitter = 0.0f;
    stats->packet_loss = 0.0f;

    // calculate packet loss

    int num_pings_sent = 0;
    int num_pongs_received = 0;

    for ( int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++ )
    {
        const relay_ping_history_entry_t * entry = &history->entries[i];

        if ( entry->time_ping_sent >= start && entry->time_ping_sent <= end - ping_safety )
        {
            num_pings_sent++;

            if ( entry->time_pong_received >= entry->time_ping_sent )
                num_pongs_received++;
        }
    }

    if ( num_pings_sent > 0 )
    {
        stats->packet_loss = (float) ( 100.0 * ( 1.0 - ( double( num_pongs_received ) / double( num_pings_sent ) ) ) );
    }

    // calculate mean RTT

    double mean_rtt = 0.0;
    int num_pings = 0;
    int num_pongs = 0;

    for ( int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++ )
    {
        const relay_ping_history_entry_t * entry = &history->entries[i];

        if ( entry->time_ping_sent >= start && entry->time_ping_sent <= end )
        {
            if ( entry->time_pong_received > entry->time_ping_sent )
            {
                mean_rtt += 1000.0 * ( entry->time_pong_received - entry->time_ping_sent );
                num_pongs++;
            }
            num_pings++;
        }
    }

    mean_rtt = ( num_pongs > 0 ) ? ( mean_rtt / num_pongs ) : 10000.0;

    assert( mean_rtt >= 0.0 );

    stats->rtt = float( mean_rtt );

    // calculate jitter

    int num_jitter_samples = 0;

    double stddev_rtt = 0.0;

    for ( int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++ )
    {
        const relay_ping_history_entry_t * entry = &history->entries[i];

        if ( entry->time_ping_sent >= start && entry->time_ping_sent <= end )
        {
            if ( entry->time_pong_received > entry->time_ping_sent )
            {
                // pong received
                double rtt = 1000.0 * ( entry->time_pong_received - entry->time_ping_sent );
                if ( rtt >= mean_rtt )
                {
                    double error = rtt - mean_rtt;
                    stddev_rtt += error * error;
                    num_jitter_samples++;
                }
            }
        }
    }

    if ( num_jitter_samples > 0 )
    {
        stats->jitter = 3.0f * (float) sqrt( stddev_rtt / num_jitter_samples );
    }
}

// --------------------------------------------------------------------------

int relay_write_route_request_packet_sdk5( uint8_t * packet_data, const uint8_t * token_data, int token_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_ROUTE_REQUEST_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    relay_write_bytes( &p, token_data, token_bytes );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_route_response_packet_sdk5( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_ROUTE_RESPONSE_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += RELAY_HEADER_BYTES_SDK5;
    send_sequence |= uint64_t(1) << 63;
    send_sequence |= uint64_t(1) << 62;
    if ( relay_write_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_ROUTE_RESPONSE_PACKET_SDK5, send_sequence, session_id, session_version, private_key, b ) != RELAY_OK )
        return 0;
    if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_ROUTE_RESPONSE_PACKET_SDK5, private_key, b, RELAY_HEADER_BYTES_SDK5 ) != RELAY_OK )
        return 0;
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_continue_request_packet_sdk5( uint8_t * packet_data, const uint8_t * token_data, int token_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_CONTINUE_REQUEST_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    relay_write_bytes( &p, token_data, token_bytes );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_continue_response_packet_sdk5( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_CONTINUE_RESPONSE_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += RELAY_HEADER_BYTES_SDK5;
    send_sequence |= uint64_t(1) << 63;
    send_sequence |= uint64_t(1) << 62;
    if ( relay_write_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_CONTINUE_RESPONSE_PACKET_SDK5, send_sequence, session_id, session_version, private_key, b ) != RELAY_OK )
        return 0;
    if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_CONTINUE_RESPONSE_PACKET_SDK5, private_key, b, RELAY_HEADER_BYTES_SDK5 ) != RELAY_OK )
        return 0;
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_client_to_server_packet_sdk5( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * game_packet_data, int game_packet_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    assert( packet_data );
    assert( private_key );
    assert( game_packet_data );
    assert( game_packet_bytes >= 0 );
    assert( game_packet_bytes <= RELAY_MTU );
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_CLIENT_TO_SERVER_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += RELAY_HEADER_BYTES_SDK5;
    if ( relay_write_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, RELAY_CLIENT_TO_SERVER_PACKET_SDK5, send_sequence, session_id, session_version, private_key, b ) != RELAY_OK )
        return 0;
    if ( relay_verify_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, RELAY_CLIENT_TO_SERVER_PACKET_SDK5, private_key, b, RELAY_HEADER_BYTES_SDK5 ) != RELAY_OK )
        return 0;
    relay_write_bytes( &p, game_packet_data, game_packet_bytes );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_server_to_client_packet_sdk5( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * game_packet_data, int game_packet_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    assert( packet_data );
    assert( private_key );
    assert( game_packet_data );
    assert( game_packet_bytes >= 0 );
    assert( game_packet_bytes <= RELAY_MTU );
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_SERVER_TO_CLIENT_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += RELAY_HEADER_BYTES_SDK5;
    send_sequence |= uint64_t(1) << 63;
    if ( relay_write_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_SERVER_TO_CLIENT_PACKET_SDK5, send_sequence, session_id, session_version, private_key, b ) != RELAY_OK )
        return 0;
    if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_SERVER_TO_CLIENT_PACKET_SDK5, private_key, b, RELAY_HEADER_BYTES_SDK5 ) != RELAY_OK )
        return 0;
    relay_write_bytes( &p, game_packet_data, game_packet_bytes );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_session_ping_packet_sdk5( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint64_t ping_sequence, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    assert( packet_data );
    assert( private_key );
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_SESSION_PING_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += RELAY_HEADER_BYTES_SDK5;
    send_sequence |= uint64_t(1) << 62;
    if ( relay_write_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, RELAY_SESSION_PING_PACKET_SDK5, send_sequence, session_id, session_version, private_key, b ) != RELAY_OK )
        return 0;
    if ( relay_verify_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, RELAY_SESSION_PING_PACKET_SDK5, private_key, b, RELAY_HEADER_BYTES_SDK5 ) != RELAY_OK )
        return 0;
    relay_write_uint64( &p, ping_sequence );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_session_pong_packet_sdk5( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint64_t ping_sequence, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    assert( packet_data );
    assert( private_key );
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_SESSION_PONG_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += RELAY_HEADER_BYTES_SDK5;
    send_sequence |= uint64_t(1) << 63;
    send_sequence |= uint64_t(1) << 62;
    if ( relay_write_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_SESSION_PONG_PACKET_SDK5, send_sequence, session_id, session_version, private_key, b ) != RELAY_OK )
        return 0;
    if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, RELAY_SESSION_PONG_PACKET_SDK5, private_key, b, RELAY_HEADER_BYTES_SDK5 ) != RELAY_OK )
        return 0;
    relay_write_uint64( &p, ping_sequence );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int relay_write_pong_packet_sdk5( uint8_t * packet_data, uint64_t ping_sequence, uint64_t session_id, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    relay_write_uint8( &p, RELAY_NEAR_PONG_PACKET_SDK5 );
    uint8_t * a = p; p += 15;
    relay_write_uint64( &p, ping_sequence );
    relay_write_uint64( &p, session_id );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    relay_generate_chonkle_sdk5( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle_sdk5( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

// --------------------------------------------------------------------------

#define MAX_RELAYS 1024

struct relay_stats_t
{
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    float relay_rtt[MAX_RELAYS];
    float relay_jitter[MAX_RELAYS];
    float relay_packet_loss[MAX_RELAYS];
};

struct relay_manager_t
{
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    double relay_last_ping_time[MAX_RELAYS];
    relay_address_t relay_addresses[MAX_RELAYS];
    relay_ping_history_t * relay_ping_history[MAX_RELAYS];
    relay_ping_history_t ping_history_array[MAX_RELAYS];
};

void relay_manager_reset( relay_manager_t * manager );

relay_manager_t * relay_manager_create()
{
    relay_manager_t * manager = (relay_manager_t*) malloc( sizeof(relay_manager_t) );
    if ( !manager )
        return NULL;
    relay_manager_reset( manager );
    return manager;
}

void relay_manager_reset( relay_manager_t * manager )
{
    assert( manager );
    manager->num_relays = 0;
    memset( manager->relay_ids, 0, sizeof(manager->relay_ids) );
    memset( manager->relay_last_ping_time, 0, sizeof(manager->relay_last_ping_time) );
    memset( manager->relay_addresses, 0, sizeof(manager->relay_addresses) );
    memset( manager->relay_ping_history, 0, sizeof(manager->relay_ping_history) );
    for ( int i = 0; i < MAX_RELAYS; ++i )
    {
        relay_ping_history_clear( &manager->ping_history_array[i] );
    }
}

void relay_manager_update( relay_manager_t * manager, int num_relays, const uint64_t * relay_ids, const relay_address_t * relay_addresses )
{
    assert( manager );
    assert( num_relays >= 0 );
    assert( num_relays <= MAX_RELAYS );
    assert( relay_ids );
    assert( relay_addresses );

    // first copy all current relays that are also in the updated relay relay list

    bool history_slot_taken[MAX_RELAYS];
    memset( history_slot_taken, 0, sizeof(history_slot_taken) );

    bool found[MAX_RELAYS];
    memset( found, 0, sizeof(found) );

    uint64_t new_relay_ids[MAX_RELAYS];
    double new_relay_last_ping_time[MAX_RELAYS];
    relay_address_t new_relay_addresses[MAX_RELAYS];
    relay_ping_history_t * new_relay_ping_history[MAX_RELAYS];

    int index = 0;

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        for ( int j = 0; j < num_relays; ++j )
        {
            if ( manager->relay_ids[i] == relay_ids[j] )
            {
                found[j] = true;
                new_relay_ids[index] = manager->relay_ids[i];
                new_relay_last_ping_time[index] = manager->relay_last_ping_time[i];
                new_relay_addresses[index] = manager->relay_addresses[i];
                new_relay_ping_history[index] = manager->relay_ping_history[i];
                const int slot = manager->relay_ping_history[i] - manager->ping_history_array;
                assert( slot >= 0 );
                assert( slot < MAX_RELAYS );
                history_slot_taken[slot] = true;
                index++;
                break;
            }
        }
    }

    // now copy all near relays not found in the current relay list

    for ( int i = 0; i < num_relays; ++i )
    {
        if ( !found[i] )
        {
            new_relay_ids[index] = relay_ids[i];
            new_relay_last_ping_time[index] = -10000.0;
            new_relay_addresses[index] = relay_addresses[i];
            new_relay_ping_history[index] = NULL;
            for ( int j = 0; j < MAX_RELAYS; ++j )
            {
                if ( !history_slot_taken[j] )
                {
                    new_relay_ping_history[index] = &manager->ping_history_array[j];
                    relay_ping_history_clear( new_relay_ping_history[index] );
                    history_slot_taken[j] = true;
                    break;
                }
            }
            assert( new_relay_ping_history[index] );
            index++;
        }
    }

    // commit the updated relay array

    manager->num_relays = index;
    memcpy( manager->relay_ids, new_relay_ids, 8 * index );
    memcpy( manager->relay_last_ping_time, new_relay_last_ping_time, 8 * index );
    memcpy( manager->relay_addresses, new_relay_addresses, sizeof(relay_address_t) * index );
    memcpy( manager->relay_ping_history, new_relay_ping_history, sizeof(relay_ping_history_t*) * index );

    // make sure all ping times are evenly distributed to avoid clusters of ping packets

    double current_time = relay_platform_time();

    if ( manager->num_relays > 0 )
    {
        for ( int i = 0; i < manager->num_relays; ++i )
        {
            manager->relay_last_ping_time[i] = current_time - RELAY_PING_TIME + i * RELAY_PING_TIME / manager->num_relays;
        }
    }

#ifndef NDEBUG

    // make sure everything is correct

    assert( num_relays == index );

    int num_found = 0;
    for ( int i = 0; i < num_relays; ++i )
    {
        for ( int j = 0; j < manager->num_relays; ++j )
        {
            if ( relay_ids[i] == manager->relay_ids[j] && relay_address_equal( &relay_addresses[i], &manager->relay_addresses[j] ) == 1 )
            {
                num_found++;
                break;
            }
        }
    }
    assert( num_found == num_relays );

    for ( int i = 0; i < num_relays; ++i )
    {
        for ( int j = 0; j < num_relays; ++j )
        {
            if ( i == j )
                continue;
            assert( manager->relay_ping_history[i] != manager->relay_ping_history[j] );
        }
    }

#endif // #ifndef DEBUG
}

bool relay_manager_process_pong( relay_manager_t * manager, const relay_address_t * from, uint64_t sequence )
{
    assert( manager );
    assert( from );

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        if ( relay_address_equal( from, &manager->relay_addresses[i] ) )
        {
            relay_ping_history_pong_received( manager->relay_ping_history[i], sequence, relay_platform_time() );
            return true;
        }
    }

    return false;
}

void relay_manager_get_stats( relay_manager_t * manager, relay_stats_t * stats )
{
    assert( manager );
    assert( stats );

    double current_time = relay_platform_time();

    stats->num_relays = manager->num_relays;

    for ( int i = 0; i < stats->num_relays; ++i )
    {
        relay_route_stats_t route_stats;
        relay_route_stats_from_ping_history( manager->relay_ping_history[i], current_time - RELAY_STATS_WINDOW, current_time, &route_stats, RELAY_PING_SAFETY );
        stats->relay_ids[i] = manager->relay_ids[i];
        stats->relay_rtt[i] = route_stats.rtt;
        stats->relay_jitter[i] = route_stats.jitter;
        stats->relay_packet_loss[i] = route_stats.packet_loss;
    }
}

void relay_manager_destroy( relay_manager_t * manager )
{
    free( manager );
}

// -----------------------------------------------------------------------------

#define RELAY_TOKEN_BYTES 32
#define RESPONSE_MAX_BYTES 1024 * 1024

#define RELAY_PING_PACKET 75
#define RELAY_PONG_PACKET 76

struct relay_session_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint64_t client_to_server_sequence;
    uint64_t server_to_client_sequence;
    int kbps_up;
    int kbps_down;
    relay_address_t prev_address;
    relay_address_t next_address;
    uint8_t prev_internal;
    uint8_t next_internal;
    uint8_t private_key[crypto_box_SECRETKEYBYTES];
    relay_replay_protection_t replay_protection_server_to_client;
    relay_replay_protection_t replay_protection_client_to_server;
};

struct relay_t
{
    relay_manager_t * relay_manager;
    relay_address_t relay_public_address;
    relay_address_t relay_internal_address;
    bool has_internal_address;
    relay_platform_socket_t * socket;
    relay_platform_mutex_t * mutex;
    double initialize_time;
    uint64_t initialize_router_timestamp;
    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    uint8_t router_public_key[RELAY_PUBLIC_KEY_BYTES];
    std::map<uint64_t, relay_session_t*> * sessions;
    bool relays_dirty;
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    relay_address_t relay_addresses[MAX_RELAYS];
    uint8_t relay_internal[MAX_RELAYS];
    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];
    std::atomic<uint64_t> envelope_bandwidth_kbps_up;               // todo: these should probably be in bytes up/down instead
    std::atomic<uint64_t> envelope_bandwidth_kbps_down;
    std::atomic<uint64_t> counters[NUM_RELAY_COUNTERS];
#if RELAY_DEVELOPMENT
    float fake_packet_loss_percent;
    float fake_packet_loss_start_time;
#endif // #if RELAY_DEVELOPMENT
};

struct curl_buffer_t
{
    int size;
    int max_size;
    uint8_t * data;
};

size_t curl_buffer_write_function( char * ptr, size_t size, size_t nmemb, void * userdata )
{
    curl_buffer_t * buffer = (curl_buffer_t*) userdata;
    assert( buffer );
    assert( size == 1 );
    if ( int( buffer->size + size*nmemb ) > buffer->max_size )
        return 0;
    memcpy( buffer->data + buffer->size, ptr, size*nmemb );
    buffer->size += size * nmemb;
    return size * nmemb;
}

int relay_init( CURL * curl, const char * hostname, uint8_t * relay_token, const char * relay_address, const uint8_t * router_public_key, const uint8_t * relay_private_key, uint64_t * router_timestamp )
{
    const uint32_t init_request_magic = 0x9083708f;

    uint32_t init_request_version = 0;

    uint8_t init_data[1024];
    memset( init_data, 0, sizeof(init_data) );

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes( nonce, crypto_box_NONCEBYTES );

    uint8_t * p = init_data;

    relay_write_uint32( &p, init_request_magic );
    relay_write_uint32( &p, init_request_version );
    relay_write_bytes( &p, nonce, crypto_box_NONCEBYTES );
    relay_write_string( &p, relay_address, RELAY_MAX_ADDRESS_STRING_LENGTH );

    uint8_t * q = p;

    relay_write_bytes( &p, relay_token, RELAY_TOKEN_BYTES );

    int encrypt_length = int( p - q );

    if ( crypto_box_easy( q, q, encrypt_length, nonce, router_public_key, relay_private_key ) != 0 )
    {
        return RELAY_ERROR;
    }

    int init_length = (int) ( p - init_data ) + encrypt_length + crypto_box_MACBYTES;

    struct curl_slist * slist = curl_slist_append( NULL, "Content-Type:application/octet-stream" );

    curl_buffer_t init_response_buffer;
    init_response_buffer.size = 0;
    init_response_buffer.max_size = 1024;
    init_response_buffer.data = (uint8_t*) alloca( init_response_buffer.max_size );

    char init_url[1024];
    snprintf( init_url, sizeof(init_url), "%s/relay_init", hostname );

    curl_easy_setopt( curl, CURLOPT_BUFFERSIZE, 102400L );
    curl_easy_setopt( curl, CURLOPT_URL, init_url );
    curl_easy_setopt( curl, CURLOPT_NOPROGRESS, 1L );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDS, init_data );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)init_length );
    curl_easy_setopt( curl, CURLOPT_HTTPHEADER, slist );
    curl_easy_setopt( curl, CURLOPT_USERAGENT, "network next relay" );
    curl_easy_setopt( curl, CURLOPT_MAXREDIRS, 50L );
    curl_easy_setopt( curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS );
    curl_easy_setopt( curl, CURLOPT_TCP_KEEPALIVE, 1L );
    curl_easy_setopt( curl, CURLOPT_TIMEOUT_MS, long( 1000 ) );
    curl_easy_setopt( curl, CURLOPT_WRITEDATA, &init_response_buffer );
    curl_easy_setopt( curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function );

    CURLcode ret = curl_easy_perform( curl );

    curl_slist_free_all( slist );
    slist = NULL;

    if ( ret != 0 )
    {
        return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo( curl, CURLINFO_RESPONSE_CODE, &code );
    if ( code != 200 )
    {
        return RELAY_ERROR;
    }

    if ( init_response_buffer.size < 4 )
    {
        printf( "error: bad relay init response size. too small to have valid data (%d)\n", init_response_buffer.size );
        return RELAY_ERROR;
    }

    const uint8_t * r = init_response_buffer.data;

    uint32_t version = relay_read_uint32( &r );

    const uint32_t init_response_version = 0;

    if ( version != init_response_version )
    {
        printf( "error: bad relay init response version. expected %d, got %d\n", init_response_version, version );
        return RELAY_ERROR;
    }

    if ( init_response_buffer.size != 4 + 8 + RELAY_TOKEN_BYTES )
    {
        printf( "error: bad relay init response size. expected %d bytes, got %d\n", RELAY_TOKEN_BYTES, init_response_buffer.size );
        return RELAY_ERROR;
    }

    *router_timestamp = relay_read_uint64( &r );

    memcpy( relay_token, init_response_buffer.data + 4 + 8, RELAY_TOKEN_BYTES );

    return RELAY_OK;
}

int relay_update( CURL * curl, const char * hostname, const uint8_t * relay_token, const char * relay_address, uint8_t * update_response_memory, relay_t * relay, bool shutdown )
{
    // build update data

    uint32_t update_version = 5;

    uint8_t update_data[10*1024];

    uint8_t * p = update_data;
    relay_write_uint32( &p, update_version );
    relay_write_string( &p, relay_address, 256 );
    relay_write_bytes( &p, relay_token, RELAY_TOKEN_BYTES );

    relay_platform_mutex_acquire( relay->mutex );
    relay_stats_t stats;
    relay_manager_get_stats( relay->relay_manager, &stats );
    relay_platform_mutex_release( relay->mutex );

    relay_write_uint32( &p, stats.num_relays );
    for ( int i = 0; i < stats.num_relays; ++i )
    {
        relay_write_uint64( &p, stats.relay_ids[i] );
        relay_write_float32( &p, stats.relay_rtt[i] );
        relay_write_float32( &p, stats.relay_jitter[i] );
        relay_write_float32( &p, stats.relay_packet_loss[i] );
    }

    relay_platform_mutex_acquire( relay->mutex );
    uint64_t sessions = relay->sessions->size();
    relay_platform_mutex_release( relay->mutex );
    relay_write_uint64(&p, sessions);

    relay_write_uint8(&p, uint8_t(shutdown));
    relay_write_string(&p, RELAY_VERSION, 32);

    uint8_t cpu = 0;
    relay_write_uint8(&p, cpu);

    relay_write_uint64(&p, relay->envelope_bandwidth_kbps_up);
    relay_write_uint64(&p, relay->envelope_bandwidth_kbps_down);

    // todo: redo bandwidth
    uint64_t bandwidth_tx = 0;
    uint64_t bandwidth_rx = 0;
    relay_write_uint64(&p, bandwidth_tx);
    relay_write_uint64(&p, bandwidth_rx);

    relay_write_uint32( &p, NUM_RELAY_COUNTERS );
    for ( int i = 0; i < NUM_RELAY_COUNTERS; ++i )
    {
        relay_write_uint64(&p, relay->counters[i]);
    }

    int update_data_length = (int) ( p - update_data );

    // post it to backend

    struct curl_slist * slist = curl_slist_append( NULL, "Content-Type:application/octet-stream" );

    curl_buffer_t update_response_buffer;
    update_response_buffer.size = 0;
    update_response_buffer.max_size = RESPONSE_MAX_BYTES;
    update_response_buffer.data = (uint8_t*) update_response_memory;

    char update_url[1024];
    snprintf( update_url, sizeof(update_url), "%s/relay_update", hostname );

    curl_easy_setopt( curl, CURLOPT_BUFFERSIZE, 102400L );
    curl_easy_setopt( curl, CURLOPT_URL, update_url );
    curl_easy_setopt( curl, CURLOPT_NOPROGRESS, 1L );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDS, update_data );
    curl_easy_setopt( curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)update_data_length );
    curl_easy_setopt( curl, CURLOPT_HTTPHEADER, slist );
    curl_easy_setopt( curl, CURLOPT_USERAGENT, "network next relay" );
    curl_easy_setopt( curl, CURLOPT_MAXREDIRS, 50L );
    curl_easy_setopt( curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS );
    curl_easy_setopt( curl, CURLOPT_TCP_KEEPALIVE, 1L );
    curl_easy_setopt( curl, CURLOPT_TIMEOUT_MS, long( 1000 ) );
    curl_easy_setopt( curl, CURLOPT_WRITEDATA, &update_response_buffer );
    curl_easy_setopt( curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function );

    CURLcode ret = curl_easy_perform( curl );

    curl_slist_free_all( slist );
    slist = NULL;

    if ( ret != 0 )
    {
        printf( "error: could not post relay update\n" );
        return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo( curl, CURLINFO_RESPONSE_CODE, &code );
    if ( code != 200 )
    {
        printf( "error: relay update response was %d, expected 200\n", int(code) );
        return RELAY_ERROR;
    }

    // parse update response

    const uint8_t * q = update_response_buffer.data;

    uint32_t version = relay_read_uint32( &q );

    const uint32_t update_response_version = 1;

    if ( version > update_response_version )
    {
        printf( "error: bad relay update response version. expected %d, got %d\n", update_response_version, version );
        return RELAY_ERROR;
    }

    uint64_t timestamp = relay_read_uint64( &q );
    if ( relay->initialize_router_timestamp == 0 )
    {
        printf( "Relay initialized\n" );
        fflush( stdout );
        relay->initialize_router_timestamp = timestamp;
    }

    uint32_t num_relays = relay_read_uint32( &q );

    if ( num_relays > MAX_RELAYS )
    {
        printf( "error: too many relays to ping. max is %d, got %d\n", MAX_RELAYS, num_relays );
        return RELAY_ERROR;
    }

    bool error = false;

    struct relay_ping_data_t
    {
        uint64_t id;
        relay_address_t address;
        uint8_t internal;
    };

    relay_ping_data_t relay_ping_data[MAX_RELAYS];
    memset( relay_ping_data, 0, sizeof(relay_ping_data) );

    for ( uint32_t i = 0; i < num_relays; ++i )
    {
        char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
        relay_ping_data[i].id = relay_read_uint64( &q );
        relay_read_string( &q, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH );
        if ( relay_address_parse( &relay_ping_data[i].address, address_string ) != RELAY_OK )
        {
            error = true;
            break;
        }
        relay_ping_data[i].internal = relay_read_uint8( &q );
    }

    if ( error )
    {
        printf( "error: error while reading set of relays to ping in update response\n" );
        return RELAY_ERROR;
    }

    char target_version[RELAY_MAX_VERSION_STRING_LENGTH];
    relay_read_string( &q, target_version, RELAY_MAX_VERSION_STRING_LENGTH);

    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    if ( version >= 1 )
    {
        relay_read_bytes( &q, upcoming_magic, 8 );
        relay_read_bytes( &q, current_magic, 8 );
        relay_read_bytes( &q, previous_magic, 8 );
    }

    relay_platform_mutex_acquire( relay->mutex );
    relay->num_relays = num_relays;
    for ( int i = 0; i < int(num_relays); ++i )
    {
        relay->relay_ids[i] = relay_ping_data[i].id;
        relay->relay_addresses[i] = relay_ping_data[i].address;
        relay->relay_internal[i] = relay_ping_data[i].internal;
    }
    relay->relays_dirty = true;
    memcpy( relay->upcoming_magic, &upcoming_magic, 8 );
    memcpy( relay->current_magic, &current_magic, 8 );
    memcpy( relay->previous_magic, &previous_magic, 8 );
    relay_platform_mutex_release( relay->mutex );

    return RELAY_OK;
}

static volatile uint64_t quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

static volatile bool relay_clean_shutdown = false;

void clean_shutdown_handler( int signal )
{
    (void) signal;
    relay_clean_shutdown = true;
    quit = 1;
}

uint64_t relay_timestamp( relay_t * relay )
{
    assert( relay );
    double current_time = relay_platform_time();
    uint64_t seconds_since_initialize = uint64_t( current_time - relay->initialize_time );
    return relay->initialize_router_timestamp + seconds_since_initialize;
}

uint64_t relay_clean_sequence( uint64_t sequence )
{
    uint64_t mask = ~( (1ULL<<63) | (1ULL<<62) );
    return sequence & mask;
}

static relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC receive_thread_function( void * context )
{
    relay_t * relay = (relay_t*) context;

    uint8_t packet_data[RELAY_MAX_PACKET_BYTES];

    while ( !quit )
    {
        relay_address_t from;
        int packet_bytes = relay_platform_socket_receive_packet( relay->socket, &from, packet_data, sizeof(packet_data) );
        if ( packet_bytes == 0 )
            continue;

        if ( relay->initialize_router_timestamp == 0 )
            continue;

        relay->counters[RELAY_COUNTER_PACKETS_RECEIVED]++;
        relay->counters[RELAY_COUNTER_BYTES_RECEIVED] += packet_bytes;

#if RELAY_DEVELOPMENT
        if ( relay->fake_packet_loss_start_time >= 0.0f )
        {
            const double current_time = relay_platform_time();
            if ( current_time >= relay->fake_packet_loss_start_time && ( ( rand() % 100 ) < relay->fake_packet_loss_percent ) )
            {
                continue;
            }
        }
#endif // #if RELAY_DEVELOPMENT

        int packet_id = packet_data[0];

        if ( packet_id == RELAY_PING_PACKET && packet_bytes == 1 + 8 )
        {
#if INTENSIVE_RELAY_DEBUGGING
            printf("relay ping packet\n");
#endif // #if INTENSIVE_RELAY_DEBUGGING

            relay->counters[RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED]++;

            packet_data[0] = RELAY_PONG_PACKET;
            int packet_bytes = 1 + 8;
            relay_platform_socket_send_packet( relay->socket, &from, packet_data, packet_bytes );
            relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
            relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
        }
        else if ( packet_id == RELAY_PONG_PACKET && packet_bytes == 1 + 8 )
        {
#if INTENSIVE_RELAY_DEBUGGING
            printf("relay pong packet\n");
#endif // #if INTENSIVE_RELAY_DEBUGGING

            relay->counters[RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED]++;

            relay_platform_mutex_acquire( relay->mutex );
            const uint8_t * p = packet_data + 1;
            uint64_t sequence = relay_read_uint64( &p );
            relay_manager_process_pong( relay->relay_manager, &from, sequence );
            relay_platform_mutex_release( relay->mutex );
        }

// ==================================================================================================================================================================================

        else if ( packet_id >= RELAY_ROUTE_REQUEST_PACKET_SDK5 && packet_id <= RELAY_NEAR_PONG_PACKET_SDK5 )
        {
#if INTENSIVE_RELAY_DEBUGGING
        	char from_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
        	relay_address_to_string( &from, from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING

            if ( !relay_basic_packet_filter_sdk5( packet_data, packet_bytes ) )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] basic packet filter dropped packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	
	            relay->counters[RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET]++;

                continue;
            }

            uint8_t from_address_data[32];
            uint8_t relay_public_address_data[32];
            uint8_t relay_internal_address_data[32];
            uint16_t from_address_port;
            uint16_t relay_public_address_port;
            uint16_t relay_internal_address_port;
            int from_address_bytes;
            int relay_public_address_bytes;
            int relay_internal_address_bytes;

            relay_address_data_sdk5( &from, from_address_data, &from_address_bytes, &from_address_port );
            relay_address_data_sdk5( &relay->relay_public_address, relay_public_address_data, &relay_public_address_bytes, &relay_public_address_port );
            relay_address_data_sdk5( &relay->relay_internal_address, relay_internal_address_data, &relay_internal_address_bytes, &relay_internal_address_port );

            uint8_t upcoming_magic[8];
            uint8_t current_magic[8];
            uint8_t previous_magic[8];

            relay_platform_mutex_acquire( relay->mutex );
            memcpy( &upcoming_magic, relay->upcoming_magic, 8 );
            memcpy( &current_magic, relay->current_magic, 8 );
            memcpy( &previous_magic, relay->previous_magic, 8 );
            relay_platform_mutex_release( relay->mutex );

            if ( ! ( relay_advanced_packet_filter_sdk5( packet_data, current_magic, from_address_data, from_address_bytes, from_address_port, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, packet_bytes ) ||
                     relay_advanced_packet_filter_sdk5( packet_data, previous_magic, from_address_data, from_address_bytes, from_address_port, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, packet_bytes ) ||
                     relay_advanced_packet_filter_sdk5( packet_data, upcoming_magic, from_address_data, from_address_bytes, from_address_port, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, packet_bytes ) ||
                     ( relay->has_internal_address && 
                       ( relay_advanced_packet_filter_sdk5( packet_data, current_magic, from_address_data, from_address_bytes, from_address_port, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, packet_bytes ) ||
                         relay_advanced_packet_filter_sdk5( packet_data, previous_magic, from_address_data, from_address_bytes, from_address_port, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, packet_bytes ) ||
                         relay_advanced_packet_filter_sdk5( packet_data, upcoming_magic, from_address_data, from_address_bytes, from_address_port, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, packet_bytes ) 
                       ) 
                     ) 
                   ) 
               )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] advanced packet filter dropped packet %d [sdk5]\n", from_string, packet_id );
#endif // #if INTENSIVE_RELAY_DEBUGGING

	            relay->counters[RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET]++;

                continue;
            }

            uint8_t * p = packet_data;
            p += 16;

            packet_bytes -= 18;

            if ( packet_id == RELAY_ROUTE_REQUEST_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
            	printf( "[%s] received route request packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            relay->counters[RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED]++;

                if ( packet_bytes < int( RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES * 2 ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignoring route request. bad packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_ROUTE_REQUEST_PACKET_BAD_SIZE]++;
                    continue;
                }

                relay_route_token_t token;
                if ( relay_read_encrypted_route_token( &p, &token, relay->router_public_key, relay->relay_private_key ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignoring route request. could not read route token [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_READ_TOKEN]++;
                    continue;
                }

                uint64_t current_timestamp = relay_timestamp( relay );
                if ( token.expire_timestamp < current_timestamp )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignoring route request. route token expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED]++;
                    continue;
                }

                uint64_t hash = token.session_id ^ token.session_version;

                relay_platform_mutex_acquire( relay->mutex );
                if ( relay->sessions->find(hash) == relay->sessions->end() )
                {
                    relay_session_t * session = (relay_session_t*) malloc( sizeof( relay_session_t ) );
                    assert( session );
                    session->expire_timestamp = token.expire_timestamp;
                    session->session_id = token.session_id;
                    session->session_version = token.session_version;
                    session->client_to_server_sequence = 0;
                    session->server_to_client_sequence = 0;
                    session->kbps_up = token.kbps_up;
                    session->kbps_down = token.kbps_down;
                    session->prev_address = from;
                    session->next_address = token.next_address;
                    session->prev_internal = token.prev_internal;
                    session->next_internal = token.next_internal;
                    memcpy( session->private_key, token.private_key, crypto_box_SECRETKEYBYTES );
                    relay_replay_protection_reset( &session->replay_protection_client_to_server );
                    relay_replay_protection_reset( &session->replay_protection_server_to_client );
                    relay->sessions->insert( std::make_pair(hash, session) );
                    relay->envelope_bandwidth_kbps_up += session->kbps_up;
                    relay->envelope_bandwidth_kbps_down += session->kbps_down;
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "session created: %" PRIx64 ".%d [sdk5]\n", token.session_id, token.session_version );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_CREATED]++;
                }
                relay_platform_mutex_release( relay->mutex );

                const uint8_t * token_data = p;
                int token_bytes = packet_bytes - RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES;

                uint8_t next_address_data[32];
                uint16_t next_address_port;
                int next_address_bytes;

                relay_address_data_sdk5( &token.next_address, next_address_data, &next_address_bytes, &next_address_port );

                if ( !token.next_internal )
                {
                    uint8_t route_request_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_route_request_packet_sdk5( route_request_packet, token_data, token_bytes, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( route_request_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( route_request_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &token.next_address, next_hop_address );
                        printf( "[%s] forwarding route request packet to next hop %s (public address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    
	                    relay->counters[RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &token.next_address, route_request_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t route_request_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_route_request_packet_sdk5( route_request_packet, token_data, token_bytes, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay->has_internal_address );
                        assert( relay_basic_packet_filter_sdk5( route_request_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( route_request_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &token.next_address, next_hop_address );
                        printf( "[%s] forwarding route request packet to next hop %s (internal address)\n", from_string, next_hop_address );
#endif // #if #if INTENSIVE_RELAY_DEBUGGING
                    
	                    relay->counters[RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &token.next_address, route_request_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_ROUTE_RESPONSE_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received route response packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED]++;

                if ( packet_bytes != RELAY_HEADER_BYTES_SDK5 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored route response packet. wrong packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_BAD_SIZE]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if ( relay_peek_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, &sequence, &session_id, &session_version, const_p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored route response packet. could not peek header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER]++;
                    continue;
                }

                // todo: this hash trick here has to go
                uint64_t hash = session_id ^ session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );

                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored route response packet. could not find session [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION]++;
                    continue;
                }

                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored route response packet. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }

                uint64_t clean_sequence = relay_clean_sequence( sequence );

                if ( clean_sequence <= session->server_to_client_sequence )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored route response packet. packet already received (%d <= %d) [sdk5]\n", from_string, clean_sequence, session->server_to_client_sequence );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED]++;
                    continue;
                }

                if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, session->private_key, p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored route response packet. header did not verify [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY]++;
                    continue;
                }

                session->server_to_client_sequence = clean_sequence;

                uint8_t prev_address_data[32];
                uint16_t prev_address_port;
                int prev_address_bytes;

                relay_address_data_sdk5( &session->prev_address, prev_address_data, &prev_address_bytes, &prev_address_port );

                if ( !session->prev_internal )
                {
                    uint8_t route_response_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_route_response_packet_sdk5( route_response_packet, sequence, session_id, session_version, session->private_key, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( route_response_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( route_response_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding route response packet to previous hop %s\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING

			            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, route_response_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t route_response_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_route_response_packet_sdk5( route_response_packet, sequence, session_id, session_version, session->private_key, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( route_response_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( route_response_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding route response packet to previous hop %s\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING

			            relay->counters[RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, route_response_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_CONTINUE_REQUEST_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received route continue request packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED]++;

                if ( packet_bytes < int( RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES * 2 ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignoring continue request. bad packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_BAD_SIZE]++;
                    continue;
                }

                relay_continue_token_t token;
                if ( relay_read_encrypted_continue_token( &p, &token, relay->router_public_key, relay->relay_private_key ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignoring continue request. could not read continue token [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_READ_TOKEN]++;
                    continue;
                }

                if ( token.expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "ignored continue request. token expired [sdk5]\n" );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED]++;
                    continue;
                }

                uint64_t hash = token.session_id ^ token.session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );

                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue request. could not find session [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    continue;
                }

                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue request. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }

                if ( session->expire_timestamp != token.expire_timestamp )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "session continued: %" PRIx64 ".%d [sdk5]\n", token.session_id, token.session_version );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            relay->counters[RELAY_COUNTER_SESSION_CONTINUED]++;
                }

                session->expire_timestamp = token.expire_timestamp;

                const uint8_t * token_data = p;
                int token_bytes = packet_bytes - RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES;

                uint8_t next_address_data[32];
                uint16_t next_address_port;
                int next_address_bytes;

                relay_address_data_sdk5( &session->next_address, next_address_data, &next_address_bytes, &next_address_port );

                if ( !session->next_internal )
                {
                    uint8_t continue_request_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_continue_request_packet_sdk5( continue_request_packet, token_data, token_bytes, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( continue_request_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( continue_request_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->next_address, next_hop_address );
                        printf( "[%s] forwarding continue request packet to next hop %s (public address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
			            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->next_address, continue_request_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t continue_request_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_continue_request_packet_sdk5( continue_request_packet, token_data, token_bytes, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( continue_request_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( continue_request_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->next_address, next_hop_address );
                        printf( "[%s] forwarding continue request packet to next hop %s (internal address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
			            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->next_address, continue_request_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                }
                }
            }
            else if ( packet_id == RELAY_CONTINUE_RESPONSE_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received route continue response packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED]++;

                if ( packet_bytes != RELAY_HEADER_BYTES_SDK5 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue response packet. wrong packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_BAD_SIZE]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if ( relay_peek_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, &sequence, &session_id, &session_version, const_p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue response packet. could not peek header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER]++;
                    continue;
                }

                uint64_t hash = session_id ^ session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );

                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue response packet. could not find session [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION]++;
                    continue;
                }

                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue response packet. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }

                uint64_t clean_sequence = relay_clean_sequence( sequence );

                if ( clean_sequence <= session->server_to_client_sequence )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue response packet. packet already received (%d <= %d) [sdk5]\n", from_string, clean_sequence, session->server_to_client_sequence );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED]++;
                    continue;
                }

                if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, session->private_key, p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored continue response packet. header did not verify [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
	            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY]++;
                    continue;
                }

                session->server_to_client_sequence = clean_sequence;

                uint8_t prev_address_data[32];
                uint16_t prev_address_port;
                int prev_address_bytes;

                relay_address_data_sdk5( &session->prev_address, prev_address_data, &prev_address_bytes, &prev_address_port );

                if ( !session->prev_internal )
                {
                    uint8_t continue_response_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_continue_response_packet_sdk5( continue_response_packet, sequence, session_id, session_version, session->private_key, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( continue_response_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( continue_response_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding continue response packet to previous hop %s (public address)\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
		            	relay->counters[RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS]++;
                 
                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, continue_response_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t continue_response_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_continue_response_packet_sdk5( continue_response_packet, sequence, session_id, session_version, session->private_key, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( continue_response_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( continue_response_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding continue response packet to previous hop %s (internal address)\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
			            relay->counters[RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS]++;
                 
                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, continue_response_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_CLIENT_TO_SERVER_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received client to server packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
			    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED]++;

                if ( packet_bytes <= RELAY_HEADER_BYTES_SDK5 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. packet too small (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL]++;
                    continue;
                }

                if ( packet_bytes > RELAY_HEADER_BYTES_SDK5 + RELAY_MTU )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. packet too big (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if ( relay_peek_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, packet_id, &sequence, &session_id, &session_version, const_p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. could not peek header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_PEEK_HEADER]++;
                    continue;
                }

                uint64_t hash = session_id ^ session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );
                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. could not find session [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION]++;
                    continue;
                }

                /*
                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }
                */

                uint64_t clean_sequence = relay_clean_sequence( sequence );

                if ( relay_replay_protection_already_received( &session->replay_protection_client_to_server, clean_sequence ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. already received [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED]++;
                    continue;
                }

                if ( relay_verify_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, packet_id, session->private_key, p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored client to server packet. could not verify header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_VERIFY_HEADER]++;
                    continue;
                }

                relay_replay_protection_advance_sequence( &session->replay_protection_client_to_server, clean_sequence );

                const_p += RELAY_HEADER_BYTES_SDK5;
                int game_packet_bytes = packet_bytes - RELAY_HEADER_BYTES_SDK5;
                uint8_t game_packet_data[game_packet_bytes];
                relay_read_bytes( &const_p, game_packet_data, game_packet_bytes );

                uint8_t next_address_data[32];
                uint16_t next_address_port;
                int next_address_bytes;

                relay_address_data_sdk5( &session->next_address, next_address_data, &next_address_bytes, &next_address_port );

                if ( !session->next_internal )
                {
                    uint8_t client_to_server_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_client_to_server_packet_sdk5( client_to_server_packet, sequence, session_id, session_version, session->private_key, game_packet_data, game_packet_bytes, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( client_to_server_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( client_to_server_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->next_address, next_hop_address );
                        printf( "[%s] forwarding client to server packet to next hop %s (public address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
					    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS]++;
                    
                        relay_platform_socket_send_packet( relay->socket, &session->next_address, client_to_server_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t client_to_server_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_client_to_server_packet_sdk5( client_to_server_packet, sequence, session_id, session_version, session->private_key, game_packet_data, game_packet_bytes, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( client_to_server_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( client_to_server_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->next_address, next_hop_address );
                        printf( "[%s] forwarding client to server packet to next hop %s (internal address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
					    relay->counters[RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS]++;
                    
                        relay_platform_socket_send_packet( relay->socket, &session->next_address, client_to_server_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_SERVER_TO_CLIENT_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received server to client packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
			    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED]++;

                if ( packet_bytes <= RELAY_HEADER_BYTES_SDK5 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. packet too small (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL]++;
                    continue;
                }

                if ( packet_bytes > RELAY_HEADER_BYTES_SDK5 + RELAY_MTU )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. packet too big (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if ( relay_peek_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, &sequence, &session_id, &session_version, const_p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. could not peek header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_PEEK_HEADER]++;
                    continue;
                }

                uint64_t hash = session_id ^ session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );
                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. could not find session [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION]++;
                    continue;
                }

                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }

                uint64_t clean_sequence = relay_clean_sequence( sequence );

                if ( relay_replay_protection_already_received( &session->replay_protection_server_to_client, clean_sequence ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. already received [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED]++;
                    continue;
                }

                if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, session->private_key, p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored server to client packet. could not verify header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
				    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_VERIFY_HEADER]++;
                    continue;
                }

                relay_replay_protection_advance_sequence( &session->replay_protection_server_to_client, clean_sequence );

                const_p += RELAY_HEADER_BYTES_SDK5;
                int game_packet_bytes = packet_bytes - RELAY_HEADER_BYTES_SDK5;
                uint8_t game_packet_data[game_packet_bytes];
                relay_read_bytes( &const_p, game_packet_data, game_packet_bytes );

                uint8_t prev_address_data[32];
                uint16_t prev_address_port;
                int prev_address_bytes;

                relay_address_data_sdk5( &session->prev_address, prev_address_data, &prev_address_bytes, &prev_address_port );

                if ( !session->prev_internal )
                {
                    uint8_t server_to_client_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_server_to_client_packet_sdk5( server_to_client_packet, sequence, session_id, session_version, session->private_key, game_packet_data, game_packet_bytes, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port );

                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( server_to_client_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( server_to_client_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding server to client packet to previous hop %s (public address)\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
					    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, server_to_client_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t server_to_client_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_server_to_client_packet_sdk5( server_to_client_packet, sequence, session_id, session_version, session->private_key, game_packet_data, game_packet_bytes, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port );

                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( server_to_client_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( server_to_client_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding server to client packet to previous hop %s (internal address)\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
					    relay->counters[RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, server_to_client_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_SESSION_PING_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received session ping packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED]++;

                if ( packet_bytes != RELAY_HEADER_BYTES_SDK5 + 8 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session ping packet. bad packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_BAD_PACKET_SIZE]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if ( relay_peek_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, packet_id, &sequence, &session_id, &session_version, const_p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session ping packet. could not peek header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_PEEK_HEADER]++;
                    continue;
                }

                uint64_t hash = session_id ^ session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );
                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session ping packet. session does not exist [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_SESSION_DOES_NOT_EXIST]++;
                    continue;
                }

                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session ping packet. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }

                uint64_t clean_sequence = relay_clean_sequence( sequence );

                if ( clean_sequence <= session->client_to_server_sequence )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session ping packet. already received (%d <= %d) [sdk5]\n", from_string, int(clean_sequence), int(session->client_to_server_sequence) );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED]++;
                    continue;
                }

                if ( relay_verify_header_sdk5( RELAY_DIRECTION_CLIENT_TO_SERVER, packet_id, session->private_key, p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session ping packet. could not verify header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_VERIFY_HEADER]++;
                    continue;
                }

                session->client_to_server_sequence = clean_sequence;

                const_p += RELAY_HEADER_BYTES_SDK5;
                uint64_t ping_sequence = relay_read_uint64( &const_p );

                uint8_t next_address_data[32];
                uint16_t next_address_port;
                int next_address_bytes;

                relay_address_data_sdk5( &session->next_address, next_address_data, &next_address_bytes, &next_address_port );

                if ( !session->next_internal )
                {
                    uint8_t session_ping_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_session_ping_packet_sdk5( session_ping_packet, sequence, session_id, session_version, session->private_key, ping_sequence, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port );

                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( session_ping_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( session_ping_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->next_address, next_hop_address );
                        printf( "[%s] forwarding session ping packet to next hop %s (public address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                        relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->next_address, session_ping_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t session_ping_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_session_ping_packet_sdk5( session_ping_packet, sequence, session_id, session_version, session->private_key, ping_sequence, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port );

                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( session_ping_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( session_ping_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, next_address_data, next_address_bytes, next_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                        char next_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->next_address, next_hop_address );
                        printf( "[%s] forwarding session ping packet to next hop %s (internal address)\n", from_string, next_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                        relay->counters[RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->next_address, session_ping_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_SESSION_PONG_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "received session pong packet [sdk5]\n" );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED]++;

                if ( packet_bytes != RELAY_HEADER_BYTES_SDK5 + 8 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session pong packet. bad packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_BAD_SIZE]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;

                if ( relay_peek_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, &sequence, &session_id, &session_version, const_p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session pong packet. could not peek header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_PEEK_HEADER]++;
                    continue;
                }

                uint64_t hash = session_id ^ session_version;

                relay_platform_mutex_acquire( relay->mutex );
                relay_session_t * session = (*(relay->sessions))[hash];
                relay_platform_mutex_release( relay->mutex );
                if ( !session )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session pong packet. session does not exist [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_DOES_NOT_EXIST]++;
                    continue;
                }

                if ( session->expire_timestamp < relay_timestamp( relay ) )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session pong packet. session expired [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_EXPIRED]++;
                    relay_platform_mutex_acquire( relay->mutex );
                    relay->sessions->erase(hash);
                    relay_platform_mutex_release( relay->mutex );
                    continue;
                }

                uint64_t clean_sequence = relay_clean_sequence( sequence );

                if ( clean_sequence <= session->server_to_client_sequence )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session pong packet. already received (%d <= %d) [sdk5]\n", from_string, int(clean_sequence), int(session->server_to_client_sequence) );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED]++;
                    continue;
                }

                if ( relay_verify_header_sdk5( RELAY_DIRECTION_SERVER_TO_CLIENT, packet_id, session->private_key, p, packet_bytes ) != RELAY_OK )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored session pong packet. could not verify header [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_VERIFY_HEADER]++;
                    continue;
                }

                session->server_to_client_sequence = clean_sequence;

                const_p += RELAY_HEADER_BYTES_SDK5;
                uint64_t ping_sequence = relay_read_uint64( &const_p );

                uint8_t prev_address_data[32];
                uint16_t prev_address_port;
                int prev_address_bytes;

                relay_address_data_sdk5( &session->prev_address, prev_address_data, &prev_address_bytes, &prev_address_port );

                if ( !session->prev_internal )
                {
                    uint8_t session_pong_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_session_pong_packet_sdk5( session_pong_packet, sequence, session_id, session_version, session->private_key, ping_sequence, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( session_pong_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( session_pong_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );
     
#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding session pong packet to previous hop %s (public address)\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                        relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, session_pong_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
                else
                {
                    uint8_t session_pong_packet[RELAY_MAX_PACKET_BYTES];
                    packet_bytes = relay_write_session_pong_packet_sdk5( session_pong_packet, sequence, session_id, session_version, session->private_key, ping_sequence, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port );
                    if ( packet_bytes > 0 )
                    {
                        assert( relay_basic_packet_filter_sdk5( session_pong_packet, packet_bytes ) );
                        assert( relay_advanced_packet_filter_sdk5( session_pong_packet, current_magic, relay_internal_address_data, relay_internal_address_bytes, relay_internal_address_port, prev_address_data, prev_address_bytes, prev_address_port, packet_bytes ) );
     
#if INTENSIVE_RELAY_DEBUGGING
                        char prev_hop_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
                        relay_address_to_string( &session->prev_address, prev_hop_address );
                        printf( "[%s] forwarding session pong packet to previous hop %s (internal address)\n", from_string, prev_hop_address );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                        relay->counters[RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS]++;

                        relay_platform_socket_send_packet( relay->socket, &session->prev_address, session_pong_packet, packet_bytes );
                        relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                        relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                    }
                }
            }
            else if ( packet_id == RELAY_NEAR_PING_PACKET_SDK5 )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "[%s] received near relay ping packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                relay->counters[RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED]++;

                if ( packet_bytes != 8 + 8 + RELAY_ENCRYPTED_PING_TOKEN_BYTES_SDK5 )
                {
#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] ignored relay near ping packet. bad packet size (%d) [sdk5]\n", from_string, packet_bytes );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_NEAR_PING_PACKET_BAD_SIZE]++;
                    continue;
                }

                const uint8_t * const_p = p;

                uint64_t ping_sequence = relay_read_uint64( &const_p );
                uint64_t session_id = relay_read_uint64( &const_p );

                uint8_t pong_packet[RELAY_MAX_PACKET_BYTES];
                packet_bytes = relay_write_pong_packet_sdk5( pong_packet, ping_sequence, session_id, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, from_address_data, from_address_bytes, from_address_port );
                if ( packet_bytes > 0 )
                {
                    assert( relay_basic_packet_filter_sdk5( pong_packet, packet_bytes ) );
                    assert( relay_advanced_packet_filter_sdk5( pong_packet, current_magic, relay_public_address_data, relay_public_address_bytes, relay_public_address_port, from_address_data, from_address_bytes, from_address_port, packet_bytes ) );

#if INTENSIVE_RELAY_DEBUGGING
                    printf( "[%s] responded with near relay pong packet [sdk5]\n", from_string );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                    relay->counters[RELAY_COUNTER_NEAR_PING_PACKET_RESPONDED_WITH_PONG]++;

                    relay_platform_socket_send_packet( relay->socket, &from, pong_packet, packet_bytes );
                    relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
                    relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
                }
            }
        }
    }

    RELAY_PLATFORM_THREAD_RETURN();
}

// ========================================================================================================================================

static relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC ping_thread_function( void * context )
{
    relay_t * relay = (relay_t*) context;

    while ( !quit )
    {
        relay_platform_mutex_acquire( relay->mutex );

        if ( relay->relays_dirty )
        {
            relay_manager_update( relay->relay_manager, relay->num_relays, relay->relay_ids, relay->relay_addresses );
            relay->relays_dirty = false;
        }

        double current_time = relay_platform_time();

        struct ping_data_t
        {
            uint64_t sequence;
            relay_address_t address;
        };

        int num_pings = 0;
        ping_data_t pings[MAX_RELAYS];

        for ( int i = 0; i < relay->relay_manager->num_relays; ++i )
        {
            if ( relay->relay_manager->relay_last_ping_time[i] + RELAY_PING_TIME <= current_time )
            {
                pings[num_pings].sequence = relay_ping_history_ping_sent( relay->relay_manager->relay_ping_history[i], current_time );
                pings[num_pings].address = relay->relay_manager->relay_addresses[i];
                relay->relay_manager->relay_last_ping_time[i] = current_time;
                num_pings++;
            }
        }

        relay_platform_mutex_release( relay->mutex );

        for ( int i = 0; i < num_pings; ++i )
        {
            uint8_t packet_data[1+8];
            packet_data[0] = RELAY_PING_PACKET;
            uint8_t * p = packet_data + 1;
            relay_write_uint64( &p, pings[i].sequence );

#if INTENSIVE_RELAY_DEBUGGING
            char to_address[RELAY_MAX_ADDRESS_STRING_LENGTH];
            relay_address_to_string( &pings[i].address, to_address);
            printf("sending relay ping packet to %s\n", to_address);
#endif // #if INTENSIVE_RELAY_DEBUGGING
            relay->counters[RELAY_COUNTER_RELAY_PING_PACKET_SENT]++;

            const int packet_bytes = 1 + 8;

            relay_platform_socket_send_packet( relay->socket, &pings[i].address, packet_data, packet_bytes );
            relay->counters[RELAY_COUNTER_PACKETS_SENT]++;
            relay->counters[RELAY_COUNTER_BYTES_SENT] += packet_bytes;
        }

        relay_platform_sleep( 1.0 / 100.0 );
    }

    RELAY_PLATFORM_THREAD_RETURN();
}

// ========================================================================================================================================

int main( int argc, const char ** argv )
{
    if ( argc == 2 && strcmp(argv[1], "version" ) == 0 ) {
        printf( "%s\n", RELAY_VERSION );
        fflush( stdout );
        exit(0);
    }

    printf( "\nNetwork Next Relay (%s)\n", RELAY_VERSION );

    printf( "\nEnvironment:\n\n" );

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_name = relay_platform_getenv( "RELAY_NAME" );
    if ( !relay_name )
    {
        printf( "\nerror: RELAY_NAME not set\n\n" );
        return 1;
    }

    printf( "    relay name is '%s'\n", relay_name );

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_public_address_env = relay_platform_getenv( "RELAY_PUBLIC_ADDRESS" );
    if ( !relay_public_address_env )
    {
        printf( "\nerror: RELAY_PUBLIC_ADDRESS not set\n\n" );
        return 1;
    }

    relay_address_t relay_public_address;
    if ( relay_address_parse( &relay_public_address, relay_public_address_env ) != RELAY_OK )
    {
        printf( "\nerror: invalid relay public address '%s'\n\n", relay_public_address_env );
        return 1;
    }

    char public_address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
    printf( "    relay public address is '%s'\n", relay_address_to_string( &relay_public_address, public_address_buffer ) );

    // -----------------------------------------------------------------------------------------------------------------------------

    bool has_internal_address = false;
    relay_address_t relay_internal_address;

    const char * relay_internal_address_env = relay_platform_getenv( "RELAY_INTERNAL_ADDRESS" );
    if ( relay_internal_address_env )
    {
        if ( relay_address_parse( &relay_internal_address, relay_internal_address_env ) != RELAY_OK )
        {
            printf( "\nerror: invalid relay internal address '%s'\n\n", relay_internal_address_env );
            return 1;
        }

        char internal_address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
        printf( "    relay internal address is '%s'\n", relay_address_to_string( &relay_internal_address, internal_address_buffer ) );
        has_internal_address = true;
    }

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_public_key_env = relay_platform_getenv( "RELAY_PUBLIC_KEY" );
    if ( !relay_public_key_env )
    {
        printf( "\nerror: RELAY_PUBLIC_KEY not set\n\n" );
        return 1;
    }

    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    if ( relay_base64_decode_data( relay_public_key_env, relay_public_key, RELAY_PUBLIC_KEY_BYTES ) != RELAY_PUBLIC_KEY_BYTES )
    {
        printf( "\nerror: invalid relay public key\n\n" );
        return 1;
    }

    printf( "    relay public key is '%s'\n", relay_public_key_env );

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_private_key_env = relay_platform_getenv( "RELAY_PRIVATE_KEY" );
    if ( !relay_private_key_env )
    {
        printf( "\nerror: RELAY_PRIVATE_KEY not set\n\n" );
        return 1;
    }

    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    if ( relay_base64_decode_data( relay_private_key_env, relay_private_key, RELAY_PRIVATE_KEY_BYTES ) != RELAY_PRIVATE_KEY_BYTES )
    {
        printf( "\nerror: invalid relay private key\n\n" );
        return 1;
    }

    printf( "    relay private key is '%s'\n", relay_private_key_env );

    // -----------------------------------------------------------------------------------------------------------------------------

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

    printf( "    router public key is '%s'\n", router_public_key_env );

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_backend_hostname = relay_platform_getenv( "RELAY_BACKEND_HOSTNAME" );
    if ( !relay_backend_hostname )
    {
        printf( "\nerror: RELAY_BACKEND_HOSTNAME not set\n\n" );
        return 1;
    }

    printf( "    relay backend hostname is '%s'\n", relay_backend_hostname );

    if ( relay_initialize() != RELAY_OK )
    {
        printf( "\nerror: failed to initialize relay\n\n" );
        return 1;
    }

    // -----------------------------------------------------------------------------------------------------------------------------

#if RELAY_DEVELOPMENT

    float relay_fake_packet_loss_percent = 0.0f;
    const char * fake_packet_loss_percent_env = relay_platform_getenv( "RELAY_FAKE_PACKET_LOSS_PERCENT" );
    if ( fake_packet_loss_percent_env )
    {
        relay_fake_packet_loss_percent = atof( fake_packet_loss_percent_env );
    }

    if ( relay_fake_packet_loss_percent > 0.0f )
    {
        printf( "    fake packet loss is %.1f percent\n", relay_fake_packet_loss_percent );
    }

    float relay_fake_packet_loss_start_time = -1.0f;
    const char * fake_packet_loss_start_time_env = relay_platform_getenv( "RELAY_FAKE_PACKET_LOSS_START_TIME" );
    if ( fake_packet_loss_start_time_env )
    {
        relay_fake_packet_loss_start_time = atof( fake_packet_loss_start_time_env );
    }

    if ( relay_fake_packet_loss_start_time >= 0.0f )
    {
        printf( "    fake packet loss starts at %.1f seconds\n", relay_fake_packet_loss_start_time );
    }

#endif // #if RELAY_DEVELOPMENT

    // -----------------------------------------------------------------------------------------------------------------------------

    // IMPORTANT: Bind to 127.0.0.1 if specified, otherwise bind to 0.0.0.0
    relay_address_t bind_address;
    if ( relay_public_address.data.ipv4[0] == 127 && relay_public_address.data.ipv4[1] == 0 && relay_public_address.data.ipv4[2] == 0 && relay_public_address.data.ipv4[3] == 1 )
    {
        printf( "\nBinding to 127.0.0.1:%d\n", relay_public_address.port );
        bind_address = relay_public_address;
    }
    else
    {
        printf( "\nBinding to 0.0.0.0:%d\n", relay_public_address.port );
        memset( &bind_address, 0, sizeof(bind_address) );
        bind_address.type = RELAY_ADDRESS_IPV4;
        bind_address.port = relay_public_address.port;
    }

    relay_platform_socket_t * socket = relay_platform_socket_create( &bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.1f, 100 * 1024, 100 * 1024 );
    if ( socket == NULL )
    {
        printf( "\ncould not create socket\n\n" );
        relay_term();
        return 1;
    }

    relay_public_address.port = bind_address.port;

    printf( "\nRelay socket opened on port %d\n\n", relay_public_address.port );

    char relay_public_address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
    const char * relay_address_string = relay_address_to_string( &relay_public_address, relay_public_address_buffer );
    printf( "Relay public address is '%s'\n", relay_public_address_buffer );

    fflush( stdout );

    CURL * curl = curl_easy_init();
    if ( !curl )
    {
        printf( "\nerror: could not initialize curl\n\n" );
        relay_platform_socket_destroy( socket );
        curl_easy_cleanup( curl );
        relay_term();
        return 1;
    }

    uint8_t relay_token[RELAY_TOKEN_BYTES];
    relay_random_bytes( relay_token, RELAY_TOKEN_BYTES );

    relay_t relay;

    memset( &relay, 0, sizeof(relay_t) );

    relay.relay_manager = nullptr;
    relay.relay_public_address = relay_public_address;
    relay.relay_internal_address = relay_internal_address;
    relay.has_internal_address = has_internal_address;
    relay.socket = nullptr;
    relay.mutex = nullptr;
    relay.initialize_time = relay_platform_time();
    relay.sessions = new std::map<uint64_t, relay_session_t*>();
    memcpy( relay.relay_public_key, relay_public_key, RELAY_PUBLIC_KEY_BYTES );
    memcpy( relay.relay_private_key, relay_private_key, RELAY_PRIVATE_KEY_BYTES );
    memcpy( relay.router_public_key, router_public_key, crypto_sign_PUBLICKEYBYTES );
    relay.relays_dirty = false;
    relay.num_relays = 0;
    memset( relay.relay_ids, 0, sizeof(relay.relay_ids) );
    memset( relay.relay_addresses, 0, sizeof(relay.relay_addresses) );
    relay.envelope_bandwidth_kbps_up = 0;
    relay.envelope_bandwidth_kbps_down = 0;
#if RELAY_DEVELOPMENT
    relay.fake_packet_loss_percent = relay_fake_packet_loss_percent;
    relay.fake_packet_loss_start_time = relay_fake_packet_loss_start_time;
#endif // #if RELAY_DEVELOPMENT

    relay.socket = socket;
    relay.mutex = relay_platform_mutex_create();
    if ( !relay.mutex )
    {
        printf( "\nerror: could not create ping thread\n\n" );
        quit = 1;
    }

    relay.relay_manager = relay_manager_create();
    if ( !relay.relay_manager )
    {
        printf( "\nerror: could not create relay manager\n\n" );
        quit = 1;
    }

    relay_platform_thread_t * receive_thread = relay_platform_thread_create( receive_thread_function, &relay );
    if ( !receive_thread )
    {
        printf( "\nerror: could not create receive thread\n\n" );
        quit = 1;
    }

    relay_platform_thread_t * ping_thread = relay_platform_thread_create( ping_thread_function, &relay );
    if ( !ping_thread )
    {
        printf( "\nerror: could not create ping thread\n\n" );
        quit = 1;
    }

    signal( SIGINT, interrupt_handler );
    signal( SIGTERM, interrupt_handler );
    signal( SIGHUP, clean_shutdown_handler );

    uint8_t * update_response_memory = (uint8_t*) malloc( RESPONSE_MAX_BYTES );

    bool aborted = false;

    int update_attempts = 0;

    while ( !quit )
    {
        if ( relay_update( curl, relay_backend_hostname, relay_token, relay_address_string, update_response_memory, &relay, false ) == RELAY_OK )
        {
            update_attempts = 0;
        }
        else
        {
            if ( update_attempts++ >= RELAY_MAX_UPDATE_ATTEMPTS )
            {
                printf( "error: could not update relay %d times in a row. shutting down", RELAY_MAX_UPDATE_ATTEMPTS );
                aborted = true;
                quit = 1;
                break;
            }
        }

        relay_platform_mutex_acquire( relay.mutex );
        std::map<uint64_t, relay_session_t*>::iterator iter = relay.sessions->begin();
        while ( iter != relay.sessions->end() )
        {
            if ( iter->second && iter->second->expire_timestamp < relay_timestamp( &relay ) )
            {
#if INTENSIVE_RELAY_DEBUGGING
                printf( "session destroyed: %" PRIx64 ".%d\n", iter->second->session_id, iter->second->session_version );
#endif // #if INTENSIVE_RELAY_DEBUGGING
                relay.counters[RELAY_COUNTER_SESSION_DESTROYED]++;
                relay.envelope_bandwidth_kbps_up -= iter->second->kbps_up;          // todo: why?!
                relay.envelope_bandwidth_kbps_down -= iter->second->kbps_down;
                iter = relay.sessions->erase( iter );
            }
            else
            {
                iter++;
            }
        }
        relay_platform_mutex_release( relay.mutex );

        relay_platform_sleep( 1.0 );
    }

    if ( relay_clean_shutdown )
    {
        uint seconds = 0;
        while ( seconds++ < 60 && relay_update( curl, relay_backend_hostname, relay_token, relay_address_string, update_response_memory, &relay, false ) != RELAY_OK )
        {
            relay_platform_sleep( 1.0 );
        }

        if ( seconds < 60 )
        {
            relay_platform_sleep( 30.0 );
        }
    }

    printf( "\n\nCleaning up\n\n" );
    fflush( stdout );

    if ( receive_thread )
    {
        relay_platform_thread_join( receive_thread );
        relay_platform_thread_destroy( receive_thread );
    }

    if ( ping_thread )
    {
        relay_platform_thread_join( ping_thread );
        relay_platform_thread_destroy( ping_thread );
    }

    free( update_response_memory );

    for ( std::map<uint64_t, relay_session_t*>::iterator itor = relay.sessions->begin(); itor != relay.sessions->end(); ++itor )
    {
        delete itor->second;
    }

    delete relay.sessions;

    relay_manager_destroy( relay.relay_manager );

    relay_platform_mutex_destroy( relay.mutex );

    relay_platform_socket_destroy( socket );

    curl_easy_cleanup( curl );

    relay_term();

    return aborted ? 1 : 0;
}
