/*
    Network Next: $(NEXT_VERSION_FULL)
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#ifndef NEXT_INTERNAL_H
#define NEXT_INTERNAL_H

#include "relay.h"
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdint.h>
#include <math.h>
#include <cstring>
#include "rapidjson/document.h"
#include "rapidjson/writer.h"
#include "rapidjson/stringbuffer.h"

#if defined(__386__) || defined(i386)   || defined(__i386__)        \
    || defined(__X86)   || defined(_M_IX86)                         \
    || defined(_M_X64)  || defined(__x86_64__)                      \
    || defined(alpha)   || defined(__alpha) || defined(__alpha__)   \
    || defined(_M_ALPHA)                                            \
    || defined(ARM)  || defined(_ARM)   || defined(__arm__)         \
    || defined(WIN32)   || defined(_WIN32)  || defined(__WIN32__)   \
    || defined(_WIN32_WCE) || defined(__NT__)                       \
    || defined(__MIPSEL__)
  #define NEXT_LITTLE_ENDIAN 1
#else
  #define NEXT_BIG_ENDIAN 1
#endif

#define NEXT_PLATFORM_WINDOWS   1
#define NEXT_PLATFORM_MAC       2
#define NEXT_PLATFORM_UNIX      3
#define NEXT_PLATFORM_SWITCH    4
#define NEXT_PLATFORM_PS4       5
#define NEXT_PLATFORM_IOS       6

#if defined(NN_NINTENDO_SDK)
    #define NEXT_PLATFORM NEXT_PLATFORM_SWITCH
#elif defined(__ORBIS__)
    #define NEXT_PLATFORM NEXT_PLATFORM_PS4
#elif defined(_WIN32)
    #define NEXT_PLATFORM NEXT_PLATFORM_WINDOWS
#elif defined(__APPLE__)
    #include "TargetConditionals.h"
    #if TARGET_OS_IPHONE
        #define NEXT_PLATFORM NEXT_PLATFORM_IOS
    #else
        #define NEXT_PLATFORM NEXT_PLATFORM_MAC
    #endif
#else
    #define NEXT_PLATFORM NEXT_PLATFORM_UNIX
#endif

#if defined( _MSC_VER ) && _MSC_VER < 1700
typedef __int32 int32_t;
typedef __int64 int64_t;
#define PRId64 "I64d"
#define SCNd64 "I64d"
#define PRIx64 "I64x"
#define SCNx64 "I64x"
#else
#include <inttypes.h>
#endif

#define NEXT_ONE_SECOND_NS 1000000000LL

#define NEXT_CANARY_DISABLED                                    0
#define NEXT_CANARY_ROUTE_STATE_FAIL                            1
#define NEXT_CANARY_NEAR_FAIL                                   2
#define NEXT_CANARY_DIRECT                                      3
#define NEXT_CANARY_RANDOM_ROUTE                                4
#define NEXT_CANARY_BACKUP_FLOW                                 5
#define NEXT_CANARY_KEEPALIVE                                   6

#define NEXT_SOCKET_IPV6                                        1
#define NEXT_SOCKET_IPV4                                        2

#define NEXT_SYMMETRIC_KEY_BYTES                               32
#define NEXT_SYMMETRIC_MAC_BYTES                               16

#define NEXT_PING_MAC_BYTES                                    32
#define NEXT_PING_TOKEN_BYTES     ( 8 + 8 + NEXT_PING_MAC_BYTES )

#define NEXT_PACKET_TYPE_V2_DIRECT                              0
#define NEXT_PACKET_TYPE_V2_ROUTE_REQUEST                       1
#define NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE                      2
#define NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER                    3
#define NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT                    4
#define NEXT_PACKET_TYPE_V2_CLIENT_RELAY_PING                   7
#define NEXT_PACKET_TYPE_V2_CLIENT_RELAY_PONG                   8
#define NEXT_PACKET_TYPE_V2_DIRECT_SERVER_PING                  9
#define NEXT_PACKET_TYPE_V2_DIRECT_SERVER_PONG                 10
#define NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING                   11
#define NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG                   12
#define NEXT_PACKET_TYPE_V2_CONTINUE_REQUEST                   13
#define NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE                  14
#define NEXT_PACKET_TYPE_V2_MIGRATE                            15
#define NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE                   16
#define NEXT_PACKET_TYPE_V2_DESTROY                            17
#define NEXT_PACKET_TYPE_V2_BACKUP                             18
//#define NEXT_PACKET_TYPE_V3_MASTER_UPDATE_REQUEST              19
//#define NEXT_PACKET_TYPE_V3_MASTER_UPDATE_RESPONSE             20
//#define NEXT_PACKET_TYPE_V3_MASTER_INIT_REQUEST                21
//#define NEXT_PACKET_TYPE_V3_MASTER_INIT_RESPONSE               22
//#define NEXT_PACKET_TYPE_V3_MASTER_ROUTE_REQUEST               23
//#define NEXT_PACKET_TYPE_V3_MASTER_ROUTE_RESPONSE              24
//#define NEXT_PACKET_TYPE_V3_MASTER_COUNTER_POST                25
//#define NEXT_PACKET_TYPE_V3_MASTER_COUNTER_RESPONSE            26

#define NEXT_PACKET_TYPE_V3_OFFSET                             27

#define NEXT_PACKET_TYPE_V3_CLIENT_RELAY_PING                  27
#define NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE                     28
#define NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER                   29
#define NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT                   30
#define NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING                   31
#define NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG                   32
#define NEXT_PACKET_TYPE_V3_CLIENT_RELAY_PONG                  33
#define NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE                  34
#define NEXT_PACKET_TYPE_V3_MIGRATE                            35
#define NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE                   36
#define NEXT_PACKET_TYPE_V3_DESTROY                            37
#define NEXT_PACKET_TYPE_V3_MASTER_NEAR_REQUEST                38
#define NEXT_PACKET_TYPE_V3_MASTER_NEAR_RESPONSE               39
#define NEXT_PACKET_TYPE_V3_CLIENT_MEASUREMENT_RELAY_PING      40
#define NEXT_PACKET_TYPE_V3_CLIENT_MEASUREMENT_RELAY_PONG      41

#define NEXT_PACKET_TYPE_V4_OFFSET                             42
#define NEXT_PACKET_TYPE_V4_MASTER_UPDATE_REQUEST              42
#define NEXT_PACKET_TYPE_V4_MASTER_INIT_REQUEST                43
#define NEXT_PACKET_TYPE_V4_MASTER_ROUTE_REQUEST               44
#define NEXT_PACKET_TYPE_V4_MASTER_COUNTER_POST                45
#define NEXT_PACKET_TYPE_V4_MASTER_NEAR_REQUEST                46
#define NEXT_PACKET_TYPE_V4_MASTER_KEEPALIVE                   47
#define NEXT_PACKET_TYPE_V4_MASTER_RELAY_REPORT                48
#define NEXT_PACKET_TYPE_V4_MASTER_RELAY_RESPONSE              49
#define NEXT_PACKET_TYPE_V4_MASTER_RELAY_CONFIG_REQUEST        50
#define NEXT_PACKET_TYPE_V4_MASTER_RELAY_CONFIG_RESPONSE       51
#define NEXT_PACKET_TYPE_V4_MASTER_INIT_RESPONSE               52
#define NEXT_PACKET_TYPE_V4_MASTER_ROUTE_RESPONSE              53
#define NEXT_PACKET_TYPE_V4_MASTER_UPDATE_RESPONSE             54
#define NEXT_PACKET_TYPE_V4_MASTER_COUNTER_RESPONSE            55
#define NEXT_PACKET_TYPE_V4_MASTER_NEAR_RESPONSE               56

#define NEXT_PACKET_TYPE_V3_SDK_CLIENT_RELAY_PING              73
#define NEXT_PACKET_TYPE_V3_SDK_CLIENT_RELAY_PONG              74

#define NEXT_PACKET_TYPE_PASSTHROUGH                          255

#define NEXT_PACKET_V2_PING_PONG_BYTES                    (1+8+8)
#define NEXT_BACKUP_FLOW_BYTES                              (1+8)

#define NEXT_CLIENT_INFO_VERSION                                2

#define NEXT_ERROR_SOCKET_CREATE_FAILED                        -1
#define NEXT_ERROR_SOCKET_SET_NON_BLOCKING_FAILED              -2
#define NEXT_ERROR_SOCKET_SOCKOPT_IPV6_ONLY_FAILED             -3
#define NEXT_ERROR_SOCKET_SOCKOPT_RCVBUF_FAILED                -4
#define NEXT_ERROR_SOCKET_SOCKOPT_SNDBUF_FAILED                -5
#define NEXT_ERROR_SOCKET_SOCKOPT_RCVTIMEO_FAILED              -6
#define NEXT_ERROR_SOCKET_BIND_IPV4_FAILED                     -7
#define NEXT_ERROR_SOCKET_BIND_IPV6_FAILED                     -8
#define NEXT_ERROR_SOCKET_GET_SOCKNAME_IPV4_FAILED             -9
#define NEXT_ERROR_SOCKET_GET_SOCKNAME_IPV6_FAILED            -10

#define NEXT_SOCKET_ERROR_NONE                                  0
#define NEXT_SOCKET_ERROR_CREATE_FAILED                         1
#define NEXT_SOCKET_ERROR_SET_NON_BLOCKING_FAILED               2
#define NEXT_SOCKET_ERROR_SOCKOPT_IPV6_ONLY_FAILED              3
#define NEXT_SOCKET_ERROR_SOCKOPT_RCVBUF_FAILED                 4
#define NEXT_SOCKET_ERROR_SOCKOPT_SNDBUF_FAILED                 5
#define NEXT_SOCKET_ERROR_SOCKOPT_RCVTIMEO_FAILED               6
#define NEXT_SOCKET_ERROR_BIND_IPV4_FAILED                      7
#define NEXT_SOCKET_ERROR_BIND_IPV6_FAILED                      8
#define NEXT_SOCKET_ERROR_GET_SOCKNAME_IPV4_FAILED              9
#define NEXT_SOCKET_ERROR_GET_SOCKNAME_IPV6_FAILED             10

#define NEXT_SOCKET_FLAG_REUSEPORT                         (1<<0)

#define NEXT_REPLAY_PROTECTION_BUFFER_SIZE                    256

#define NEXT_ADDRESS_BYTES                                     19

#define NEXT_ADDRESS_BUFFER_SAFETY                             32

#define NEXT_BILLING_SLICE_SECONDS                             10

#define NEXT_FLOW_FLAG_FLOW_CREATE       ( ( uint8_t ) ( 1<<0 ) )
#define NEXT_FLOW_FLAG_FLOW_FORCED       ( ( uint8_t ) ( 1<<1 ) )

#define NEXT_FLOW_TOKEN_BYTES ( 8 + 8 + 1 + 1 + 4 + 4 + NEXT_ADDRESS_BYTES + NEXT_SYMMETRIC_KEY_BYTES )

#define NEXT_ENCRYPTED_ROUTE_STATE_MAX_BYTES                 1500

#define NEXT_ENCRYPTED_FLOW_TOKEN_BYTES ( crypto_box_NONCEBYTES + NEXT_FLOW_TOKEN_BYTES + crypto_box_MACBYTES )

#define NEXT_CONTINUE_TOKEN_BYTES ( 8 + 8 + 1 + 1 )

#define NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES ( crypto_box_NONCEBYTES + NEXT_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES )

#define NEXT_SERVER_TOKEN_BYTES ( 8 + 8 + 1 + 1 )

#define NEXT_ENCRYPTED_SERVER_TOKEN_BYTES ( crypto_box_NONCEBYTES + NEXT_SERVER_TOKEN_BYTES + crypto_box_MACBYTES )

#define NEXT_ROUTE_PREFIX_BYTES ( 1 + 4 )
#define NEXT_ROUTE_PREFIX_TYPE_NULL_BYTES ( NEXT_ROUTE_PREFIX_BYTES )
#define NEXT_ROUTE_PREFIX_TYPE_SERVER_ADDRESS_BYTES ( NEXT_ROUTE_PREFIX_BYTES + NEXT_ADDRESS_BYTES )
#define NEXT_ROUTE_PREFIX_TYPE_FORCED_ROUTE_BYTES ( NEXT_ROUTE_PREFIX_BYTES )
#define NEXT_ROUTE_PREFIX_BYTES_MAX           1024
#define NEXT_ROUTE_PREFIX_TYPE_NULL              0
#define NEXT_ROUTE_PREFIX_TYPE_SERVER_ADDRESS    1
#define NEXT_ROUTE_PREFIX_TYPE_DIRECT            2
#define NEXT_ROUTE_PREFIX_TYPE_FORCED_ROUTE      3

#define NEXT_HEADER_BYTES ( (int) ( 1 + 8 + 8 + 1 + 1 + NEXT_SYMMETRIC_MAC_BYTES ) )

#define NEXT_SOCKET_SNDBUF_SIZE ( 1 * 1024 * 1024 )
#define NEXT_SOCKET_RCVBUF_SIZE ( 1 * 1024 * 1024 )

#define NEXT_MAX_ROUTE_REQUEST_BYTES ( NEXT_ENCRYPTED_FLOW_TOKEN_BYTES * NEXT_MAX_FLOW_TOKENS )

#define NEXT_MAX_ROUTES 1

#ifndef NEXT_ENABLE_TESTS
#define NEXT_ENABLE_TESTS 0
#endif // #ifndef NEXT_ENABLE_TESTS

#ifndef NEXT_HTTP_LOG
#define NEXT_HTTP_LOG 0
#endif // #ifndef NEXT_HTTP_LOG

#ifndef NEXT_ENABLE_LOGGING
#define NEXT_ENABLE_LOGGING 1
#endif // #ifndef NEXT_ENABLE_LOGGING

NEXT_EXPORT_FUNC void (*next_assert_function)( const char *, const char *, const char * file, int line );

extern uint8_t NEXT_KEY_MASTER[];

#ifndef NDEBUG
#define next_assert( condition )                                                            \
do                                                                                          \
{                                                                                           \
    if ( !(condition) )                                                                     \
    {                                                                                       \
        next_assert_function( #condition, __FUNCTION__, __FILE__, __LINE__ );               \
    }                                                                                       \
} while(0)
#else
#define next_assert( ignore ) ((void)0)
#endif

extern void next_flow_log( int level, uint64_t flow_id, uint8_t flow_version, const char *, ... );

NEXT_EXPORT_FUNC void * next_alloc( size_t bytes );
NEXT_EXPORT_FUNC void * next_realloc( void *, size_t bytes_new );
NEXT_EXPORT_FUNC void next_free( void * p );

#if defined( _MSC_VER ) && _MSC_VER < 1900

#define snprintf c99_snprintf
#define vsnprintf c99_vsnprintf

__inline int c99_vsnprintf(char *outBuf, size_t size, const char *format, va_list ap)
{
    int count = -1;

    if (size != 0)
        count = _vsnprintf_s(outBuf, size, _TRUNCATE, format, ap);
    if (count == -1)
        count = _vscprintf(format, ap);

    return count;
}

__inline int c99_snprintf(char *outBuf, size_t size, const char *format, ...)
{
    int count;
    va_list ap;

    va_start(ap, format);
    count = c99_vsnprintf(outBuf, size, format, ap);
    va_end(ap);

    return count;
}

#endif

extern int next_internal_init();

extern void next_internal_term();

extern uint16_t next_ntohs( uint16_t in );
extern uint16_t next_htons( uint16_t in );

template <typename T> struct next_vector_t
{
    T * data;
    int length;
    int reserved;

    inline next_vector_t( int reserve_count = 0 )
    {
        next_assert( reserve_count >= 0 );
        data = 0;
        length = 0;
        reserved = 0;
        if ( reserve_count > 0 )
        {
            reserve( reserve_count );
        }
    }

    inline ~next_vector_t()
    {
        clear();
    }

    inline void clear() 
    {
        if ( data )
        {
            next_free( data );
        }
        data = NULL;
        length = 0;
        reserved = 0;
    }

    inline const T & operator [] ( int i ) const
    {
        next_assert( data );
        next_assert( i >= 0 && i < length );
        return *( data + i );
    }

    inline T & operator [] ( int i )
    {
        next_assert( data );
        next_assert( i >= 0 && i < length );
        return *( data + i );
    }

    void reserve( int size )
    {
        next_assert( size >= 0 );
        if ( size > reserved )
        {
            const double VECTOR_GROWTH_FACTOR = 1.5;
            const int VECTOR_INITIAL_RESERVATION = 1;
            unsigned int next_size = (unsigned int)( pow( VECTOR_GROWTH_FACTOR, int( log( double( size ) ) / log( double( VECTOR_GROWTH_FACTOR ) ) ) + 1 ) );
            if ( !reserved )
            {
                next_size = next_size > VECTOR_INITIAL_RESERVATION ? next_size : VECTOR_INITIAL_RESERVATION;
                data = (T*)( next_alloc( next_size * sizeof(T) ) );
            }
            else
            {
                data = (T*)( next_realloc( data, next_size * sizeof(T) ) );
            }
            next_assert( data );
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
        next_assert( data );
        next_assert( i >= 0 && i < length );
        if ( i != length - 1 )
        {
            data[i] = data[length - 1];
        }
        length--;
    }

    void remove_ordered( int i )
    {
        next_assert( data );
        next_assert( i >= 0 && i < length );
        memmove( &data[i], &data[i + 1], sizeof( T ) * ( length - ( i + 1 ) ) );
        length--;
    }

    T * insert( int i )
    {
        next_assert( i >= 0 && i <= length );
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

struct next_json_allocator_t
{
public:
    static const bool kNeedFree = true;

    void * Malloc( size_t size )
    {
        if ( size )
            return next_alloc( size );
        else
            return NULL;
    }

    void * Realloc( void * original_ptr, size_t original_size, size_t new_size )
    {
        (void) original_size;
        if (new_size == 0)
        {
            next_free( original_ptr );
            return NULL;
        }
        return next_realloc( original_ptr, new_size );
    }

    static void Free( void * ptr )
    {
        next_free( ptr );
    }
};

typedef rapidjson::GenericDocument<rapidjson::UTF8<char>, next_json_allocator_t> next_json_document_t;
typedef rapidjson::GenericStringBuffer<rapidjson::UTF8<char>, next_json_allocator_t> next_json_string_buffer_t;
typedef rapidjson::GenericValue<rapidjson::UTF8<char>, next_json_allocator_t> next_json_value_t;
typedef rapidjson::Writer<next_json_string_buffer_t, rapidjson::UTF8<char>, rapidjson::UTF8<char>, next_json_allocator_t> next_json_writer_t;
typedef rapidjson::SizeType next_json_size_t;

#ifdef _MSC_VER
#define SODIUM_STATIC
#pragma warning(disable:4996)
#pragma warning(push)
#pragma warning(disable:4324)
#endif // #ifdef _MSC_VER

#include <sodium.h>

#ifdef _MSC_VER
#pragma warning(pop)
#endif

struct next_replay_protection_t
{
    uint64_t most_recent_sequence;
    uint64_t received_packet[NEXT_REPLAY_PROTECTION_BUFFER_SIZE];
};

extern void next_replay_protection_reset( next_replay_protection_t * replay_protection );

extern int next_replay_protection_already_received( next_replay_protection_t * replay_protection, uint64_t sequence );

extern void next_replay_protection_advance_sequence( next_replay_protection_t * replay_protection, uint64_t sequence );

#define NEXT_FLOW_BANDWIDTH_CHECK_INVALID_TIME ( -1 )
#define NEXT_FLOW_BANDWIDTH_CHECK_INVALID_BITS ( (uint64_t)(-1) )
#define NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS(_bytes) ( ( 14 + 20 + 8 + (_bytes) + 4 ) * 8 )
#define NEXT_FLOW_BANDWIDTH_MAX_BITS_ALLOWED(_interval, _kbps) ( (uint64_t)( double( (_kbps) * 1000 * _interval ) / double( NEXT_ONE_SECOND_NS ) ) )

struct next_flow_bandwidth_t
{
    int64_t last_bandwidth_check_timestamp;
    uint64_t bits_sent_in_current_interval;
};

extern void next_flow_bandwidth_reset( next_flow_bandwidth_t * flow_bandwidth );

extern bool next_flow_bandwidth_over_budget( next_flow_bandwidth_t * flow_bandwidth, int64_t current_time, int64_t interval, uint32_t kbps_allowed, uint32_t packet_bytes_sent );

struct next_flow_token_t
{
    uint64_t expire_timestamp;
    uint64_t flow_id;
    uint8_t flow_version;
    uint8_t flow_flags;
    int kbps_up;
    int kbps_down;
    next_address_t next_address;
    uint8_t private_key[NEXT_SYMMETRIC_KEY_BYTES];
};

struct next_route_stats_t
{
    float rtt;
    float cost;
    float jitter;
    float packet_loss;
};

extern int next_base64_encode_string( const char * input, char * output, size_t output_size );

extern int next_base64_decode_string( const char * input, char * output, size_t output_size );

extern int next_base64_encode_data( const uint8_t * input, size_t input_length, char * output, size_t output_size );

extern int next_base64_decode_data( const char * input, uint8_t * output, size_t output_size );

extern void next_write_flow_token( next_flow_token_t * token, uint8_t * buffer, int buffer_length );

extern void next_read_flow_token( next_flow_token_t * token, uint8_t * buffer );

extern int next_encrypt_flow_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length );

extern int next_decrypt_flow_token( uint8_t * sender_public_key, uint8_t * receiver_private_key, uint8_t * nonce, uint8_t * buffer );

extern int next_write_encrypted_flow_token( uint8_t ** buffer, next_flow_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key );

extern int next_read_encrypted_flow_token( uint8_t ** buffer, next_flow_token_t * token, uint8_t * sender_public_key, uint8_t * receiver_private_key );

extern void next_write_address( uint8_t ** buffer, next_address_t * address );

extern void next_read_address( uint8_t ** buffer, next_address_t * address );

struct next_continue_token_t
{
    uint64_t expire_timestamp;
    uint64_t flow_id;
    uint8_t flow_version;
    uint8_t flow_flags;
};

extern void next_write_continue_token( next_continue_token_t * token, uint8_t * buffer, int buffer_length );

extern void next_read_continue_token( next_continue_token_t * token, uint8_t * buffer );

extern int next_encrypt_continue_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length );

extern int next_decrypt_continue_token( uint8_t * sender_public_key, uint8_t * receiver_private_key, uint8_t * nonce, uint8_t * buffer );

extern int next_write_encrypted_continue_token( uint8_t ** buffer, next_continue_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key );

extern int next_read_encrypted_continue_token( uint8_t ** buffer, next_continue_token_t * token, uint8_t * sender_public_key, uint8_t * receiver_private_key );

struct next_server_token_t
{
    uint64_t expire_timestamp;
    uint64_t flow_id;
    uint8_t flow_version;
    uint8_t flow_flags;
};

extern void next_write_server_token( next_server_token_t * token, uint8_t * buffer, int buffer_length );

extern void next_read_server_token( next_server_token_t * token, uint8_t * buffer );

extern int next_encrypt_server_token( uint8_t * sender_private_key, uint8_t * receiver_public_key, uint8_t * nonce, uint8_t * buffer, int buffer_length );

extern int next_decrypt_server_token( uint8_t * sender_public_key, uint8_t * receiver_private_key, uint8_t * nonce, uint8_t * buffer );

extern int next_write_encrypted_server_token( uint8_t ** buffer, next_server_token_t * token, uint8_t * sender_private_key, uint8_t * receiver_public_key );

extern int next_read_encrypted_server_token( uint8_t ** buffer, next_server_token_t * token, uint8_t * sender_public_key, uint8_t * receiver_private_key );

extern int next_write_header( uint8_t type, uint64_t sequence, uint64_t flow_id, uint8_t flow_version, uint8_t flow_flags, uint8_t * private_key, uint8_t * buffer, int buffer_length );

extern int next_peek_header( uint8_t * type, uint64_t * sequence, uint64_t * flow_id, uint8_t * flow_version, uint8_t * flow_flags, uint8_t * buffer, int buffer_length );

extern int next_read_header( uint8_t * type, uint64_t * sequence, uint64_t * flow_id, uint8_t * flow_version, uint8_t * flow_flags, uint8_t * private_key, uint8_t * buffer, int buffer_length );

inline int next_sequence_greater_than( uint8_t s1, uint8_t s2 )
{
    return ( ( s1 > s2 ) && ( s1 - s2 <= 128 ) ) || 
           ( ( s1 < s2 ) && ( s2 - s1  > 128 ) );
}

inline int next_sequence_less_than( uint8_t s1, uint8_t s2 )
{
    return next_sequence_greater_than( s2, s1 );
}

extern int next_encrypt_aead( uint8_t * message, uint64_t message_length, 
  uint8_t * additional, uint64_t additional_length,
  uint8_t * nonce,
  uint8_t * key );

extern int next_decrypt_aead( uint8_t * message, uint64_t message_length, 
  uint8_t * additional, uint64_t additional_length,
  uint8_t * nonce,
  uint8_t * key );

extern void next_print_bytes( const char * label, const uint8_t * data, int data_bytes );

extern uint64_t next_direct_address_to_flow_id( next_address_t * direct_address );

extern void next_direct_address_from_flow_id( uint64_t flow_id, next_address_t * direct_address );

extern void next_session_to_address( int session_index, uint8_t session_sequence, next_address_t * session_address );

extern void next_session_from_address( next_address_t * session_address, int * session_index, uint8_t * session_sequence );

extern void next_write_uint8( uint8_t ** p, uint8_t value );

extern void next_write_uint16( uint8_t ** p, uint16_t value );

extern void next_write_uint32( uint8_t ** p, uint32_t value );

extern void next_write_uint64( uint8_t ** p, uint64_t value );

extern void next_write_float32( uint8_t ** p, float value );

extern void next_write_float64( uint8_t ** p, double value );

extern void next_write_bytes( uint8_t ** p, uint8_t * byte_array, int num_bytes );

extern uint8_t next_read_uint8( uint8_t ** p );

extern uint16_t next_read_uint16( uint8_t ** p );

extern uint32_t next_read_uint32( uint8_t ** p );

extern uint64_t next_read_uint64( uint8_t ** p );

extern float next_read_float32( uint8_t ** p );

extern double next_read_float64( uint8_t ** p );

extern void next_read_bytes( uint8_t ** p, uint8_t * byte_array, int num_bytes );

extern void next_random_bytes( uint8_t * data, int bytes );

#include "relay_unix.h"

struct next_socket_t;

extern int next_inet_pton4( const char * address_string, uint32_t * address_out );

extern int next_inet_pton6( const char * address_string, uint16_t * address_out );

extern int next_inet_ntop6( const uint16_t * address, char * address_string, size_t address_string_size );

extern int next_socket_create( next_socket_t * socket, const next_address_t * address, int non_blocking, int send_buffer_size, int receive_buffer_size, int flags );

extern void next_socket_destroy( next_socket_t * socket );

extern void next_socket_send_packet( next_socket_t * socket, const next_address_t * to, void * packet_data, int packet_bytes );

extern int next_socket_receive_packet( next_socket_t * socket, next_address_t * from, void * packet_data, int max_packet_size );

extern int next_address_resolve( const char * in, next_address_t * out );

extern int next_platform_init( );

extern void next_platform_term();

extern int64_t next_platform_time();

extern int next_log_level();

extern const char * next_log_level_str( int level );

typedef next_thread_return_t (NEXT_THREAD_FUNC *next_thread_func_t)(void*);

extern int next_thread_create( next_thread_t * thread, next_thread_func_t fn, void * arg );

extern void next_thread_join( next_thread_t * thread );

extern int next_mutex_init( next_mutex_t * mutex );

extern void next_mutex_destroy( next_mutex_t * mutex );

extern void _next_mutex_acquire( next_mutex_t* mutex );

extern void _next_mutex_release( next_mutex_t* mutex );

#define NEXT_MUTEX_DEBUG 0

#if NEXT_MUTEX_DEBUG
#define next_mutex_acquire( mutex )\
do\
{\
    _next_mutex_acquire( mutex );\
    printf("%s:%d: acquired: level %d\n", __func__, __LINE__, (mutex)->level);\
} while ( 0 )
#else
#define next_mutex_acquire( mutex ) _next_mutex_acquire( mutex )
#endif

#if NEXT_MUTEX_DEBUG
#define next_mutex_release( mutex )\
do\
{\
    printf("%s:%d: releasing: level %d\n", __func__, __LINE__, (mutex)->level - 1);\
    _next_mutex_release( mutex );\
} while ( 0 )
#else
#define next_mutex_release( mutex ) _next_mutex_release( mutex )
#endif

struct resolver_t
{
    int64_t resolve_last;
    next_address_t address;
    char address_string[512];
    next_address_t thread_local_address;
    next_thread_t thread;
    next_mutex_t mutex;
    bool resolve_needed;
    bool resolving;
};

void resolver_init( resolver_t * resolver, const char * address );

const next_address_t * resolver_address( resolver_t * resolver );

const char * resolver_address_string( resolver_t * resolver );

void resolver_update( resolver_t * resolver );

void resolver_destroy( resolver_t * resolver );

#endif // #ifndef NEXT_INTERNAL_H
