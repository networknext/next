
// Relay testbed

#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include <string.h>

#define RELAY_PING_PACKET          11
#define RELAY_PING_TOKEN_BYTES     32
#define RELAY_PING_KEY_BYTES       32

#if !defined ( RELAY_LITTLE_ENDIAN ) && !defined( RELAY_BIG_ENDIAN )

  #ifdef __BYTE_ORDER__
    #if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
      #define RELAY_LITTLE_ENDIAN 1
    #elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
      #define RELAY_BIG_ENDIAN 1
    #else
      #error Unknown machine endianess detected. Please define RELAY_LITTLE_ENDIAN or RELAY_BIG_ENDIAN.
    #endif // __BYTE_ORDER__

  // Detect with GLIBC's endian.h
  #elif defined(__GLIBC__)
    #include <endian.h>
    #if (__BYTE_ORDER == __LITTLE_ENDIAN)
      #define RELAY_LITTLE_ENDIAN 1
    #elif (__BYTE_ORDER == __BIG_ENDIAN)
      #define RELAY_BIG_ENDIAN 1
    #else
      #error Unknown machine endianess detected. Please define RELAY_LITTLE_ENDIAN or RELAY_BIG_ENDIAN.
    #endif // __BYTE_ORDER

  // Detect with _LITTLE_ENDIAN and _BIG_ENDIAN macro
  #elif defined(_LITTLE_ENDIAN) && !defined(_BIG_ENDIAN)
    #define RELAY_LITTLE_ENDIAN 1
  #elif defined(_BIG_ENDIAN) && !defined(_LITTLE_ENDIAN)
    #define RELAY_BIG_ENDIAN 1

  // Detect with architecture macros
  #elif    defined(__sparc)     || defined(__sparc__)                           \
        || defined(_POWER)      || defined(__powerpc__)                         \
        || defined(__ppc__)     || defined(__hpux)      || defined(__hppa)      \
        || defined(_MIPSEB)     || defined(_POWER)      || defined(__s390__)
    #define RELAY_BIG_ENDIAN 1
  #elif    defined(__i386__)    || defined(__alpha__)   || defined(__ia64)      \
        || defined(__ia64__)    || defined(_M_IX86)     || defined(_M_IA64)     \
        || defined(_M_ALPHA)    || defined(__amd64)     || defined(__amd64__)   \
        || defined(_M_AMD64)    || defined(__x86_64)    || defined(__x86_64__)  \
        || defined(_M_X64)      || defined(__bfin__)
    #define RELAY_LITTLE_ENDIAN 1
  #elif defined(_MSC_VER) && defined(_M_ARM)
    #define RELAY_LITTLE_ENDIAN 1
  #else
    #error Unknown machine endianess detected. Please define RELAY_LITTLE_ENDIAN or RELAY_BIG_ENDIAN.
  #endif

#endif

inline uint64_t bswap( uint32_t value )
{
#ifdef __GNUC__
    return __builtin_bswap32( value );
#else // #ifdef __GNUC__
    uint32_t output;
    output  = ( value & 0xFF000000 ) >> 24;
    output |= ( value & 0x00FF0000 ) >> 8;
    output |= ( value & 0x0000FF00 ) << 8;
    output |= ( value & 0x000000FF ) << 24;
#endif // #ifdef __GNUC__
}

uint16_t relay_ntohs( uint16_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
#endif // #if RELAY_BIG_ENDIAN
}

uint16_t relay_htons( uint16_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
#endif // #if RELAY_BIG_ENDIAN
}

inline uint32_t relay_ntohl( uint32_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return bswap( in );
#endif // #if RELAY_BIG_ENDIAN
}

inline uint32_t relay_htonl( uint32_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return bswap( in );
#endif // #if RELAY_BIG_ENDIAN
}

#pragma pack(push, 1)
struct ping_token_data
{
    uint8_t ping_key[RELAY_PING_KEY_BYTES];
    uint64_t expire_timestamp;                         
    uint32_t source_address;                                                   // big endian
    uint32_t dest_address;                                                     // big endian
    uint16_t source_port;                                                      // big endian
    uint16_t dest_port;                                                        // big endian
};
#pragma pack(pop)

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

