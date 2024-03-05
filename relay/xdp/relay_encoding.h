
/*
    Network Next XDP Relay
*/

#ifndef RELAY_ENCODING_H
#define RELAY_ENCODING_H

#include "arpa/inet.h"
#include <stdio.h>
#include <memory.h>

inline void relay_write_uint8( uint8_t ** p, uint8_t value )
{
    **p = value;
    ++(*p);
}

inline void relay_write_uint16( uint8_t ** p, uint16_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = value >> 8;
    *p += 2;
}

inline void relay_write_uint32( uint8_t ** p, uint32_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = value >> 24;
    *p += 4;
}

inline void relay_write_uint64( uint8_t ** p, uint64_t value )
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

inline void relay_write_float32( uint8_t ** p, float value )
{
    uint32_t value_int = 0;
    char * p_value = (char*)(&value);
    char * p_value_int = (char*)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint32_t));
    relay_write_uint32( p, value_int);
}

inline void relay_write_float64( uint8_t ** p, double value )
{
    uint64_t value_int = 0;
    char * p_value = (char *)(&value);
    char * p_value_int = (char *)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint64_t));
    relay_write_uint64( p, value_int);
}

inline void relay_write_bytes( uint8_t ** p, const uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        relay_write_uint8( p, byte_array[i] );
    }
}

inline void relay_write_string( uint8_t ** p, const char * string_data, uint32_t max_length )
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

inline void relay_write_address( uint8_t ** p, uint32_t address, uint16_t port )
{
    relay_write_uint8( p, RELAY_ADDRESS_IPV4 );
    relay_write_uint32( p, htonl( address ) );
    relay_write_uint16( p, port );
}

inline uint8_t relay_read_uint8( const uint8_t ** p )
{
    uint8_t value = **p;
    ++(*p);
    return value;
}

inline uint16_t relay_read_uint16( const uint8_t ** p )
{
    uint16_t value;
    value = (*p)[0];
    value |= ( ( (uint16_t)( (*p)[1] ) ) << 8 );
    *p += 2;
    return value;
}

inline uint32_t relay_read_uint32( const uint8_t ** p )
{
    uint32_t value;
    value  = (*p)[0];
    value |= ( ( (uint32_t)( (*p)[1] ) ) << 8 );
    value |= ( ( (uint32_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint32_t)( (*p)[3] ) ) << 24 );
    *p += 4;
    return value;
}

inline uint64_t relay_read_uint64( const uint8_t ** p )
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

inline float relay_read_float32( const uint8_t ** p )
{
    uint32_t value_int = relay_read_uint32( p );
    float value_float = 0.0f;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

inline double relay_read_float64( const uint8_t ** p )
{
    uint64_t value_int = relay_read_uint64( p );
    double value_float = 0.0;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

inline void relay_read_bytes( const uint8_t ** p, uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        byte_array[i] = relay_read_uint8( p );
    }
}

inline void relay_read_string( const uint8_t ** p, char * string_data, uint32_t max_length )
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

inline bool relay_read_address( const uint8_t ** p, uint32_t * address, uint16_t * port )
{
    uint8_t type = relay_read_uint8( p );
    if ( type != RELAY_ADDRESS_IPV4 )
    {
        printf( "error: only ipv4 addresses are currently supported\n" );
        return false;
    }
    *address = relay_read_uint32( p );
    *port = relay_read_uint16( p );

    *address = ntohl( *address );
}

#endif // #ifndef RELAY_ENCODING_H