static void relay_generate_pittle( uint8_t * output, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length )
{
    assert( output );
    assert( from_address );
    assert( to_address );
    assert( packet_length > 0 );
#if RELAY_BIG_ENDIAN
    relay_bswap( packet_length );
#endif // #if RELAY_BIG_ENDIAN
    uint16_t sum = 0;
    for ( int i = 0; i < 4; ++i ) { sum += (uint8_t) from_address[i]; }
    for ( int i = 0; i < 4; ++i ) { sum += (uint8_t) to_address[i]; }
    const char * packet_length_data = (const char*) &packet_length;
    sum += (uint8_t) packet_length_data[0];
    sum += (uint8_t) packet_length_data[1];
#if RELAY_BIG_ENDIAN
    relay_bswap( sum );
#endif // #if RELAY_BIG_ENDIAN
    const char * sum_data = (const char*) &sum;
    output[0] = 1 | ( (uint8_t)sum_data[0] ^ (uint8_t)sum_data[1] ^ 193 );
    output[1] = 1 | ( ( 255 - output[0] ) ^ 113 );
}

static void relay_generate_chonkle( uint8_t * output, const uint8_t * magic, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length )
{
    assert( output );
    assert( magic );
    assert( from_address );
    assert( to_address );
    assert( packet_length > 0 );
#if RELAY_BIG_ENDIAN
    relay_bswap( packet_length );
#endif // #if RELAY_BIG_ENDIAN
    relay_fnv_t fnv;
    relay_fnv_init( &fnv );
    relay_fnv_write( &fnv, magic, 8 );
    relay_fnv_write( &fnv, from_address, 4 );
    relay_fnv_write( &fnv, to_address, 4 );
    relay_fnv_write( &fnv, (const uint8_t*) &packet_length, 2 );
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

void relay_address_data( uint32_t address, uint8_t * output )
{
    output[0] = address & 0xFF;
    output[1] = ( address >> 8  ) & 0xFF;
    output[2] = ( address >> 16 ) & 0xFF;
    output[3] = ( address >> 24 ) & 0xFF;
}

void relay_write_ping_packet( uint8_t * packet_data, int & packet_length, uint64_t sequence, uint64_t expire_timestamp, uint32_t from, uint32_t to, uint8_t internal, uint8_t * magic )
{
    assert( packet_data );

    uint8_t ping_token[RELAY_PING_TOKEN_BYTES];

    for ( int i = 0; i < RELAY_PING_TOKEN_BYTES; i++ )
    {
        ping_token[i] = (uint8_t) i;
    }

    packet_data[0] = RELAY_PING_PACKET;
    uint8_t * a = packet_data + 1;
    uint8_t * b = packet_data + 3;
    uint8_t * p = packet_data + 18;

    relay_write_uint64( &p, sequence );
    relay_write_uint64( &p, expire_timestamp );
    relay_write_uint8( &p, internal );
    relay_write_bytes( &p, ping_token, RELAY_PING_TOKEN_BYTES );

    packet_length = p - packet_data;

    uint8_t to_address_data[4];
    uint8_t from_address_data[4];

    relay_address_data( to, to_address_data );
    relay_address_data( from, from_address_data );

    relay_generate_pittle( a, from_address_data, to_address_data, packet_length );
    relay_generate_chonkle( b, magic, from_address_data, to_address_data, packet_length );
}

// --------------------------------------------

static void ref_generate_pittle( uint8_t * output, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length )
{
    assert( output );
    assert( from_address );
    assert( to_address );
    assert( packet_length > 0 );
#if RELAY_BIG_ENDIAN
    relay_bswap( packet_length );
#endif // #if RELAY_BIG_ENDIAN
    uint16_t sum = 0;
    for ( int i = 0; i < 4; ++i ) { sum += uint8_t(from_address[i]); }
    for ( int i = 0; i < 4; ++i ) { sum += uint8_t(to_address[i]); }
    const char * packet_length_data = (const char*) &packet_length;
    sum += uint8_t(packet_length_data[0]);
    sum += uint8_t(packet_length_data[1]);
#if RELAY_BIG_ENDIAN
    relay_bswap( sum );
#endif // #if RELAY_BIG_ENDIAN
    const char * sum_data = (const char*) &sum;
    output[0] = 1 | ( uint8_t(sum_data[0]) ^ uint8_t(sum_data[1]) ^ 193 );
    output[1] = 1 | ( ( 255 - output[0] ) ^ 113 );
}

static void ref_generate_chonkle( uint8_t * output, const uint8_t * magic, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length )
{
    assert( output );
    assert( magic );
    assert( from_address );
    assert( to_address );
    assert( packet_length > 0 );
#if RELAY_BIG_ENDIAN
    relay_bswap( packet_length );
#endif // #if RELAY_BIG_ENDIAN
    relay_fnv_t fnv;
    relay_fnv_init( &fnv );
    relay_fnv_write( &fnv, magic, 8 );
    relay_fnv_write( &fnv, from_address, 4 );
    relay_fnv_write( &fnv, to_address, 4 );
    relay_fnv_write( &fnv, (const uint8_t*) &packet_length, 2 );
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

bool ref_basic_packet_filter( const uint8_t * data, uint16_t packet_length )
{
    if ( packet_length < 18 )
        return false;

    if ( data[0] < 0x01 || data[0] > 0x0E )
        return false;

    if ( data[2] != ( 1 | ( ( 255 - data[1] ) ^ 113 ) ) )
        return false;

    if ( data[3] < 0x2A || data[3] > 0x2D )
        return false;

    if ( data[4] < 0xC8 || data[4] > 0xE7 )
        return false;

    if ( data[5] < 0x05 || data[5] > 0x44 )
        return false;

    if ( data[7] < 0x4E || data[7] > 0x51 )
        return false;

    if ( data[8] < 0x60 || data[8] > 0xDF )
        return false;

    if ( data[9] < 0x64 || data[9] > 0xE3 )
        return false;

    if ( data[10] != 0x07 && data[10] != 0x4F )
        return false;

    if ( data[11] != 0x25 && data[11] != 0x53 )
        return false;

    if ( data[12] < 0x7C || data[12] > 0x83 )
        return false;

    if ( data[13] < 0xAF || data[13] > 0xB6 )
        return false;

    if ( data[14] < 0x21 || data[14] > 0x60 )
        return false;

    if ( data[15] != 0x61 && data[15] != 0x05 && data[15] != 0x2B && data[15] != 0x0D )
        return false;

    if ( data[16] < 0xD2 || data[16] > 0xF1 )
        return false;

    if ( data[17] < 0x11 || data[17] > 0x90 )
        return false;

    return true;
}

bool ref_advanced_packet_filter( const uint8_t * data, const uint8_t * magic, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length )
{
    if ( packet_length < 18 )
        return false;
    uint8_t a[2];
    uint8_t b[15];
    ref_generate_pittle( a, from_address, to_address, packet_length );
    ref_generate_chonkle( b, magic, from_address, to_address, packet_length );
    if ( memcmp( a, data + 1, 2 ) != 0 )
        return false;
    if ( memcmp( b, data + 3, 15 ) != 0 )
        return false;
    return true;
}

// ----------------------------------------------------------------

void relay_write_ping_token( uint8_t * data, uint32_t source_address, uint16_t source_port, uint32_t dest_address, uint16_t dest_port, uint64_t expire_timestamp, uint8_t * ping_key )
{
    struct ping_token_data token_data;

    token_data.source_address = relay_htonl( source_address );
    token_data.source_port = relay_htonl( source_port );
    token_data.dest_address = relay_htonl( dest_address );
    token_data.dest_port = relay_htons( dest_port );
    token_data.expire_timestamp = expire_timestamp;

    /*
    memcpy( token_data.ping_key, ping->ping_key, RELAY_PING_KEY_BYTES );

    uint8_t ping_token[RELAY_PING_TOKEN_BYTES];
    */

    // crypto_hash_sha256( ping_token, (const unsigned char*) &token_data, sizeof(struct ping_token_data) );
}

// ----------------------------------------------------------------

int main( int argc, char *argv[] )
{
    printf( "test relay\n" );

    uint8_t packet_data[256];

    int packet_length = 0;

    uint64_t sequence = 1000000ULL;
    uint64_t expire_timestamp = 123456789101112ULL;
    uint32_t from = 0x11223344;
    uint32_t to = 0x55667788;
    uint8_t internal = 1;
    uint8_t magic[8] = { 1, 2, 3, 4, 5, 6, 7, 8 };

    relay_write_ping_packet( packet_data, packet_length, sequence, expire_timestamp, from, to, internal, magic );

    if ( !ref_basic_packet_filter( packet_data, packet_length ) )
    {
        printf( "basic packet filter failed\n" );
        exit( 1 );
    }

    uint8_t to_address_data[4];
    uint8_t from_address_data[4];

    relay_address_data( to, to_address_data );
    relay_address_data( from, from_address_data );

    if ( !ref_advanced_packet_filter( packet_data, magic, from_address_data, to_address_data, packet_length ) )
    {
        printf( "advanced packet filter failed\n" );
        exit( 1 );
    }

    char network_interface_name[1024];
    network_interface_name[0] = 'a';
    network_interface_name[1] = 's';
    network_interface_name[2] = 's';
    network_interface_name[3] = '\0';

    char command[2048];
    snprintf( command, sizeof(command), "sudo ifconfig %s mtu 1500 up", (const char*) &network_interface_name[0] );

    return 0;
}
