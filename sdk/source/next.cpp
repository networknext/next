/*
    Network Next Accelerate. Copyright © 2017 - 2023 Network Next, Inc.

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
#include "next_crypto.h"
#include "next_platform.h"
#include "next_address.h"
#include "next_read_write.h"
#include "next_base64.h"
#include "next_bitpacker.h"
#include "next_serialize.h"
#include "next_stream.h"
#include "next_queue.h"
#include "next_hash.h"
#include "next_config.h"
#include "next_replay_protection.h"
#include "next_ping_history.h"
#include "next_upgrade_token.h"
#include "next_route_token.h"
#include "next_continue_token.h"

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdint.h>
#include <math.h>
#include <float.h>
#include <string.h>
#include <inttypes.h>
#if defined( _MSC_VER )
#include <malloc.h>
#endif // #if defined( _MSC_VER )
#include <time.h>
#include <atomic>

#if defined( _MSC_VER )
#pragma warning(push)
#pragma warning(disable:4996)
#pragma warning(disable:4127)
#pragma warning(disable:4244)
#pragma warning(disable:4668)
#endif

// -------------------------------------------------

void next_printf( const char * format, ... );

static void default_assert_function( const char * condition, const char * function, const char * file, int line )
{
    next_printf( "assert failed: ( %s ), function %s, file %s, line %d\n", condition, function, file, line );
    fflush( stdout );
    #if defined(_MSC_VER)
        __debugbreak();
    #elif defined(__ORBIS__)
        __builtin_trap();
    #elif defined(__PROSPERO__)
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

void (*next_assert_function_pointer)( const char * condition, const char * function, const char * file, int line ) = default_assert_function;

void next_assert_function( void (*function)( const char * condition, const char * function, const char * file, int line ) )
{
    next_assert_function_pointer = function;
}

// -------------------------------------------------------------

static int log_quiet = 0;

void next_quiet( bool flag )
{
    log_quiet = flag;
}

static int log_level = NEXT_LOG_LEVEL_INFO;

void next_log_level( int level )
{
    log_level = level;
}

const char * next_log_level_string( int level )
{
    if ( level == NEXT_LOG_LEVEL_SPAM )
        return "spam";
    else if ( level == NEXT_LOG_LEVEL_DEBUG )
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

static void default_log_function( int level, const char * format, ... )
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    if ( level != NEXT_LOG_LEVEL_NONE )
    {
        if ( !log_quiet )
        {
            const char * level_string = next_log_level_string( level );
            printf( "%.6f: %s: %s\n", next_platform_time(), level_string, buffer );
        }
    }
    else
    {
        printf( "%s\n", buffer );
    }
    va_end( args );
    fflush( stdout );
}

static void (*log_function)( int level, const char * format, ... ) = default_log_function;

void next_log_function( void (*function)( int level, const char * format, ... ) )
{
    log_function = function;
}

void next_printf( const char * format, ... )
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof(buffer), format, args );
    log_function( NEXT_LOG_LEVEL_NONE, "%s", buffer );
    va_end( args );
}

void next_printf( int level, const char * format, ... )
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

// ------------------------------------------------------------

static void * next_default_malloc_function( void * context, size_t bytes )
{
    (void) context;
    return malloc( bytes );
}

static void next_default_free_function( void * context, void * p )
{
    (void) context;
    free( p );
}

static void * (*next_malloc_function)( void * context, size_t bytes ) = next_default_malloc_function;
static void (*next_free_function)( void * context, void * p ) = next_default_free_function;

void next_allocator( void * (*malloc_function)( void * context, size_t bytes ), void (*free_function)( void * context, void * p ) )
{
    next_assert( malloc_function );
    next_assert( free_function );
    next_malloc_function = malloc_function;
    next_free_function = free_function;
}

void * next_malloc( void * context, size_t bytes )
{
    next_assert( next_malloc_function );
    return next_malloc_function( context, bytes );
}

void next_free( void * context, void * p )
{
    next_assert( next_free_function );
    return next_free_function( context, p );
}

void next_clear_and_free( void * context, void * p, size_t p_size )
{
    memset( p, 0, p_size );
    next_free( context, p );
}

// -------------------------------------------------------------

uint16_t next_ntohs( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

uint16_t next_htons( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

const char * next_user_id_string( uint64_t user_id, char * buffer, size_t buffer_size )
{
    snprintf( buffer, buffer_size, "%" PRIx64, user_id );
    return buffer;
}

uint64_t next_protocol_version()
{
#if !NEXT_DEVELOPMENT
    #define VERSION_STRING(major,minor) #major #minor
    return next_hash_string( VERSION_STRING(NEXT_VERSION_MAJOR_INT, NEXT_VERSION_MINOR_INT) );
#else // #if !NEXT_DEVELOPMENT
    return 0;
#endif // #if !NEXT_DEVELOPMENT
}

float next_random_float()
{
    uint32_t uint32_value;
    next_crypto_random_bytes( (uint8_t*)&uint32_value, sizeof(uint32_value) );
    uint64_t uint64_value = uint64_t(uint32_value);
    double double_value = double(uint64_value) / 0xFFFFFFFF;
    return float(double_value);
}

uint64_t next_random_uint64()
{
    uint64_t value;
    next_crypto_random_bytes( (uint8_t*)&value, sizeof(value) );
    return value;
}

// -------------------------------------------------------------

int next_wire_packet_bits( int payload_bytes )
{
    return ( NEXT_ETHERNET_HEADER_BYTES + NEXT_IPV4_HEADER_BYTES + NEXT_UDP_HEADER_BYTES + 1 + 15 + NEXT_HEADER_BYTES + payload_bytes + 2 ) * 8;
}

struct next_bandwidth_limiter_t
{
    uint64_t bits_sent;
    double last_check_time;
    double average_kbps;
};

void next_bandwidth_limiter_reset( next_bandwidth_limiter_t * bandwidth_limiter )
{
    next_assert( bandwidth_limiter );
    bandwidth_limiter->last_check_time = -100.0;
    bandwidth_limiter->bits_sent = 0;
    bandwidth_limiter->average_kbps = 0.0;
}

bool next_bandwidth_limiter_add_packet( next_bandwidth_limiter_t * bandwidth_limiter, double current_time, uint32_t kbps_allowed, uint32_t packet_bits )
{
    next_assert( bandwidth_limiter );
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if ( invalid || current_time - bandwidth_limiter->last_check_time >= NEXT_BANDWIDTH_LIMITER_INTERVAL - 0.001f )
    {
        bandwidth_limiter->bits_sent = 0;
        bandwidth_limiter->last_check_time = current_time;
    }
    bandwidth_limiter->bits_sent += packet_bits;
    return bandwidth_limiter->bits_sent > uint64_t(kbps_allowed) * 1000 * NEXT_BANDWIDTH_LIMITER_INTERVAL;
}

void next_bandwidth_limiter_add_sample( next_bandwidth_limiter_t * bandwidth_limiter, double kbps )
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

double next_bandwidth_limiter_usage_kbps( next_bandwidth_limiter_t * bandwidth_limiter, double current_time )
{
    next_assert( bandwidth_limiter );
    const bool invalid = bandwidth_limiter->last_check_time < 0.0;
    if ( !invalid )
    {
        const double delta_time = current_time - bandwidth_limiter->last_check_time;
        if ( delta_time > 0.1f )
        {
            const double kbps = bandwidth_limiter->bits_sent / delta_time / 1000.0;
            next_bandwidth_limiter_add_sample( bandwidth_limiter, kbps );
        }
    }
    return bandwidth_limiter->average_kbps;
}

// -------------------------------------------------------------

struct next_packet_loss_tracker_t
{
    NEXT_DECLARE_SENTINEL(0)

    uint64_t last_packet_processed;
    uint64_t most_recent_packet_received;

    NEXT_DECLARE_SENTINEL(1)

    uint64_t received_packets[NEXT_PACKET_LOSS_TRACKER_HISTORY];

    NEXT_DECLARE_SENTINEL(2)
};

void next_packet_loss_tracker_initialize_sentinels( next_packet_loss_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_INITIALIZE_SENTINEL( tracker, 0 )
    NEXT_INITIALIZE_SENTINEL( tracker, 1 )
    NEXT_INITIALIZE_SENTINEL( tracker, 2 )
}

void next_packet_loss_tracker_verify_sentinels( next_packet_loss_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_VERIFY_SENTINEL( tracker, 0 )
    NEXT_VERIFY_SENTINEL( tracker, 1 )
    NEXT_VERIFY_SENTINEL( tracker, 2 )
}

void next_packet_loss_tracker_reset( next_packet_loss_tracker_t * tracker )
{
    next_assert( tracker );

    next_packet_loss_tracker_initialize_sentinels( tracker );

    tracker->last_packet_processed = 0;
    tracker->most_recent_packet_received = 0;

    for ( int i = 0; i < NEXT_PACKET_LOSS_TRACKER_HISTORY; ++i )
    {
        tracker->received_packets[i] = 0xFFFFFFFFFFFFFFFFUL;
    }

    next_packet_loss_tracker_verify_sentinels( tracker );
}

void next_packet_loss_tracker_packet_received( next_packet_loss_tracker_t * tracker, uint64_t sequence )
{
    next_packet_loss_tracker_verify_sentinels( tracker );

    sequence++;

    const int index = int( sequence % NEXT_PACKET_LOSS_TRACKER_HISTORY );

    tracker->received_packets[index] = sequence;
    tracker->most_recent_packet_received = sequence;
}

int next_packet_loss_tracker_update( next_packet_loss_tracker_t * tracker )
{
    next_packet_loss_tracker_verify_sentinels( tracker );

    int lost_packets = 0;

    uint64_t start = tracker->last_packet_processed + 1;
    uint64_t finish = ( tracker->most_recent_packet_received > NEXT_PACKET_LOSS_TRACKER_SAFETY ) ? ( tracker->most_recent_packet_received - NEXT_PACKET_LOSS_TRACKER_SAFETY ) : 0;

    if ( finish > start && finish - start > NEXT_PACKET_LOSS_TRACKER_HISTORY )
    {
        tracker->last_packet_processed = tracker->most_recent_packet_received;
        return 0;
    }

    for ( uint64_t sequence = start; sequence <= finish; ++sequence )
    {
        const int index = int( sequence % NEXT_PACKET_LOSS_TRACKER_HISTORY );
        if ( tracker->received_packets[index] != sequence )
        {
            lost_packets++;
        }
    }

    tracker->last_packet_processed = finish;

    return lost_packets;
}

// -------------------------------------------------------------

struct next_out_of_order_tracker_t
{
    NEXT_DECLARE_SENTINEL(0)

    uint64_t last_packet_processed;
    uint64_t num_out_of_order_packets;

    NEXT_DECLARE_SENTINEL(1)
};

void next_out_of_order_tracker_initialize_sentinels( next_out_of_order_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_INITIALIZE_SENTINEL( tracker, 0 )
    NEXT_INITIALIZE_SENTINEL( tracker, 1 )
}

void next_out_of_order_tracker_verify_sentinels( next_out_of_order_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_VERIFY_SENTINEL( tracker, 0 )
    NEXT_VERIFY_SENTINEL( tracker, 1 )
}

void next_out_of_order_tracker_reset( next_out_of_order_tracker_t * tracker )
{
    next_assert( tracker );

    next_out_of_order_tracker_initialize_sentinels( tracker );

    tracker->last_packet_processed = 0;
    tracker->num_out_of_order_packets = 0;

    next_out_of_order_tracker_verify_sentinels( tracker );
}

void next_out_of_order_tracker_packet_received( next_out_of_order_tracker_t * tracker, uint64_t sequence )
{
    next_out_of_order_tracker_verify_sentinels( tracker );

    if ( sequence < tracker->last_packet_processed )
    {
        tracker->num_out_of_order_packets++;
        return;
    }

    tracker->last_packet_processed = sequence;
}

// -------------------------------------------------------------

struct next_jitter_tracker_t
{
    NEXT_DECLARE_SENTINEL(0)

    uint64_t last_packet_processed;
    double last_packet_time;
    double last_packet_delta;
    double jitter;

    NEXT_DECLARE_SENTINEL(1)
};

void next_jitter_tracker_initialize_sentinels( next_jitter_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_INITIALIZE_SENTINEL( tracker, 0 )
    NEXT_INITIALIZE_SENTINEL( tracker, 1 )
}

void next_jitter_tracker_verify_sentinels( next_jitter_tracker_t * tracker )
{
    (void) tracker;
    next_assert( tracker );
    NEXT_VERIFY_SENTINEL( tracker, 0 )
    NEXT_VERIFY_SENTINEL( tracker, 1 )
}

void next_jitter_tracker_reset( next_jitter_tracker_t * tracker )
{
    next_assert( tracker );

    next_jitter_tracker_initialize_sentinels( tracker );

    tracker->last_packet_processed = 0;
    tracker->last_packet_time = 0.0;
    tracker->last_packet_delta = 0.0;
    tracker->jitter = 0.0;

    next_jitter_tracker_verify_sentinels( tracker );
}

void next_jitter_tracker_packet_received( next_jitter_tracker_t * tracker, uint64_t sequence, double time )
{
    next_jitter_tracker_verify_sentinels( tracker );

    if ( sequence == tracker->last_packet_processed + 1 && tracker->last_packet_time > 0.0 )
    {
        const double delta = time - tracker->last_packet_time;
        const double jitter = fabs( delta - tracker->last_packet_delta );
        tracker->last_packet_delta = delta;

        if ( fabs( jitter - tracker->jitter ) > 0.00001 )
        {
            tracker->jitter += ( jitter - tracker->jitter ) * 0.01;
        }
        else
        {
            tracker->jitter = jitter;
        }
    }

    tracker->last_packet_processed = sequence;
    tracker->last_packet_time = time;
}

// -------------------------------------------------------------

struct NextUpgradeRequestPacket
{
    uint64_t protocol_version;
    uint64_t session_id;
    next_address_t client_address;
    next_address_t server_address;
    uint8_t server_kx_public_key[NEXT_CRYPTO_KX_PUBLICKEYBYTES];
    uint8_t upgrade_token[NEXT_UPGRADE_TOKEN_BYTES];
    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    NextUpgradeRequestPacket()
    {
        memset( this, 0, sizeof(NextUpgradeRequestPacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, protocol_version );
        serialize_uint64( stream, session_id );
        serialize_address( stream, client_address );
        serialize_address( stream, server_address );
        serialize_bytes( stream, server_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
        serialize_bytes( stream, upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );
        serialize_bytes( stream, upcoming_magic, 8 );
        serialize_bytes( stream, current_magic, 8 );
        serialize_bytes( stream, previous_magic, 8 );
        return true;
    }
};

struct NextUpgradeResponsePacket
{
    uint8_t client_open_session_sequence;
    uint8_t client_kx_public_key[NEXT_CRYPTO_KX_PUBLICKEYBYTES];
    uint8_t client_route_public_key[NEXT_CRYPTO_BOX_PUBLICKEYBYTES];
    uint8_t upgrade_token[NEXT_UPGRADE_TOKEN_BYTES];
    int platform_id;
    int connection_type;

    NextUpgradeResponsePacket()
    {
        memset( this, 0, sizeof(NextUpgradeResponsePacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_bits( stream, client_open_session_sequence, 8 );
        serialize_bytes( stream, client_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
        serialize_bytes( stream, client_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );
        serialize_bytes( stream, upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );
        serialize_int( stream, platform_id, NEXT_PLATFORM_UNKNOWN, NEXT_PLATFORM_MAX );
        serialize_int( stream, connection_type, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_MAX );
        return true;
    }
};

struct NextUpgradeConfirmPacket
{
    uint64_t upgrade_sequence;
    uint64_t session_id;
    next_address_t server_address;
    uint8_t client_kx_public_key[NEXT_CRYPTO_KX_PUBLICKEYBYTES];
    uint8_t server_kx_public_key[NEXT_CRYPTO_KX_PUBLICKEYBYTES];

    NextUpgradeConfirmPacket()
    {
        memset( this, 0, sizeof(NextUpgradeConfirmPacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, upgrade_sequence );
        serialize_uint64( stream, session_id );
        serialize_address( stream, server_address );
        serialize_bytes( stream, client_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
        serialize_bytes( stream, server_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
        return true;
    }
};

struct NextDirectPingPacket
{
    uint64_t ping_sequence;

    NextDirectPingPacket()
    {
        ping_sequence = 0;
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, ping_sequence );
        return true;
    }
};

struct NextDirectPongPacket
{
    uint64_t ping_sequence;

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, ping_sequence );
        return true;
    }
};

struct NextClientStatsPacket
{
    bool fallback_to_direct;
    bool next;
    bool multipath;
    bool reported;
    bool next_bandwidth_over_limit;
    int platform_id;
    int connection_type;
    float direct_kbps_up;
    float direct_kbps_down;
    float next_kbps_up;
    float next_kbps_down;
    float direct_rtt;
    float direct_jitter;
    float direct_packet_loss;
    float direct_max_packet_loss_seen;
    float next_rtt;
    float next_jitter;
    float next_packet_loss;
    float max_jitter_seen;
    int num_near_relays;
    uint64_t near_relay_ids[NEXT_MAX_NEAR_RELAYS];
    uint8_t near_relay_rtt[NEXT_MAX_NEAR_RELAYS];
    uint8_t near_relay_jitter[NEXT_MAX_NEAR_RELAYS];
    float near_relay_packet_loss[NEXT_MAX_NEAR_RELAYS];
    uint64_t packets_sent_client_to_server;
    uint64_t packets_lost_server_to_client;
    uint64_t packets_out_of_order_server_to_client;
    float jitter_server_to_client;

    NextClientStatsPacket()
    {
        memset( this, 0, sizeof(NextClientStatsPacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_bool( stream, fallback_to_direct );
        serialize_bool( stream, next );
        serialize_bool( stream, multipath );
        serialize_bool( stream, reported );
        serialize_bool( stream, next_bandwidth_over_limit );
        serialize_int( stream, platform_id, NEXT_PLATFORM_UNKNOWN, NEXT_PLATFORM_MAX );
        serialize_int( stream, connection_type, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_MAX );
        serialize_float( stream, direct_kbps_up );
        serialize_float( stream, direct_kbps_down );
        serialize_float( stream, next_kbps_up );
        serialize_float( stream, next_kbps_down );
        serialize_float( stream, direct_rtt );
        serialize_float( stream, direct_jitter );
        serialize_float( stream, direct_packet_loss );
        serialize_float( stream, direct_max_packet_loss_seen );
        if ( next )
        {
            serialize_float( stream, next_rtt );
            serialize_float( stream, next_jitter );
            serialize_float( stream, next_packet_loss );
        }
        serialize_int( stream, num_near_relays, 0, NEXT_MAX_NEAR_RELAYS );
        bool has_near_relay_pings = false;
        if ( Stream::IsWriting )
        {
            has_near_relay_pings = num_near_relays > 0;
        }
        serialize_bool( stream, has_near_relay_pings );
        if ( has_near_relay_pings )
        {
            for ( int i = 0; i < num_near_relays; ++i )
            {
                serialize_uint64( stream, near_relay_ids[i] );
                serialize_int( stream, near_relay_rtt[i], 0, 255 );
                serialize_int( stream, near_relay_jitter[i], 0, 255 );
                serialize_float( stream, near_relay_packet_loss[i] );
            }
        }
        serialize_uint64( stream, packets_sent_client_to_server );
        serialize_uint64( stream, packets_lost_server_to_client );
        serialize_uint64( stream, packets_out_of_order_server_to_client );
        serialize_float( stream, jitter_server_to_client );
        return true;
    }
};

struct NextRouteUpdatePacket
{
    uint64_t sequence;
    bool multipath;
    bool has_near_relays;
    int num_near_relays;
    uint64_t near_relay_ids[NEXT_MAX_NEAR_RELAYS];
    next_address_t near_relay_addresses[NEXT_MAX_NEAR_RELAYS];
    uint8_t near_relay_ping_tokens[NEXT_MAX_NEAR_RELAYS*NEXT_PING_TOKEN_BYTES];
    uint64_t near_relay_expire_timestamp;
    uint8_t update_type;
    int num_tokens;
    uint8_t tokens[NEXT_MAX_TOKENS*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES];
    uint64_t packets_sent_server_to_client;
    uint64_t packets_lost_client_to_server;
    uint64_t packets_out_of_order_client_to_server;
    float jitter_client_to_server;
    bool has_debug;
    char debug[NEXT_MAX_SESSION_DEBUG];
    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    NextRouteUpdatePacket()
    {
        memset( this, 0, sizeof(NextRouteUpdatePacket ) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, sequence );

        serialize_bool( stream, has_near_relays );
        if ( has_near_relays )
        {
            serialize_int( stream, num_near_relays, 0, NEXT_MAX_NEAR_RELAYS );
            for ( int i = 0; i < num_near_relays; ++i )
            {
                serialize_uint64( stream, near_relay_ids[i] );
                serialize_address( stream, near_relay_addresses[i] );
                serialize_bytes( stream, near_relay_ping_tokens + i * NEXT_PING_TOKEN_BYTES, NEXT_PING_TOKEN_BYTES );
            }
            serialize_uint64( stream, near_relay_expire_timestamp );
        }

        serialize_int( stream, update_type, 0, NEXT_UPDATE_TYPE_CONTINUE );

        if ( update_type != NEXT_UPDATE_TYPE_DIRECT )
        {
            serialize_int( stream, num_tokens, 0, NEXT_MAX_TOKENS );
            serialize_bool( stream, multipath );
        }
        if ( update_type == NEXT_UPDATE_TYPE_ROUTE )
        {
            serialize_bytes( stream, tokens, num_tokens * NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES );
        }
        else if ( update_type == NEXT_UPDATE_TYPE_CONTINUE )
        {
            serialize_bytes( stream, tokens, num_tokens * NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES );
        }

        serialize_uint64( stream, packets_sent_server_to_client );
        serialize_uint64( stream, packets_lost_client_to_server );

        serialize_uint64( stream, packets_out_of_order_client_to_server );

        serialize_float( stream, jitter_client_to_server );

        serialize_bool( stream, has_debug );
        if ( has_debug )
        {
            serialize_string( stream, debug, NEXT_MAX_SESSION_DEBUG );
        }

        serialize_bytes( stream, upcoming_magic, 8 );
        serialize_bytes( stream, current_magic, 8 );
        serialize_bytes( stream, previous_magic, 8 );

        return true;
    }
};

struct NextRouteUpdateAckPacket
{
    uint64_t sequence;

    NextRouteUpdateAckPacket()
    {
        sequence = 0;
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, sequence );
        return true;
    }
};

static void next_generate_pittle( uint8_t * output, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    next_assert( output );
    next_assert( from_address );
    next_assert( from_address_bytes > 0 );
    next_assert( to_address );
    next_assert( to_address_bytes >= 0 );
    next_assert( packet_length > 0 );
#if NEXT_BIG_ENDIAN
    next_bswap( from_port );
    next_bswap( to_port );
    next_bswap( packet_length );
#endif // #if NEXT_BIG_ENDIAN
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
#if NEXT_BIG_ENDIAN
    next_bswap( sum );
#endif // #if NEXT_BIG_ENDIAN
    const char * sum_data = (const char*) &sum;
    output[0] = 1 | ( uint8_t(sum_data[0]) ^ uint8_t(sum_data[1]) ^ 193 );
    output[1] = 1 | ( ( 255 - output[0] ) ^ 113 );
}

static void next_generate_chonkle( uint8_t * output, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    next_assert( output );
    next_assert( magic );
    next_assert( from_address );
    next_assert( from_address_bytes >= 0 );
    next_assert( to_address );
    next_assert( to_address_bytes >= 0 );
    next_assert( packet_length > 0 );
#if NEXT_BIG_ENDIAN
    next_bswap( from_port );
    next_bswap( to_port );
    next_bswap( packet_length );
#endif // #if NEXT_BIG_ENDIAN
    next_fnv_t fnv;
    next_fnv_init( &fnv );
    next_fnv_write( &fnv, magic, 8 );
    next_fnv_write( &fnv, from_address, from_address_bytes );
    next_fnv_write( &fnv, (const uint8_t*) &from_port, 2 );
    next_fnv_write( &fnv, to_address, to_address_bytes );
    next_fnv_write( &fnv, (const uint8_t*) &to_port, 2 );
    next_fnv_write( &fnv, (const uint8_t*) &packet_length, 4 );
    uint64_t hash = next_fnv_finalize( &fnv );
#if NEXT_BIG_ENDIAN
    next_bswap( hash );
#endif // #if NEXT_BIG_ENDIAN
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

bool next_basic_packet_filter( const uint8_t * data, int packet_length )
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

void next_address_data( const next_address_t * address, uint8_t * address_data, int * address_bytes, uint16_t * address_port )
{
    next_assert( address );
    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        address_data[0] = address->data.ipv4[0];
        address_data[1] = address->data.ipv4[1];
        address_data[2] = address->data.ipv4[2];
        address_data[3] = address->data.ipv4[3];
        *address_bytes = 4;
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
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

bool next_advanced_packet_filter( const uint8_t * data, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    if ( data[0] == 0 )
        return true;

    if ( packet_length < 18 )
        return false;
    
    uint8_t a[15];
    uint8_t b[2];

    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    if ( memcmp( a, data + 1, 15 ) != 0 )
        return false;
    if ( memcmp( b, data + packet_length - 2, 2 ) != 0 )
        return false;
    return true;
}

// --------------------------------------------------

int next_write_direct_packet( uint8_t * packet_data, uint8_t open_session_sequence, uint64_t send_sequence, const uint8_t * game_packet_data, int game_packet_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_data );
    next_assert( game_packet_data );
    next_assert( game_packet_bytes >= 0 );
    next_assert( game_packet_bytes <= NEXT_MTU );
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_DIRECT_PACKET );
    uint8_t * a = p; p += 15;
    next_write_uint8( &p, open_session_sequence );
    next_write_uint64( &p, send_sequence );
    next_write_bytes( &p, game_packet_data, game_packet_bytes );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_route_request_packet( uint8_t * packet_data, const uint8_t * token_data, int token_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_ROUTE_REQUEST_PACKET );
    uint8_t * a = p; p += 15;
    next_write_bytes( &p, token_data, token_bytes );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_continue_request_packet( uint8_t * packet_data, const uint8_t * token_data, int token_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_CONTINUE_REQUEST_PACKET );
    uint8_t * a = p; p += 15;
    next_write_bytes( &p, token_data, token_bytes );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_header( uint8_t type, uint64_t sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint8_t * buffer )
{
    next_assert( private_key );
    next_assert( buffer );

    uint8_t * start = buffer;

    (void) start;

    next_write_uint64( &buffer, sequence );

    uint8_t * additional = buffer;
    const int additional_length = 8 + 1;

    next_write_uint64( &buffer, session_id );
    next_write_uint8( &buffer, session_version );

    uint8_t nonce[12];
    {
        uint8_t * p = nonce;
        next_write_uint32( &p, type );
        next_write_uint64( &p, sequence );
    }

    unsigned long long encrypted_length = 0;

    int result = next_crypto_aead_chacha20poly1305_ietf_encrypt( buffer, &encrypted_length,
                                                                 buffer, 0,
                                                                 additional, (unsigned long long) additional_length,
                                                                 NULL, nonce, private_key );

    if ( result != 0 )
        return NEXT_ERROR;

    buffer += encrypted_length;

    next_assert( int( buffer - start ) == NEXT_HEADER_BYTES );

    return NEXT_OK;
}

int next_write_route_response_packet( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_ROUTE_RESPONSE_PACKET );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += NEXT_HEADER_BYTES;
    if ( next_write_header( NEXT_ROUTE_RESPONSE_PACKET, send_sequence, session_id, session_version, private_key, b ) != NEXT_OK )
        return 0;
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_client_to_server_packet( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * game_packet_data, int game_packet_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_data );
    next_assert( private_key );
    next_assert( game_packet_data );
    next_assert( game_packet_bytes >= 0 );
    next_assert( game_packet_bytes <= NEXT_MTU );
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_CLIENT_TO_SERVER_PACKET );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += NEXT_HEADER_BYTES;
    if ( next_write_header( NEXT_CLIENT_TO_SERVER_PACKET, send_sequence, session_id, session_version, private_key, b ) != NEXT_OK )
        return 0;
    next_write_bytes( &p, game_packet_data, game_packet_bytes );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_server_to_client_packet( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * game_packet_data, int game_packet_bytes, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_data );
    next_assert( private_key );
    next_assert( game_packet_data );
    next_assert( game_packet_bytes >= 0 );
    next_assert( game_packet_bytes <= NEXT_MTU );
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_SERVER_TO_CLIENT_PACKET );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += NEXT_HEADER_BYTES;
    if ( next_write_header( NEXT_SERVER_TO_CLIENT_PACKET, send_sequence, session_id, session_version, private_key, b ) != NEXT_OK )
        return 0;
    next_write_bytes( &p, game_packet_data, game_packet_bytes );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_ping_packet( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint64_t ping_sequence, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_data );
    next_assert( private_key );
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_PING_PACKET );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += NEXT_HEADER_BYTES;
    if ( next_write_header( NEXT_PING_PACKET, send_sequence, session_id, session_version, private_key, b ) != NEXT_OK )
        return 0;
    next_write_uint64( &p, ping_sequence );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_pong_packet( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint64_t ping_sequence, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_data );
    next_assert( private_key );
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_PONG_PACKET );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += NEXT_HEADER_BYTES;
    if ( next_write_header( NEXT_PONG_PACKET, send_sequence, session_id, session_version, private_key, b ) != NEXT_OK )
        return 0;
    next_write_uint64( &p, ping_sequence );
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_continue_response_packet( uint8_t * packet_data, uint64_t send_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_CONTINUE_RESPONSE_PACKET );
    uint8_t * a = p; p += 15;
    uint8_t * b = p; p += NEXT_HEADER_BYTES;
    if ( next_write_header( NEXT_CONTINUE_RESPONSE_PACKET, send_sequence, session_id, session_version, private_key, b ) != NEXT_OK )
        return 0;
    uint8_t * c = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( c, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_relay_ping_packet( uint8_t * packet_data, const uint8_t * ping_token, uint64_t ping_sequence, uint64_t session_id, uint64_t expire_timestamp, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_RELAY_PING_PACKET );
    uint8_t * a = p; p += 15;
    next_write_uint64( &p, ping_sequence );
    next_write_uint64( &p, session_id );
    next_write_uint64( &p, expire_timestamp );
    next_write_bytes( &p, ping_token, NEXT_PING_TOKEN_BYTES );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_relay_pong_packet( uint8_t * packet_data, uint64_t ping_sequence, uint64_t session_id, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_RELAY_PONG_PACKET );
    uint8_t * a = p; p += 15;
    next_write_uint64( &p, ping_sequence );
    next_write_uint64( &p, session_id );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    return packet_length;
}

int next_write_packet( uint8_t packet_id, void * packet_object, uint8_t * packet_data, int * packet_bytes, const int * signed_packet, const int * encrypted_packet, uint64_t * sequence, const uint8_t * sign_private_key, const uint8_t * encrypt_private_key, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_object );
    next_assert( packet_data );
    next_assert( packet_bytes );

    next::WriteStream stream( packet_data, NEXT_MAX_PACKET_BYTES );

    typedef next::WriteStream Stream;

    serialize_bits( stream, packet_id, 8 );

    for ( int i = 0; i < 15; ++i )
    {
        uint8_t dummy = 0;
        serialize_bits( stream, dummy, 8 );
    }

    if ( encrypted_packet && encrypted_packet[packet_id] != 0 )
    {
        next_assert( sequence );
        next_assert( encrypt_private_key );
        serialize_uint64( stream, *sequence );
    }

    switch ( packet_id )
    {
        case NEXT_UPGRADE_REQUEST_PACKET:
        {
            NextUpgradeRequestPacket * packet = (NextUpgradeRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write upgrade request packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_UPGRADE_RESPONSE_PACKET:
        {
            NextUpgradeResponsePacket * packet = (NextUpgradeResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write upgrade response packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_UPGRADE_CONFIRM_PACKET:
        {
            NextUpgradeConfirmPacket * packet = (NextUpgradeConfirmPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write upgrade confirm packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_DIRECT_PING_PACKET:
        {
            NextDirectPingPacket * packet = (NextDirectPingPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write direct ping packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_DIRECT_PONG_PACKET:
        {
            NextDirectPongPacket * packet = (NextDirectPongPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write direct pong packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_CLIENT_STATS_PACKET:
        {
            NextClientStatsPacket * packet = (NextClientStatsPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write client stats packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_ROUTE_UPDATE_PACKET:
        {
            NextRouteUpdatePacket * packet = (NextRouteUpdatePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write route update packet" );
                return NEXT_ERROR;
            }
        }
        break;

        case NEXT_ROUTE_UPDATE_ACK_PACKET:
        {
            NextRouteUpdateAckPacket * packet = (NextRouteUpdateAckPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "failed to write route update ack packet" );
                return NEXT_ERROR;
            }
        }
        break;

        default:
            return NEXT_ERROR;
    }

    stream.Flush();

    *packet_bytes = stream.GetBytesProcessed();

    if ( signed_packet && signed_packet[packet_id] )
    {
        next_assert( sign_private_key );
        next_crypto_sign_state_t state;
        next_crypto_sign_init( &state );
        next_crypto_sign_update( &state, packet_data, 1 );
        next_crypto_sign_update( &state, packet_data + 16, *packet_bytes - 16 );
        next_crypto_sign_final_create( &state, packet_data + *packet_bytes, NULL, sign_private_key );
        *packet_bytes += NEXT_CRYPTO_SIGN_BYTES;
    }

    if ( encrypted_packet && encrypted_packet[packet_id] )
    {
        next_assert( !( signed_packet && signed_packet[packet_id] ) );

        uint8_t * additional = packet_data;
        uint8_t * nonce = packet_data + 16;
        uint8_t * message = packet_data + 16 + 8;
        int message_length = *packet_bytes - 16 - 8;

        unsigned long long encrypted_bytes = 0;

        next_crypto_aead_chacha20poly1305_encrypt( message, &encrypted_bytes,
                                                   message, message_length,
                                                   additional, 1,
                                                   NULL, nonce, encrypt_private_key );

        next_assert( encrypted_bytes == uint64_t(message_length) + NEXT_CRYPTO_AEAD_CHACHA20POLY1305_ABYTES );

        *packet_bytes = 1 + 15 + 8 + encrypted_bytes;

        (*sequence)++;
    }

    *packet_bytes += 2;

    int packet_length = *packet_bytes;

    next_generate_chonkle( packet_data + 1, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( packet_data + packet_length - 2, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );

    return NEXT_OK;
}

bool next_is_payload_packet( uint8_t packet_id )
{
    return packet_id == NEXT_CLIENT_TO_SERVER_PACKET ||
           packet_id == NEXT_SERVER_TO_CLIENT_PACKET;
}

int next_read_packet( uint8_t packet_id, uint8_t * packet_data, int begin, int end, void * packet_object, const int * signed_packet, const int * encrypted_packet, uint64_t * sequence, const uint8_t * sign_public_key, const uint8_t * encrypt_private_key, next_replay_protection_t * replay_protection )
{
    next_assert( packet_data );
    next_assert( packet_object );

    next::ReadStream stream( packet_data, end );

    uint8_t * dummy = (uint8_t*) alloca( begin );
    serialize_bytes( stream, dummy, begin );

    if ( signed_packet && signed_packet[packet_id] )
    {
        next_assert( sign_public_key );

        const int packet_bytes = end - begin;

        if ( packet_bytes < int( NEXT_CRYPTO_SIGN_BYTES ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "signed packet is too small to be valid" );
            return NEXT_ERROR;
        }

        next_crypto_sign_state_t state;
        next_crypto_sign_init( &state );
        next_crypto_sign_update( &state, &packet_id, 1 );
        next_crypto_sign_update( &state, packet_data + begin, packet_bytes - NEXT_CRYPTO_SIGN_BYTES );
        if ( next_crypto_sign_final_verify( &state, packet_data + end - NEXT_CRYPTO_SIGN_BYTES, sign_public_key ) != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "signed packet did not verify" );
            return NEXT_ERROR;
        }
    }

    if ( encrypted_packet && encrypted_packet[packet_id] )
    {
        next_assert( !( signed_packet && signed_packet[packet_id] ) );

        next_assert( sequence );
        next_assert( encrypt_private_key );
        next_assert( replay_protection );

        const int packet_bytes = end - begin;

        if ( packet_bytes <= (int) ( 8 + NEXT_CRYPTO_AEAD_CHACHA20POLY1305_ABYTES ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "encrypted packet is too small to be valid" );
            return NEXT_ERROR;
        }

        const uint8_t * p = packet_data + begin;

        *sequence = next_read_uint64( &p );

        uint8_t * nonce = packet_data + begin;
        uint8_t * message = packet_data + begin + 8;
        uint8_t * additional = &packet_id;

        int message_length = end - ( begin + 8 );

        unsigned long long decrypted_bytes;

        if ( next_crypto_aead_chacha20poly1305_decrypt( message, &decrypted_bytes,
                                                        NULL,
                                                        message, message_length,
                                                        additional, 1,
                                                        nonce, encrypt_private_key ) != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "encrypted packet failed to decrypt" );
            return NEXT_ERROR;
        }

        next_assert( decrypted_bytes == uint64_t(message_length) - NEXT_CRYPTO_AEAD_CHACHA20POLY1305_ABYTES );

        serialize_bytes( stream, dummy, 8 );

        if ( next_replay_protection_already_received( replay_protection, *sequence ) )
            return NEXT_ERROR;
    }

    switch ( packet_id )
    {
        case NEXT_UPGRADE_REQUEST_PACKET:
        {
            NextUpgradeRequestPacket * packet = (NextUpgradeRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_UPGRADE_RESPONSE_PACKET:
        {
            NextUpgradeResponsePacket * packet = (NextUpgradeResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_UPGRADE_CONFIRM_PACKET:
        {
            NextUpgradeConfirmPacket * packet = (NextUpgradeConfirmPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_DIRECT_PING_PACKET:
        {
            NextDirectPingPacket * packet = (NextDirectPingPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_DIRECT_PONG_PACKET:
        {
            NextDirectPongPacket * packet = (NextDirectPongPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_CLIENT_STATS_PACKET:
        {
            NextClientStatsPacket * packet = (NextClientStatsPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_ROUTE_UPDATE_PACKET:
        {
            NextRouteUpdatePacket * packet = (NextRouteUpdatePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_ROUTE_UPDATE_ACK_PACKET:
        {
            NextRouteUpdateAckPacket * packet = (NextRouteUpdateAckPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        default:
            return NEXT_ERROR;
    }

    return (int) packet_id;
}

void next_post_validate_packet( uint8_t packet_id, const int * encrypted_packet, uint64_t * sequence, next_replay_protection_t * replay_protection )
{
    const bool payload_packet = next_is_payload_packet( packet_id );

    if ( payload_packet && encrypted_packet && encrypted_packet[packet_id] )
    {
        next_replay_protection_advance_sequence( replay_protection, *sequence );
    }
}

// -------------------------------------------------------------

// *le sigh* ...

void next_copy_string( char * dest, const char * source, size_t dest_size )
{
    next_assert( dest );
    next_assert( source );
    next_assert( dest_size >= 1 );
    memset( dest, 0, dest_size );
    for ( size_t i = 0; i < dest_size - 1; i++ )
    {
        if ( source[i] == '\0' )
            break;
        dest[i] = source[i];
    }
}

// -------------------------------------------------------------

static int next_signed_packets[256];

static int next_encrypted_packets[256];

void * next_global_context = NULL;

struct next_config_internal_t
{
    char server_backend_hostname[256];
    uint64_t client_customer_id;
    uint64_t server_customer_id;
    uint8_t customer_public_key[NEXT_CRYPTO_SIGN_PUBLICKEYBYTES];
    uint8_t customer_private_key[NEXT_CRYPTO_SIGN_SECRETKEYBYTES];
    bool valid_customer_private_key;
    bool valid_customer_public_key;
    int socket_send_buffer_size;
    int socket_receive_buffer_size;
    bool disable_network_next;
    bool disable_autodetect;
};

static next_config_internal_t next_global_config;

void next_default_config( next_config_t * config )
{
    next_assert( config );
    memset( config, 0, sizeof(next_config_t) );
    next_copy_string( config->server_backend_hostname, NEXT_SERVER_BACKEND_HOSTNAME, sizeof(config->server_backend_hostname) );
    config->server_backend_hostname[sizeof(config->server_backend_hostname)-1] = '\0';
    config->socket_send_buffer_size = NEXT_DEFAULT_SOCKET_SEND_BUFFER_SIZE;
    config->socket_receive_buffer_size = NEXT_DEFAULT_SOCKET_RECEIVE_BUFFER_SIZE;
}

int next_init( void * context, next_config_t * config_in )
{
    next_assert( next_global_context == NULL );

    next_global_context = context;

    if ( next_platform_init() != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to initialize platform" );
        return NEXT_ERROR;
    }

    if ( next_crypto_init() == -1 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to initialize sodium" );
        return NEXT_ERROR;
    }

    const char * log_level_override = next_platform_getenv( "NEXT_LOG_LEVEL" );
    if ( log_level_override )
    {
        log_level = atoi( log_level_override );
        next_printf( NEXT_LOG_LEVEL_INFO, "log level overridden to %d", log_level );
    }

    next_config_internal_t config;

    memset( &config, 0, sizeof(next_config_internal_t) );

    config.socket_send_buffer_size = NEXT_DEFAULT_SOCKET_SEND_BUFFER_SIZE;
    config.socket_receive_buffer_size = NEXT_DEFAULT_SOCKET_RECEIVE_BUFFER_SIZE;

    const char * customer_public_key_env = next_platform_getenv( "NEXT_CUSTOMER_PUBLIC_KEY" );
    if ( customer_public_key_env )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "customer public key override: '%s'", customer_public_key_env );
    }

    const char * customer_public_key = customer_public_key_env ? customer_public_key_env : ( config_in ? config_in->customer_public_key : "" );
    if ( customer_public_key )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "customer public key is '%s'", customer_public_key );
        uint8_t decode_buffer[8+NEXT_CRYPTO_SIGN_PUBLICKEYBYTES];
        if ( next_base64_decode_data( customer_public_key, decode_buffer, sizeof(decode_buffer) ) == sizeof(decode_buffer) )
        {
            const uint8_t * p = decode_buffer;
            config.client_customer_id = next_read_uint64( &p );
            memcpy( config.customer_public_key, decode_buffer + 8, NEXT_CRYPTO_SIGN_PUBLICKEYBYTES );
            next_printf( NEXT_LOG_LEVEL_INFO, "found valid customer public key: '%s'", customer_public_key );
            config.valid_customer_public_key = true;
        }
        else
        {
            if ( customer_public_key[0] != '\0' )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "customer public key is invalid: '%s'", customer_public_key );
            }
        }
    }

    const char * customer_private_key_env = next_platform_getenv( "NEXT_CUSTOMER_PRIVATE_KEY" );
    if ( customer_private_key_env )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "customer private key override" );
    }

    const char * customer_private_key = customer_private_key_env ? customer_private_key_env : ( config_in ? config_in->customer_private_key : "" );
    if ( customer_private_key )
    {
        uint8_t decode_buffer[8+NEXT_CRYPTO_SIGN_SECRETKEYBYTES];
        if ( customer_private_key && next_base64_decode_data( customer_private_key, decode_buffer, sizeof(decode_buffer) ) == sizeof(decode_buffer) )
        {
            const uint8_t * p = decode_buffer;
            config.server_customer_id = next_read_uint64( &p );
            memcpy( config.customer_private_key, decode_buffer + 8, NEXT_CRYPTO_SIGN_SECRETKEYBYTES );
            config.valid_customer_private_key = true;
            next_printf( NEXT_LOG_LEVEL_INFO, "found valid customer private key" );
        }
        else
        {
            if ( customer_private_key[0] != '\0' )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "customer private key is invalid" );
            }
        }
    }

    if ( config.valid_customer_private_key && config.valid_customer_public_key && config.client_customer_id != config.server_customer_id )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "mismatch between client and server customer id. please check the private and public keys are part of the same keypair!" );
        config.valid_customer_public_key = false;
        config.valid_customer_private_key = false;
        memset( config.customer_public_key, 0, sizeof(config.customer_public_key) );
        memset( config.customer_private_key, 0, sizeof(config.customer_private_key) );
    }

    next_copy_string( config.server_backend_hostname, config_in ? config_in->server_backend_hostname : NEXT_SERVER_BACKEND_HOSTNAME, sizeof(config.server_backend_hostname) );

    if ( config_in )
    {
        config.socket_send_buffer_size = config_in->socket_send_buffer_size;
        config.socket_receive_buffer_size = config_in->socket_receive_buffer_size;
    }

    config.disable_network_next = config_in ? config_in->disable_network_next : false;

    const char * next_disable_override = next_platform_getenv( "NEXT_DISABLE_NETWORK_NEXT" );
    {
        if ( next_disable_override != NULL )
        {
            int value = atoi( next_disable_override );
            if ( value > 0 )
            {
                config.disable_network_next = true;
            }
        }
    }

    if ( config.disable_network_next )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "network next is disabled" );
    }

    config.disable_autodetect = config_in ? config_in->disable_autodetect : false;

    const char * next_disable_autodetect_override = next_platform_getenv( "NEXT_DISABLE_AUTODETECT" );
    {
        if ( next_disable_autodetect_override != NULL )
        {
            int value = atoi( next_disable_autodetect_override );
            if ( value > 0 )
            {
                config.disable_autodetect = true;
            }
        }
    }

    if ( config.disable_autodetect )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "autodetect is disabled" );
    }

    const char * socket_send_buffer_size_override = next_platform_getenv( "NEXT_SOCKET_SEND_BUFFER_SIZE" );
    if ( socket_send_buffer_size_override != NULL )
    {
        int value = atoi( socket_send_buffer_size_override );
        if ( value > 0 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "override socket send buffer size: %d", value );
            config.socket_send_buffer_size = value;
        }
    }

    const char * socket_receive_buffer_size_override = next_platform_getenv( "NEXT_SOCKET_RECEIVE_BUFFER_SIZE" );
    if ( socket_receive_buffer_size_override != NULL )
    {
        int value = atoi( socket_receive_buffer_size_override );
        if ( value > 0 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "override socket receive buffer size: %d", value );
            config.socket_receive_buffer_size = value;
        }
    }

    const char * next_server_backend_hostname_override = next_platform_getenv( "NEXT_SERVER_BACKEND_HOSTNAME" );

    if ( next_server_backend_hostname_override )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "override server backend hostname: '%s'", next_server_backend_hostname_override );
        next_copy_string( config.server_backend_hostname, next_server_backend_hostname_override, sizeof(config.server_backend_hostname) );
    }

    const char * server_backend_public_key_env = next_platform_getenv( "NEXT_SERVER_BACKEND_PUBLIC_KEY" );
    if ( server_backend_public_key_env )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server backend public key override: %s", server_backend_public_key_env );

        if ( next_base64_decode_data( server_backend_public_key_env, next_server_backend_public_key, NEXT_CRYPTO_SIGN_PUBLICKEYBYTES ) == NEXT_CRYPTO_SIGN_PUBLICKEYBYTES )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "valid server backend public key" );
        }
        else
        {
            if ( server_backend_public_key_env[0] != '\0' )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "server backend public key is invalid: \"%s\"", server_backend_public_key_env );
            }
        }
    }

    const char * router_public_key_env = next_platform_getenv( "NEXT_ROUTER_PUBLIC_KEY" );
    if ( router_public_key_env )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "router public key override: %s", router_public_key_env );

        if ( next_base64_decode_data( router_public_key_env, next_router_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES ) == NEXT_CRYPTO_BOX_PUBLICKEYBYTES )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "valid router public key" );
        }
        else
        {
            if ( router_public_key_env[0] != '\0' )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "router public key is invalid: \"%s\"", router_public_key_env );
            }
        }
    }

    next_global_config = config;

    next_signed_packets[NEXT_UPGRADE_REQUEST_PACKET] = 1;
    next_signed_packets[NEXT_UPGRADE_CONFIRM_PACKET] = 1;

    next_signed_packets[NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_SERVER_UPDATE_REQUEST_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_SERVER_UPDATE_RESPONSE_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_SESSION_UPDATE_REQUEST_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_SESSION_UPDATE_RESPONSE_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_MATCH_DATA_REQUEST_PACKET] = 1;
    next_signed_packets[NEXT_BACKEND_MATCH_DATA_RESPONSE_PACKET] = 1;

    next_encrypted_packets[NEXT_DIRECT_PING_PACKET] = 1;
    next_encrypted_packets[NEXT_DIRECT_PONG_PACKET] = 1;
    next_encrypted_packets[NEXT_CLIENT_STATS_PACKET] = 1;
    next_encrypted_packets[NEXT_ROUTE_UPDATE_PACKET] = 1;
    next_encrypted_packets[NEXT_ROUTE_UPDATE_ACK_PACKET] = 1;

    return NEXT_OK;
}

void next_term()
{
    next_platform_term();

    next_global_context = NULL;
}

// ---------------------------------------------------------------

struct next_relay_stats_t
{
    NEXT_DECLARE_SENTINEL(0)

    bool has_pings;
    int num_relays;

    NEXT_DECLARE_SENTINEL(1)

    uint64_t relay_ids[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(2)

    float relay_rtt[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(3)

    float relay_jitter[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(4)

    float relay_packet_loss[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(5)
};

void next_relay_stats_initialize_sentinels( next_relay_stats_t * stats )
{
    (void) stats;
    next_assert( stats );
    NEXT_INITIALIZE_SENTINEL( stats, 0 )
    NEXT_INITIALIZE_SENTINEL( stats, 1 )
    NEXT_INITIALIZE_SENTINEL( stats, 2 )
    NEXT_INITIALIZE_SENTINEL( stats, 3 )
    NEXT_INITIALIZE_SENTINEL( stats, 4 )
    NEXT_INITIALIZE_SENTINEL( stats, 5 )
}

void next_relay_stats_verify_sentinels( next_relay_stats_t * stats )
{
    (void) stats;
    next_assert( stats );
    NEXT_VERIFY_SENTINEL( stats, 0 )
    NEXT_VERIFY_SENTINEL( stats, 1 )
    NEXT_VERIFY_SENTINEL( stats, 2 )
    NEXT_VERIFY_SENTINEL( stats, 3 )
    NEXT_VERIFY_SENTINEL( stats, 4 )
    NEXT_VERIFY_SENTINEL( stats, 5 )
}

// ---------------------------------------------------------------

struct next_relay_manager_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int num_relays;

    NEXT_DECLARE_SENTINEL(1)

    uint64_t relay_ids[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(2)

    double relay_last_ping_time[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(3)

    next_address_t relay_addresses[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(4)

    uint8_t relay_ping_tokens[NEXT_MAX_NEAR_RELAYS * NEXT_PING_TOKEN_BYTES];

    NEXT_DECLARE_SENTINEL(5)

    uint64_t relay_ping_expire_timestamp;

    NEXT_DECLARE_SENTINEL(6)

    next_ping_history_t relay_ping_history[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(7)
};

void next_relay_manager_initialize_sentinels( next_relay_manager_t * manager )
{
    (void) manager;

    next_assert( manager );

    NEXT_INITIALIZE_SENTINEL( manager, 0 )
    NEXT_INITIALIZE_SENTINEL( manager, 1 )
    NEXT_INITIALIZE_SENTINEL( manager, 2 )
    NEXT_INITIALIZE_SENTINEL( manager, 3 )
    NEXT_INITIALIZE_SENTINEL( manager, 4 )
    NEXT_INITIALIZE_SENTINEL( manager, 5 )
    NEXT_INITIALIZE_SENTINEL( manager, 6 )
    NEXT_INITIALIZE_SENTINEL( manager, 7 )

    for ( int i = 0; i < NEXT_MAX_NEAR_RELAYS; ++i )
        next_ping_history_initialize_sentinels( &manager->relay_ping_history[i] );
}

void next_relay_manager_verify_sentinels( next_relay_manager_t * manager )
{
    (void) manager;
#if NEXT_ENABLE_MEMORY_CHECKS
    next_assert( manager );
    NEXT_VERIFY_SENTINEL( manager, 0 )
    NEXT_VERIFY_SENTINEL( manager, 1 )
    NEXT_VERIFY_SENTINEL( manager, 2 )
    NEXT_VERIFY_SENTINEL( manager, 3 )
    NEXT_VERIFY_SENTINEL( manager, 4 )
    NEXT_VERIFY_SENTINEL( manager, 5 )
    NEXT_VERIFY_SENTINEL( manager, 6 )
    NEXT_VERIFY_SENTINEL( manager, 7 )
    for ( int i = 0; i < NEXT_MAX_NEAR_RELAYS; ++i )
        next_ping_history_verify_sentinels( &manager->relay_ping_history[i] );
#endif // #if NEXT_ENABLE_MEMORY_CHECKS
}

void next_relay_manager_reset( next_relay_manager_t * manager );

next_relay_manager_t * next_relay_manager_create( void * context )
{
    next_relay_manager_t * manager = (next_relay_manager_t*) next_malloc( context, sizeof(next_relay_manager_t) );
    if ( !manager )
        return NULL;

    memset( manager, 0, sizeof(next_relay_manager_t) );

    manager->context = context;

    next_relay_manager_initialize_sentinels( manager );

    next_relay_manager_reset( manager );

    next_relay_manager_verify_sentinels( manager );

    return manager;
}

void next_relay_manager_reset( next_relay_manager_t * manager )
{
    next_relay_manager_verify_sentinels( manager );

    manager->num_relays = 0;

    memset( manager->relay_ids, 0, sizeof(manager->relay_ids) );
    memset( manager->relay_last_ping_time, 0, sizeof(manager->relay_last_ping_time) );
    memset( manager->relay_addresses, 0, sizeof(manager->relay_addresses) );
    memset( manager->relay_ping_tokens, 0, sizeof(manager->relay_ping_tokens) );
    manager->relay_ping_expire_timestamp = 0;

    for ( int i = 0; i < NEXT_MAX_NEAR_RELAYS; ++i )
    {
        next_ping_history_clear( &manager->relay_ping_history[i] );
    }
}

void next_relay_manager_update( next_relay_manager_t * manager, int num_relays, const uint64_t * relay_ids, const next_address_t * relay_addresses, const uint8_t * relay_ping_tokens, uint64_t relay_ping_expire_timestamp )
{
    next_relay_manager_verify_sentinels( manager );

    next_assert( num_relays >= 0 );
    next_assert( num_relays <= NEXT_MAX_NEAR_RELAYS );
    next_assert( relay_ids );
    next_assert( relay_addresses );

    // reset relay manager

    next_relay_manager_reset( manager );

    // copy across all relay data

    manager->num_relays = num_relays;

    for ( int i = 0; i < num_relays; ++i )
    {
        manager->relay_ids[i] = relay_ids[i];
        manager->relay_addresses[i] = relay_addresses[i];
    }

    memcpy( manager->relay_ping_tokens, relay_ping_tokens, num_relays * NEXT_PING_TOKEN_BYTES );

    manager->relay_ping_expire_timestamp = relay_ping_expire_timestamp;

    // make sure all ping times are evenly distributed to avoid clusters of ping packets

    double current_time = next_platform_time();

    const double ping_time = 1.0 / NEXT_PING_RATE;

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        manager->relay_last_ping_time[i] = current_time - ping_time + i * ping_time / manager->num_relays;
    }

    next_relay_manager_verify_sentinels( manager );
}

void next_relay_manager_send_pings( next_relay_manager_t * manager, next_platform_socket_t * socket, uint64_t session_id, const uint8_t * magic, const next_address_t * client_external_address )
{
    next_relay_manager_verify_sentinels( manager );

    next_assert( socket );

    uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

    double current_time = next_platform_time();

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        const double ping_time = 1.0 / NEXT_PING_RATE;

        if ( manager->relay_last_ping_time[i] + ping_time <= current_time )
        {
            uint64_t ping_sequence = next_ping_history_ping_sent( &manager->relay_ping_history[i], next_platform_time() );

            const uint8_t * ping_token = manager->relay_ping_tokens + i * NEXT_PING_TOKEN_BYTES;

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( client_external_address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &manager->relay_addresses[i], to_address_data, &to_address_bytes, &to_address_port );

            int packet_bytes = next_write_relay_ping_packet( packet_data, ping_token, ping_sequence, session_id, manager->relay_ping_expire_timestamp, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

            next_assert( packet_bytes > 0 );

            next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
            next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

#if NEXT_SPIKE_TRACKING
            double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

            next_platform_socket_send_packet( socket, &manager->relay_addresses[i], packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
            double finish_time = next_platform_time();
            if ( finish_time - start_time > 0.001 )
            {
                next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
            }
#endif // #if NEXT_SPIKE_TRACKING

            manager->relay_last_ping_time[i] = current_time;
        }
    }
}

void next_relay_manager_process_pong( next_relay_manager_t * manager, const next_address_t * from, uint64_t sequence )
{
    next_relay_manager_verify_sentinels( manager );

    next_assert( from );

    for ( int i = 0; i < manager->num_relays; ++i )
    {
        if ( next_address_equal( from, &manager->relay_addresses[i] ) )
        {
            next_ping_history_pong_received( &manager->relay_ping_history[i], sequence, next_platform_time() );
            return;
        }
    }
}

void next_relay_manager_get_stats( next_relay_manager_t * manager, next_relay_stats_t * stats )
{
    next_relay_manager_verify_sentinels( manager );

    next_assert( stats );

    double current_time = next_platform_time();

    next_relay_stats_initialize_sentinels( stats );

    stats->num_relays = manager->num_relays;
    stats->has_pings = stats->num_relays > 0;

    for ( int i = 0; i < stats->num_relays; ++i )
    {
        next_route_stats_t route_stats;

        next_route_stats_from_ping_history( &manager->relay_ping_history[i], current_time - NEXT_CLIENT_STATS_WINDOW, current_time, &route_stats );

        stats->relay_ids[i] = manager->relay_ids[i];
        stats->relay_rtt[i] = route_stats.rtt;
        stats->relay_jitter[i] = route_stats.jitter;
        stats->relay_packet_loss[i] = route_stats.packet_loss;
    }

    next_relay_stats_verify_sentinels( stats );
}

void next_relay_manager_destroy( next_relay_manager_t * manager )
{
    next_relay_manager_verify_sentinels( manager );

    next_free( manager->context, manager );
}

// ----------------------------------------------------------------------

void next_peek_header( uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, const uint8_t * buffer, int buffer_length )
{
    uint64_t packet_sequence;

    next_assert( buffer );
    next_assert( buffer_length >= NEXT_HEADER_BYTES );

    packet_sequence = next_read_uint64( &buffer );

    *sequence = packet_sequence;
    *session_id = next_read_uint64( &buffer );
    *session_version = next_read_uint8( &buffer );
}

int next_read_header( int packet_type, uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, const uint8_t * private_key, uint8_t * buffer, int buffer_length )
{
    next_assert( private_key );
    next_assert( buffer );

    if ( buffer_length < NEXT_HEADER_BYTES )
    {
        return NEXT_ERROR;
    }

    const uint8_t * p = buffer;

    uint64_t packet_sequence = next_read_uint64( &p );

    const uint8_t * additional = p;

    const int additional_length = 8 + 1;

    uint64_t packet_session_id = next_read_uint64( &p );
    uint8_t packet_session_version = next_read_uint8( &p );

    uint8_t nonce[12];
    {
        uint8_t * q = nonce;
        next_write_uint32( &q, packet_type );
        next_write_uint64( &q, packet_sequence );
    }

    unsigned long long decrypted_length;

    int result = next_crypto_aead_chacha20poly1305_ietf_decrypt( buffer + 17, &decrypted_length, NULL,
                                                                 buffer + 17, (unsigned long long) NEXT_CRYPTO_AEAD_CHACHA20POLY1305_IETF_ABYTES,
                                                                 additional, (unsigned long long) additional_length,
                                                                 nonce, private_key );

    if ( result != 0 )
    {
        return NEXT_ERROR;
    }

    *sequence = packet_sequence;
    *session_id = packet_session_id;
    *session_version = packet_session_version;

    return NEXT_OK;
}

// ---------------------------------------------------------------

struct next_route_data_t
{
    NEXT_DECLARE_SENTINEL(0)

    bool current_route;
    double current_route_expire_time;
    uint64_t current_route_session_id;
    uint8_t current_route_session_version;
    int current_route_kbps_up;
    int current_route_kbps_down;
    next_address_t current_route_next_address;

    NEXT_DECLARE_SENTINEL(1)

    uint8_t current_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(2)

    bool previous_route;
    uint64_t previous_route_session_id;
    uint8_t previous_route_session_version;

    NEXT_DECLARE_SENTINEL(3)

    uint8_t previous_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(4)

    bool pending_route;
    double pending_route_start_time;
    double pending_route_last_send_time;
    uint64_t pending_route_session_id;
    uint8_t pending_route_session_version;
    int pending_route_kbps_up;
    int pending_route_kbps_down;
    int pending_route_request_packet_bytes;
    next_address_t pending_route_next_address;

    NEXT_DECLARE_SENTINEL(5)

    uint8_t pending_route_request_packet_data[NEXT_MAX_PACKET_BYTES];

    NEXT_DECLARE_SENTINEL(6)

    uint8_t pending_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(7)

    bool pending_continue;
    double pending_continue_start_time;
    double pending_continue_last_send_time;
    int pending_continue_request_packet_bytes;

    NEXT_DECLARE_SENTINEL(8)

    uint8_t pending_continue_request_packet_data[NEXT_MAX_PACKET_BYTES];

    NEXT_DECLARE_SENTINEL(9)
};

void next_route_data_initialize_sentinels( next_route_data_t * route_data )
{
    (void) route_data;
    next_assert( route_data );
    NEXT_INITIALIZE_SENTINEL( route_data, 0 )
    NEXT_INITIALIZE_SENTINEL( route_data, 1 )
    NEXT_INITIALIZE_SENTINEL( route_data, 2 )
    NEXT_INITIALIZE_SENTINEL( route_data, 3 )
    NEXT_INITIALIZE_SENTINEL( route_data, 4 )
    NEXT_INITIALIZE_SENTINEL( route_data, 5 )
    NEXT_INITIALIZE_SENTINEL( route_data, 6 )
    NEXT_INITIALIZE_SENTINEL( route_data, 7 )
    NEXT_INITIALIZE_SENTINEL( route_data, 8 )
    NEXT_INITIALIZE_SENTINEL( route_data, 9 )
}

void next_route_data_verify_sentinels( next_route_data_t * route_data )
{
    (void) route_data;
    next_assert( route_data );
    NEXT_VERIFY_SENTINEL( route_data, 0 )
    NEXT_VERIFY_SENTINEL( route_data, 1 )
    NEXT_VERIFY_SENTINEL( route_data, 2 )
    NEXT_VERIFY_SENTINEL( route_data, 3 )
    NEXT_VERIFY_SENTINEL( route_data, 4 )
    NEXT_VERIFY_SENTINEL( route_data, 5 )
    NEXT_VERIFY_SENTINEL( route_data, 6 )
    NEXT_VERIFY_SENTINEL( route_data, 7 )
    NEXT_VERIFY_SENTINEL( route_data, 8 )
    NEXT_VERIFY_SENTINEL( route_data, 9 )
}

struct next_route_manager_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    uint64_t send_sequence;
    bool fallback_to_direct;
    next_route_data_t route_data;
    double last_route_update_time;
    uint32_t flags;

    NEXT_DECLARE_SENTINEL(1)
};

void next_route_manager_initialize_sentinels( next_route_manager_t * route_manager )
{
    (void) route_manager;
    next_assert( route_manager );
    NEXT_INITIALIZE_SENTINEL( route_manager, 0 )
    NEXT_INITIALIZE_SENTINEL( route_manager, 1 )
    next_route_data_initialize_sentinels( &route_manager->route_data );
}

void next_route_manager_verify_sentinels( next_route_manager_t * route_manager )
{
    (void) route_manager;
    next_assert( route_manager );
    NEXT_VERIFY_SENTINEL( route_manager, 0 )
    NEXT_VERIFY_SENTINEL( route_manager, 1 )
    next_route_data_verify_sentinels( &route_manager->route_data );
}

next_route_manager_t * next_route_manager_create( void * context )
{
    next_route_manager_t * route_manager = (next_route_manager_t*) next_malloc( context, sizeof(next_route_manager_t) );
    if ( !route_manager )
        return NULL;
    memset( route_manager, 0, sizeof(next_route_manager_t) );
    next_route_manager_initialize_sentinels( route_manager );
    route_manager->context = context;
    return route_manager;
}

void next_route_manager_reset( next_route_manager_t * route_manager )
{
    next_route_manager_verify_sentinels( route_manager );

    route_manager->send_sequence = 0;
    route_manager->fallback_to_direct = false;
    route_manager->last_route_update_time = 0.0;

    memset( &route_manager->route_data, 0, sizeof(next_route_data_t) );

    next_route_manager_initialize_sentinels( route_manager );

    route_manager->flags = 0;

    next_route_manager_verify_sentinels( route_manager );
}

void next_route_manager_fallback_to_direct( next_route_manager_t * route_manager, uint32_t flags )
{
    next_route_manager_verify_sentinels( route_manager );

    route_manager->flags |= flags;

    if ( route_manager->fallback_to_direct )
        return;

    route_manager->fallback_to_direct = true;

    next_printf( NEXT_LOG_LEVEL_INFO, "client fallback to direct" );

    route_manager->route_data.previous_route = route_manager->route_data.current_route;
    route_manager->route_data.previous_route_session_id = route_manager->route_data.current_route_session_id;
    route_manager->route_data.previous_route_session_version = route_manager->route_data.current_route_session_version;
    memcpy( route_manager->route_data.previous_route_private_key, route_manager->route_data.current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );

    route_manager->route_data.current_route = false;
}

void next_route_manager_direct_route( next_route_manager_t * route_manager, bool quiet )
{
    next_route_manager_verify_sentinels( route_manager );

    if ( route_manager->fallback_to_direct )
        return;

    if ( !quiet )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "client direct route" );
    }

    route_manager->route_data.previous_route = route_manager->route_data.current_route;
    route_manager->route_data.previous_route_session_id = route_manager->route_data.current_route_session_id;
    route_manager->route_data.previous_route_session_version = route_manager->route_data.current_route_session_version;
    memcpy( route_manager->route_data.previous_route_private_key, route_manager->route_data.current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );

    route_manager->route_data.current_route = false;
}

void next_route_manager_begin_next_route( next_route_manager_t * route_manager, int num_tokens, uint8_t * tokens, const uint8_t * public_key, const uint8_t * private_key, const uint8_t * magic, const next_address_t * client_external_address )
{
    next_route_manager_verify_sentinels( route_manager );

    next_assert( tokens );
    next_assert( num_tokens >= 2 );
    next_assert( num_tokens <= NEXT_MAX_TOKENS );

    if ( route_manager->fallback_to_direct )
        return;

    uint8_t * p = tokens;
    next_route_token_t route_token;
    if ( next_read_encrypted_route_token( &p, &route_token, public_key, private_key ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client received bad route token" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_BAD_ROUTE_TOKEN );
        return;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "client next route" );

    route_manager->route_data.pending_route = true;
    route_manager->route_data.pending_route_start_time = next_platform_time();
    route_manager->route_data.pending_route_last_send_time = -1000.0;
    route_manager->route_data.pending_route_next_address = route_token.next_address;
    route_manager->route_data.pending_route_session_id = route_token.session_id;
    route_manager->route_data.pending_route_session_version = route_token.session_version;
    route_manager->route_data.pending_route_kbps_up = route_token.kbps_up;
    route_manager->route_data.pending_route_kbps_down = route_token.kbps_down;

    memcpy( route_manager->route_data.pending_route_request_packet_data + 1, tokens + NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES, ( size_t(num_tokens) - 1 ) * NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES );
    memcpy( route_manager->route_data.pending_route_private_key, route_token.private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );

    const uint8_t * token_data = tokens + NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES;
    const int token_bytes = ( num_tokens - 1 ) * NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES;

    uint8_t from_address_data[32];
    uint8_t to_address_data[32];
    uint16_t from_address_port;
    uint16_t to_address_port;
    int from_address_bytes;
    int to_address_bytes;

    next_address_data( client_external_address, from_address_data, &from_address_bytes, &from_address_port );
    next_address_data( &route_token.next_address, to_address_data, &to_address_bytes, &to_address_port );

    route_manager->route_data.pending_route_request_packet_bytes = next_write_route_request_packet( route_manager->route_data.pending_route_request_packet_data, token_data, token_bytes, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

    next_assert( route_manager->route_data.pending_route_request_packet_bytes > 0 );
    next_assert( route_manager->route_data.pending_route_request_packet_bytes <= NEXT_MAX_PACKET_BYTES );

    const uint8_t * packet_data = route_manager->route_data.pending_route_request_packet_data;
    const int packet_bytes = route_manager->route_data.pending_route_request_packet_bytes;

    next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
    next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

    (void) packet_data;
    (void) packet_bytes;
}

void next_route_manager_continue_next_route( next_route_manager_t * route_manager, int num_tokens, uint8_t * tokens, const uint8_t * public_key, const uint8_t * private_key, const uint8_t * magic, const next_address_t * client_external_address )
{
    next_route_manager_verify_sentinels( route_manager );

    next_assert( tokens );
    next_assert( num_tokens >= 2 );
    next_assert( num_tokens <= NEXT_MAX_TOKENS );

    if ( route_manager->fallback_to_direct )
        return;

    if ( !route_manager->route_data.current_route )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client has no route to continue" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_NO_ROUTE_TO_CONTINUE );
        return;
    }

    if ( route_manager->route_data.pending_route || route_manager->route_data.pending_continue )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client previous update still pending" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_PREVIOUS_UPDATE_STILL_PENDING );
        return;
    }

    uint8_t * p = tokens;
    next_continue_token_t continue_token;
    if ( next_read_encrypted_continue_token( &p, &continue_token, public_key, private_key ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client received bad continue token" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_BAD_CONTINUE_TOKEN );
        return;
    }

    route_manager->route_data.pending_continue = true;
    route_manager->route_data.pending_continue_start_time = next_platform_time();
    route_manager->route_data.pending_continue_last_send_time = -1000.0;

    uint8_t from_address_data[32];
    uint8_t to_address_data[32];
    uint16_t from_address_port;
    uint16_t to_address_port;
    int from_address_bytes;
    int to_address_bytes;

    next_address_data( client_external_address, from_address_data, &from_address_bytes, &from_address_port );
    next_address_data( &route_manager->route_data.current_route_next_address, to_address_data, &to_address_bytes, &to_address_port );

    const uint8_t * token_data = tokens + NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES;
    const int token_bytes = ( num_tokens - 1 ) * NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES;

    route_manager->route_data.pending_continue_request_packet_bytes = next_write_continue_request_packet( route_manager->route_data.pending_continue_request_packet_data, token_data, token_bytes, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

    next_assert( route_manager->route_data.pending_continue_request_packet_bytes >= 0 );
    next_assert( route_manager->route_data.pending_continue_request_packet_bytes <= NEXT_MAX_PACKET_BYTES );

    const uint8_t * packet_data = route_manager->route_data.pending_continue_request_packet_data;
    const int packet_bytes = route_manager->route_data.pending_continue_request_packet_bytes;

    next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
    next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

    (void) packet_data;
    (void) packet_bytes;

    next_printf( NEXT_LOG_LEVEL_INFO, "client continues route" );
}

void next_route_manager_update( next_route_manager_t * route_manager, int update_type, int num_tokens, uint8_t * tokens, const uint8_t * public_key, const uint8_t * private_key, const uint8_t * magic, const next_address_t * client_external_address )
{
    next_route_manager_verify_sentinels( route_manager );

    next_assert( public_key );
    next_assert( private_key );

    if ( update_type == NEXT_UPDATE_TYPE_DIRECT )
    {
        next_route_manager_direct_route( route_manager, false );
    }
    else if ( update_type == NEXT_UPDATE_TYPE_ROUTE )
    {
        next_route_manager_begin_next_route( route_manager, num_tokens, tokens, public_key, private_key, magic, client_external_address );
    }
    else if ( update_type == NEXT_UPDATE_TYPE_CONTINUE )
    {
        next_route_manager_continue_next_route( route_manager, num_tokens, tokens, public_key, private_key, magic, client_external_address );
    }
}

bool next_route_manager_has_network_next_route( next_route_manager_t * route_manager )
{
    next_route_manager_verify_sentinels( route_manager );
    return route_manager->route_data.current_route;
}

uint64_t next_route_manager_next_send_sequence( next_route_manager_t * route_manager )
{
    next_route_manager_verify_sentinels( route_manager );
    return route_manager->send_sequence++;
}

bool next_route_manager_prepare_send_packet( next_route_manager_t * route_manager, uint64_t sequence, next_address_t * to, const uint8_t * payload_data, int payload_bytes, uint8_t * packet_data, int * packet_bytes, const uint8_t * magic, const next_address_t * client_external_address )
{
    next_route_manager_verify_sentinels( route_manager );

    if ( !route_manager->route_data.current_route )
        return false;

    next_assert( route_manager->route_data.current_route );
    next_assert( to );
    next_assert( payload_data );
    next_assert( payload_bytes );
    next_assert( packet_data );
    next_assert( packet_bytes );

    *to = route_manager->route_data.current_route_next_address;

    uint8_t from_address_data[32];
    uint8_t to_address_data[32];
    uint16_t from_address_port;
    uint16_t to_address_port;
    int from_address_bytes;
    int to_address_bytes;

    next_address_data( client_external_address, from_address_data, &from_address_bytes, &from_address_port );
    next_address_data( to, to_address_data, &to_address_bytes, &to_address_port );

    *packet_bytes = next_write_client_to_server_packet( packet_data, sequence, route_manager->route_data.current_route_session_id, route_manager->route_data.current_route_session_version, route_manager->route_data.current_route_private_key, payload_data, payload_bytes, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

    if ( *packet_bytes == 0 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client failed to write client to server packet header" );
        return false;
    }

    next_assert( next_basic_packet_filter( packet_data, *packet_bytes ) );
    next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, *packet_bytes ) );

    next_assert( *packet_bytes < NEXT_MAX_PACKET_BYTES );

    return true;
}

bool next_route_manager_process_server_to_client_packet( next_route_manager_t * route_manager, uint8_t packet_type, uint8_t * packet_data, int packet_bytes, uint64_t * payload_sequence )
{
    next_route_manager_verify_sentinels( route_manager );

    next_assert( packet_data );
    next_assert( payload_sequence );

    uint64_t packet_sequence = 0;
    uint64_t packet_session_id = 0;
    uint8_t packet_session_version = 0;

    bool from_current_route = true;

    if ( next_read_header( packet_type, &packet_sequence, &packet_session_id, &packet_session_version, route_manager->route_data.current_route_private_key, packet_data, packet_bytes ) != NEXT_OK )
    {
        from_current_route = false;
        if ( next_read_header( packet_type, &packet_sequence, &packet_session_id, &packet_session_version, route_manager->route_data.previous_route_private_key, packet_data, packet_bytes ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. could not read header" );
            return false;
        }
    }

    if ( !route_manager->route_data.current_route && !route_manager->route_data.previous_route )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. no current or previous route" );
        return false;
    }

    if ( from_current_route )
    {
        if ( packet_session_id != route_manager->route_data.current_route_session_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. session id mismatch (current route)" );
            return false;
        }

        if ( packet_session_version != route_manager->route_data.current_route_session_version )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. session version mismatch (current route)" );
            return false;
        }
    }
    else
    {
        if ( packet_session_id != route_manager->route_data.previous_route_session_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. session id mismatch (previous route)" );
            return false;
        }

        if ( packet_session_version != route_manager->route_data.previous_route_session_version )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. session version mismatch (previous route)" );
            return false;
        }
    }

    *payload_sequence = packet_sequence;

    int payload_bytes = packet_bytes - NEXT_HEADER_BYTES;

    if ( payload_bytes > NEXT_MTU )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. too large (%d>%d)", payload_bytes, NEXT_MTU );
        return false;
    }

    (void) payload_bytes;

    return true;
}

void next_route_manager_check_for_timeouts( next_route_manager_t * route_manager )
{
    next_route_manager_verify_sentinels( route_manager );

    if ( route_manager->fallback_to_direct )
        return;

    const double current_time = next_platform_time();

    if ( route_manager->last_route_update_time > 0.0 && route_manager->last_route_update_time + NEXT_CLIENT_ROUTE_TIMEOUT < current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client route timed out" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_ROUTE_TIMED_OUT );
        return;
    }

    if ( route_manager->route_data.current_route && route_manager->route_data.current_route_expire_time <= current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client route expired" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_ROUTE_EXPIRED );
        return;
    }

    if ( route_manager->route_data.pending_route && route_manager->route_data.pending_route_start_time + NEXT_ROUTE_REQUEST_TIMEOUT <= current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client route request timed out" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_ROUTE_REQUEST_TIMED_OUT );
        return;
    }

    if ( route_manager->route_data.pending_continue && route_manager->route_data.pending_continue_start_time + NEXT_CONTINUE_REQUEST_TIMEOUT <= current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client continue request timed out" );
        next_route_manager_fallback_to_direct( route_manager, NEXT_FLAGS_CONTINUE_REQUEST_TIMED_OUT );
        return;
    }
}

bool next_route_manager_send_route_request( next_route_manager_t * route_manager, next_address_t * to, uint8_t * packet_data, int * packet_bytes )
{
    next_route_manager_verify_sentinels( route_manager );

    next_assert( to );
    next_assert( packet_data );
    next_assert( packet_bytes );

    if ( route_manager->fallback_to_direct )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client not sending route request. fallback to direct" );
        return false;
    }

    if ( !route_manager->route_data.pending_route )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client not sending route request. pending route" );
        return false;
    }

    double current_time = next_platform_time();

    if ( route_manager->route_data.pending_route_last_send_time + NEXT_ROUTE_REQUEST_SEND_TIME > current_time )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client not sending route request. not yet" );
        return false;
    }

    *to = route_manager->route_data.pending_route_next_address;
    route_manager->route_data.pending_route_last_send_time = current_time;
    *packet_bytes = route_manager->route_data.pending_route_request_packet_bytes;
    memcpy( packet_data, route_manager->route_data.pending_route_request_packet_data, route_manager->route_data.pending_route_request_packet_bytes );

    return true;
}

bool next_route_manager_send_continue_request( next_route_manager_t * route_manager, next_address_t * to, uint8_t * packet_data, int * packet_bytes )
{
    next_route_manager_verify_sentinels( route_manager );

    next_assert( to );
    next_assert( packet_data );
    next_assert( packet_bytes );

    if ( route_manager->fallback_to_direct )
        return false;

    if ( !route_manager->route_data.current_route || !route_manager->route_data.pending_continue )
        return false;

    double current_time = next_platform_time();

    if ( route_manager->route_data.pending_continue_last_send_time + NEXT_CONTINUE_REQUEST_SEND_TIME > current_time )
        return false;

    *to = route_manager->route_data.current_route_next_address;
    route_manager->route_data.pending_continue_last_send_time = current_time;
    *packet_bytes = route_manager->route_data.pending_continue_request_packet_bytes;
    memcpy( packet_data, route_manager->route_data.pending_continue_request_packet_data, route_manager->route_data.pending_continue_request_packet_bytes );

    return true;
}

void next_route_manager_destroy( next_route_manager_t * route_manager )
{
    next_route_manager_verify_sentinels( route_manager );

    next_free( route_manager->context, route_manager );
}

// ---------------------------------------------------------------

#define NEXT_CLIENT_COMMAND_OPEN_SESSION            0
#define NEXT_CLIENT_COMMAND_CLOSE_SESSION           1
#define NEXT_CLIENT_COMMAND_DESTROY                 2
#define NEXT_CLIENT_COMMAND_REPORT_SESSION          3

struct next_client_command_t
{
    int type;
};

struct next_client_command_open_session_t : public next_client_command_t
{
    next_address_t server_address;
};

struct next_client_command_close_session_t : public next_client_command_t
{
    // ...
};

struct next_client_command_destroy_t : public next_client_command_t
{
    // ...
};

struct next_client_command_report_session_t : public next_client_command_t
{
    // ...
};

// ---------------------------------------------------------------

#define NEXT_CLIENT_NOTIFY_PACKET_RECEIVED          0
#define NEXT_CLIENT_NOTIFY_UPGRADED                 1
#define NEXT_CLIENT_NOTIFY_STATS_UPDATED            2
#define NEXT_CLIENT_NOTIFY_MAGIC_UPDATED            3
#define NEXT_CLIENT_NOTIFY_READY                    4

struct next_client_notify_t
{
    int type;
};

struct next_client_notify_packet_received_t : public next_client_notify_t
{
    bool direct;
    int payload_bytes;
    uint8_t payload_data[NEXT_MAX_PACKET_BYTES-1];
};

struct next_client_notify_upgraded_t : public next_client_notify_t
{
    uint64_t session_id;
    next_address_t client_external_address;
    uint8_t current_magic[8];
};

struct next_client_notify_stats_updated_t : public next_client_notify_t
{
    next_client_stats_t stats;
    bool fallback_to_direct;
};

struct next_client_notify_magic_updated_t : public next_client_notify_t
{
    uint8_t current_magic[8];
};

struct next_client_notify_ready_t : public next_client_notify_t
{
};

// ---------------------------------------------------------------

struct next_client_internal_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    next_queue_t * command_queue;
    next_queue_t * notify_queue;
    next_platform_socket_t * socket;
    next_platform_mutex_t command_mutex;
    next_platform_mutex_t notify_mutex;
    next_address_t server_address;
    next_address_t client_external_address;     // IMPORTANT: only known post-upgrade
    uint16_t bound_port;
    bool session_open;
    bool upgraded;
    bool reported;
    bool fallback_to_direct;
    bool multipath;
    uint8_t open_session_sequence;
    uint64_t upgrade_sequence;
    uint64_t session_id;
    uint64_t special_send_sequence;
    uint64_t internal_send_sequence;
    double last_next_ping_time;
    double last_next_pong_time;
    double last_direct_ping_time;
    double last_direct_pong_time;
    double last_stats_update_time;
    double last_stats_report_time;
    double last_route_switch_time;
    double route_update_timeout_time;
    uint64_t route_update_sequence;
    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    NEXT_DECLARE_SENTINEL(1)

    std::atomic<uint64_t> packets_sent;

    NEXT_DECLARE_SENTINEL(2)

    next_relay_manager_t * near_relay_manager;
    next_route_manager_t * route_manager;
    next_platform_mutex_t route_manager_mutex;

    NEXT_DECLARE_SENTINEL(3)

    next_packet_loss_tracker_t packet_loss_tracker;
    next_out_of_order_tracker_t out_of_order_tracker;
    next_jitter_tracker_t jitter_tracker;

    NEXT_DECLARE_SENTINEL(4)

    uint8_t customer_public_key[NEXT_CRYPTO_SIGN_PUBLICKEYBYTES];
    uint8_t client_kx_public_key[NEXT_CRYPTO_KX_PUBLICKEYBYTES];
    uint8_t client_kx_private_key[NEXT_CRYPTO_KX_SECRETKEYBYTES];
    uint8_t client_send_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
    uint8_t client_receive_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
    uint8_t client_route_public_key[NEXT_CRYPTO_BOX_PUBLICKEYBYTES];
    uint8_t client_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(5)

    next_client_stats_t client_stats;

    NEXT_DECLARE_SENTINEL(6)

    next_relay_stats_t near_relay_stats;

    NEXT_DECLARE_SENTINEL(7)

    next_ping_history_t next_ping_history;
    next_ping_history_t direct_ping_history;

    NEXT_DECLARE_SENTINEL(8)

    next_replay_protection_t payload_replay_protection;
    next_replay_protection_t special_replay_protection;
    next_replay_protection_t internal_replay_protection;

    NEXT_DECLARE_SENTINEL(9)

    next_platform_mutex_t direct_bandwidth_mutex;
    float direct_bandwidth_usage_kbps_up;
    float direct_bandwidth_usage_kbps_down;

    NEXT_DECLARE_SENTINEL(10)

    next_platform_mutex_t next_bandwidth_mutex;
    bool next_bandwidth_over_limit;
    float next_bandwidth_usage_kbps_up;
    float next_bandwidth_usage_kbps_down;
    float next_bandwidth_envelope_kbps_up;
    float next_bandwidth_envelope_kbps_down;

    NEXT_DECLARE_SENTINEL(11)

    bool sending_upgrade_response;
    double upgrade_response_start_time;
    double last_upgrade_response_send_time;
    int upgrade_response_packet_bytes;
    uint8_t upgrade_response_packet_data[NEXT_MAX_PACKET_BYTES];

    NEXT_DECLARE_SENTINEL(12)

    std::atomic<uint64_t> counters[NEXT_CLIENT_COUNTER_MAX];

    NEXT_DECLARE_SENTINEL(13)
};

void next_client_internal_initialize_sentinels( next_client_internal_t * client )
{
    (void) client;
    next_assert( client );
    NEXT_INITIALIZE_SENTINEL( client, 0 )
    NEXT_INITIALIZE_SENTINEL( client, 1 )
    NEXT_INITIALIZE_SENTINEL( client, 2 )
    NEXT_INITIALIZE_SENTINEL( client, 3 )
    NEXT_INITIALIZE_SENTINEL( client, 4 )
    NEXT_INITIALIZE_SENTINEL( client, 5 )
    NEXT_INITIALIZE_SENTINEL( client, 6 )
    NEXT_INITIALIZE_SENTINEL( client, 7 )
    NEXT_INITIALIZE_SENTINEL( client, 8 )
    NEXT_INITIALIZE_SENTINEL( client, 9 )
    NEXT_INITIALIZE_SENTINEL( client, 10 )
    NEXT_INITIALIZE_SENTINEL( client, 11 )
    NEXT_INITIALIZE_SENTINEL( client, 12 )
    NEXT_INITIALIZE_SENTINEL( client, 13 )

    next_replay_protection_initialize_sentinels( &client->payload_replay_protection );
    next_replay_protection_initialize_sentinels( &client->special_replay_protection );
    next_replay_protection_initialize_sentinels( &client->internal_replay_protection );

    next_relay_stats_initialize_sentinels( &client->near_relay_stats );

    next_ping_history_initialize_sentinels( &client->next_ping_history );
    next_ping_history_initialize_sentinels( &client->direct_ping_history );
}

void next_client_internal_verify_sentinels( next_client_internal_t * client )
{
    (void) client;

    next_assert( client );

    NEXT_VERIFY_SENTINEL( client, 0 )
    NEXT_VERIFY_SENTINEL( client, 1 )
    NEXT_VERIFY_SENTINEL( client, 2 )
    NEXT_VERIFY_SENTINEL( client, 3 )
    NEXT_VERIFY_SENTINEL( client, 4 )
    NEXT_VERIFY_SENTINEL( client, 5 )
    NEXT_VERIFY_SENTINEL( client, 6 )
    NEXT_VERIFY_SENTINEL( client, 7 )
    NEXT_VERIFY_SENTINEL( client, 8 )
    NEXT_VERIFY_SENTINEL( client, 9 )
    NEXT_VERIFY_SENTINEL( client, 10 )
    NEXT_VERIFY_SENTINEL( client, 11 )
    NEXT_VERIFY_SENTINEL( client, 12 )
    NEXT_VERIFY_SENTINEL( client, 13 )

    if ( client->command_queue )
        next_queue_verify_sentinels( client->command_queue );

    if ( client->notify_queue )
        next_queue_verify_sentinels( client->notify_queue );

    next_replay_protection_verify_sentinels( &client->payload_replay_protection );
    next_replay_protection_verify_sentinels( &client->special_replay_protection );
    next_replay_protection_verify_sentinels( &client->internal_replay_protection );

    next_relay_stats_verify_sentinels( &client->near_relay_stats );

    if ( client->near_relay_manager )
        next_relay_manager_verify_sentinels( client->near_relay_manager );

    next_ping_history_verify_sentinels( &client->next_ping_history );
    next_ping_history_verify_sentinels( &client->direct_ping_history );

    if ( client->route_manager )
        next_route_manager_verify_sentinels( client->route_manager );
}

void next_client_internal_destroy( next_client_internal_t * client );

next_client_internal_t * next_client_internal_create( void * context, const char * bind_address_string )
{
#if !NEXT_DEVELOPMENT
    next_printf( NEXT_LOG_LEVEL_INFO, "client sdk version is %s", NEXT_VERSION_FULL );
#endif // #if !NEXT_DEVELOPMENT

    next_printf( NEXT_LOG_LEVEL_INFO, "client buyer id is %" PRIx64, next_global_config.client_customer_id );

    next_address_t bind_address;
    if ( next_address_parse( &bind_address, bind_address_string ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client failed to parse bind address: %s", bind_address_string );
        return NULL;
    }

    next_client_internal_t * client = (next_client_internal_t*) next_malloc( context, sizeof(next_client_internal_t) );
    if ( !client )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not create internal client" );
        return NULL;
    }

    char * just_clear_it_and_dont_complain = (char*) client;
    memset( just_clear_it_and_dont_complain, 0, sizeof(next_client_internal_t) );

    next_client_internal_initialize_sentinels( client );

    next_client_internal_verify_sentinels( client );

    client->context = context;

    memcpy( client->customer_public_key, next_global_config.customer_public_key, NEXT_CRYPTO_SIGN_PUBLICKEYBYTES );

    next_client_internal_verify_sentinels( client );

    client->command_queue = next_queue_create( context, NEXT_COMMAND_QUEUE_LENGTH );
    if ( !client->command_queue )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create client command queue" );
        next_client_internal_destroy( client );
        return NULL;
    }

    next_client_internal_verify_sentinels( client );

    client->notify_queue = next_queue_create( context, NEXT_NOTIFY_QUEUE_LENGTH );
    if ( !client->notify_queue )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create client notify queue" );
        next_client_internal_destroy( client );
        return NULL;
    }

    next_client_internal_verify_sentinels( client );

    client->socket = next_platform_socket_create( client->context, &bind_address, NEXT_PLATFORM_SOCKET_BLOCKING, 0.1f, next_global_config.socket_send_buffer_size, next_global_config.socket_receive_buffer_size, true );
    if ( client->socket == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create socket" );
        next_client_internal_destroy( client );
        return NULL;
    }

    char address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_INFO, "client bound to %s", next_address_to_string( &bind_address, address_string ) );
    client->bound_port = bind_address.port;

    int result = next_platform_mutex_create( &client->command_mutex );
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create command mutex" );
        next_client_internal_destroy( client );
        return NULL;
    }

    result = next_platform_mutex_create( &client->notify_mutex );
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create notify mutex" );
        next_client_internal_destroy( client );
        return NULL;
    }

    client->near_relay_manager = next_relay_manager_create( context );
    if ( !client->near_relay_manager )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create near relay manager" );
        next_client_internal_destroy( client );
        return NULL;
    }

    client->route_manager = next_route_manager_create( context );
    if ( !client->route_manager )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create route manager" );
        next_client_internal_destroy( client );
        return NULL;
    }

    result = next_platform_mutex_create( &client->route_manager_mutex );
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create client route manager mutex" );
        next_client_internal_destroy( client );
        return NULL;
    }

    result = next_platform_mutex_create( &client->direct_bandwidth_mutex );
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create direct bandwidth mutex" );
        next_client_internal_destroy( client );
        return NULL;
    }

    result = next_platform_mutex_create( &client->next_bandwidth_mutex );
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create next bandwidth mutex" );
        next_client_internal_destroy( client );
        return NULL;
    }

    next_ping_history_clear( &client->next_ping_history );
    next_ping_history_clear( &client->direct_ping_history );

    next_replay_protection_reset( &client->payload_replay_protection );
    next_replay_protection_reset( &client->special_replay_protection );
    next_replay_protection_reset( &client->internal_replay_protection );

    next_packet_loss_tracker_reset( &client->packet_loss_tracker );
    next_out_of_order_tracker_reset( &client->out_of_order_tracker );
    next_jitter_tracker_reset( &client->jitter_tracker );

    next_client_internal_verify_sentinels( client );

    client->special_send_sequence = 1;
    client->internal_send_sequence = 1;

    return client;
}

void next_client_internal_destroy( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    if ( client->socket )
    {
        next_platform_socket_destroy( client->socket );
    }
    if ( client->command_queue )
    {
        next_queue_destroy( client->command_queue );
    }
    if ( client->notify_queue )
    {
        next_queue_destroy( client->notify_queue );
    }
    if ( client->near_relay_manager )
    {
        next_relay_manager_destroy( client->near_relay_manager );
    }
    if ( client->route_manager )
    {
        next_route_manager_destroy( client->route_manager );
    }

    next_platform_mutex_destroy( &client->command_mutex );
    next_platform_mutex_destroy( &client->notify_mutex );
    next_platform_mutex_destroy( &client->route_manager_mutex );
    next_platform_mutex_destroy( &client->direct_bandwidth_mutex );
    next_platform_mutex_destroy( &client->next_bandwidth_mutex );

    next_clear_and_free( client->context, client, sizeof(next_client_internal_t) );
}

int next_client_internal_send_packet_to_server( next_client_internal_t * client, uint8_t packet_id, void * packet_object )
{
    next_client_internal_verify_sentinels( client );

    next_assert( packet_object );
    next_assert( client->session_open );

    if ( !client->session_open )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client can't send internal packet to server because no session is open" );
        return NEXT_ERROR;
    }

    int packet_bytes = 0;

    uint8_t buffer[NEXT_MAX_PACKET_BYTES];

    uint8_t from_address_data[32];
    uint8_t to_address_data[32];
    uint16_t from_address_port;
    uint16_t to_address_port;
    int from_address_bytes;
    int to_address_bytes;

    next_address_data( &client->client_external_address, from_address_data, &from_address_bytes, &from_address_port );
    next_address_data( &client->server_address, to_address_data, &to_address_bytes, &to_address_port );

    if ( next_write_packet( packet_id, packet_object, buffer, &packet_bytes, next_signed_packets, next_encrypted_packets, &client->internal_send_sequence, NULL, client->client_send_key, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client failed to write internal packet type %d", packet_id );
        return NEXT_ERROR;
    }

    next_assert( next_basic_packet_filter( buffer, sizeof(buffer) ) );
    next_assert( next_advanced_packet_filter( buffer, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( client->socket, &client->server_address, buffer, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING

    return NEXT_OK;
}

void next_client_internal_process_network_next_packet( next_client_internal_t * client, const next_address_t * from, uint8_t * packet_data, int packet_bytes, double packet_receive_time )
{
    next_client_internal_verify_sentinels( client );

    next_assert( from );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );
    next_assert( packet_bytes <= NEXT_MAX_PACKET_BYTES );

    const bool from_server_address = client->server_address.type != 0 && next_address_equal( from, &client->server_address );

    const int packet_id = packet_data[0];

#if NEXT_ASSERT
    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_SPAM, "client processing packet type %d from %s (%d bytes)", packet_id, next_address_to_string( &client->server_address, address_buffer ), packet_bytes );
#endif // #if NEXT_ASSERT

    // run packet filters
    {
        if ( !next_basic_packet_filter( packet_data, packet_bytes ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client basic packet filter dropped packet (%d)", packet_id );
            return;
        }

        uint8_t from_address_data[32];
        uint8_t to_address_data[32];
        uint16_t from_address_port = 0;
        uint16_t to_address_port = 0;
        int from_address_bytes = 0;
        int to_address_bytes = 0;

        next_address_data( from, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( &client->client_external_address, to_address_data, &to_address_bytes, &to_address_port );

        if ( packet_id != NEXT_UPGRADE_REQUEST_PACKET )
        {
            if ( !next_advanced_packet_filter( packet_data, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) )
            {
                if ( !next_advanced_packet_filter( packet_data, client->upcoming_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) )
                {
                    if ( !next_advanced_packet_filter( packet_data, client->previous_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) )
                    {
                        next_printf( NEXT_LOG_LEVEL_DEBUG, "client advanced packet filter dropped packet (%d)", packet_id );
                    }
                    return;
                }
            }
        }
        else
        {
            uint8_t magic[8];
            memset( magic, 0, sizeof(magic) );
            to_address_bytes = 0;
            to_address_port = 0;
            if ( !next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "client advanced packet filter dropped packet (upgrade request)" );
                return;
            }
        }
    }

    // upgrade request packet (not encrypted)

    if ( !client->upgraded && from_server_address && packet_id == NEXT_UPGRADE_REQUEST_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing upgrade request packet" );

        if ( !next_address_equal( from, &client->server_address ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. packet does not come from server address" );
            return;
        }

        if ( client->fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. in fallback to direct state" );
            return;
        }

        if ( next_global_config.disable_network_next )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. network next is disabled" );
            return;
        }

        NextUpgradeRequestPacket packet;
        int begin = 16;
        int end = packet_bytes - 2;
        if ( next_read_packet( NEXT_UPGRADE_REQUEST_PACKET, packet_data, begin, end, &packet, NULL, NULL, NULL, NULL, NULL, NULL ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. failed to read" );
            return;
        }

        if ( packet.protocol_version != next_protocol_version() )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. protocol version mismatch" );
            return;
        }

        next_post_validate_packet( NEXT_UPGRADE_REQUEST_PACKET, NULL, NULL, NULL );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client received upgrade request packet from server" );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client initial magic: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
            packet.upcoming_magic[0],
            packet.upcoming_magic[1],
            packet.upcoming_magic[2],
            packet.upcoming_magic[3],
            packet.upcoming_magic[4],
            packet.upcoming_magic[5],
            packet.upcoming_magic[6],
            packet.upcoming_magic[7],
            packet.current_magic[0],
            packet.current_magic[1],
            packet.current_magic[2],
            packet.current_magic[3],
            packet.current_magic[4],
            packet.current_magic[5],
            packet.current_magic[6],
            packet.current_magic[7],
            packet.previous_magic[0],
            packet.previous_magic[1],
            packet.previous_magic[2],
            packet.previous_magic[3],
            packet.previous_magic[4],
            packet.previous_magic[5],
            packet.previous_magic[6],
            packet.previous_magic[7] );

        memcpy( client->upcoming_magic, packet.upcoming_magic, 8 );
        memcpy( client->current_magic, packet.current_magic, 8 );
        memcpy( client->previous_magic, packet.previous_magic, 8 );

        client->client_external_address = packet.client_address;

        char address_buffer[256];
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client external address is %s", next_address_to_string( &client->client_external_address, address_buffer ) );

        NextUpgradeResponsePacket response;

        response.client_open_session_sequence = client->open_session_sequence;
        memcpy( response.client_kx_public_key, client->client_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
        memcpy( response.client_route_public_key, client->client_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );
        memcpy( response.upgrade_token, packet.upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );
        response.platform_id = next_platform_id();

        // todo
        // response.connection_type = next_platform_connection_type();

        if ( next_client_internal_send_packet_to_server( client, NEXT_UPGRADE_RESPONSE_PACKET, &response ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_WARN, "client failed to send upgrade response packet to server" );
            return;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client sent upgrade response packet to server" );

        // IMPORTANT: Cache upgrade response and keep sending it until we get an upgrade confirm.
        // Without this, under very rare packet loss conditions it's possible for the client to get
        // stuck in an undefined state.

        uint8_t from_address_data[32];
        uint8_t to_address_data[32];
        uint16_t from_address_port = 0;
        uint16_t to_address_port = 0;
        int from_address_bytes = 0;
        int to_address_bytes = 0;

        next_address_data( &client->client_external_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( &client->server_address, to_address_data, &to_address_bytes, &to_address_port );

        client->upgrade_response_packet_bytes = 0;
        const int result = next_write_packet( NEXT_UPGRADE_RESPONSE_PACKET, &response, client->upgrade_response_packet_data, &client->upgrade_response_packet_bytes, NULL, NULL, NULL, NULL, NULL, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

        if ( result != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "client failed to write upgrade response packet" );
            return;
        }

#if NEXT_DEBUG

        const uint8_t * packet_data = client->upgrade_response_packet_data;
        const int packet_bytes = client->upgrade_response_packet_bytes;

        next_assert( packet_data );
        next_assert( packet_bytes > 0 );

        next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
        next_assert( next_advanced_packet_filter( packet_data, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

#endif // #if NEXT_DEBUG

        client->sending_upgrade_response = true;
        client->upgrade_response_start_time = next_platform_time();
        client->last_upgrade_response_send_time = next_platform_time();

        return;
    }

    // upgrade confirm packet

    if ( !client->upgraded && packet_id == NEXT_UPGRADE_CONFIRM_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing upgrade confirm packet" );

        if ( !client->sending_upgrade_response )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade confirm packet from server. unexpected" );
            return;
        }

        if ( client->fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. in fallback to direct state" );
            return;
        }

        if ( !next_address_equal( from, &client->server_address ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. not from server address" );
            return;
        }

        NextUpgradeConfirmPacket packet;
        int begin = 16;
        int end = packet_bytes - 2;
        if ( next_read_packet( NEXT_UPGRADE_CONFIRM_PACKET, packet_data, begin, end, &packet, NULL, NULL, NULL, NULL, NULL, NULL ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade request packet from server. could not read packet" );
            return;
        }

        if ( memcmp( packet.client_kx_public_key, client->client_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES ) != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade confirm packet from server. client public key does not match" );
            return;
        }

        if ( client->upgraded && client->upgrade_sequence >= packet.upgrade_sequence )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade confirm packet from server. client already upgraded" );
            return;
        }

        uint8_t client_send_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
        uint8_t client_receive_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
        if ( next_crypto_kx_client_session_keys( client_receive_key, client_send_key, client->client_kx_public_key, client->client_kx_private_key, packet.server_kx_public_key ) != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored upgrade confirm packet from server. could not generate session keys from server public key" );
            return;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client received upgrade confirm packet from server" );

        next_post_validate_packet( NEXT_UPGRADE_CONFIRM_PACKET, NULL, NULL, NULL );

        client->upgraded = true;
        client->upgrade_sequence = packet.upgrade_sequence;
        client->session_id = packet.session_id;
        client->last_direct_pong_time = next_platform_time();
        client->last_next_pong_time = next_platform_time();
        memcpy( client->client_send_key, client_send_key, NEXT_CRYPTO_KX_SESSIONKEYBYTES );
        memcpy( client->client_receive_key, client_receive_key, NEXT_CRYPTO_KX_SESSIONKEYBYTES );

        next_client_notify_upgraded_t * notify = (next_client_notify_upgraded_t*) next_malloc( client->context, sizeof(next_client_notify_upgraded_t) );
        next_assert( notify );
        notify->type = NEXT_CLIENT_NOTIFY_UPGRADED;
        notify->session_id = client->session_id;
        notify->client_external_address = client->client_external_address;
        memcpy( notify->current_magic, client->current_magic, 8 );
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_UPGRADED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &client->notify_mutex );
            next_queue_push( client->notify_queue, notify );
        }

        client->counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION]++;

        client->sending_upgrade_response = false;

        client->route_update_timeout_time = next_platform_time() + NEXT_CLIENT_ROUTE_UPDATE_TIMEOUT;

        return;
    }

    // direct packet

    if ( packet_id == NEXT_DIRECT_PACKET && client->upgraded && from_server_address )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing direct packet" );

        packet_data += 16;
        packet_bytes -= 18;

        if ( packet_bytes <= 9 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored direct packet. packet is too small to be valid" );
            return;
        }

        if ( packet_bytes > NEXT_MTU + 9 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored direct packet. packet is too large to be valid" );
            return;
        }

        const uint8_t * p = packet_data;

        uint8_t packet_session_sequence = next_read_uint8( &p );

        if ( packet_session_sequence != client->open_session_sequence )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored direct packet. session mismatch" );
            return;
        }

        uint64_t packet_sequence = next_read_uint64( &p );

        if ( next_replay_protection_already_received( &client->payload_replay_protection, packet_sequence ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored direct packet. already received" );
            return;
        }

        next_replay_protection_advance_sequence( &client->payload_replay_protection, packet_sequence );

        next_packet_loss_tracker_packet_received( &client->packet_loss_tracker, packet_sequence );

        next_out_of_order_tracker_packet_received( &client->out_of_order_tracker, packet_sequence );

        next_jitter_tracker_packet_received( &client->jitter_tracker, packet_sequence, packet_receive_time );

        next_client_notify_packet_received_t * notify = (next_client_notify_packet_received_t*) next_malloc( client->context, sizeof( next_client_notify_packet_received_t ) );
        notify->type = NEXT_CLIENT_NOTIFY_PACKET_RECEIVED;
        notify->direct = true;
        notify->payload_bytes = packet_bytes - 9;
        next_assert( notify->payload_bytes > 0 );
        memcpy( notify->payload_data, packet_data + 9, size_t(notify->payload_bytes) );
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_PACKET_RECEIVED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &client->notify_mutex );
            next_queue_push( client->notify_queue, notify );
        }
        client->counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]++;

        return;
    }

    // -------------------
    // PACKETS FROM RELAYS
    // -------------------

    if ( packet_id == NEXT_ROUTE_RESPONSE_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing route response packet" );

        packet_data += 16;
        packet_bytes -= 18;

        if ( packet_bytes != NEXT_HEADER_BYTES )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. bad packet size" );
            return;
        }

        bool fallback_to_direct;
        bool pending_route;
        uint64_t pending_route_session_id;
        uint8_t pending_route_session_version;
        uint8_t route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            memcpy( route_private_key, client->route_manager->route_data.pending_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
            fallback_to_direct = client->route_manager->fallback_to_direct;
            pending_route = client->route_manager->route_data.pending_route;
            pending_route_session_id = client->route_manager->route_data.pending_route_session_id;
            pending_route_session_version = client->route_manager->route_data.pending_route_session_version;
        }

        uint64_t packet_sequence = 0;
        uint64_t packet_session_id = 0;
        uint8_t packet_session_version = 0;

        if ( next_read_header( packet_id, &packet_sequence, &packet_session_id, &packet_session_version, route_private_key, packet_data, packet_bytes ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. could not read header" );
            return;
        }

        if ( fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. in fallback to direct state" );
            return;
        }

        if ( !pending_route )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. no pending route" );
            return;
        }

        next_platform_mutex_guard( &client->route_manager_mutex );

        next_route_manager_t * route_manager = client->route_manager;

        next_replay_protection_t * replay_protection = &client->special_replay_protection;

        if ( next_replay_protection_already_received( replay_protection, packet_sequence ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. sequence already received (%" PRIx64 " vs. %" PRIx64 ")", packet_sequence, replay_protection->most_recent_sequence );
            return;
        }

        if ( packet_session_id != pending_route_session_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. session id mismatch" );
            return;
        }

        if ( packet_session_version != pending_route_session_version )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route response packet from relay. session version mismatch" );
            return;
        }

        next_replay_protection_advance_sequence( replay_protection, packet_sequence );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client received route response from relay" );

        if ( route_manager->route_data.current_route )
        {
            route_manager->route_data.previous_route = route_manager->route_data.current_route;
            route_manager->route_data.previous_route_session_id = route_manager->route_data.current_route_session_id;
            route_manager->route_data.previous_route_session_version = route_manager->route_data.current_route_session_version;
            memcpy( route_manager->route_data.previous_route_private_key, route_manager->route_data.current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
        }

        route_manager->route_data.current_route_session_id = route_manager->route_data.pending_route_session_id;
        route_manager->route_data.current_route_session_version = route_manager->route_data.pending_route_session_version;
        route_manager->route_data.current_route_kbps_up = route_manager->route_data.pending_route_kbps_up;
        route_manager->route_data.current_route_kbps_down = route_manager->route_data.pending_route_kbps_down;
        route_manager->route_data.current_route_next_address = route_manager->route_data.pending_route_next_address;
        memcpy( route_manager->route_data.current_route_private_key, route_manager->route_data.pending_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );

        if ( !route_manager->route_data.current_route )
        {
            route_manager->route_data.current_route_expire_time = route_manager->route_data.pending_route_start_time + 2.0 * NEXT_SLICE_SECONDS;
        }
        else
        {
            route_manager->route_data.current_route_expire_time += 2.0 * NEXT_SLICE_SECONDS;
        }

        route_manager->route_data.current_route = true;
        route_manager->route_data.pending_route = false;

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client network next route is confirmed" );

        client->last_route_switch_time = next_platform_time();

        const bool route_established = route_manager->route_data.current_route;

        const int route_kbps_up = route_manager->route_data.current_route_kbps_up;
        const int route_kbps_down = route_manager->route_data.current_route_kbps_down;

        if ( route_established )
        {
            client->next_bandwidth_envelope_kbps_up = route_kbps_up;
            client->next_bandwidth_envelope_kbps_down = route_kbps_down;
        }
        else
        {
            client->next_bandwidth_envelope_kbps_up = 0;
            client->next_bandwidth_envelope_kbps_down = 0;
        }

        return;
    }

    // continue response packet

    if ( packet_id == NEXT_CONTINUE_RESPONSE_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing continue response packet" );

        packet_data += 16;
        packet_bytes -= 18;

        if ( packet_bytes != NEXT_HEADER_BYTES )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. bad packet size" );
            return;
        }

        uint8_t current_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];
        bool fallback_to_direct = client->route_manager->fallback_to_direct;
        bool current_route = client->route_manager->route_data.current_route;
        bool pending_continue = client->route_manager->route_data.pending_continue;
        uint64_t current_route_session_id = client->route_manager->route_data.current_route_session_id;
        uint8_t current_route_session_version = client->route_manager->route_data.current_route_session_version;
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            memcpy( current_route_private_key, client->route_manager->route_data.current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
            fallback_to_direct = client->route_manager->fallback_to_direct;
            current_route = client->route_manager->route_data.current_route;
            pending_continue = client->route_manager->route_data.pending_continue;
            current_route_session_id = client->route_manager->route_data.current_route_session_id;
            current_route_session_version = client->route_manager->route_data.current_route_session_version;
        }

        if ( fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. in fallback to direct state" );
            return;
        }

        if ( !current_route )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. no current route" );
            return;
        }

        if ( !pending_continue )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. no pending continue" );
            return;
        }

        uint64_t packet_sequence = 0;
        uint64_t packet_session_id = 0;
        uint8_t packet_session_version = 0;

        if ( next_read_header( packet_id, &packet_sequence, &packet_session_id, &packet_session_version, current_route_private_key, packet_data, packet_bytes ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. could not read header" );
            return;
        }

        next_replay_protection_t * replay_protection = &client->special_replay_protection;

        if ( next_replay_protection_already_received( replay_protection, packet_sequence ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. sequence already received (%" PRIx64 " vs. %" PRIx64 ")", packet_sequence, replay_protection->most_recent_sequence );
            return;
        }

        if ( packet_session_id != current_route_session_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. session id mismatch" );
            return;
        }

        if ( packet_session_version != current_route_session_version )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored continue response packet from relay. session version mismatch" );
            return;
        }

        next_replay_protection_advance_sequence( replay_protection, packet_sequence );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client received continue response from relay" );

        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            client->route_manager->route_data.current_route_expire_time += NEXT_SLICE_SECONDS;
            client->route_manager->route_data.pending_continue = false;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client continue network next route is confirmed" );

        return;
    }

    // server to client packet

    if ( packet_id == NEXT_SERVER_TO_CLIENT_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing server to client packet" );

        packet_data += 16;
        packet_bytes -= 18;

        uint64_t payload_sequence = 0;

        bool result = false;
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            result = next_route_manager_process_server_to_client_packet( client->route_manager, packet_id, packet_data, packet_bytes, &payload_sequence );
        }

        if ( !result )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. could not verify" );
            return;
        }

        const bool already_received = next_replay_protection_already_received( &client->payload_replay_protection, payload_sequence ) != 0;

        if ( already_received && !client->multipath )
        {
            return;
        }

        if ( already_received && client->multipath )
        {
            client->counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]++;
            return;
        }

        next_replay_protection_advance_sequence( &client->payload_replay_protection, payload_sequence );

        next_packet_loss_tracker_packet_received( &client->packet_loss_tracker, payload_sequence );

        next_out_of_order_tracker_packet_received( &client->out_of_order_tracker, payload_sequence );

        next_jitter_tracker_packet_received( &client->jitter_tracker, payload_sequence, next_platform_time() );

        next_client_notify_packet_received_t * notify = (next_client_notify_packet_received_t*) next_malloc( client->context, sizeof( next_client_notify_packet_received_t ) );
        notify->type = NEXT_CLIENT_NOTIFY_PACKET_RECEIVED;
        notify->direct = false;
        notify->payload_bytes = packet_bytes - NEXT_HEADER_BYTES;
        memcpy( notify->payload_data, packet_data + NEXT_HEADER_BYTES, size_t(packet_bytes) - NEXT_HEADER_BYTES );
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_PACKET_RECEIVED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &client->notify_mutex );
            next_queue_push( client->notify_queue, notify );
        }

        client->counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]++;

        return;
    }

    // next pong packet

    if ( packet_id == NEXT_PONG_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing next pong packet" );

        packet_data += 16;
        packet_bytes -= 18;

        uint64_t payload_sequence = 0;

        bool result = false;
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            result = next_route_manager_process_server_to_client_packet( client->route_manager, packet_id, packet_data, packet_bytes, &payload_sequence );
        }

        if ( !result )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored server to client packet. could not verify" );
            return;
        }

        if ( next_replay_protection_already_received( &client->special_replay_protection, payload_sequence ) )
            return;

        next_replay_protection_advance_sequence( &client->special_replay_protection, payload_sequence );

        const uint8_t * p = packet_data + NEXT_HEADER_BYTES;

        uint64_t ping_sequence = next_read_uint64( &p );

        next_ping_history_pong_received( &client->next_ping_history, ping_sequence, next_platform_time() );

        client->last_next_pong_time = next_platform_time();

        return;
    }

    // relay pong packet

    if ( packet_id == NEXT_RELAY_PONG_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing relay pong packet" );

        if ( !client->upgraded )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored relay pong packet. not upgraded yet" );
            return;
        }

        packet_data += 16;
        packet_bytes -= 18;

        const uint8_t * p = packet_data;

        uint64_t ping_sequence = next_read_uint64( &p );
        uint64_t ping_session_id = next_read_uint64( &p );

        if ( ping_session_id != client->session_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignoring relay pong packet. session id does not match" );
            return;
        }

        next_relay_manager_process_pong( client->near_relay_manager, from, ping_sequence );

        return;
    }

    // -------------------
    // PACKETS FROM SERVER
    // -------------------

    if ( !next_address_equal( from, &client->server_address ) )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client ignoring packet because it's not from the server" );
        return;
    }

    // direct pong packet

    if ( packet_id == NEXT_DIRECT_PONG_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing direct packet" );

        NextDirectPongPacket packet;

        uint64_t packet_sequence = 0;

        int begin = 16;
        int end = packet_bytes - 2;

        if ( next_read_packet( NEXT_DIRECT_PONG_PACKET, packet_data, begin, end, &packet, next_signed_packets, next_encrypted_packets, &packet_sequence, NULL, client->client_receive_key, &client->internal_replay_protection ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored direct pong packet. could not read" );
            return;
        }

        next_ping_history_pong_received( &client->direct_ping_history, packet.ping_sequence, next_platform_time() );

        next_post_validate_packet( NEXT_DIRECT_PONG_PACKET, next_encrypted_packets, &packet_sequence, &client->internal_replay_protection );

        client->last_direct_pong_time = next_platform_time();

        return;
    }

    // route update packet

    if ( packet_id == NEXT_ROUTE_UPDATE_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "client processing route update packet" );

        if ( client->fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route update packet from server. in fallback to direct state (1)" );
            return;
        }

        NextRouteUpdatePacket packet;

        uint64_t packet_sequence = 0;

        int begin = 16;
        int end = packet_bytes - 2;

        if ( next_read_packet( NEXT_ROUTE_UPDATE_PACKET, packet_data, begin, end, &packet, next_signed_packets, next_encrypted_packets, &packet_sequence, NULL, client->client_receive_key, &client->internal_replay_protection ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route update packet. could not read" );
            return;
        }

        if ( packet.sequence < client->route_update_sequence )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route update packet from server. sequence is old" );
            return;
        }

        next_post_validate_packet( NEXT_ROUTE_UPDATE_PACKET, next_encrypted_packets, &packet_sequence, &client->internal_replay_protection );

        bool fallback_to_direct = false;

        if ( packet.sequence > client->route_update_sequence )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client received route update packet from server" );

            if ( packet.has_debug )
            {
                next_printf( "--------------------------------------\n%s--------------------------------------", packet.debug );
            }

            if ( packet.has_near_relays )
            {
                // enable near relay pings

                next_printf( NEXT_LOG_LEVEL_INFO, "client pinging %d near relays", packet.num_near_relays );
                
                next_relay_manager_update( client->near_relay_manager, packet.num_near_relays, packet.near_relay_ids, packet.near_relay_addresses, packet.near_relay_ping_tokens, packet.near_relay_expire_timestamp );
            }
            else
            {
                // disable near relay pings (and clear any ping data)
                
                if ( client->near_relay_manager->num_relays != 0 )
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "client near relay pings completed" );
                
                    next_relay_manager_update( client->near_relay_manager, 0, packet.near_relay_ids, packet.near_relay_addresses, NULL, 0 );
                }
            }

            {
                next_platform_mutex_guard( &client->route_manager_mutex );
                next_route_manager_update( client->route_manager, packet.update_type, packet.num_tokens, packet.tokens, next_router_public_key, client->client_route_private_key, client->current_magic, &client->client_external_address );
                fallback_to_direct = client->route_manager->fallback_to_direct;
            }

            if ( !client->fallback_to_direct && fallback_to_direct )
            {
                client->counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT]++;
            }

            client->fallback_to_direct = fallback_to_direct;

            if ( !fallback_to_direct )
            {
                if ( packet.multipath && !client->multipath )
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "client multipath enabled" );
                    client->multipath = true;
                    client->counters[NEXT_CLIENT_COUNTER_MULTIPATH]++;
                }

                client->route_update_sequence = packet.sequence;
                client->client_stats.packets_sent_server_to_client = packet.packets_sent_server_to_client;
                client->client_stats.packets_lost_client_to_server = packet.packets_lost_client_to_server;
                client->client_stats.packets_out_of_order_client_to_server = packet.packets_out_of_order_client_to_server;
                client->client_stats.jitter_client_to_server = packet.jitter_client_to_server;
                client->counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] = packet.packets_lost_client_to_server;
                client->counters[NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_CLIENT_TO_SERVER] = packet.packets_out_of_order_client_to_server;
                client->route_update_timeout_time = next_platform_time() + NEXT_CLIENT_ROUTE_UPDATE_TIMEOUT;

                if ( memcmp( client->upcoming_magic, packet.upcoming_magic, 8 ) != 0 )
                {
                    next_printf( NEXT_LOG_LEVEL_DEBUG, "client updated magic: %x,%x,%x,%x,%x,%x,%x,%x | %x,%x,%x,%x,%x,%x,%x,%x | %x,%x,%x,%x,%x,%x,%x,%x",
                        packet.upcoming_magic[0],
                        packet.upcoming_magic[1],
                        packet.upcoming_magic[2],
                        packet.upcoming_magic[3],
                        packet.upcoming_magic[4],
                        packet.upcoming_magic[5],
                        packet.upcoming_magic[6],
                        packet.upcoming_magic[7],

                        packet.current_magic[0],
                        packet.current_magic[1],
                        packet.current_magic[2],
                        packet.current_magic[3],
                        packet.current_magic[4],
                        packet.current_magic[5],
                        packet.current_magic[6],
                        packet.current_magic[7],

                        packet.previous_magic[0],
                        packet.previous_magic[1],
                        packet.previous_magic[2],
                        packet.previous_magic[3],
                        packet.previous_magic[4],
                        packet.previous_magic[5],
                        packet.previous_magic[6],
                        packet.previous_magic[7] );

                    memcpy( client->upcoming_magic, packet.upcoming_magic, 8 );
                    memcpy( client->current_magic, packet.current_magic, 8 );
                    memcpy( client->previous_magic, packet.previous_magic, 8 );

                    next_client_notify_magic_updated_t * notify = (next_client_notify_magic_updated_t*) next_malloc( client->context, sizeof(next_client_notify_magic_updated_t) );
                    next_assert( notify );
                    notify->type = NEXT_CLIENT_NOTIFY_MAGIC_UPDATED;
                    memcpy( notify->current_magic, client->current_magic, 8 );
                    {
#if NEXT_SPIKE_TRACKING
                        next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_MAGIC_UPDATED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
                        next_platform_mutex_guard( &client->notify_mutex );
                        next_queue_push( client->notify_queue, notify );
                    }
                }
            }
        }

        if ( fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "client ignored route update packet from server. in fallback to direct state (2)" );
            return;
        }

        NextRouteUpdateAckPacket ack;
        ack.sequence = packet.sequence;

        if ( next_client_internal_send_packet_to_server( client, NEXT_ROUTE_UPDATE_ACK_PACKET, &ack ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_WARN, "client failed to send route update ack packet to server" );
            return;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "client sent route update ack packet to server" );

        return;
    }
}

void next_client_internal_process_passthrough_packet( next_client_internal_t * client, const next_address_t * from, uint8_t * packet_data, int packet_bytes )
{
    next_client_internal_verify_sentinels( client );

    next_printf( NEXT_LOG_LEVEL_SPAM, "client processing passthrough packet" );

    next_assert( from );
    next_assert( packet_data );

    const bool from_server_address = client->server_address.type != 0 && next_address_equal( from, &client->server_address );

#if NEXT_SPIKE_TRACKING
    next_printf( NEXT_LOG_LEVEL_SPAM, "client drops passthrough packet at %s:%d because it does not think it comes from the server", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING

    if ( packet_bytes <= NEXT_MAX_PACKET_BYTES - 1 && from_server_address )
    {
        next_client_notify_packet_received_t * notify = (next_client_notify_packet_received_t*) next_malloc( client->context, sizeof( next_client_notify_packet_received_t ) );
        notify->type = NEXT_CLIENT_NOTIFY_PACKET_RECEIVED;
        notify->direct = true;
        notify->payload_bytes = packet_bytes;
        next_assert( notify->payload_bytes >= 0 );
        next_assert( notify->payload_bytes <= NEXT_MAX_PACKET_BYTES - 1 );
        memcpy( notify->payload_data, packet_data, size_t(packet_bytes) );
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_PACKET_RECEIVED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &client->notify_mutex );
            next_queue_push( client->notify_queue, notify );
        }
        client->counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH]++;
    }
}

#if NEXT_DEVELOPMENT
bool next_packet_loss = false;
#endif // #if NEXT_DEVELOPMENT

void next_client_internal_block_and_receive_packet( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

    next_assert( ( size_t(packet_data) % 4 ) == 0 );

    next_address_t from;

#if NEXT_SPIKE_TRACKING
    next_printf( NEXT_LOG_LEVEL_SPAM, "client calls next_platform_socket_receive_packet on internal thread" );
#endif // #if NEXT_SPIKE_TRACKING

    int packet_bytes = next_platform_socket_receive_packet( client->socket, &from, packet_data, NEXT_MAX_PACKET_BYTES );

#if NEXT_SPIKE_TRACKING
    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_SPAM, "client next_platform_socket_receive_packet returns with a %d byte packet from %s", packet_bytes, next_address_to_string( &from, address_buffer ) );
#endif // #if NEXT_SPIKE_TRACKING

    double packet_receive_time = next_platform_time();

    next_assert( packet_bytes >= 0 );

    if ( packet_bytes <= 1 )
        return;
    
#if NEXT_DEVELOPMENT
    if ( next_packet_loss && ( rand() % 10 ) == 0 )
        return;
#endif // #if NEXT_DEVELOPMENT

    if ( packet_data[0] != NEXT_PASSTHROUGH_PACKET )
    {
        next_client_internal_process_network_next_packet( client, &from, packet_data, packet_bytes, packet_receive_time );
    }
    else
    {
        next_client_internal_process_passthrough_packet( client, &from, packet_data + 1, packet_bytes - 1 );
    }
}

bool next_client_internal_pump_commands( next_client_internal_t * client )
{
#if NEXT_SPIKE_TRACKING
    next_printf( NEXT_LOG_LEVEL_SPAM, "next_client_internal_pump_commands" );
#endif // #if NEXT_SPIKE_TRACKING

    next_client_internal_verify_sentinels( client );

    bool quit = false;

    while ( true )
    {
        void * entry = NULL;
        {
            next_platform_mutex_guard( &client->command_mutex );
            entry = next_queue_pop( client->command_queue );
        }

        if ( entry == NULL )
            break;

        next_client_command_t * command = (next_client_command_t*) entry;

        switch ( command->type )
        {
            case NEXT_CLIENT_COMMAND_OPEN_SESSION:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread received NEXT_CLIENT_COMMAND_OPEN_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING

                next_client_command_open_session_t * open_session_command = (next_client_command_open_session_t*) entry;
                client->server_address = open_session_command->server_address;
                client->session_open = true;
                client->open_session_sequence++;
                client->last_direct_ping_time = next_platform_time();
                client->last_stats_update_time = next_platform_time();
                client->last_stats_report_time = next_platform_time() + next_random_float();
                next_crypto_kx_keypair( client->client_kx_public_key, client->client_kx_private_key );
                next_crypto_box_keypair( client->client_route_public_key, client->client_route_private_key );
                char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_INFO, "client opened session to %s", next_address_to_string( &open_session_command->server_address, buffer ) );
                client->counters[NEXT_CLIENT_COUNTER_OPEN_SESSION]++;
                {
                    next_platform_mutex_guard( &client->route_manager_mutex );
                    next_route_manager_reset( client->route_manager );
                    next_route_manager_direct_route( client->route_manager, true );
                }

                // IMPORTANT: Fire back ready when the client is ready to start sending packets and we're all dialed in for this session
                next_client_notify_ready_t * notify = (next_client_notify_ready_t*) next_malloc( client->context, sizeof(next_client_notify_ready_t) );
                next_assert( notify );
                notify->type = NEXT_CLIENT_NOTIFY_READY;
                {
#if NEXT_SPIKE_TRACKING
                    next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_READY at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
                    next_platform_mutex_guard( &client->notify_mutex );
                    next_queue_push( client->notify_queue, notify );
                }
            }
            break;

            case NEXT_CLIENT_COMMAND_CLOSE_SESSION:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread received NEXT_CLIENT_COMMAND_CLOSE_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING

                if ( !client->session_open )
                    break;

                char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_INFO, "client closed session to %s", next_address_to_string( &client->server_address, buffer ) );

                memset( client->upcoming_magic, 0, 8 );
                memset( client->current_magic, 0, 8 );
                memset( client->previous_magic, 0, 8 );
                memset( &client->server_address, 0, sizeof(next_address_t) );
                memset( &client->client_external_address, 0, sizeof(next_address_t) );

                client->session_open = false;
                client->upgraded = false;
                client->reported = false;
                client->fallback_to_direct = false;
                client->multipath = false;
                client->upgrade_sequence = 0;
                client->session_id = 0;
                client->internal_send_sequence = 0;
                client->last_next_ping_time = 0.0;
                client->last_next_pong_time = 0.0;
                client->last_direct_ping_time = 0.0;
                client->last_direct_pong_time = 0.0;
                client->last_stats_update_time = 0.0;
                client->last_stats_report_time = 0.0;
                client->last_route_switch_time = 0.0;
                client->route_update_timeout_time = 0.0;
                client->route_update_sequence = 0;
                client->sending_upgrade_response = false;
                client->upgrade_response_packet_bytes = 0;
                memset( client->upgrade_response_packet_data, 0, sizeof(client->upgrade_response_packet_data) );
                client->upgrade_response_start_time = 0.0;
                client->last_upgrade_response_send_time = 0.0;

                client->packets_sent = 0;

                memset( &client->client_stats, 0, sizeof(next_client_stats_t) );
                memset( &client->near_relay_stats, 0, sizeof(next_relay_stats_t ) );
                next_relay_stats_initialize_sentinels( &client->near_relay_stats );

                next_relay_manager_reset( client->near_relay_manager );

                memset( client->client_kx_public_key, 0, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
                memset( client->client_kx_private_key, 0, NEXT_CRYPTO_KX_SECRETKEYBYTES );
                memset( client->client_send_key, 0, NEXT_CRYPTO_KX_SESSIONKEYBYTES );
                memset( client->client_receive_key, 0, NEXT_CRYPTO_KX_SESSIONKEYBYTES );
                memset( client->client_route_public_key, 0, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );
                memset( client->client_route_private_key, 0, NEXT_CRYPTO_BOX_SECRETKEYBYTES );

                next_ping_history_clear( &client->next_ping_history );
                next_ping_history_clear( &client->direct_ping_history );

                next_replay_protection_reset( &client->payload_replay_protection );
                next_replay_protection_reset( &client->special_replay_protection );
                next_replay_protection_reset( &client->internal_replay_protection );

                {
                    next_platform_mutex_guard( &client->direct_bandwidth_mutex );
                    client->direct_bandwidth_usage_kbps_up = 0;
                    client->direct_bandwidth_usage_kbps_down = 0;
                }

                {
                    next_platform_mutex_guard( &client->next_bandwidth_mutex );
                    client->next_bandwidth_over_limit = false;
                    client->next_bandwidth_usage_kbps_up = 0;
                    client->next_bandwidth_usage_kbps_down = 0;
                    client->next_bandwidth_envelope_kbps_up = 0;
                    client->next_bandwidth_envelope_kbps_down = 0;
                }

                {
                    next_platform_mutex_guard( &client->route_manager_mutex );
                    next_route_manager_reset( client->route_manager );
                }

                next_packet_loss_tracker_reset( &client->packet_loss_tracker );
                next_out_of_order_tracker_reset( &client->out_of_order_tracker );
                next_jitter_tracker_reset( &client->jitter_tracker );

                client->counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION]++;
            }
            break;

            case NEXT_CLIENT_COMMAND_DESTROY:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread received NEXT_CLIENT_COMMAND_DESTROY" );
#endif // #if NEXT_SPIKE_TRACKING
                quit = true;
            }
            break;

            case NEXT_CLIENT_COMMAND_REPORT_SESSION:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread received NEXT_CLIENT_COMMAND_REPORT_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING
                if ( client->session_id != 0 && !client->reported )
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "client reported session %" PRIx64, client->session_id );
                    client->reported = true;
                }
            }
            break;

            default: break;
        }

        next_free( client->context, command );
    }

    return quit;
}

#if NEXT_DEVELOPMENT
bool next_fake_fallback_to_direct = false;
float next_fake_direct_packet_loss = 0.0f;
float next_fake_direct_rtt = 0.0f;
float next_fake_next_packet_loss = 0.0f;
float next_fake_next_rtt = 0.0f;
#endif // #if !NEXT_DEVELOPMENT

void next_client_internal_update_stats( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    next_assert( !next_global_config.disable_network_next );

    double current_time = next_platform_time();

    if ( client->last_stats_update_time + ( 1.0 / NEXT_CLIENT_STATS_UPDATES_PER_SECOND ) < current_time )
    {
        bool network_next = false;
        bool fallback_to_direct = false;
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            network_next = client->route_manager->route_data.current_route;
            fallback_to_direct = client->route_manager->fallback_to_direct;
        }
        
        client->client_stats.next = network_next;
        client->client_stats.upgraded = client->upgraded;
        client->client_stats.reported = client->reported;
        client->client_stats.fallback_to_direct = client->fallback_to_direct;
        client->client_stats.multipath = client->multipath;
        client->client_stats.platform_id = next_platform_id();

        // todo
        // client->client_stats.connection_type = next_platform_connection_type();

        double start_time = current_time - NEXT_CLIENT_STATS_WINDOW;
        if ( start_time < client->last_route_switch_time + NEXT_PING_SAFETY )
        {
            start_time = client->last_route_switch_time + NEXT_PING_SAFETY;
        }

        next_route_stats_t next_route_stats;
        next_route_stats_from_ping_history( &client->next_ping_history, current_time - NEXT_CLIENT_STATS_WINDOW, current_time, &next_route_stats );

        next_route_stats_t direct_route_stats;
        next_route_stats_from_ping_history( &client->direct_ping_history, current_time - NEXT_CLIENT_STATS_WINDOW, current_time, &direct_route_stats );

        {
            next_platform_mutex_guard( &client->direct_bandwidth_mutex );
            client->client_stats.direct_kbps_up = client->direct_bandwidth_usage_kbps_up;
            client->client_stats.direct_kbps_down = client->direct_bandwidth_usage_kbps_down;
        }

        if ( network_next )
        {
            client->client_stats.next_rtt = next_route_stats.rtt;
            client->client_stats.next_jitter = next_route_stats.jitter;
            client->client_stats.next_packet_loss = next_route_stats.packet_loss;
            {
                next_platform_mutex_guard( &client->next_bandwidth_mutex );
                client->client_stats.next_kbps_up = client->next_bandwidth_usage_kbps_up;
                client->client_stats.next_kbps_down = client->next_bandwidth_usage_kbps_down;
            }
        }
        else
        {
            client->client_stats.next_rtt = 0.0f;
            client->client_stats.next_jitter = 0.0f;
            client->client_stats.next_packet_loss = 0.0f;
            client->client_stats.next_kbps_up = 0;
            client->client_stats.next_kbps_down = 0;
        }

        client->client_stats.direct_rtt = direct_route_stats.rtt;
        client->client_stats.direct_jitter = direct_route_stats.jitter;
        client->client_stats.direct_packet_loss = direct_route_stats.packet_loss;

        if ( direct_route_stats.packet_loss > client->client_stats.direct_max_packet_loss_seen )
        {
            client->client_stats.direct_max_packet_loss_seen = direct_route_stats.packet_loss;
        }

#if NEXT_DEVELOPMENT
        if ( !fallback_to_direct && next_fake_fallback_to_direct )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "client fakes fallback to direct" );
            {
                next_platform_mutex_guard( &client->route_manager_mutex );
                next_route_manager_fallback_to_direct( client->route_manager, NEXT_FLAGS_ROUTE_UPDATE_TIMED_OUT );
            }
        }
        client->client_stats.direct_rtt += next_fake_direct_rtt;
        client->client_stats.direct_packet_loss += next_fake_direct_packet_loss;
        client->client_stats.next_rtt += next_fake_next_rtt;
        client->client_stats.next_packet_loss += next_fake_next_packet_loss;
 #endif // #if NEXT_DEVELOPMENT

        if ( !fallback_to_direct )
        {
            const int packets_lost = next_packet_loss_tracker_update( &client->packet_loss_tracker );
            client->client_stats.packets_lost_server_to_client += packets_lost;
            client->counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] += packets_lost;

            client->client_stats.packets_out_of_order_server_to_client = client->out_of_order_tracker.num_out_of_order_packets;
            client->counters[NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_SERVER_TO_CLIENT] = client->out_of_order_tracker.num_out_of_order_packets;

            client->client_stats.jitter_server_to_client = float( client->jitter_tracker.jitter * 1000.0 );
        }

        client->client_stats.packets_sent_client_to_server = client->packets_sent;

        next_relay_manager_get_stats( client->near_relay_manager, &client->near_relay_stats );

        next_client_notify_stats_updated_t * notify = (next_client_notify_stats_updated_t*) next_malloc( client->context, sizeof( next_client_notify_stats_updated_t ) );
        notify->type = NEXT_CLIENT_NOTIFY_STATS_UPDATED;
        notify->stats = client->client_stats;
        notify->fallback_to_direct = fallback_to_direct;
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "client internal thread queues up NEXT_CLIENT_NOTIFY_STATS_UPDATED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &client->notify_mutex );
            next_queue_push( client->notify_queue, notify );
        }

        client->last_stats_update_time = current_time;
    }

    if ( client->last_stats_report_time + 1.0 < current_time && client->client_stats.direct_rtt > 0.0f )
    {
        NextClientStatsPacket packet;

        packet.reported = client->reported;
        packet.fallback_to_direct = client->fallback_to_direct;
        packet.multipath = client->multipath;
        packet.platform_id = client->client_stats.platform_id;
        packet.connection_type = client->client_stats.connection_type;

        {
            next_platform_mutex_guard( &client->direct_bandwidth_mutex );
            packet.direct_kbps_up = (int) ceil( client->direct_bandwidth_usage_kbps_up );
            packet.direct_kbps_down = (int) ceil( client->direct_bandwidth_usage_kbps_down );
        }

        {
            next_platform_mutex_guard( &client->next_bandwidth_mutex );
            packet.next_bandwidth_over_limit = client->next_bandwidth_over_limit;
            packet.next_kbps_up = (int) ceil( client->next_bandwidth_usage_kbps_up );
            packet.next_kbps_down = (int) ceil( client->next_bandwidth_usage_kbps_down );
            client->next_bandwidth_over_limit = false;
        }

        if ( !client->client_stats.next )
        {
            packet.next_kbps_up = 0;
            packet.next_kbps_down = 0;
        }

        packet.next = client->client_stats.next;
        packet.next_rtt = client->client_stats.next_rtt;
        packet.next_jitter = client->client_stats.next_jitter;
        packet.next_packet_loss = client->client_stats.next_packet_loss;

        packet.direct_rtt = client->client_stats.direct_rtt;
        packet.direct_jitter = client->client_stats.direct_jitter;
        packet.direct_packet_loss = client->client_stats.direct_packet_loss;
        packet.direct_max_packet_loss_seen = client->client_stats.direct_max_packet_loss_seen;

        if ( !client->fallback_to_direct )
        {
            packet.num_near_relays = client->near_relay_stats.num_relays;
            for ( int i = 0; i < packet.num_near_relays; ++i )
            {
                packet.near_relay_ids[i] = client->near_relay_stats.relay_ids[i];

                int rtt = (int) ceil( client->near_relay_stats.relay_rtt[i] );
                int jitter = (int) ceil( client->near_relay_stats.relay_jitter[i] );
                float packet_loss = client->near_relay_stats.relay_packet_loss[i];

                if ( rtt > 255 )
                    rtt = 255;

                if ( jitter > 255 )
                    jitter = 255;

                if ( packet_loss > 100 )
                    packet_loss = 100;

                packet.near_relay_rtt[i] = uint8_t( rtt );
                packet.near_relay_jitter[i] = uint8_t( jitter );
                packet.near_relay_packet_loss[i] = packet_loss;
            }
        }

        packet.packets_sent_client_to_server = client->packets_sent;

        packet.packets_lost_server_to_client = client->client_stats.packets_lost_server_to_client;
        packet.packets_out_of_order_server_to_client = client->client_stats.packets_out_of_order_server_to_client;
        packet.jitter_server_to_client = client->client_stats.jitter_server_to_client;

        if ( next_client_internal_send_packet_to_server( client, NEXT_CLIENT_STATS_PACKET, &packet ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "client failed to send stats packet to server" );
            return;
        }

        client->last_stats_report_time = current_time;
    }
}

void next_client_internal_update_direct_pings( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    next_assert( !next_global_config.disable_network_next );

    if ( !client->upgraded )
        return;

    if ( client->fallback_to_direct )
        return;

    double current_time = next_platform_time();

    if ( client->last_direct_pong_time + NEXT_CLIENT_SESSION_TIMEOUT < current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client direct pong timed out. falling back to direct" );
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            next_route_manager_fallback_to_direct( client->route_manager, NEXT_FLAGS_DIRECT_PONG_TIMED_OUT );
        }
        return;
    }

    if ( client->last_direct_ping_time + ( 1.0 / NEXT_DIRECT_PINGS_PER_SECOND ) <= current_time )
    {
        NextDirectPingPacket packet;
        packet.ping_sequence = next_ping_history_ping_sent( &client->direct_ping_history, current_time );

        if ( next_client_internal_send_packet_to_server( client, NEXT_DIRECT_PING_PACKET, &packet ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "client failed to send direct ping packet to server" );
            return;
        }

        client->last_direct_ping_time = current_time;
    }
}

void next_client_internal_update_next_pings( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    next_assert( !next_global_config.disable_network_next );

    if ( !client->upgraded )
        return;

    if ( client->fallback_to_direct )
        return;

    double current_time = next_platform_time();

    bool has_next_route = false;
    {
        next_platform_mutex_guard( &client->route_manager_mutex );
        has_next_route = client->route_manager->route_data.current_route;
    }

    if ( !has_next_route )
    {
        client->last_next_pong_time = current_time;
    }

    if ( client->last_next_pong_time + NEXT_CLIENT_SESSION_TIMEOUT < current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client next pong timed out. falling back to direct" );
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            next_route_manager_fallback_to_direct( client->route_manager, NEXT_FLAGS_NEXT_PONG_TIMED_OUT );
        }
        return;
    }

    if ( client->last_next_ping_time + ( 1.0 / NEXT_PINGS_PER_SECOND ) <= current_time )
    {
        bool send_over_network_next = false;
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            send_over_network_next = client->route_manager->route_data.current_route;
        }

        if ( !send_over_network_next )
            return;

        uint64_t session_id;
        uint8_t session_version;
        next_address_t to;
        uint8_t private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            session_id = client->route_manager->route_data.current_route_session_id;
            session_version = client->route_manager->route_data.current_route_session_version;
            to = client->route_manager->route_data.current_route_next_address;
            memcpy( private_key, client->route_manager->route_data.current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
        }

        uint64_t sequence = client->special_send_sequence++;

        uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

        uint8_t from_address_data[32];
        uint8_t to_address_data[32];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( &client->client_external_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( &to, to_address_data, &to_address_bytes, &to_address_port );

        const uint64_t ping_sequence = next_ping_history_ping_sent( &client->next_ping_history, current_time );

        int packet_bytes = next_write_ping_packet( packet_data, sequence, session_id, session_version, private_key, ping_sequence, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

        next_assert( packet_bytes > 0 );

        next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
        next_assert( next_advanced_packet_filter( packet_data, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

#if NEXT_SPIKE_TRACKING
        double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

        next_platform_socket_send_packet( client->socket, &to, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
        double finish_time = next_platform_time();
        if ( finish_time - start_time > 0.001 )
        {
            next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
        }
#endif // #if NEXT_SPIKE_TRACKING

        client->last_next_ping_time = current_time;
    }
}

void next_client_internal_send_pings_to_near_relays( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    if ( next_global_config.disable_network_next )
        return;

    if ( !client->upgraded )
        return;

    if ( client->fallback_to_direct )
        return;

    next_relay_manager_send_pings( client->near_relay_manager, client->socket, client->session_id, client->current_magic, &client->client_external_address );
}

void next_client_internal_update_fallback_to_direct( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    next_assert( !next_global_config.disable_network_next );

    bool fallback_to_direct = false;
    {
        next_platform_mutex_guard( &client->route_manager_mutex );
        if ( client->upgraded )
        {
            next_route_manager_check_for_timeouts( client->route_manager );
        }
        fallback_to_direct = client->route_manager->fallback_to_direct;
    }

    if ( !client->fallback_to_direct && fallback_to_direct )
    {
        client->counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT]++;
        client->fallback_to_direct = fallback_to_direct;
        return;
    }

    if ( !client->fallback_to_direct && client->upgraded && client->route_update_timeout_time > 0.0 )
    {
        if ( next_platform_time() > client->route_update_timeout_time )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "client route update timeout. falling back to direct" );
            {
                next_platform_mutex_guard( &client->route_manager_mutex );
                next_route_manager_fallback_to_direct( client->route_manager, NEXT_FLAGS_ROUTE_UPDATE_TIMED_OUT );
            }
            client->counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT]++;
            client->fallback_to_direct = true;
        }
    }
}

void next_client_internal_update_route_manager( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    next_assert( !next_global_config.disable_network_next );

    if ( !client->upgraded )
        return;

    if ( client->fallback_to_direct )
        return;

    next_address_t route_request_to;
    next_address_t continue_request_to;

    int route_request_packet_bytes;
    int continue_request_packet_bytes;

    uint8_t route_request_packet_data[NEXT_MAX_PACKET_BYTES];
    uint8_t continue_request_packet_data[NEXT_MAX_PACKET_BYTES];

    bool send_route_request = false;
    bool send_continue_request = false;
    {
        next_platform_mutex_guard( &client->route_manager_mutex );
        send_route_request = next_route_manager_send_route_request( client->route_manager, &route_request_to, route_request_packet_data, &route_request_packet_bytes );
        send_continue_request = next_route_manager_send_continue_request( client->route_manager, &continue_request_to, continue_request_packet_data, &continue_request_packet_bytes );
    }

    if ( send_route_request )
    {
        char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client sent route request to relay: %s", next_address_to_string( &route_request_to, buffer ) );

#if NEXT_SPIKE_TRACKING
        double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

        next_platform_socket_send_packet( client->socket, &route_request_to, route_request_packet_data, route_request_packet_bytes );

#if NEXT_SPIKE_TRACKING
        double finish_time = next_platform_time();
        if ( finish_time - start_time > 0.001 )
        {
            next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
        }
#endif // #if NEXT_SPIKE_TRACKING
    }

    if ( send_continue_request )
    {
        char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client sent continue request to relay: %s", next_address_to_string( &continue_request_to, buffer ) );

#if NEXT_SPIKE_TRACKING
        double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

        next_platform_socket_send_packet( client->socket, &continue_request_to, continue_request_packet_data, continue_request_packet_bytes );

#if NEXT_SPIKE_TRACKING
        double finish_time = next_platform_time();
        if ( finish_time - start_time > 0.001 )
        {
            next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
        }
#endif // #if NEXT_SPIKE_TRACKING
    }
}

void next_client_internal_update_upgrade_response( next_client_internal_t * client )
{
    next_client_internal_verify_sentinels( client );

    next_assert( !next_global_config.disable_network_next );

    if ( client->fallback_to_direct )
        return;

    if ( !client->sending_upgrade_response )
        return;

    const double current_time = next_platform_time();

    if ( client->last_upgrade_response_send_time + 1.0 > current_time )
        return;

    next_assert( client->upgrade_response_packet_bytes > 0 );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( client->socket, &client->server_address, client->upgrade_response_packet_data, client->upgrade_response_packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING

    next_printf( NEXT_LOG_LEVEL_DEBUG, "client sent cached upgrade response packet to server" );

    client->last_upgrade_response_send_time = current_time;

    if ( client->upgrade_response_start_time + 5.0 <= current_time )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "upgrade response timed out" );
        {
            next_platform_mutex_guard( &client->route_manager_mutex );
            next_route_manager_fallback_to_direct( client->route_manager, NEXT_FLAGS_UPGRADE_RESPONSE_TIMED_OUT );
        }
        client->fallback_to_direct = true;
    }
}

static void next_client_internal_update( next_client_internal_t * client )
{
    if ( next_global_config.disable_network_next )
        return;

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_client_internal_update_direct_pings( client );

    next_client_internal_update_next_pings( client );

    next_client_internal_send_pings_to_near_relays( client );

    next_client_internal_update_stats( client );

    next_client_internal_update_fallback_to_direct( client );

    next_client_internal_update_route_manager( client );

    next_client_internal_update_upgrade_response( client );

#if NEXT_SPIKE_TRACKING

    double finish_time = next_platform_time();

    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_client_internal_update spike %.2f milliseconds", ( finish_time - start_time ) * 1000.0 );
    }

#endif // #if NEXT_SPIKE_TRACKING
}

static void next_client_internal_thread_function( void * context )
{
    next_client_internal_t * client = (next_client_internal_t*) context;

    next_assert( client );

    bool quit = false;

    double last_update_time = next_platform_time();

    while ( !quit )
    {
        next_client_internal_block_and_receive_packet( client );

        if ( next_platform_time() > last_update_time + 0.01 )
        {
            next_client_internal_update( client );

            quit = next_client_internal_pump_commands( client );

            last_update_time = next_platform_time();
        }
    }
}

// ---------------------------------------------------------------

struct next_client_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int state;
    bool ready;
    bool upgraded;
    bool fallback_to_direct;
    uint8_t open_session_sequence;
    uint8_t current_magic[8];
    uint16_t bound_port;
    uint64_t session_id;
    next_address_t server_address;
    next_address_t client_external_address;
    next_client_internal_t * internal;
    next_platform_thread_t * thread;
    void (*packet_received_callback)( next_client_t * client, void * context, const struct next_address_t * from, const uint8_t * packet_data, int packet_bytes );

    NEXT_DECLARE_SENTINEL(1)

    next_client_stats_t client_stats;

    NEXT_DECLARE_SENTINEL(2)

    next_bandwidth_limiter_t direct_send_bandwidth;
    next_bandwidth_limiter_t direct_receive_bandwidth;
    next_bandwidth_limiter_t next_send_bandwidth;
    next_bandwidth_limiter_t next_receive_bandwidth;

    NEXT_DECLARE_SENTINEL(3)

    uint64_t counters[NEXT_CLIENT_COUNTER_MAX];

    NEXT_DECLARE_SENTINEL(4)
};

void next_client_initialize_sentinels( next_client_t * client )
{
    (void) client;
    next_assert( client );
    NEXT_INITIALIZE_SENTINEL( client, 0 )
    NEXT_INITIALIZE_SENTINEL( client, 1 )
    NEXT_INITIALIZE_SENTINEL( client, 2 )
    NEXT_INITIALIZE_SENTINEL( client, 3 )
    NEXT_INITIALIZE_SENTINEL( client, 4 )
}

void next_client_verify_sentinels( next_client_t * client )
{
    (void) client;
    next_assert( client );
    NEXT_VERIFY_SENTINEL( client, 0 )
    NEXT_VERIFY_SENTINEL( client, 1 )
    NEXT_VERIFY_SENTINEL( client, 2 )
    NEXT_VERIFY_SENTINEL( client, 3 )
    NEXT_VERIFY_SENTINEL( client, 4 )
}

void next_client_destroy( next_client_t * client );

next_client_t * next_client_create( void * context, const char * bind_address, void (*packet_received_callback)( next_client_t * client, void * context, const struct next_address_t * from, const uint8_t * packet_data, int packet_bytes ) )
{
    next_assert( bind_address );
    next_assert( packet_received_callback );

    next_client_t * client = (next_client_t*) next_malloc( context, sizeof(next_client_t) );
    if ( !client )
        return NULL;

    memset( client, 0, sizeof( next_client_t) );

    next_client_initialize_sentinels( client );

    client->context = context;
    client->packet_received_callback = packet_received_callback;

    client->internal = next_client_internal_create( client->context, bind_address );
    if ( !client->internal )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create internal client" );
        next_client_destroy( client );
        return NULL;
    }

    client->bound_port = client->internal->bound_port;

    client->thread = next_platform_thread_create( client->context, next_client_internal_thread_function, client->internal );
    next_assert( client->thread );
    if ( !client->thread )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not create thread" );
        next_client_destroy( client );
        return NULL;
    }

    // todo
    /*
    if ( next_platform_thread_high_priority( client->thread ) )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "client increased thread priority" );
    }
    */

    next_bandwidth_limiter_reset( &client->direct_send_bandwidth );
    next_bandwidth_limiter_reset( &client->direct_receive_bandwidth );
    next_bandwidth_limiter_reset( &client->next_send_bandwidth );
    next_bandwidth_limiter_reset( &client->next_receive_bandwidth );

    next_client_verify_sentinels( client );

    return client;
}

uint16_t next_client_port( next_client_t * client )
{
    next_client_verify_sentinels( client );

    return client->bound_port;
}

void next_client_destroy( next_client_t * client )
{
    next_client_verify_sentinels( client );

    if ( client->thread )
    {
        next_client_command_destroy_t * command = (next_client_command_destroy_t*) next_malloc( client->context, sizeof( next_client_command_destroy_t ) );
        if ( !command )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "client destroy failed. could not create destroy command" );
            return;
        }
        command->type = NEXT_CLIENT_COMMAND_DESTROY;
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "client sent NEXT_CLIENT_COMMAND_DESTROY" );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &client->internal->command_mutex );
            next_queue_push( client->internal->command_queue, command );
        }

        next_platform_thread_join( client->thread );
        next_platform_thread_destroy( client->thread );
    }

    if ( client->internal )
    {
        next_client_internal_destroy( client->internal );
    }

    next_clear_and_free( client->context, client, sizeof(next_client_t) );
}

void next_client_open_session( next_client_t * client, const char * server_address_string )
{
    next_client_verify_sentinels( client );

    next_assert( client->internal );

    next_client_close_session( client );

    next_address_t server_address;
    if ( next_address_parse( &server_address, server_address_string ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client open session failed. could not parse server address: %s", server_address_string );
        client->state = NEXT_CLIENT_STATE_ERROR;
        return;
    }

    next_client_command_open_session_t * command = (next_client_command_open_session_t*) next_malloc( client->context, sizeof( next_client_command_open_session_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client open session failed. could not create open session command" );
        client->state = NEXT_CLIENT_STATE_ERROR;
        return;
    }

    command->type = NEXT_CLIENT_COMMAND_OPEN_SESSION;
    command->server_address = server_address;

    {
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "client sent NEXT_CLIENT_COMMAND_OPEN_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &client->internal->command_mutex );
        next_queue_push( client->internal->command_queue, command );
    }

    client->state = NEXT_CLIENT_STATE_OPEN;
    client->server_address = server_address;
    client->open_session_sequence++;
}

bool next_client_is_session_open( next_client_t * client )
{
    next_client_verify_sentinels( client );

    return client->state == NEXT_CLIENT_STATE_OPEN;
}

int next_client_state( next_client_t * client )
{
    next_client_verify_sentinels( client );

    return client->state;
}

void next_client_close_session( next_client_t * client )
{
    next_client_verify_sentinels( client );

    next_assert( client->internal );

    next_client_command_close_session_t * command = (next_client_command_close_session_t*) next_malloc( client->context, sizeof( next_client_command_close_session_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client close session failed. could not create close session command" );
        client->state = NEXT_CLIENT_STATE_ERROR;
        return;
    }

    command->type = NEXT_CLIENT_COMMAND_CLOSE_SESSION;
    {
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "client sent NEXT_CLIENT_COMMAND_CLOSE_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &client->internal->command_mutex );
        next_queue_push( client->internal->command_queue, command );
    }

    client->ready = false;
    client->upgraded = false;
    client->fallback_to_direct = false;
    client->session_id = 0;
    memset( &client->client_stats, 0, sizeof(next_client_stats_t ) );
    memset( &client->server_address, 0, sizeof(next_address_t) );
    memset( &client->client_external_address, 0, sizeof(next_address_t) );
    next_bandwidth_limiter_reset( &client->direct_send_bandwidth );
    next_bandwidth_limiter_reset( &client->direct_receive_bandwidth );
    next_bandwidth_limiter_reset( &client->next_send_bandwidth );
    next_bandwidth_limiter_reset( &client->next_receive_bandwidth );
    client->state = NEXT_CLIENT_STATE_CLOSED;
    memset( client->current_magic, 0, sizeof(client->current_magic) );
}

void next_client_update( next_client_t * client )
{
    next_client_verify_sentinels( client );

#if NEXT_SPIKE_TRACKING
    next_printf( NEXT_LOG_LEVEL_SPAM, "next_client_update" );
#endif // #if NEXT_SPIKE_TRACKING

    while ( true )
    {
        void * entry = NULL;
        {
            next_platform_mutex_guard( &client->internal->notify_mutex );
            entry = next_queue_pop( client->internal->notify_queue );
        }

        if ( entry == NULL )
            break;

        next_client_notify_t * notify = (next_client_notify_t*) entry;

        switch ( notify->type )
        {
            case NEXT_CLIENT_NOTIFY_PACKET_RECEIVED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "NEXT_CLIENT_NOTIFY_STATS_UPGRADED" );
#endif // #if NEXT_SPIKE_TRACKING

                next_client_notify_packet_received_t * packet_received = (next_client_notify_packet_received_t*) notify;

#if NEXT_SPIKE_TRACKING
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_SPAM, "client calling packet received callback: from = %s, payload = %d bytes", next_address_to_string( &client->server_address, address_buffer ), packet_received->payload_bytes );
#endif // #if NEXT_SPIKE_TRACKING

                client->packet_received_callback( client, client->context, &client->server_address, packet_received->payload_data, packet_received->payload_bytes );

                const int wire_packet_bits = next_wire_packet_bits( packet_received->payload_bytes );

                next_bandwidth_limiter_add_packet( &client->direct_receive_bandwidth, next_platform_time(), 0, wire_packet_bits );

                double direct_kbps_down = next_bandwidth_limiter_usage_kbps( &client->direct_receive_bandwidth, next_platform_time() );

                {
                    next_platform_mutex_guard( &client->internal->direct_bandwidth_mutex );
                    client->internal->direct_bandwidth_usage_kbps_down = direct_kbps_down;
                }

                if ( !packet_received->direct )
                {
                    int envelope_kbps_down;
                    {
                        next_platform_mutex_guard( &client->internal->next_bandwidth_mutex );
                        envelope_kbps_down = client->internal->next_bandwidth_envelope_kbps_down;
                    }

                    next_bandwidth_limiter_add_packet( &client->next_receive_bandwidth, next_platform_time(), envelope_kbps_down, wire_packet_bits );

                    double next_kbps_down = next_bandwidth_limiter_usage_kbps( &client->next_receive_bandwidth, next_platform_time() );

                    {
                        next_platform_mutex_guard( &client->internal->next_bandwidth_mutex );
                        client->internal->next_bandwidth_usage_kbps_down = next_kbps_down;
                    }
                }
            }
            break;

            case NEXT_CLIENT_NOTIFY_UPGRADED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "NEXT_CLIENT_NOTIFY_STATS_UPGRADED" );
#endif // #if NEXT_SPIKE_TRACKING
                next_client_notify_upgraded_t * upgraded = (next_client_notify_upgraded_t*) notify;
                client->upgraded = true;
                client->session_id = upgraded->session_id;
                client->client_external_address = upgraded->client_external_address;
                memcpy( client->current_magic, upgraded->current_magic, 8 );
                next_printf( NEXT_LOG_LEVEL_INFO, "client upgraded to session %" PRIx64, client->session_id );
            }
            break;

            case NEXT_CLIENT_NOTIFY_STATS_UPDATED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "NEXT_CLIENT_NOTIFY_STATS_UPDATED" );
#endif // #if NEXT_SPIKE_TRACKING
                next_client_notify_stats_updated_t * stats_updated = (next_client_notify_stats_updated_t*) notify;
                client->client_stats = stats_updated->stats;
                client->fallback_to_direct = stats_updated->fallback_to_direct;
            }
            break;

            case NEXT_CLIENT_NOTIFY_MAGIC_UPDATED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "NEXT_CLIENT_NOTIFY_MAGIC_UPDATED" );
#endif // #if NEXT_SPIKE_TRACKING
                next_client_notify_magic_updated_t * magic_updated = (next_client_notify_magic_updated_t*) notify;
                memcpy( client->current_magic, magic_updated->current_magic, 8 );
            }
            break;

            case NEXT_CLIENT_NOTIFY_READY:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "NEXT_CLIENT_NOTIFY_READY" );
#endif // #if NEXT_SPIKE_TRACKING
                client->ready = true;
            }
            break;

            default: break;
        }

        next_free( client->context, entry );
    }
}

bool next_client_ready( next_client_t * client )
{
    next_assert( client );
    return ( next_global_config.disable_network_next || client->ready ) ? true : false;
}

bool next_client_fallback_to_direct( struct next_client_t * client )
{
    next_assert( client );
    return client->client_stats.fallback_to_direct;
}

void next_client_send_packet( next_client_t * client, const uint8_t * packet_data, int packet_bytes )
{
    next_client_verify_sentinels( client );

    next_assert( client->internal );
    next_assert( client->internal->socket );
    next_assert( packet_bytes > 0 );

    if ( client->state != NEXT_CLIENT_STATE_OPEN )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client can't send packet because no session is open" );
        return;
    }

    if ( packet_bytes > NEXT_MAX_PACKET_BYTES - 1 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client can't send packet because packet is too large" );
        return;
    }

    if ( next_global_config.disable_network_next || client->fallback_to_direct )
    {
        next_client_send_packet_direct( client, packet_data, packet_bytes );
        return;
    }

#if NEXT_DEVELOPMENT
    if ( next_packet_loss && ( rand() % 10 ) == 0 )
        return;
#endif // #if NEXT_DEVELOPMENT

    if ( client->upgraded && packet_bytes <= NEXT_MTU )
    {
        uint64_t send_sequence = 0;
        bool send_over_network_next = false;
        {
            next_platform_mutex_guard( &client->internal->route_manager_mutex );
            send_sequence = next_route_manager_next_send_sequence( client->internal->route_manager );
            send_over_network_next = next_route_manager_has_network_next_route( client->internal->route_manager );
        }

        bool send_direct = !send_over_network_next;
        bool multipath = client->client_stats.multipath;
        if ( send_over_network_next && multipath )
        {
            send_direct = true;
        }

        // track direct send bandwidth

        const int wire_packet_bits = next_wire_packet_bits( packet_bytes );

        next_bandwidth_limiter_add_packet( &client->direct_send_bandwidth, next_platform_time(), 0, wire_packet_bits );

        double direct_usage_kbps_up = next_bandwidth_limiter_usage_kbps( &client->direct_send_bandwidth, next_platform_time() );

        {
            next_platform_mutex_guard( &client->internal->direct_bandwidth_mutex );
            client->internal->direct_bandwidth_usage_kbps_up = direct_usage_kbps_up;
        }

        // track next send backend and don't send over network next if we're over the bandwidth budget

        if ( send_over_network_next )
        {
            int next_envelope_kbps_up;
            {
                next_platform_mutex_guard( &client->internal->next_bandwidth_mutex );
                next_envelope_kbps_up = client->internal->next_bandwidth_envelope_kbps_up;
            }

            bool over_budget = next_bandwidth_limiter_add_packet( &client->next_send_bandwidth, next_platform_time(), next_envelope_kbps_up, wire_packet_bits );

            double next_usage_kbps_up = next_bandwidth_limiter_usage_kbps( &client->next_send_bandwidth, next_platform_time() );

            {
                next_platform_mutex_guard( &client->internal->next_bandwidth_mutex );
                client->internal->next_bandwidth_usage_kbps_up = next_usage_kbps_up;
                if ( over_budget )
                    client->internal->next_bandwidth_over_limit = true;
            }

            if ( over_budget )
            {
                next_printf( NEXT_LOG_LEVEL_WARN, "client exceeded bandwidth budget (%d kbps)", next_envelope_kbps_up );
                send_over_network_next = false;
                send_direct = true;
            }
        }

        if ( send_over_network_next )
        {
            // send over network next

            int next_packet_bytes = 0;
            next_address_t next_to;
            uint8_t next_packet_data[NEXT_MAX_PACKET_BYTES];

            bool result = false;
            {
                next_platform_mutex_guard( &client->internal->route_manager_mutex );
                result = next_route_manager_prepare_send_packet( client->internal->route_manager, send_sequence, &next_to, packet_data, packet_bytes, next_packet_data, &next_packet_bytes, client->current_magic, &client->client_external_address );
            }

            if ( result )
            {
#if NEXT_SPIKE_TRACKING
                double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

                next_platform_socket_send_packet( client->internal->socket, &next_to, next_packet_data, next_packet_bytes );

#if NEXT_SPIKE_TRACKING
                double finish_time = next_platform_time();
                if ( finish_time - start_time > 0.001 )
                {
                    next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
                }
#endif // #if NEXT_SPIKE_TRACKING

                client->counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]++;
            }
            else
            {
                // could not send over network next
                send_direct = true;
            }
        }

        if ( send_direct )
        {
            // send direct from client to server

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( &client->client_external_address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &client->server_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t direct_packet_data[NEXT_MAX_PACKET_BYTES];

            const int direct_packet_bytes = next_write_direct_packet( direct_packet_data, client->open_session_sequence, send_sequence, packet_data, packet_bytes, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

            next_assert( direct_packet_bytes >= 0 );

            next_assert( next_basic_packet_filter( direct_packet_data, direct_packet_bytes ) );
            next_assert( next_advanced_packet_filter( direct_packet_data, client->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, direct_packet_bytes ) );

            (void) direct_packet_data;
            (void) direct_packet_bytes;

#if NEXT_SPIKE_TRACKING 
            double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

            next_platform_socket_send_packet( client->internal->socket, &client->server_address, direct_packet_data, direct_packet_bytes );

#if NEXT_SPIKE_TRACKING
            double finish_time = next_platform_time();
            if ( finish_time - start_time > 0.001 )
            {
                next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
            }
#endif // #if NEXT_SPIKE_TRACKING

            client->counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]++;
        }

        client->internal->packets_sent++;
    }
    else
    {
        // passthrough packet

        next_client_send_packet_direct( client, packet_data, packet_bytes );
    }
}

void next_client_send_packet_direct( next_client_t * client, const uint8_t * packet_data, int packet_bytes )
{
    next_client_verify_sentinels( client );

    next_assert( client->internal );
    next_assert( client->internal->socket );
    next_assert( packet_bytes > 0 );

    if ( client->state != NEXT_CLIENT_STATE_OPEN )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "client can't send packet because no session is open" );
        return;
    }

    if ( packet_bytes <= 0 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client can't send packet because packet size <= 0" );
        return;
    }

    if ( packet_bytes > NEXT_MAX_PACKET_BYTES - 1 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client can't send packet because packet is too large" );
        return;
    }

    uint8_t buffer[NEXT_MAX_PACKET_BYTES];
    buffer[0] = NEXT_PASSTHROUGH_PACKET;
    memcpy( buffer + 1, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( client->internal->socket, &client->server_address, buffer, packet_bytes + 1 );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING

    client->counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH]++;

    client->internal->packets_sent++;
}

void next_client_send_packet_raw( next_client_t * client, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes )
{
    next_client_verify_sentinels( client );

    next_assert( client->internal );
    next_assert( client->internal->socket );
    next_assert( to_address );
    next_assert( packet_bytes > 0 );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( client->internal->socket, to_address, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING
}

void next_client_report_session( next_client_t * client )
{
    next_client_verify_sentinels( client );

    next_client_command_report_session_t * command = (next_client_command_report_session_t*) next_malloc( client->context, sizeof( next_client_command_report_session_t ) );

    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "report session failed. could not create report session command" );
        return;
    }

    command->type = NEXT_CLIENT_COMMAND_REPORT_SESSION;
    {
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "client sent NEXT_CLIENT_COMMAND_REPORT_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &client->internal->command_mutex );
        next_queue_push( client->internal->command_queue, command );
    }
}

uint64_t next_client_session_id( next_client_t * client )
{
    next_client_verify_sentinels( client );

    return client->session_id;
}

const next_client_stats_t * next_client_stats( next_client_t * client )
{
    next_client_verify_sentinels( client );

    return &client->client_stats;
}

const next_address_t * next_client_server_address( next_client_t * client )
{
    next_assert( client );
    return &client->server_address;
}

void next_client_counters( next_client_t * client, uint64_t * counters )
{
    next_client_verify_sentinels( client );
    memcpy( counters, client->counters, sizeof(uint64_t) * NEXT_CLIENT_COUNTER_MAX );
    for ( int i = 0; i < NEXT_CLIENT_COUNTER_MAX; ++i )
        counters[i] += client->internal->counters[i];
}

// ---------------------------------------------------------------

struct next_pending_session_entry_t
{
    NEXT_DECLARE_SENTINEL(0)

    next_address_t address;
    uint64_t session_id;
    uint64_t user_hash;
    // todo: internal_events
    double upgrade_time;
    double last_packet_send_time;
    uint8_t private_key[NEXT_CRYPTO_SECRETBOX_KEYBYTES];
    uint8_t upgrade_token[NEXT_UPGRADE_TOKEN_BYTES];

    NEXT_DECLARE_SENTINEL(1)
};

void next_pending_session_entry_initialize_sentinels( next_pending_session_entry_t * entry )
{
    (void) entry;
    next_assert( entry );
    NEXT_INITIALIZE_SENTINEL( entry, 0 )
    NEXT_INITIALIZE_SENTINEL( entry, 1 )
}

void next_pending_session_entry_verify_sentinels( next_pending_session_entry_t * entry )
{
    (void) entry;
    next_assert( entry );
    NEXT_VERIFY_SENTINEL( entry, 0 )
    NEXT_VERIFY_SENTINEL( entry, 1 )
}

struct next_pending_session_manager_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int size;
    int max_entry_index;
    next_address_t * addresses;
    next_pending_session_entry_t * entries;

    NEXT_DECLARE_SENTINEL(1)
};

void next_pending_session_manager_initialize_sentinels( next_pending_session_manager_t * session_manager )
{
    (void) session_manager;
    next_assert( session_manager );
    NEXT_INITIALIZE_SENTINEL( session_manager, 0 )
    NEXT_INITIALIZE_SENTINEL( session_manager, 1 )
}

void next_pending_session_manager_verify_sentinels( next_pending_session_manager_t * session_manager )
{
    (void) session_manager;
#if NEXT_ENABLE_MEMORY_CHECKS
    next_assert( session_manager );
    NEXT_VERIFY_SENTINEL( session_manager, 0 )
    NEXT_VERIFY_SENTINEL( session_manager, 1 )
    const int size = session_manager->size;
    for ( int i = 0; i < size; ++i )
    {
        if ( session_manager->addresses[i].type != 0 )
        {
            next_pending_session_entry_verify_sentinels( &session_manager->entries[i] );
        }
    }
#endif // #if NEXT_ENABLE_MEMORY_CHECKS
}

void next_pending_session_manager_destroy( next_pending_session_manager_t * pending_session_manager );

next_pending_session_manager_t * next_pending_session_manager_create( void * context, int initial_size )
{
    next_pending_session_manager_t * pending_session_manager = (next_pending_session_manager_t*) next_malloc( context, sizeof(next_pending_session_manager_t) );

    next_assert( pending_session_manager );
    if ( !pending_session_manager )
        return NULL;

    memset( pending_session_manager, 0, sizeof(next_pending_session_manager_t) );

    next_pending_session_manager_initialize_sentinels( pending_session_manager );

    pending_session_manager->context = context;
    pending_session_manager->size = initial_size;
    pending_session_manager->addresses = (next_address_t*) next_malloc( context, initial_size * sizeof(next_address_t) );
    pending_session_manager->entries = (next_pending_session_entry_t*) next_malloc( context, initial_size * sizeof(next_pending_session_entry_t) );

    next_assert( pending_session_manager->addresses );
    next_assert( pending_session_manager->entries );

    if ( pending_session_manager->addresses == NULL || pending_session_manager->entries == NULL )
    {
        next_pending_session_manager_destroy( pending_session_manager );
        return NULL;
    }

    memset( pending_session_manager->addresses, 0, initial_size * sizeof(next_address_t) );
    memset( pending_session_manager->entries, 0, initial_size * sizeof(next_pending_session_entry_t) );

    for ( int i = 0; i < initial_size; i++ )
        next_pending_session_entry_initialize_sentinels( &pending_session_manager->entries[i] );

    next_pending_session_manager_verify_sentinels( pending_session_manager );

    return pending_session_manager;
}

void next_pending_session_manager_destroy( next_pending_session_manager_t * pending_session_manager )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    next_free( pending_session_manager->context, pending_session_manager->addresses );
    next_free( pending_session_manager->context, pending_session_manager->entries );

    next_clear_and_free( pending_session_manager->context, pending_session_manager, sizeof(next_pending_session_manager_t) );
}

bool next_pending_session_manager_expand( next_pending_session_manager_t * pending_session_manager )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    int new_size = pending_session_manager->size * 2;

    next_address_t * new_addresses = (next_address_t*) next_malloc( pending_session_manager->context, new_size * sizeof(next_address_t) );

    next_pending_session_entry_t * new_entries = (next_pending_session_entry_t*) next_malloc( pending_session_manager->context, new_size * sizeof(next_pending_session_entry_t) );

    next_assert( pending_session_manager->addresses );
    next_assert( pending_session_manager->entries );

    if ( pending_session_manager->addresses == NULL || pending_session_manager->entries == NULL )
    {
        next_free( pending_session_manager->context, new_addresses );
        next_free( pending_session_manager->context, new_entries );
        return false;
    }

    memset( new_addresses, 0, new_size * sizeof(next_address_t) );
    memset( new_entries, 0, new_size * sizeof(next_pending_session_entry_t) );

    for ( int i = 0; i < new_size; ++i )
        next_pending_session_entry_initialize_sentinels( &new_entries[i] );

    int index = 0;
    const int current_size = pending_session_manager->size;
    for ( int i = 0; i < current_size; ++i )
    {
        if ( pending_session_manager->addresses[i].type != NEXT_ADDRESS_NONE )
        {
            memcpy( &new_addresses[index], &pending_session_manager->addresses[i], sizeof(next_address_t) );
            memcpy( &new_entries[index], &pending_session_manager->entries[i], sizeof(next_pending_session_entry_t) );
            index++;
        }
    }

    next_free( pending_session_manager->context, pending_session_manager->addresses );
    next_free( pending_session_manager->context, pending_session_manager->entries );

    pending_session_manager->addresses = new_addresses;
    pending_session_manager->entries = new_entries;
    pending_session_manager->size = new_size;
    pending_session_manager->max_entry_index = index - 1;

    next_pending_session_manager_verify_sentinels( pending_session_manager );

    return true;
}

next_pending_session_entry_t * next_pending_session_manager_add( next_pending_session_manager_t * pending_session_manager, const next_address_t * address, uint64_t session_id, const uint8_t * private_key, const uint8_t * upgrade_token, double current_time )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    next_assert( session_id != 0 );
    next_assert( address );
    next_assert( address->type != NEXT_ADDRESS_NONE );

    // first scan existing entries and see if we can insert there

    const int size = pending_session_manager->size;

    for ( int i = 0; i < size; ++i )
    {
        if ( pending_session_manager->addresses[i].type == NEXT_ADDRESS_NONE )
        {
            pending_session_manager->addresses[i] = *address;
            next_pending_session_entry_t * entry = &pending_session_manager->entries[i];
            entry->address = *address;
            entry->session_id = session_id;
            entry->upgrade_time = current_time;
            entry->last_packet_send_time = -1000.0;
            memcpy( entry->private_key, private_key, NEXT_CRYPTO_SECRETBOX_KEYBYTES );
            memcpy( entry->upgrade_token, upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );
            if ( i > pending_session_manager->max_entry_index )
            {
                pending_session_manager->max_entry_index = i;
            }
            return entry;
        }
    }

    // ok, we need to grow, expand and add at the end (expand compacts existing entries)

    if ( !next_pending_session_manager_expand( pending_session_manager ) )
        return NULL;

    const int i = ++pending_session_manager->max_entry_index;
    pending_session_manager->addresses[i] = *address;
    next_pending_session_entry_t * entry = &pending_session_manager->entries[i];
    entry->address = *address;
    entry->session_id = session_id;
    entry->upgrade_time = current_time;
    entry->last_packet_send_time = -1000.0;
    memcpy( entry->private_key, private_key, NEXT_CRYPTO_SECRETBOX_KEYBYTES );
    memcpy( entry->upgrade_token, upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );

    next_pending_session_manager_verify_sentinels( pending_session_manager );

    return entry;
}

void next_pending_session_manager_remove_at_index( next_pending_session_manager_t * pending_session_manager, int index )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    next_assert( index >= 0 );
    next_assert( index <= pending_session_manager->max_entry_index );

    const int max_index = pending_session_manager->max_entry_index;

    pending_session_manager->addresses[index].type = NEXT_ADDRESS_NONE;

    if ( index == max_index )
    {
        while ( index > 0 && pending_session_manager->addresses[index].type == NEXT_ADDRESS_NONE )
        {
            index--;
        }
        pending_session_manager->max_entry_index = index;
    }
}

void next_pending_session_manager_remove_by_address( next_pending_session_manager_t * pending_session_manager, const next_address_t * address )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    next_assert( address );

    const int max_index = pending_session_manager->max_entry_index;

    for ( int i = 0; i <= max_index; ++i )
    {
        if ( next_address_equal( address, &pending_session_manager->addresses[i] ) == 1 )
        {
            next_pending_session_manager_remove_at_index( pending_session_manager, i );
            return;
        }
    }
}

next_pending_session_entry_t * next_pending_session_manager_find( next_pending_session_manager_t * pending_session_manager, const next_address_t * address )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    next_assert( address );

    const int max_index = pending_session_manager->max_entry_index;

    for ( int i = 0; i <= max_index; ++i )
    {
        if ( next_address_equal( address, &pending_session_manager->addresses[i] ) == 1 )
        {
            return &pending_session_manager->entries[i];
        }
    }

    return NULL;
}

int next_pending_session_manager_num_entries( next_pending_session_manager_t * pending_session_manager )
{
    next_pending_session_manager_verify_sentinels( pending_session_manager );

    int num_entries = 0;

    const int max_index = pending_session_manager->max_entry_index;

    for ( int i = 0; i <= max_index; ++i )
    {
        if ( pending_session_manager->addresses[i].type != 0 )
        {
            num_entries++;
        }
    }

    return num_entries;
}

// ---------------------------------------------------------------

struct next_proxy_session_entry_t
{
    NEXT_DECLARE_SENTINEL(0)

    next_address_t address;
    uint64_t session_id;

    NEXT_DECLARE_SENTINEL(1)

    next_bandwidth_limiter_t send_bandwidth;

    NEXT_DECLARE_SENTINEL(2)
};

void next_proxy_session_entry_initialize_sentinels( next_proxy_session_entry_t * entry )
{
    (void) entry;
    next_assert( entry );
    NEXT_INITIALIZE_SENTINEL( entry, 0 )
    NEXT_INITIALIZE_SENTINEL( entry, 1 )
    NEXT_INITIALIZE_SENTINEL( entry, 2 )
}

void next_proxy_session_entry_verify_sentinels( next_proxy_session_entry_t * entry )
{
    (void) entry;
    next_assert( entry );
    NEXT_VERIFY_SENTINEL( entry, 0 )
    NEXT_VERIFY_SENTINEL( entry, 1 )
    NEXT_VERIFY_SENTINEL( entry, 2 )
}

struct next_proxy_session_manager_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int size;
    int max_entry_index;
    next_address_t * addresses;
    next_proxy_session_entry_t * entries;

    NEXT_DECLARE_SENTINEL(1)
};

void next_proxy_session_manager_initialize_sentinels( next_proxy_session_manager_t * session_manager )
{
    (void) session_manager;
    next_assert( session_manager );
    NEXT_INITIALIZE_SENTINEL( session_manager, 0 )
    NEXT_INITIALIZE_SENTINEL( session_manager, 1 )
}

void next_proxy_session_manager_verify_sentinels( next_proxy_session_manager_t * session_manager )
{
    (void) session_manager;
#if NEXT_ENABLE_MEMORY_CHECKS
    next_assert( session_manager );
    NEXT_VERIFY_SENTINEL( session_manager, 0 )
    NEXT_VERIFY_SENTINEL( session_manager, 1 )
    const int size = session_manager->size;
    for ( int i = 0; i < size; ++i )
    {
        if ( session_manager->addresses[i].type != 0 )
        {
            next_proxy_session_entry_verify_sentinels( &session_manager->entries[i] );
        }
    }
#endif // #if NEXT_ENABLE_MEMORY_CHECKS
}

void next_proxy_session_manager_destroy( next_proxy_session_manager_t * session_manager );

next_proxy_session_manager_t * next_proxy_session_manager_create( void * context, int initial_size )
{
    next_proxy_session_manager_t * session_manager = (next_proxy_session_manager_t*) next_malloc( context, sizeof(next_proxy_session_manager_t) );

    next_assert( session_manager );

    if ( !session_manager )
        return NULL;

    memset( session_manager, 0, sizeof(next_proxy_session_manager_t) );

    next_proxy_session_manager_initialize_sentinels( session_manager );

    session_manager->context = context;
    session_manager->size = initial_size;
    session_manager->addresses = (next_address_t*) next_malloc( context, initial_size * sizeof(next_address_t) );
    session_manager->entries = (next_proxy_session_entry_t*) next_malloc( context, initial_size * sizeof(next_proxy_session_entry_t) );

    next_assert( session_manager->addresses );
    next_assert( session_manager->entries );

    if ( session_manager->addresses == NULL || session_manager->entries == NULL )
    {
        next_proxy_session_manager_destroy( session_manager );
        return NULL;
    }

    memset( session_manager->addresses, 0, initial_size * sizeof(next_address_t) );
    memset( session_manager->entries, 0, initial_size * sizeof(next_proxy_session_entry_t) );

    for ( int i = 0; i < initial_size; ++i )
        next_proxy_session_entry_initialize_sentinels( &session_manager->entries[i] );

    next_proxy_session_manager_verify_sentinels( session_manager );

    return session_manager;
}

void next_proxy_session_manager_destroy( next_proxy_session_manager_t * session_manager )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    next_free( session_manager->context, session_manager->addresses );
    next_free( session_manager->context, session_manager->entries );

    next_clear_and_free( session_manager->context, session_manager, sizeof(next_proxy_session_manager_t) );
}

bool next_proxy_session_manager_expand( next_proxy_session_manager_t * session_manager )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    int new_size = session_manager->size * 2;
    next_address_t * new_addresses = (next_address_t*) next_malloc( session_manager->context, new_size * sizeof(next_address_t) );
    next_proxy_session_entry_t * new_entries = (next_proxy_session_entry_t*) next_malloc( session_manager->context, new_size * sizeof(next_proxy_session_entry_t) );

    next_assert( session_manager->addresses );
    next_assert( session_manager->entries );

    if ( session_manager->addresses == NULL || session_manager->entries == NULL )
    {
        next_free( session_manager->context, new_addresses );
        next_free( session_manager->context, new_entries );
        return false;
    }

    memset( new_addresses, 0, new_size * sizeof(next_address_t) );
    memset( new_entries, 0, new_size * sizeof(next_proxy_session_entry_t) );

    for ( int i = 0; i < new_size; ++i )
        next_proxy_session_entry_initialize_sentinels( &new_entries[i] );

    int index = 0;
    const int current_size = session_manager->size;
    for ( int i = 0; i < current_size; ++i )
    {
        if ( session_manager->addresses[i].type != NEXT_ADDRESS_NONE )
        {
            memcpy( &new_addresses[index], &session_manager->addresses[i], sizeof(next_address_t) );
            memcpy( &new_entries[index], &session_manager->entries[i], sizeof(next_proxy_session_entry_t) );
            index++;
        }
    }

    next_free( session_manager->context, session_manager->addresses );
    next_free( session_manager->context, session_manager->entries );

    session_manager->addresses = new_addresses;
    session_manager->entries = new_entries;
    session_manager->size = new_size;
    session_manager->max_entry_index = index - 1;

    next_proxy_session_manager_verify_sentinels( session_manager );

    return true;
}

next_proxy_session_entry_t * next_proxy_session_manager_add( next_proxy_session_manager_t * session_manager, const next_address_t * address, uint64_t session_id )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    next_assert( session_id != 0 );
    next_assert( address );
    next_assert( address->type != NEXT_ADDRESS_NONE );

    // first scan existing entries and see if we can insert there

    const int size = session_manager->size;

    for ( int i = 0; i < size; ++i )
    {
        if ( session_manager->addresses[i].type == NEXT_ADDRESS_NONE )
        {
            session_manager->addresses[i] = *address;
            next_proxy_session_entry_t * entry = &session_manager->entries[i];
            entry->address = *address;
            entry->session_id = session_id;
            next_bandwidth_limiter_reset( &entry->send_bandwidth );
            if ( i > session_manager->max_entry_index )
            {
                session_manager->max_entry_index = i;
            }
            return entry;
        }
    }

    // ok, we need to grow, expand and add at the end (expand compacts existing entries)

    if ( !next_proxy_session_manager_expand( session_manager ) )
        return NULL;

    const int i = ++session_manager->max_entry_index;
    session_manager->addresses[i] = *address;
    next_proxy_session_entry_t * entry = &session_manager->entries[i];
    entry->address = *address;
    entry->session_id = session_id;
    next_bandwidth_limiter_reset( &entry->send_bandwidth );

    next_proxy_session_manager_verify_sentinels( session_manager );

    return entry;
}

void next_proxy_session_manager_remove_at_index( next_proxy_session_manager_t * session_manager, int index )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    next_assert( index >= 0 );
    next_assert( index <= session_manager->max_entry_index );
    const int max_index = session_manager->max_entry_index;
    session_manager->addresses[index].type = NEXT_ADDRESS_NONE;
    if ( index == max_index )
    {
        while ( index > 0 && session_manager->addresses[index].type == NEXT_ADDRESS_NONE )
        {
            index--;
        }
        session_manager->max_entry_index = index;
    }

    next_proxy_session_manager_verify_sentinels( session_manager );
}

void next_proxy_session_manager_remove_by_address( next_proxy_session_manager_t * session_manager, const next_address_t * address )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    next_assert( address );

    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( next_address_equal( address, &session_manager->addresses[i] ) == 1 )
        {
            next_proxy_session_manager_remove_at_index( session_manager, i );
            next_proxy_session_manager_verify_sentinels( session_manager );
            return;
        }
    }
}

next_proxy_session_entry_t * next_proxy_session_manager_find( next_proxy_session_manager_t * session_manager, const next_address_t * address )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    next_assert( address );

    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( next_address_equal( address, &session_manager->addresses[i] ) == 1 )
        {
            return &session_manager->entries[i];
        }
    }

    return NULL;
}

int next_proxy_session_manager_num_entries( next_proxy_session_manager_t * session_manager )
{
    next_proxy_session_manager_verify_sentinels( session_manager );

    int num_entries = 0;

    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( session_manager->addresses[i].type != 0 )
        {
            num_entries++;
        }
    }

    return num_entries;
}

// ---------------------------------------------------------------

struct NextBackendServerInitRequestPacket
{
    int version_major;
    int version_minor;
    int version_patch;
    uint64_t customer_id;
    uint64_t request_id;
    uint64_t datacenter_id;
    char datacenter_name[NEXT_MAX_DATACENTER_NAME_LENGTH];

    NextBackendServerInitRequestPacket()
    {
        version_major = NEXT_VERSION_MAJOR_INT;
        version_minor = NEXT_VERSION_MINOR_INT;
        version_patch = NEXT_VERSION_PATCH_INT;
        customer_id = 0;
        request_id = 0;
        datacenter_id = 0;
        datacenter_name[0] = '\0';
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_bits( stream, version_major, 8 );
        serialize_bits( stream, version_minor, 8 );
        serialize_bits( stream, version_patch, 8 );
        serialize_uint64( stream, customer_id );
        serialize_uint64( stream, request_id );
        serialize_uint64( stream, datacenter_id );
        serialize_string( stream, datacenter_name, NEXT_MAX_DATACENTER_NAME_LENGTH );
        return true;
    }
};

// ---------------------------------------------------------------

struct NextBackendServerInitResponsePacket
{
    uint64_t request_id;
    uint32_t response;
    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    NextBackendServerInitResponsePacket()
    {
        memset( this, 0, sizeof(NextBackendServerInitResponsePacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, request_id );
        serialize_bits( stream, response, 8 );
        serialize_bytes( stream, upcoming_magic, 8 );
        serialize_bytes( stream, current_magic, 8 );
        serialize_bytes( stream, previous_magic, 8 );
        return true;
    }
};

// ---------------------------------------------------------------

struct NextBackendServerUpdateRequestPacket
{
    int version_major;
    int version_minor;
    int version_patch;
    uint64_t customer_id;
    uint64_t request_id;
    uint64_t datacenter_id;
    uint64_t match_id;
    uint32_t num_sessions;
    next_address_t server_address;

    NextBackendServerUpdateRequestPacket()
    {
        version_major = NEXT_VERSION_MAJOR_INT;
        version_minor = NEXT_VERSION_MINOR_INT;
        version_patch = NEXT_VERSION_PATCH_INT;
        customer_id = 0;
        request_id = 0;
        datacenter_id = 0;
        match_id = 0;
        num_sessions = 0;
        memset( &server_address, 0, sizeof(next_address_t) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_bits( stream, version_major, 8 );
        serialize_bits( stream, version_minor, 8 );
        serialize_bits( stream, version_patch, 8 );
        serialize_uint64( stream, customer_id );
        serialize_uint64( stream, request_id );
        serialize_uint64( stream, datacenter_id );
        serialize_uint64( stream, match_id );
        serialize_uint32( stream, num_sessions );
        serialize_address( stream, server_address );
        return true;
    }
};

// ---------------------------------------------------------------

struct NextBackendServerUpdateResponsePacket
{
    uint64_t request_id;
    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    NextBackendServerUpdateResponsePacket()
    {
        memset( this, 0, sizeof(NextBackendServerUpdateResponsePacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, request_id );
        serialize_bytes( stream, upcoming_magic, 8 );
        serialize_bytes( stream, current_magic, 8 );
        serialize_bytes( stream, previous_magic, 8 );
        return true;
    }
};

// ---------------------------------------------------------------

struct NextBackendSessionUpdateRequestPacket
{
    int version_major;
    int version_minor;
    int version_patch;
    uint64_t customer_id;
    uint64_t datacenter_id;
    uint64_t session_id;
    uint32_t slice_number;
    uint32_t retry_number;
    int session_data_bytes;
    uint8_t session_data[NEXT_MAX_SESSION_DATA_BYTES];
    uint8_t session_data_signature[NEXT_CRYPTO_SIGN_BYTES];
    next_address_t client_address;
    next_address_t server_address;
    uint8_t client_route_public_key[NEXT_CRYPTO_BOX_PUBLICKEYBYTES];
    uint8_t server_route_public_key[NEXT_CRYPTO_BOX_PUBLICKEYBYTES];
    uint64_t user_hash;
    int platform_id;
    int connection_type;
    bool next;
    bool reported;
    bool fallback_to_direct;
    bool client_bandwidth_over_limit;
    bool server_bandwidth_over_limit;
    bool client_ping_timed_out;
    bool has_near_relay_pings;
    uint64_t session_events;
    uint64_t internal_events;
    float direct_rtt;
    float direct_jitter;
    float direct_packet_loss;
    float direct_max_packet_loss_seen;
    float next_rtt;
    float next_jitter;
    float next_packet_loss;
    int num_near_relays;
    uint64_t near_relay_ids[NEXT_MAX_NEAR_RELAYS];
    uint8_t near_relay_rtt[NEXT_MAX_NEAR_RELAYS];
    uint8_t near_relay_jitter[NEXT_MAX_NEAR_RELAYS];
    float near_relay_packet_loss[NEXT_MAX_NEAR_RELAYS];
    uint32_t direct_kbps_up;
    uint32_t direct_kbps_down;
    uint32_t next_kbps_up;
    uint32_t next_kbps_down;
    uint64_t packets_sent_client_to_server;
    uint64_t packets_sent_server_to_client;
    uint64_t packets_lost_client_to_server;
    uint64_t packets_lost_server_to_client;
    uint64_t packets_out_of_order_client_to_server;
    uint64_t packets_out_of_order_server_to_client;
    float jitter_client_to_server;
    float jitter_server_to_client;

    void Reset()
    {
        memset( this, 0, sizeof(NextBackendSessionUpdateRequestPacket) );
        version_major = NEXT_VERSION_MAJOR_INT;
        version_minor = NEXT_VERSION_MINOR_INT;
        version_patch = NEXT_VERSION_PATCH_INT;
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_bits( stream, version_major, 8 );
        serialize_bits( stream, version_minor, 8 );
        serialize_bits( stream, version_patch, 8 );

        serialize_uint64( stream, customer_id );
        serialize_uint64( stream, datacenter_id );
        serialize_uint64( stream, session_id );

        serialize_uint32( stream, slice_number );

        serialize_int( stream, retry_number, 0, NEXT_MAX_SESSION_UPDATE_RETRIES );

        serialize_int( stream, session_data_bytes, 0, NEXT_MAX_SESSION_DATA_BYTES );
        if ( session_data_bytes > 0 )
        {
            serialize_bytes( stream, session_data, session_data_bytes );
            serialize_bytes( stream, session_data_signature, NEXT_CRYPTO_SIGN_BYTES );
        }

        serialize_address( stream, client_address );
        serialize_address( stream, server_address );

        serialize_bytes( stream, client_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );
        serialize_bytes( stream, server_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );

        serialize_uint64( stream, user_hash );

        serialize_int( stream, platform_id, NEXT_PLATFORM_UNKNOWN, NEXT_PLATFORM_MAX );

        serialize_int( stream, connection_type, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_MAX );

        serialize_bool( stream, next );
        serialize_bool( stream, reported );
        serialize_bool( stream, fallback_to_direct );
        serialize_bool( stream, client_bandwidth_over_limit );
        serialize_bool( stream, server_bandwidth_over_limit );
        serialize_bool( stream, client_ping_timed_out );
        serialize_bool( stream, has_near_relay_pings );

        bool has_session_events = Stream::IsWriting && session_events != 0;
        bool has_internal_events = Stream::IsWriting && internal_events != 0;
        bool has_lost_packets = Stream::IsWriting && ( packets_lost_client_to_server + packets_lost_server_to_client ) > 0;
        bool has_out_of_order_packets = Stream::IsWriting && ( packets_out_of_order_client_to_server + packets_out_of_order_server_to_client ) > 0;

        serialize_bool( stream, has_session_events );
        serialize_bool( stream, has_internal_events );
        serialize_bool( stream, has_lost_packets );
        serialize_bool( stream, has_out_of_order_packets );

        if ( has_session_events )
        {
            serialize_uint64( stream, session_events );
        }

        if ( has_internal_events )
        {
            serialize_uint64( stream, internal_events );
        }

        serialize_float( stream, direct_rtt );
        serialize_float( stream, direct_jitter );
        serialize_float( stream, direct_packet_loss );
        serialize_float( stream, direct_max_packet_loss_seen );

        if ( next )
        {
            serialize_float( stream, next_rtt );
            serialize_float( stream, next_jitter );
            serialize_float( stream, next_packet_loss );
        }

        if ( has_near_relay_pings )
        {
            serialize_int( stream, num_near_relays, 0, NEXT_MAX_NEAR_RELAYS );

            for ( int i = 0; i < num_near_relays; ++i )
            {
                serialize_uint64( stream, near_relay_ids[i] );
                if ( has_near_relay_pings )
                {
                    serialize_int( stream, near_relay_rtt[i], 0, 255 );
                    serialize_int( stream, near_relay_jitter[i], 0, 255 );
                    serialize_float( stream, near_relay_packet_loss[i] );
                }
            }
        }

        serialize_uint32( stream, direct_kbps_up );
        serialize_uint32( stream, direct_kbps_down );

        if ( next )
        {
            serialize_uint32( stream, next_kbps_up );
            serialize_uint32( stream, next_kbps_down );
        }

        serialize_uint64( stream, packets_sent_client_to_server );
        serialize_uint64( stream, packets_sent_server_to_client );

        if ( has_lost_packets )
        {
            serialize_uint64( stream, packets_lost_client_to_server );
            serialize_uint64( stream, packets_lost_server_to_client );
        }

        if ( has_out_of_order_packets )
        {
            serialize_uint64( stream, packets_out_of_order_client_to_server );
            serialize_uint64( stream, packets_out_of_order_server_to_client );
        }

        serialize_float( stream, jitter_client_to_server );
        serialize_float( stream, jitter_server_to_client );

        return true;
    }
};

// ---------------------------------------------------------------

struct NextBackendMatchDataRequestPacket
{
    int version_major;
    int version_minor;
    int version_patch;
    uint64_t customer_id;
    next_address_t server_address;
    uint64_t datacenter_id;
    uint64_t user_hash;
    uint64_t session_id;
    uint32_t retry_number;
    uint64_t match_id;
    int num_match_values;
    double match_values[NEXT_MAX_MATCH_VALUES];

    void Reset()
    {
        memset( this, 0, sizeof(NextBackendMatchDataRequestPacket) );
        version_major = NEXT_VERSION_MAJOR_INT;
        version_minor = NEXT_VERSION_MINOR_INT;
        version_patch = NEXT_VERSION_PATCH_INT;
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_bits( stream, version_major, 8 );
        serialize_bits( stream, version_minor, 8 );
        serialize_bits( stream, version_patch, 8 );
        serialize_uint64( stream, customer_id );
        serialize_address( stream, server_address );
        serialize_uint64( stream, datacenter_id );
        serialize_uint64( stream, user_hash );
        serialize_uint64( stream, session_id );
        serialize_uint32( stream, retry_number );
        serialize_uint64( stream, match_id );

        bool has_match_values = Stream::IsWriting && num_match_values > 0;

        serialize_bool( stream, has_match_values );

        if ( has_match_values )
        {
            serialize_int( stream, num_match_values, 0, NEXT_MAX_MATCH_VALUES );
            for ( int i = 0; i < num_match_values; ++i )
            {
                serialize_double( stream, match_values[i] );
            }
        }

        return true;
    }
};

// ---------------------------------------------------------------

struct NextBackendMatchDataResponsePacket
{
    uint64_t session_id;

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, session_id );
        return true;
    }
};

// ---------------------------------------------------------------

struct next_session_entry_t
{
    NEXT_DECLARE_SENTINEL(0)

    next_address_t address;
    uint64_t session_id;
    uint8_t most_recent_session_version;
    uint64_t special_send_sequence;
    uint64_t internal_send_sequence;
    uint64_t stats_sequence;
    uint64_t user_hash;
    uint64_t previous_session_events;
    uint64_t current_session_events;
    uint8_t client_open_session_sequence;

    NEXT_DECLARE_SENTINEL(1)

    bool stats_reported;
    bool stats_multipath;
    bool stats_fallback_to_direct;
    bool stats_client_bandwidth_over_limit;
    bool stats_server_bandwidth_over_limit;
    bool stats_has_near_relay_pings;
    int stats_platform_id;
    int stats_connection_type;
    float stats_direct_kbps_up;
    float stats_direct_kbps_down;
    float stats_next_kbps_up;
    float stats_next_kbps_down;
    float stats_direct_rtt;
    float stats_direct_jitter;
    float stats_direct_packet_loss;
    float stats_direct_max_packet_loss_seen;
    bool stats_next;
    float stats_next_rtt;
    float stats_next_jitter;
    float stats_next_packet_loss;
    int stats_num_near_relays;

    NEXT_DECLARE_SENTINEL(2)

    uint64_t stats_near_relay_ids[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(3)

    uint8_t stats_near_relay_rtt[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(4)

    uint8_t stats_near_relay_jitter[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(5)

    float stats_near_relay_packet_loss[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(6)

    uint64_t stats_packets_sent_client_to_server;
    uint64_t stats_packets_sent_server_to_client;
    uint64_t stats_packets_lost_client_to_server;
    uint64_t stats_packets_lost_server_to_client;
    uint64_t stats_packets_out_of_order_client_to_server;
    uint64_t stats_packets_out_of_order_server_to_client;

    float stats_jitter_client_to_server;
    float stats_jitter_server_to_client;

    double next_tracker_update_time;
    double next_session_update_time;
    double next_session_resend_time;
    double last_client_stats_update;
    double last_upgraded_packet_receive_time;

    uint64_t update_sequence;
    bool update_dirty;
    bool waiting_for_update_response;
    bool multipath;
    double update_last_send_time;
    uint8_t update_type;
    int update_num_tokens;
    bool session_update_timed_out;

    NEXT_DECLARE_SENTINEL(7)

    uint8_t update_tokens[NEXT_MAX_TOKENS*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES];

    NEXT_DECLARE_SENTINEL(8)

    bool update_has_near_relays;
    int update_num_near_relays;

    NEXT_DECLARE_SENTINEL(9)

    uint64_t update_near_relay_ids[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(10)

    next_address_t update_near_relay_addresses[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(11)

    uint8_t update_near_relay_ping_tokens[NEXT_MAX_NEAR_RELAYS*NEXT_PING_TOKEN_BYTES];

    NEXT_DECLARE_SENTINEL(12)

    uint64_t update_near_relay_expire_timestamp;

    NEXT_DECLARE_SENTINEL(13)

    NextBackendSessionUpdateRequestPacket session_update_request_packet;

    NEXT_DECLARE_SENTINEL(14)

    bool has_pending_route;
    uint8_t pending_route_session_version;
    uint64_t pending_route_expire_timestamp;
    double pending_route_expire_time;
    int pending_route_kbps_up;
    int pending_route_kbps_down;
    next_address_t pending_route_send_address;

    NEXT_DECLARE_SENTINEL(15)

    uint8_t pending_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(16)

    bool has_current_route;
    uint8_t current_route_session_version;
    uint64_t current_route_expire_timestamp;
    double current_route_expire_time;
    int current_route_kbps_up;
    int current_route_kbps_down;
    next_address_t current_route_send_address;

    NEXT_DECLARE_SENTINEL(17)

    uint8_t current_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(18)

    bool has_previous_route;
    next_address_t previous_route_send_address;

    NEXT_DECLARE_SENTINEL(19)

    uint8_t previous_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(20)

    uint8_t ephemeral_private_key[NEXT_CRYPTO_SECRETBOX_KEYBYTES];
    uint8_t send_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
    uint8_t receive_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
    uint8_t client_route_public_key[NEXT_CRYPTO_BOX_PUBLICKEYBYTES];

    NEXT_DECLARE_SENTINEL(21)

    uint8_t upgrade_token[NEXT_UPGRADE_TOKEN_BYTES];

    NEXT_DECLARE_SENTINEL(22)

    next_replay_protection_t payload_replay_protection;
    next_replay_protection_t special_replay_protection;
    next_replay_protection_t internal_replay_protection;

    NEXT_DECLARE_SENTINEL(23)

    next_packet_loss_tracker_t packet_loss_tracker;
    next_out_of_order_tracker_t out_of_order_tracker;
    next_jitter_tracker_t jitter_tracker;

    NEXT_DECLARE_SENTINEL(24)

    bool mutex_multipath;
    int mutex_envelope_kbps_up;
    int mutex_envelope_kbps_down;
    uint64_t mutex_payload_send_sequence;
    uint64_t mutex_session_id;
    uint8_t mutex_session_version;
    bool mutex_send_over_network_next;
    next_address_t mutex_send_address;

    NEXT_DECLARE_SENTINEL(25)

    uint8_t mutex_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(26)

    int session_data_bytes;
    uint8_t session_data[NEXT_MAX_SESSION_DATA_BYTES];
    uint8_t session_data_signature[NEXT_CRYPTO_SIGN_BYTES];

    NEXT_DECLARE_SENTINEL(27)

    bool client_ping_timed_out;
    double last_client_direct_ping;
    double last_client_next_ping;

    NEXT_DECLARE_SENTINEL(28)

    bool has_debug;
    char debug[NEXT_MAX_SESSION_DEBUG];

    NEXT_DECLARE_SENTINEL(29)

    uint64_t match_id;
    double match_values[NEXT_MAX_MATCH_VALUES];
    int num_match_values;

    NextBackendMatchDataRequestPacket match_data_request_packet;

    bool has_match_data;
    double next_match_data_resend_time;
    bool waiting_for_match_data_response;
    bool match_data_response_received;

    NEXT_DECLARE_SENTINEL(30)

    uint32_t session_flush_update_sequence;
    bool session_update_flush;
    bool session_update_flush_finished;
    bool match_data_flush;
    bool match_data_flush_finished;

    NEXT_DECLARE_SENTINEL(31)

    int num_held_near_relays;
    uint64_t held_near_relay_ids[NEXT_MAX_NEAR_RELAYS];
    uint8_t held_near_relay_rtt[NEXT_MAX_NEAR_RELAYS];
    uint8_t held_near_relay_jitter[NEXT_MAX_NEAR_RELAYS];
    float held_near_relay_packet_loss[NEXT_MAX_NEAR_RELAYS];

    NEXT_DECLARE_SENTINEL(32)
};

void next_session_entry_initialize_sentinels( next_session_entry_t * entry )
{
    (void) entry;
    next_assert( entry );
    NEXT_INITIALIZE_SENTINEL( entry, 0 )
    NEXT_INITIALIZE_SENTINEL( entry, 1 )
    NEXT_INITIALIZE_SENTINEL( entry, 2 )
    NEXT_INITIALIZE_SENTINEL( entry, 3 )
    NEXT_INITIALIZE_SENTINEL( entry, 4 )
    NEXT_INITIALIZE_SENTINEL( entry, 5 )
    NEXT_INITIALIZE_SENTINEL( entry, 6 )
    NEXT_INITIALIZE_SENTINEL( entry, 7 )
    NEXT_INITIALIZE_SENTINEL( entry, 8 )
    NEXT_INITIALIZE_SENTINEL( entry, 9 )
    NEXT_INITIALIZE_SENTINEL( entry, 10 )
    NEXT_INITIALIZE_SENTINEL( entry, 11 )
    NEXT_INITIALIZE_SENTINEL( entry, 12 )
    NEXT_INITIALIZE_SENTINEL( entry, 13 )
    NEXT_INITIALIZE_SENTINEL( entry, 14 )
    NEXT_INITIALIZE_SENTINEL( entry, 15 )
    NEXT_INITIALIZE_SENTINEL( entry, 16 )
    NEXT_INITIALIZE_SENTINEL( entry, 17 )
    NEXT_INITIALIZE_SENTINEL( entry, 18 )
    NEXT_INITIALIZE_SENTINEL( entry, 19 )
    NEXT_INITIALIZE_SENTINEL( entry, 20 )
    NEXT_INITIALIZE_SENTINEL( entry, 21 )
    NEXT_INITIALIZE_SENTINEL( entry, 22 )
    NEXT_INITIALIZE_SENTINEL( entry, 23 )
    NEXT_INITIALIZE_SENTINEL( entry, 24 )
    NEXT_INITIALIZE_SENTINEL( entry, 25 )
    NEXT_INITIALIZE_SENTINEL( entry, 26 )
    NEXT_INITIALIZE_SENTINEL( entry, 27 )
    NEXT_INITIALIZE_SENTINEL( entry, 28 )
    NEXT_INITIALIZE_SENTINEL( entry, 29 )
    NEXT_INITIALIZE_SENTINEL( entry, 30 )
    NEXT_INITIALIZE_SENTINEL( entry, 31 )
    NEXT_INITIALIZE_SENTINEL( entry, 32 )
    next_replay_protection_initialize_sentinels( &entry->payload_replay_protection );
    next_replay_protection_initialize_sentinels( &entry->special_replay_protection );
    next_replay_protection_initialize_sentinels( &entry->internal_replay_protection );
    next_packet_loss_tracker_initialize_sentinels( &entry->packet_loss_tracker );
    next_out_of_order_tracker_initialize_sentinels( &entry->out_of_order_tracker );
    next_jitter_tracker_initialize_sentinels( &entry->jitter_tracker );
}

void next_session_entry_verify_sentinels( next_session_entry_t * entry )
{
    (void) entry;
    next_assert( entry );
    NEXT_VERIFY_SENTINEL( entry, 0 )
    NEXT_VERIFY_SENTINEL( entry, 1 )
    NEXT_VERIFY_SENTINEL( entry, 2 )
    NEXT_VERIFY_SENTINEL( entry, 3 )
    NEXT_VERIFY_SENTINEL( entry, 4 )
    NEXT_VERIFY_SENTINEL( entry, 5 )
    NEXT_VERIFY_SENTINEL( entry, 6 )
    NEXT_VERIFY_SENTINEL( entry, 7 )
    NEXT_VERIFY_SENTINEL( entry, 8 )
    NEXT_VERIFY_SENTINEL( entry, 9 )
    NEXT_VERIFY_SENTINEL( entry, 10 )
    NEXT_VERIFY_SENTINEL( entry, 11 )
    NEXT_VERIFY_SENTINEL( entry, 12 )
    NEXT_VERIFY_SENTINEL( entry, 13 )
    NEXT_VERIFY_SENTINEL( entry, 14 )
    NEXT_VERIFY_SENTINEL( entry, 15 )
    NEXT_VERIFY_SENTINEL( entry, 16 )
    NEXT_VERIFY_SENTINEL( entry, 17 )
    NEXT_VERIFY_SENTINEL( entry, 18 )
    NEXT_VERIFY_SENTINEL( entry, 19 )
    NEXT_VERIFY_SENTINEL( entry, 20 )
    NEXT_VERIFY_SENTINEL( entry, 21 )
    NEXT_VERIFY_SENTINEL( entry, 22 )
    NEXT_VERIFY_SENTINEL( entry, 23 )
    NEXT_VERIFY_SENTINEL( entry, 24 )
    NEXT_VERIFY_SENTINEL( entry, 25 )
    NEXT_VERIFY_SENTINEL( entry, 26 )
    NEXT_VERIFY_SENTINEL( entry, 27 )
    NEXT_VERIFY_SENTINEL( entry, 28 )
    NEXT_VERIFY_SENTINEL( entry, 29 )
    NEXT_VERIFY_SENTINEL( entry, 30 )
    NEXT_VERIFY_SENTINEL( entry, 31 )
    NEXT_VERIFY_SENTINEL( entry, 32 )
    next_replay_protection_verify_sentinels( &entry->payload_replay_protection );
    next_replay_protection_verify_sentinels( &entry->special_replay_protection );
    next_replay_protection_verify_sentinels( &entry->internal_replay_protection );
    next_packet_loss_tracker_verify_sentinels( &entry->packet_loss_tracker );
    next_out_of_order_tracker_verify_sentinels( &entry->out_of_order_tracker );
    next_jitter_tracker_verify_sentinels( &entry->jitter_tracker );
}

struct next_session_manager_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int size;
    int max_entry_index;
    uint64_t * session_ids;
    next_address_t * addresses;
    next_session_entry_t * entries;

    NEXT_DECLARE_SENTINEL(1)
};

void next_session_manager_initialize_sentinels( next_session_manager_t * session_manager )
{
    (void) session_manager;
    next_assert( session_manager );
    NEXT_INITIALIZE_SENTINEL( session_manager, 0 )
    NEXT_INITIALIZE_SENTINEL( session_manager, 1 )
}

void next_session_manager_verify_sentinels( next_session_manager_t * session_manager )
{
    (void) session_manager;
#if NEXT_ENABLE_MEMORY_CHECKS
    next_assert( session_manager );
    NEXT_VERIFY_SENTINEL( session_manager, 0 )
    NEXT_VERIFY_SENTINEL( session_manager, 1 )
    const int size = session_manager->size;
    for ( int i = 0; i < size; ++i )
    {
        if ( session_manager->session_ids[i] != 0 )
        {
            next_session_entry_verify_sentinels( &session_manager->entries[i] );
        }
    }
#endif // #if NEXT_ENABLE_MEMORY_CHECKS
}

void next_session_manager_destroy( next_session_manager_t * session_manager );

next_session_manager_t * next_session_manager_create( void * context, int initial_size )
{
    next_session_manager_t * session_manager = (next_session_manager_t*) next_malloc( context, sizeof(next_session_manager_t) );

    next_assert( session_manager );
    if ( !session_manager )
        return NULL;

    memset( session_manager, 0, sizeof(next_session_manager_t) );

    next_session_manager_initialize_sentinels( session_manager );

    session_manager->context = context;
    session_manager->size = initial_size;
    session_manager->session_ids = (uint64_t*) next_malloc( context, size_t(initial_size) * 8 );
    session_manager->addresses = (next_address_t*) next_malloc( context, size_t(initial_size) * sizeof(next_address_t) );
    session_manager->entries = (next_session_entry_t*) next_malloc( context, size_t(initial_size) * sizeof(next_session_entry_t) );

    next_assert( session_manager->session_ids );
    next_assert( session_manager->addresses );
    next_assert( session_manager->entries );

    if ( session_manager->session_ids == NULL || session_manager->addresses == NULL || session_manager->entries == NULL )
    {
        next_session_manager_destroy( session_manager );
        return NULL;
    }

    memset( session_manager->session_ids, 0, size_t(initial_size) * 8 );
    memset( session_manager->addresses, 0, size_t(initial_size) * sizeof(next_address_t) );
    memset( session_manager->entries, 0, size_t(initial_size) * sizeof(next_session_entry_t) );

    next_session_manager_verify_sentinels( session_manager );

    return session_manager;
}

void next_session_manager_destroy( next_session_manager_t * session_manager )
{
    next_session_manager_verify_sentinels( session_manager );

    next_free( session_manager->context, session_manager->session_ids );
    next_free( session_manager->context, session_manager->addresses );
    next_free( session_manager->context, session_manager->entries );

    next_clear_and_free( session_manager->context, session_manager, sizeof(next_session_manager_t) );
}

bool next_session_manager_expand( next_session_manager_t * session_manager )
{
    next_assert( session_manager );

    next_session_manager_verify_sentinels( session_manager );

    int new_size = session_manager->size * 2;

    uint64_t * new_session_ids = (uint64_t*) next_malloc( session_manager->context, size_t(new_size) * 8 );
    next_address_t * new_addresses = (next_address_t*) next_malloc( session_manager->context, size_t(new_size) * sizeof(next_address_t) );
    next_session_entry_t * new_entries = (next_session_entry_t*) next_malloc( session_manager->context, size_t(new_size) * sizeof(next_session_entry_t) );

    next_assert( new_session_ids );
    next_assert( new_addresses );
    next_assert( new_entries );

    if ( new_session_ids == NULL || new_addresses == NULL || new_entries == NULL )
    {
        next_free( session_manager->context, new_session_ids );
        next_free( session_manager->context, new_addresses );
        next_free( session_manager->context, new_entries );
        return false;
    }

    memset( new_session_ids, 0, size_t(new_size) * 8 );
    memset( new_addresses, 0, size_t(new_size) * sizeof(next_address_t) );
    memset( new_entries, 0, size_t(new_size) * sizeof(next_session_entry_t) );

    int index = 0;
    const int current_size = session_manager->size;
    for ( int i = 0; i < current_size; ++i )
    {
        if ( session_manager->session_ids[i] != 0 )
        {
            memcpy( &new_session_ids[index], &session_manager->session_ids[i], 8 );
            memcpy( &new_addresses[index], &session_manager->addresses[i], sizeof(next_address_t) );
            memcpy( &new_entries[index], &session_manager->entries[i], sizeof(next_session_entry_t) );
            index++;
        }
    }

    next_free( session_manager->context, session_manager->session_ids );
    next_free( session_manager->context, session_manager->addresses );
    next_free( session_manager->context, session_manager->entries );

    session_manager->session_ids = new_session_ids;
    session_manager->addresses = new_addresses;
    session_manager->entries = new_entries;
    session_manager->size = new_size;
    session_manager->max_entry_index = index - 1;

    return true;
}

void next_clear_session_entry( next_session_entry_t * entry, const next_address_t * address, uint64_t session_id )
{
    memset( entry, 0, sizeof(next_session_entry_t) );

    next_session_entry_initialize_sentinels( entry );

    entry->address = *address;
    entry->session_id = session_id;

    next_replay_protection_reset( &entry->payload_replay_protection );
    next_replay_protection_reset( &entry->special_replay_protection );
    next_replay_protection_reset( &entry->internal_replay_protection );

    next_packet_loss_tracker_reset( &entry->packet_loss_tracker );
    next_out_of_order_tracker_reset( &entry->out_of_order_tracker );
    next_jitter_tracker_reset( &entry->jitter_tracker );

    next_session_entry_verify_sentinels( entry );

    entry->special_send_sequence = 1;
    entry->internal_send_sequence = 1;

    const double current_time = next_platform_time();

    entry->last_client_direct_ping = current_time;
    entry->last_client_next_ping = current_time;
}

next_session_entry_t * next_session_manager_add( next_session_manager_t * session_manager, const next_address_t * address, uint64_t session_id, const uint8_t * ephemeral_private_key, const uint8_t * upgrade_token )
{
    next_session_manager_verify_sentinels( session_manager );

    next_assert( session_id != 0 );
    next_assert( address );
    next_assert( address->type != NEXT_ADDRESS_NONE );

    // first scan existing entries and see if we can insert there

    const int size = session_manager->size;

    for ( int i = 0; i < size; ++i )
    {
        if ( session_manager->session_ids[i] == 0 )
        {
            session_manager->session_ids[i] = session_id;
            session_manager->addresses[i] = *address;
            next_session_entry_t * entry = &session_manager->entries[i];
            next_clear_session_entry( entry, address, session_id );
            memcpy( entry->ephemeral_private_key, ephemeral_private_key, NEXT_CRYPTO_SECRETBOX_KEYBYTES );
            memcpy( entry->upgrade_token, upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );
            if ( i > session_manager->max_entry_index )
            {
                session_manager->max_entry_index = i;
            }
            return entry;
        }
    }

    // ok, we need to grow, expand and add at the end (expand compacts existing entries)

    if ( !next_session_manager_expand( session_manager ) )
        return NULL;

    const int i = ++session_manager->max_entry_index;

    session_manager->session_ids[i] = session_id;
    session_manager->addresses[i] = *address;
    next_session_entry_t * entry = &session_manager->entries[i];
    next_clear_session_entry( entry, address, session_id );
    memcpy( entry->ephemeral_private_key, ephemeral_private_key, NEXT_CRYPTO_SECRETBOX_KEYBYTES );
    memcpy( entry->upgrade_token, upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );

    next_session_manager_verify_sentinels( session_manager );

    return entry;
}

void next_session_manager_remove_at_index( next_session_manager_t * session_manager, int index )
{
    next_session_manager_verify_sentinels( session_manager );

    next_assert( index >= 0 );
    next_assert( index <= session_manager->max_entry_index );

    const int max_index = session_manager->max_entry_index;
    session_manager->session_ids[index] = 0;
    session_manager->addresses[index].type = NEXT_ADDRESS_NONE;
    if ( index == max_index )
    {
        while ( index > 0 && session_manager->session_ids[index] == 0 )
        {
            index--;
        }
        session_manager->max_entry_index = index;
    }

    next_session_manager_verify_sentinels( session_manager );
}

void next_session_manager_remove_by_address( next_session_manager_t * session_manager, const next_address_t * address )
{
    next_session_manager_verify_sentinels( session_manager );

    next_assert( address );

    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( next_address_equal( address, &session_manager->addresses[i] ) == 1 )
        {
            next_session_manager_remove_at_index( session_manager, i );
            return;
        }
    }

    next_session_manager_verify_sentinels( session_manager );
}

next_session_entry_t * next_session_manager_find_by_address( next_session_manager_t * session_manager, const next_address_t * address )
{
    next_session_manager_verify_sentinels( session_manager );
    next_assert( address );
    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( next_address_equal( address, &session_manager->addresses[i] ) == 1 )
        {
            return &session_manager->entries[i];
        }
    }
    return NULL;
}

next_session_entry_t * next_session_manager_find_by_session_id( next_session_manager_t * session_manager, uint64_t session_id )
{
    next_session_manager_verify_sentinels( session_manager );
    next_assert( session_id );
    if ( session_id == 0 )
    {
        return NULL;
    }
    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( session_id == session_manager->session_ids[i] )
        {
            return &session_manager->entries[i];
        }
    }
    return NULL;
}

int next_session_manager_num_entries( next_session_manager_t * session_manager )
{
    next_session_manager_verify_sentinels( session_manager );
    int num_entries = 0;
    const int max_index = session_manager->max_entry_index;
    for ( int i = 0; i <= max_index; ++i )
    {
        if ( session_manager->session_ids[i] != 0 )
        {
            num_entries++;
        }
    }
    return num_entries;
}

// ---------------------------------------------------------------

struct NextBackendSessionUpdateResponsePacket
{
    uint64_t session_id;
    uint32_t slice_number;
    int session_data_bytes;
    uint8_t session_data[NEXT_MAX_SESSION_DATA_BYTES];
    uint8_t session_data_signature[NEXT_CRYPTO_SIGN_BYTES];
    uint8_t response_type;
    bool has_near_relays;
    int num_near_relays;
    uint64_t near_relay_ids[NEXT_MAX_NEAR_RELAYS];
    next_address_t near_relay_addresses[NEXT_MAX_NEAR_RELAYS];
    uint8_t near_relay_ping_tokens[NEXT_MAX_NEAR_RELAYS*NEXT_PING_TOKEN_BYTES];
    uint64_t near_relay_expire_timestamp;
    int num_tokens;
    uint8_t tokens[NEXT_MAX_TOKENS*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES];
    bool multipath;
    bool has_debug;
    char debug[NEXT_MAX_SESSION_DEBUG];

    NextBackendSessionUpdateResponsePacket()
    {
        memset( this, 0, sizeof(NextBackendSessionUpdateResponsePacket) );
    }

    template <typename Stream> bool Serialize( Stream & stream )
    {
        serialize_uint64( stream, session_id );

        serialize_uint32( stream, slice_number );

        serialize_int( stream, session_data_bytes, 0, NEXT_MAX_SESSION_DATA_BYTES );
        if ( session_data_bytes > 0 )
        {
            serialize_bytes( stream, session_data, session_data_bytes );
            serialize_bytes( stream, session_data_signature, NEXT_CRYPTO_SIGN_BYTES );
        }

        serialize_int( stream, response_type, 0, NEXT_UPDATE_TYPE_CONTINUE );

        serialize_bool( stream, has_near_relays );

        if ( has_near_relays )
        {
            serialize_int( stream, num_near_relays, 0, NEXT_MAX_NEAR_RELAYS );
            for ( int i = 0; i < num_near_relays; ++i )
            {
                serialize_uint64( stream, near_relay_ids[i] );
                serialize_address( stream, near_relay_addresses[i] );
                serialize_bytes( stream, near_relay_ping_tokens + i * NEXT_PING_TOKEN_BYTES, NEXT_PING_TOKEN_BYTES );
            }
            serialize_uint64( stream, near_relay_expire_timestamp );
        }

        if ( response_type != NEXT_UPDATE_TYPE_DIRECT )
        {
            serialize_bool( stream, multipath );
            serialize_int( stream, num_tokens, 0, NEXT_MAX_TOKENS );
        }

        if ( response_type == NEXT_UPDATE_TYPE_ROUTE )
        {
            serialize_bytes( stream, tokens, num_tokens * NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES );
        }

        if ( response_type == NEXT_UPDATE_TYPE_CONTINUE )
        {
            serialize_bytes( stream, tokens, num_tokens * NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES );
        }

        serialize_bool( stream, has_debug );
        if ( has_debug )
        {
            serialize_string( stream, debug, NEXT_MAX_SESSION_DEBUG );
        }

        return true;
    }
};

// ---------------------------------------------------------------

int next_write_backend_packet( uint8_t packet_id, void * packet_object, uint8_t * packet_data, int * packet_bytes, const int * signed_packet, const uint8_t * sign_private_key, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    next_assert( packet_object );
    next_assert( packet_data );
    next_assert( packet_bytes );

    next::WriteStream stream( packet_data, NEXT_MAX_PACKET_BYTES );

    typedef next::WriteStream Stream;

    serialize_bits( stream, packet_id, 8 );

    uint8_t dummy = 0;
    for ( int i = 0; i < 15; ++i )
    {
        serialize_bits( stream, dummy, 8 );
    }

    switch ( packet_id )
    {
        case NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET:
        {
            NextBackendServerInitRequestPacket * packet = (NextBackendServerInitRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET:
        {
            NextBackendServerInitResponsePacket * packet = (NextBackendServerInitResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SERVER_UPDATE_REQUEST_PACKET:
        {
            NextBackendServerUpdateRequestPacket * packet = (NextBackendServerUpdateRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SERVER_UPDATE_RESPONSE_PACKET:
        {
            NextBackendServerUpdateResponsePacket * packet = (NextBackendServerUpdateResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SESSION_UPDATE_REQUEST_PACKET:
        {
            NextBackendSessionUpdateRequestPacket * packet = (NextBackendSessionUpdateRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SESSION_UPDATE_RESPONSE_PACKET:
        {
            NextBackendSessionUpdateResponsePacket * packet = (NextBackendSessionUpdateResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_MATCH_DATA_REQUEST_PACKET:
        {
            NextBackendMatchDataRequestPacket * packet = (NextBackendMatchDataRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_MATCH_DATA_RESPONSE_PACKET:
        {
            NextBackendMatchDataResponsePacket * packet = (NextBackendMatchDataResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        default:
            return NEXT_ERROR;
    }

    stream.Flush();

    *packet_bytes = stream.GetBytesProcessed();

    next_assert( *packet_bytes >= 0 );
    next_assert( *packet_bytes < NEXT_MAX_PACKET_BYTES );

    if ( signed_packet && signed_packet[packet_id] )
    {
        next_assert( sign_private_key );
        next_crypto_sign_state_t state;
        next_crypto_sign_init( &state );
        next_crypto_sign_update( &state, packet_data, 1 );
        next_crypto_sign_update( &state, packet_data + 16, size_t(*packet_bytes) - 16 );
        next_crypto_sign_final_create( &state, packet_data + *packet_bytes, NULL, sign_private_key );
        *packet_bytes += NEXT_CRYPTO_SIGN_BYTES;
    }

    *packet_bytes += 2;

    next_generate_chonkle( packet_data + 1, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, *packet_bytes );
    next_generate_pittle( packet_data + *packet_bytes - 2, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, *packet_bytes );

    return NEXT_OK;
}

int next_read_backend_packet( uint8_t packet_id, uint8_t * packet_data, int begin, int end, void * packet_object, const int * signed_packet, const uint8_t * sign_public_key )
{
    next_assert( packet_data );
    next_assert( packet_object );

    next::ReadStream stream( packet_data, end );

    uint8_t * dummy = (uint8_t*) alloca( begin );
    serialize_bytes( stream, dummy, begin );

    if ( signed_packet && signed_packet[packet_id] )
    {
        next_assert( sign_public_key );

        const int packet_bytes = end - begin;

        if ( packet_bytes < int( NEXT_CRYPTO_SIGN_BYTES ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "signed backend packet is too small to be valid" );
            return NEXT_ERROR;
        }

        next_crypto_sign_state_t state;
        next_crypto_sign_init( &state );
        next_crypto_sign_update( &state, &packet_id, 1 );
        next_crypto_sign_update( &state, packet_data + begin, packet_bytes - NEXT_CRYPTO_SIGN_BYTES );
        if ( next_crypto_sign_final_verify( &state, packet_data + end - NEXT_CRYPTO_SIGN_BYTES, sign_public_key ) != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "signed backend packet did not verify" );
            return NEXT_ERROR;
        }
    }

    switch ( packet_id )
    {
        case NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET:
        {
            NextBackendServerInitRequestPacket * packet = (NextBackendServerInitRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET:
        {
            NextBackendServerInitResponsePacket * packet = (NextBackendServerInitResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SERVER_UPDATE_REQUEST_PACKET:
        {
            NextBackendServerUpdateRequestPacket * packet = (NextBackendServerUpdateRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SERVER_UPDATE_RESPONSE_PACKET:
        {
            NextBackendServerUpdateResponsePacket * packet = (NextBackendServerUpdateResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SESSION_UPDATE_REQUEST_PACKET:
        {
            NextBackendSessionUpdateRequestPacket * packet = (NextBackendSessionUpdateRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_SESSION_UPDATE_RESPONSE_PACKET:
        {
            NextBackendSessionUpdateResponsePacket * packet = (NextBackendSessionUpdateResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_MATCH_DATA_REQUEST_PACKET:
        {
            NextBackendMatchDataRequestPacket * packet = (NextBackendMatchDataRequestPacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        case NEXT_BACKEND_MATCH_DATA_RESPONSE_PACKET:
        {
            NextBackendMatchDataResponsePacket * packet = (NextBackendMatchDataResponsePacket*) packet_object;
            if ( !packet->Serialize( stream ) )
                return NEXT_ERROR;
        }
        break;

        default:
            return NEXT_ERROR;
    }

    return (int) packet_id;
}

// ---------------------------------------------------------------

#define NEXT_SERVER_COMMAND_UPGRADE_SESSION                         0
#define NEXT_SERVER_COMMAND_SESSION_EVENT                           1
#define NEXT_SERVER_COMMAND_MATCH_DATA                              2
#define NEXT_SERVER_COMMAND_FLUSH                                   3
#define NEXT_SERVER_COMMAND_SET_PACKET_RECEIVE_CALLBACK             4
#define NEXT_SERVER_COMMAND_SET_SEND_PACKET_TO_ADDRESS_CALLBACK     5
#define NEXT_SERVER_COMMAND_SET_PAYLOAD_RECEIVE_CALLBACK            6

struct next_server_command_t
{
    int type;
};

struct next_server_command_upgrade_session_t : public next_server_command_t
{
    next_address_t address;
    uint64_t session_id;
    uint64_t user_hash;
};

struct next_server_command_session_event_t : public next_server_command_t
{
    next_address_t address;
    uint64_t session_events;
};

struct next_server_command_match_data_t : public next_server_command_t
{
    next_address_t address;
    uint64_t match_id;
    double match_values[NEXT_MAX_MATCH_VALUES];
    int num_match_values;
};

struct next_server_command_flush_t : public next_server_command_t
{
    // ...
};

struct next_server_command_set_packet_receive_callback_t : public next_server_command_t
{
    void (*callback) ( void * data, next_address_t * from, uint8_t * packet_data, int * begin, int * end );
    void * callback_data;
};

struct next_server_command_set_send_packet_to_address_callback_t : public next_server_command_t
{
    int (*callback) ( void * data, const next_address_t * address, const uint8_t * packet_data, int packet_bytes );
    void * callback_data;
};

struct next_server_command_set_payload_receive_callback_t : public next_server_command_t
{
    int (*callback) ( void * data, const next_address_t * client_address, const uint8_t * payload_data, int payload_bytes );
    void * callback_data;
};

// ---------------------------------------------------------------

#define NEXT_SERVER_NOTIFY_PACKET_RECEIVED                      0
#define NEXT_SERVER_NOTIFY_PENDING_SESSION_TIMED_OUT            1
#define NEXT_SERVER_NOTIFY_SESSION_UPGRADED                     2
#define NEXT_SERVER_NOTIFY_SESSION_TIMED_OUT                    3
#define NEXT_SERVER_NOTIFY_INIT_TIMED_OUT                       4
#define NEXT_SERVER_NOTIFY_READY                                5
#define NEXT_SERVER_NOTIFY_FLUSH_FINISHED                       6
#define NEXT_SERVER_NOTIFY_MAGIC_UPDATED                        7
#define NEXT_SERVER_NOTIFY_DIRECT_ONLY                          8

struct next_server_notify_t
{
    int type;
};

struct next_server_notify_packet_received_t : public next_server_notify_t
{
    next_address_t from;
    int packet_bytes;
    uint8_t packet_data[NEXT_MAX_PACKET_BYTES];
};

struct next_server_notify_pending_session_cancelled_t : public next_server_notify_t
{
    next_address_t address;
    uint64_t session_id;
};

struct next_server_notify_pending_session_timed_out_t : public next_server_notify_t
{
    next_address_t address;
    uint64_t session_id;
};

struct next_server_notify_session_upgraded_t : public next_server_notify_t
{
    next_address_t address;
    uint64_t session_id;
};

struct next_server_notify_session_timed_out_t : public next_server_notify_t
{
    next_address_t address;
    uint64_t session_id;
};

struct next_server_notify_init_timed_out_t : public next_server_notify_t
{
    // ...
};

struct next_server_notify_ready_t : public next_server_notify_t
{
    char datacenter_name[NEXT_MAX_DATACENTER_NAME_LENGTH];
};

struct next_server_notify_flush_finished_t : public next_server_notify_t
{
    // ...
};

struct next_server_notify_magic_updated_t : public next_server_notify_t
{
    uint8_t current_magic[8];
};

struct next_server_notify_direct_only_t : public next_server_notify_t
{
    // ...
};

// ---------------------------------------------------------------

struct next_server_internal_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    int state;
    uint64_t customer_id;
    uint64_t datacenter_id;
    uint64_t match_id;
    char datacenter_name[NEXT_MAX_DATACENTER_NAME_LENGTH];
    char autodetect_input[NEXT_MAX_DATACENTER_NAME_LENGTH];
    char autodetect_datacenter[NEXT_MAX_DATACENTER_NAME_LENGTH];
    bool autodetected_datacenter;

    NEXT_DECLARE_SENTINEL(1)

    uint8_t customer_private_key[NEXT_CRYPTO_SIGN_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(2)

    bool valid_customer_private_key;
    bool no_datacenter_specified;
    uint64_t upgrade_sequence;
    double next_resolve_hostname_time;
    next_address_t backend_address;
    next_address_t server_address;
    next_address_t bind_address;
    next_queue_t * command_queue;
    next_queue_t * notify_queue;
    next_platform_mutex_t session_mutex;
    next_platform_mutex_t command_mutex;
    next_platform_mutex_t notify_mutex;
    next_platform_socket_t * socket;
    next_pending_session_manager_t * pending_session_manager;
    next_session_manager_t * session_manager;

    NEXT_DECLARE_SENTINEL(3)

    bool resolving_hostname;
    bool resolve_hostname_finished;
    double resolve_hostname_start_time;
    next_address_t resolve_hostname_result;
    next_platform_mutex_t resolve_hostname_mutex;
    next_platform_thread_t * resolve_hostname_thread;

    NEXT_DECLARE_SENTINEL(4)

    bool autodetecting;
    bool autodetect_finished;
    bool autodetect_actually_did_something;
    bool autodetect_succeeded;
    double autodetect_start_time;
    char autodetect_result[NEXT_MAX_DATACENTER_NAME_LENGTH];
    next_platform_mutex_t autodetect_mutex;
    next_platform_thread_t * autodetect_thread;

    NEXT_DECLARE_SENTINEL(5)

    uint8_t server_kx_public_key[NEXT_CRYPTO_KX_PUBLICKEYBYTES];
    uint8_t server_kx_private_key[NEXT_CRYPTO_KX_SECRETKEYBYTES];
    uint8_t server_route_public_key[NEXT_CRYPTO_BOX_PUBLICKEYBYTES];
    uint8_t server_route_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

    NEXT_DECLARE_SENTINEL(6)

    uint8_t upcoming_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    NEXT_DECLARE_SENTINEL(7)

    uint64_t server_init_request_id;
    double server_init_resend_time;
    double server_init_timeout_time;
    bool received_init_response;

    NEXT_DECLARE_SENTINEL(8)

    uint64_t server_update_request_id;
    double server_update_last_time;
    double server_update_resend_time;
    int server_update_num_sessions;
    bool server_update_first;

    NEXT_DECLARE_SENTINEL(9)

    std::atomic<uint64_t> quit;

    NEXT_DECLARE_SENTINEL(10)

    bool flushing;
    bool flushed;
    uint64_t num_session_updates_to_flush;
    uint64_t num_match_data_to_flush;
    uint64_t num_flushed_session_updates;
    uint64_t num_flushed_match_data;

    NEXT_DECLARE_SENTINEL(11)

    void (*packet_receive_callback) ( void * data, next_address_t * from, uint8_t * packet_data, int * begin, int * end );
    void * packet_receive_callback_data;

    int (*send_packet_to_address_callback)( void * data, const next_address_t * address, const uint8_t * packet_data, int packet_bytes );
    void * send_packet_to_address_callback_data;

    int (*payload_receive_callback)( void * data, const next_address_t * client_address, const uint8_t * payload_data, int payload_bytes );
    void * payload_receive_callback_data;

    NEXT_DECLARE_SENTINEL(12)
};

void next_server_internal_initialize_sentinels( next_server_internal_t * server )
{
    (void) server;
    next_assert( server );
    NEXT_INITIALIZE_SENTINEL( server, 0 )
    NEXT_INITIALIZE_SENTINEL( server, 1 )
    NEXT_INITIALIZE_SENTINEL( server, 2 )
    NEXT_INITIALIZE_SENTINEL( server, 3 )
    NEXT_INITIALIZE_SENTINEL( server, 4 )
    NEXT_INITIALIZE_SENTINEL( server, 5 )
    NEXT_INITIALIZE_SENTINEL( server, 6 )
    NEXT_INITIALIZE_SENTINEL( server, 7 )
    NEXT_INITIALIZE_SENTINEL( server, 8 )
    NEXT_INITIALIZE_SENTINEL( server, 9 )
    NEXT_INITIALIZE_SENTINEL( server, 10 )
    NEXT_INITIALIZE_SENTINEL( server, 11 )
    NEXT_INITIALIZE_SENTINEL( server, 12 )
}

void next_server_internal_verify_sentinels( next_server_internal_t * server )
{
    (void) server;
    next_assert( server );
    NEXT_VERIFY_SENTINEL( server, 0 )
    NEXT_VERIFY_SENTINEL( server, 1 )
    NEXT_VERIFY_SENTINEL( server, 2 )
    NEXT_VERIFY_SENTINEL( server, 3 )
    NEXT_VERIFY_SENTINEL( server, 4 )
    NEXT_VERIFY_SENTINEL( server, 5 )
    NEXT_VERIFY_SENTINEL( server, 6 )
    NEXT_VERIFY_SENTINEL( server, 7 )
    NEXT_VERIFY_SENTINEL( server, 8 )
    NEXT_VERIFY_SENTINEL( server, 9 )
    NEXT_VERIFY_SENTINEL( server, 10 )
    NEXT_VERIFY_SENTINEL( server, 11 )
    NEXT_VERIFY_SENTINEL( server, 12 )
    if ( server->session_manager )
        next_session_manager_verify_sentinels( server->session_manager );
    if ( server->pending_session_manager )
        next_pending_session_manager_verify_sentinels( server->pending_session_manager );
}

static void next_server_internal_resolve_hostname_thread_function( void * context );
static void next_server_internal_autodetect_thread_function( void * context );

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

bool next_autodetect_google( char * output )
{
    FILE * file;
    char buffer[1024*10];

    // are we running in google cloud?
#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    file = popen( "/bin/ls /usr/bin | grep google_ 2>/dev/null", "r" );
    if ( file == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run ls" );
        return false;
    }

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    file = _popen( "dir \"C:\\Program Files (x86)\\Google\\Cloud SDK\\google-cloud-sdk\\bin\" | findstr gcloud", "r" );
    if ( file == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run dir" );
        return false;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    bool in_gcp = false;
    while ( fgets( buffer, sizeof(buffer), file ) != NULL )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: running in google cloud" );
        in_gcp = true;
        break;
    }

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    _pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    // we are not running in google cloud :(

    if ( !in_gcp )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in google cloud" );
        return false;
    }

    // we are running in google cloud, which zone are we in?

    char zone[256];
    zone[0] = '\0';

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    file = popen( "curl -s \"http://metadata.google.internal/computeMetadata/v1/instance/zone\" -H \"Metadata-Flavor: Google\" --max-time 10 -vs 2>/dev/null", "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
        return false;
    }

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    file = _popen( "powershell Invoke-RestMethod -Uri http://metadata.google.internal/computeMetadata/v1/instance/zone -Headers @{'Metadata-Flavor' = 'Google'} -TimeoutSec 10", "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
        return false;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    while ( fgets( buffer, sizeof(buffer), file ) != NULL )
    {
        size_t length = strlen( buffer );
        if ( length < 10 )
        {
            continue;
        }

        if ( buffer[0] != 'p' ||
             buffer[1] != 'r' ||
             buffer[2] != 'o' ||
             buffer[3] != 'j' ||
             buffer[4] != 'e' ||
             buffer[5] != 'c' ||
             buffer[6] != 't' ||
             buffer[7] != 's' ||
             buffer[8] != '/' )
        {
            continue;
        }

        bool found = false;
        size_t index = length - 1;
        while ( index > 10 && length  )
        {
            if ( buffer[index] == '/' )
            {
                found = true;
                break;
            }
            index--;
        }

        if ( !found )
        {
            continue;
        }

        next_copy_string( zone, buffer + index + 1, sizeof(zone) );

        size_t zone_length = strlen(zone);
        index = zone_length - 1;
        while ( index > 0 && ( zone[index] == '\n' || zone[index] == '\r' ) )
        {
            zone[index] = '\0';
            index--;
        }

        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: google zone is \"%s\"", zone );

        break;
    }

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    _pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    // we couldn't work out which zone we are in :(

    if ( zone[0] == '\0' )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not detect google zone" );
        return false;
    }

    // look up google zone -> network next datacenter via mapping in google cloud storage "google.txt" file

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    char cmd[1024];
    snprintf( cmd, sizeof(cmd), "curl -s \"https://storage.googleapis.com/network_next_sdk_config/google.txt?ts=%x\" --max-time 10 -vs 2>/dev/null", uint32_t(time(NULL)) );
    file = popen( cmd, "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
        return false;
    }

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    char cmd[1024];
    snprintf( cmd, sizeof(cmd), "powershell Invoke-RestMethod -Uri \"https://storage.googleapis.com/network_next_sdk_config/google.txt?ts=%x\" -TimeoutSec 10", uint32_t(time(NULL)) );
    file = _popen( cmd, "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
        return false;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    bool found = false;

    while ( fgets( buffer, sizeof(buffer), file ) != NULL )
    {
        const char * separators = ",\n\r";

        char * google_zone = strtok( buffer, separators );
        if ( google_zone == NULL )
        {
            continue;
        }

        char * google_datacenter = strtok( NULL, separators );
        if ( google_datacenter == NULL )
        {
            continue;
        }

        if ( strcmp( zone, google_zone ) == 0 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: \"%s\" -> \"%s\"", zone, google_datacenter );
            strcpy( output, google_datacenter );
            found = true;
            break;
        }
    }

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    _pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    return found;
}

bool next_autodetect_amazon( char * output )
{
    FILE * file;
    char buffer[1024*10];

    // Get the AZID from instance metadata
    // This is necessary because only AZ IDs are the same across different customer accounts
    // See https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html for details

    char azid[256];
    azid[0] = '\0';

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    file = popen( "curl -s \"http://169.254.169.254/latest/meta-data/placement/availability-zone-id\" --max-time 2 -vs 2>/dev/null", "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
        return false;
    }

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    file = _popen ( "powershell Invoke-RestMethod -Uri http://169.254.169.254/latest/meta-data/placement/availability-zone-id -TimeoutSec 2", "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
        return false;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    while ( fgets( buffer, sizeof(buffer), file ) != NULL )
    {
        if ( strstr( buffer, "-az" ) == NULL )
        {
            continue;
        }

        strcpy( azid, buffer );

        size_t azid_length = strlen(azid);
        size_t index = azid_length - 1;
        while ( index > 0 && ( azid[index] == '\n' || azid[index] == '\r' ) )
        {
            azid[index] = '\0';
            index--;
        }

        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: azid is \"%s\"", azid );

        break;
    }

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    _pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    // we are probably not in AWS :(

    if ( azid[0] == '\0' )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in AWS" );
        return false;
    }

    // look up AZID -> network next datacenter via mapping in google cloud storage "amazon.txt" file

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    char cmd[1024];
    snprintf( cmd, sizeof(cmd), "curl -s \"https://storage.googleapis.com/network_next_sdk_config/amazon.txt?ts=%x\" --max-time 10 -vs 2>/dev/null", uint32_t(time(NULL)) );
    file = popen( cmd, "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
        return false;
    }

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    char cmd[1024];
    snprintf( cmd, sizeof(cmd), "powershell Invoke-RestMethod -Uri \"https://storage.googleapis.com/network_next_sdk_config/amazon.txt?ts=%x\" -TimeoutSec 10", uint32_t(time(NULL)) );
    file = _popen ( cmd, "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run powershell Invoke-RestMethod" );
        return false;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    bool found = false;

    while ( fgets( buffer, sizeof(buffer), file ) != NULL )
    {
        const char * separators = ",\n\r";

        char * amazon_zone = strtok( buffer, separators );
        if ( amazon_zone == NULL )
        {
            continue;
        }

        char * amazon_datacenter = strtok( NULL, separators );
        if ( amazon_datacenter == NULL )
        {
            continue;
        }

        if ( strcmp( azid, amazon_zone ) == 0 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: \"%s\" -> \"%s\"", azid, amazon_datacenter );
            strcpy( output, amazon_datacenter );
            found = true;
            break;
        }
    }

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    pclose( file );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    _pclose( file );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    return found;
}

// --------------------------------------------------------------------------------------------------------------

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

#include <sys/cdefs.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <ctype.h>
#include <err.h>
#include <netdb.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sysexits.h>
#include <unistd.h>

#define ANICHOST        "whois.arin.net"
#define LNICHOST        "whois.lacnic.net"
#define RNICHOST        "whois.ripe.net"
#define PNICHOST        "whois.apnic.net"
#define BNICHOST        "whois.registro.br"
#define AFRINICHOST     "whois.afrinic.net"

const char *ip_whois[] = { LNICHOST, RNICHOST, PNICHOST, BNICHOST, AFRINICHOST, NULL };

bool next_whois( const char * address, const char * hostname, int recurse, char ** buffer, size_t & bytes_remaining )
{
    struct addrinfo *hostres, *res;
    char *nhost;
    int i, s;
    size_t c;

    struct addrinfo hints;
    int error;
    memset(&hints, 0, sizeof(hints));
    hints.ai_flags = 0;
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    error = getaddrinfo(hostname, "nicname", &hints, &hostres);
    if ( error != 0 )
    {
        return 0;
    }

    for (res = hostres; res; res = res->ai_next) {
        s = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
        if (s < 0)
            continue;
        if (connect(s, res->ai_addr, res->ai_addrlen) == 0)
            break;
        close(s);
    }

    freeaddrinfo(hostres);

    if (res == NULL)
        return 0;

    FILE * sfi = fdopen( s, "r" );
    FILE * sfo = fdopen( s, "w" );
    if ( sfi == NULL || sfo == NULL )
        return 0;

    if (strcmp(hostname, "de.whois-servers.net") == 0) {
#ifdef __APPLE__
        fprintf(sfo, "-T dn -C UTF-8 %s\r\n", address);
#else
        fprintf(sfo, "-T dn,ace -C US-ASCII %s\r\n", address);
#endif
    } else {
        fprintf(sfo, "%s\r\n", address);
    }
    fflush(sfo);

    nhost = NULL;

    char buf[10*1024];

    while ( fgets(buf, sizeof(buf), sfi) )
    {
        size_t len = strlen(buf);

        if ( len < bytes_remaining )
        {
            memcpy( *buffer, buf, len );
            bytes_remaining -= len;
            *buffer += len;
        }

        if ( nhost == NULL )
        {
            if ( recurse && strcmp(hostname, ANICHOST) == 0 )
            {
                for (c = 0; c <= len; c++)
                {
                    buf[c] = tolower((int)buf[c]);
                }
                for (i = 0; ip_whois[i] != NULL; i++)
                {
                    if (strstr(buf, ip_whois[i]) != NULL)
                    {
                        int result = asprintf( &nhost, "%s", ip_whois[i] );  // note: nhost is allocated here
                        if ( result == -1 )
                        {
                            nhost = NULL;
                        }
                        break;
                    }
                }
            }
        }
    }

    close( s );
    fclose( sfo );
    fclose( sfi );

    bool result = true;

    if ( nhost != NULL)
    {
        result = next_whois( address, nhost, 0, buffer, bytes_remaining );
        free( nhost );
    }

    return result;
}

bool next_autodetect_multiplay( const char * input_datacenter, const char * address, char * output, size_t output_size )
{
    FILE * file;

    // are we in a multiplay datacenter?

    if ( strlen( input_datacenter ) <= 10 ||
         input_datacenter[0] != 'm' || 
         input_datacenter[1] != 'u' || 
         input_datacenter[2] != 'l' || 
         input_datacenter[3] != 't' || 
         input_datacenter[4] != 'i' || 
         input_datacenter[5] != 'p' || 
         input_datacenter[6] != 'l' || 
         input_datacenter[7] != 'a' || 
         input_datacenter[8] != 'y' || 
         input_datacenter[9] != '.' )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: not in multiplay" );
        return false;
    }

    // is this a non-autodetect multiplay datacenter? ("multiplay.[cityname].[number]")

    const int length = strlen( input_datacenter );

    int num_periods = 0;

    for ( int i = 0; i < length; ++i )
    {
        if ( input_datacenter[i] == '.' )
        {
            num_periods++;
            if ( num_periods > 1 )
            {
                strcpy( output, input_datacenter );
                return true;
            }
        }
    }

    // capture the city name using form multiplay.[cityname]

    const char * city = input_datacenter + 10;

    // try to read in cache of whois in whois.txt first

    bool have_cached_whois = false;
    char whois_buffer[1024*64];
    memset( whois_buffer, 0, sizeof(whois_buffer) );
    FILE * f = fopen( "whois.txt", "r");
    if ( f )
    {
        fseek( f, 0, SEEK_END );
        size_t fsize = ftell( f );
        fseek( f, 0, SEEK_SET );
        if ( fsize > sizeof(whois_buffer) - 1 )
        {
            fsize = sizeof(whois_buffer) - 1;
        }
        if ( fread( whois_buffer, fsize, 1, f ) == 1 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "server successfully read cached whois.txt" );
            have_cached_whois = true;
        }
        fclose( f );
    }

    // if we couldn't read whois.txt, run whois locally and store the result to whois.txt

    if ( !have_cached_whois )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server running whois locally" );
        char * whois_output = &whois_buffer[0];
        size_t bytes_remaining = sizeof(whois_buffer) - 1;
        next_whois( address, ANICHOST, 1, &whois_output, bytes_remaining );
           FILE * f = fopen( "whois.txt", "w" );
           if ( f )
           {
               next_printf( NEXT_LOG_LEVEL_INFO, "server cached whois result to whois.txt" );
               fputs( whois_buffer, f );
               fflush( f );
               fclose( f );
           }
    }

    // check against multiplay supplier mappings

    bool found = false;
    char multiplay_line[1024];
    char multiplay_buffer[64*1024];
    multiplay_buffer[0] = '\0';
    char cmd[1024];
    snprintf( cmd, sizeof(cmd), "curl -s \"https://storage.googleapis.com/network_next_sdk_config/multiplay.txt?ts=%x\" --max-time 10 -vs 2>/dev/null", uint32_t(time(NULL)) );
    file = popen( cmd, "r" );
    if ( !file )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: could not run curl" );
        return false;
    }
    while ( fgets( multiplay_line, sizeof(multiplay_line), file ) != NULL ) 
    {
        strcat( multiplay_buffer, multiplay_line );

        if ( found )
            continue;

        const char * separators = ",\n\r\n";

        char * substring = strtok( multiplay_line, separators );
        if ( substring == NULL )
        {
            continue;
        }

        char * supplier = strtok( NULL, separators );
        if ( supplier == NULL )
        {
            continue;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "checking for supplier \"%s\" with substring \"%s\"", supplier, substring );

        if ( strstr( whois_buffer, substring ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "found supplier %s", supplier );
            snprintf( output, output_size, "%s.%s", supplier, city );
            found = true;
        }
    }
    pclose( file );

    // could not autodetect multiplay :(

    if ( !found )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "could not autodetect multiplay datacenter :(" );
        next_printf( "-------------------------\n%s-------------------------\n", multiplay_buffer );
        const char * separators = "\n\r\n";
        char * line = strtok( whois_buffer, separators );
        while ( line )
        {
            next_printf( "%s", line );
            line = strtok( NULL, separators );
        }
        next_printf( "-------------------------\n" );
        return false;
    }

    return found;
}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

bool next_autodetect_datacenter( const char * input_datacenter, const char * public_address, char * output, size_t output_size )
{
#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC
    
    // we need curl to do any autodetect. bail if we don't have it

    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: looking for curl" );

    int result = system( "curl -s >/dev/null 2>&1" );

    if ( result < 0 )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: curl not found" );
        return false;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: curl exists" );

#elif NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    // we need access to powershell and Invoke-RestMethod to do any autodetect. bail if we don't have it

    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: looking for powershell Invoke-RestMethod" );

    int result = system( "powershell Invoke-RestMethod -? > NUL 2>&1" );

    if ( result > 0 )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: powershell Invoke-RestMethod not found" );
        return false;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter: powershell Invoke-RestMethod exists" );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    // google cloud

    bool google_result = next_autodetect_google( output );
    if ( google_result )
    {
        return true;
    }

    // amazon

    bool amazon_result = next_autodetect_amazon( output );
    if ( amazon_result )
    {
        return true;
    }

    // multiplay

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    bool multiplay_result = next_autodetect_multiplay( input_datacenter, public_address, output, output_size );
    if ( multiplay_result )
    {
        return true;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC

    (void) input_datacenter;
    (void) public_address;
    (void) output_size;

    return false;
}

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

void next_server_internal_resolve_hostname( next_server_internal_t * server )
{
    if ( server->resolving_hostname )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server is already resolving hostname" );
        return;
    }

    server->resolve_hostname_thread = next_platform_thread_create( server->context, next_server_internal_resolve_hostname_thread_function, server );
    if ( !server->resolve_hostname_thread )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create resolve hostname thread" );
        return;
    }

    server->resolve_hostname_start_time = next_platform_time();
    server->resolving_hostname = true;
    server->resolve_hostname_finished = false;
}

void next_server_internal_autodetect( next_server_internal_t * server )
{
    if ( server->autodetecting )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server is already autodetecting" );
        return;
    }

    server->autodetect_thread = next_platform_thread_create( server->context, next_server_internal_autodetect_thread_function, server );
    if ( !server->autodetect_thread )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create autodetect thread" );
        return;
    }

    server->autodetect_start_time = next_platform_time();
    server->autodetecting = true;
}

void next_server_internal_initialize( next_server_internal_t * server )
{
    if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server initializing with backend" );

        server->state = NEXT_SERVER_STATE_INITIALIZING;
        server->server_init_timeout_time = next_platform_time() + NEXT_SERVER_INIT_TIMEOUT;
    }
    
    next_server_internal_resolve_hostname( server );

    next_server_internal_autodetect( server );
}

void next_server_internal_destroy( next_server_internal_t * server );

static void next_server_internal_thread_function( void * context );

next_server_internal_t * next_server_internal_create( void * context, const char * server_address_string, const char * bind_address_string, const char * datacenter_string )
{
#if !NEXT_DEVELOPMENT
    next_printf( NEXT_LOG_LEVEL_INFO, "server sdk version is %s", NEXT_VERSION_FULL );
#endif // #if !NEXT_DEVELOPMENT

    next_assert( server_address_string );
    next_assert( bind_address_string );
    next_assert( datacenter_string );

    next_printf( NEXT_LOG_LEVEL_INFO, "server buyer id is %" PRIx64, next_global_config.server_customer_id );

    const char * server_address_override = next_platform_getenv( "NEXT_SERVER_ADDRESS" );
    if ( server_address_override )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server address override: '%s'", server_address_override );
        server_address_string = server_address_override;
    }

    next_address_t server_address;
    if ( next_address_parse( &server_address, server_address_string ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to parse server address: '%s'", server_address_string );
        return NULL;
    }

    const char * bind_address_override = next_platform_getenv( "NEXT_BIND_ADDRESS" );
    if ( bind_address_override )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server bind address override: '%s'", bind_address_override );
        bind_address_string = bind_address_override;
    }

    next_address_t bind_address;
    if ( next_address_parse( &bind_address, bind_address_string ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to parse bind address: '%s'", bind_address_string );
        return NULL;
    }

    next_server_internal_t * server = (next_server_internal_t*) next_malloc( context, sizeof(next_server_internal_t) );
    if ( !server )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create internal server" );
        return NULL;
    }

    char * just_clear_it_and_dont_complain = (char*) server;
    memset( just_clear_it_and_dont_complain, 0, sizeof(next_server_internal_t) );

    next_server_internal_initialize_sentinels( server );

    next_server_internal_verify_sentinels( server );

    server->context = context;
    server->customer_id = next_global_config.server_customer_id;
    memcpy( server->customer_private_key, next_global_config.customer_private_key, NEXT_CRYPTO_SIGN_SECRETKEYBYTES );
    server->valid_customer_private_key = next_global_config.valid_customer_private_key;

    const char * datacenter = datacenter_string;

    const char * datacenter_env = next_platform_getenv( "NEXT_DATACENTER" );

    if ( datacenter_env )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server datacenter override '%s'", datacenter_env );
        datacenter = datacenter_env;
    }

    next_assert( datacenter );

    next_copy_string( server->autodetect_input, datacenter, NEXT_MAX_DATACENTER_NAME_LENGTH );

    const bool datacenter_is_empty_string = datacenter[0] == '\0';

    const bool datacenter_is_cloud = datacenter[0] == 'c' &&
                                     datacenter[1] == 'l' &&
                                     datacenter[2] == 'o' &&
                                     datacenter[3] == 'u' &&
                                     datacenter[4] == 'd' &&
                                     datacenter[5] == '\n';

    if ( !datacenter_is_empty_string && !datacenter_is_cloud )
    {
        server->datacenter_id = next_datacenter_id( datacenter );
        next_copy_string( server->datacenter_name, datacenter, NEXT_MAX_DATACENTER_NAME_LENGTH );
        next_printf( NEXT_LOG_LEVEL_INFO, "server input datacenter is '%s' [%" PRIx64 "]", server->datacenter_name, server->datacenter_id );
    }
    else
    {
        server->no_datacenter_specified = true;
    }

    server->command_queue = next_queue_create( context, NEXT_COMMAND_QUEUE_LENGTH );
    if ( !server->command_queue )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create command queue" );
        next_server_internal_destroy( server );
        return NULL;
    }

    server->notify_queue = next_queue_create( context, NEXT_NOTIFY_QUEUE_LENGTH );
    if ( !server->notify_queue )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create notify queue" );
        next_server_internal_destroy( server );
        return NULL;
    }

    server->socket = next_platform_socket_create( server->context, &bind_address, NEXT_PLATFORM_SOCKET_BLOCKING, 0.1f, next_global_config.socket_send_buffer_size, next_global_config.socket_receive_buffer_size, true );
    if ( server->socket == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create server socket" );
        next_server_internal_destroy( server );
        return NULL;
    }

    if ( server_address.port == 0 )
    {
        server_address.port = bind_address.port;
    }

    char address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_INFO, "server bound to %s", next_address_to_string( &bind_address, address_string ) );

    server->bind_address = bind_address;
    server->server_address = server_address;

    int result = next_platform_mutex_create( &server->session_mutex );
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create session mutex" );
        next_server_internal_destroy( server );
        return NULL;
    }

    result = next_platform_mutex_create( &server->command_mutex );

    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create command mutex" );
        next_server_internal_destroy( server );
        return NULL;
    }

    result = next_platform_mutex_create( &server->notify_mutex );

    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create notify mutex" );
        next_server_internal_destroy( server );
        return NULL;
    }

    result = next_platform_mutex_create( &server->resolve_hostname_mutex );

    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create resolve hostname mutex" );
        next_server_internal_destroy( server );
        return NULL;
    }

    result = next_platform_mutex_create( &server->autodetect_mutex );
    
    if ( result != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create autodetect mutex" );
        next_server_internal_destroy( server );
        return NULL;
    }

    server->pending_session_manager = next_pending_session_manager_create( context, NEXT_INITIAL_PENDING_SESSION_SIZE );
    if ( server->pending_session_manager == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create pending session manager" );
        next_server_internal_destroy( server );
        return NULL;
    }

    server->session_manager = next_session_manager_create( context, NEXT_INITIAL_SESSION_SIZE );
    if ( server->session_manager == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create session manager" );
        next_server_internal_destroy( server );
        return NULL;
    }

    if ( !next_global_config.disable_network_next && server->valid_customer_private_key )
    {
        next_server_internal_initialize( server );
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "server started on %s", next_address_to_string( &server_address, address_string ) );

    next_crypto_kx_keypair( server->server_kx_public_key, server->server_kx_private_key );

    next_crypto_box_keypair( server->server_route_public_key, server->server_route_private_key );

    server->server_update_last_time = next_platform_time() - NEXT_SECONDS_BETWEEN_SERVER_UPDATES * next_random_float();

    server->server_update_first = true;

    return server;
}

void next_server_internal_destroy( next_server_internal_t * server )
{
    next_assert( server );

    next_server_internal_verify_sentinels( server );

    if ( server->socket )
    {
        next_platform_socket_destroy( server->socket );
    }
    if ( server->resolve_hostname_thread )
    {
        next_platform_thread_destroy( server->resolve_hostname_thread );
    }
    if ( server->command_queue )
    {
        next_queue_destroy( server->command_queue );
    }
    if ( server->notify_queue )
    {
        next_queue_destroy( server->notify_queue );
    }
    if ( server->session_manager )
    {
        next_session_manager_destroy( server->session_manager );
        server->session_manager = NULL;
    }
    if ( server->pending_session_manager )
    {
        next_pending_session_manager_destroy( server->pending_session_manager );
        server->pending_session_manager = NULL;
    }

    next_platform_mutex_destroy( &server->session_mutex );
    next_platform_mutex_destroy( &server->command_mutex );
    next_platform_mutex_destroy( &server->notify_mutex );
    next_platform_mutex_destroy( &server->resolve_hostname_mutex );
    next_platform_mutex_destroy( &server->autodetect_mutex );

    next_server_internal_verify_sentinels( server );

    next_clear_and_free( server->context, server, sizeof(next_server_internal_t) );
}

void next_server_internal_quit( next_server_internal_t * server )
{
    next_assert( server );
    server->quit = 1;
}

void next_server_internal_send_packet_to_address( next_server_internal_t * server, const next_address_t * address, const uint8_t * packet_data, int packet_bytes )
{
    next_server_internal_verify_sentinels( server );

    next_assert( address );
    next_assert( address->type != NEXT_ADDRESS_NONE );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

    if ( server->send_packet_to_address_callback )
    {
        void * callback_data = server->send_packet_to_address_callback_data;
        if ( server->send_packet_to_address_callback( callback_data, address, packet_data, packet_bytes ) != 0 )
            return;
    }

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( server->socket, address, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING
}

void next_server_internal_send_packet_to_backend( next_server_internal_t * server, const uint8_t * packet_data, int packet_bytes )
{
    next_server_internal_verify_sentinels( server );

    if ( server->backend_address.type == NEXT_ADDRESS_NONE )
        return;

    next_assert( server->backend_address.type != NEXT_ADDRESS_NONE );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( server->socket, &server->backend_address, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING
}

int next_server_internal_send_packet( next_server_internal_t * server, const next_address_t * to_address, uint8_t packet_id, void * packet_object )
{
    next_assert( server );
    next_assert( server->socket );
    next_assert( packet_object );

    next_server_internal_verify_sentinels( server );

    int packet_bytes = 0;

    uint8_t buffer[NEXT_MAX_PACKET_BYTES];

    uint64_t * sequence = NULL;
    uint8_t * send_key = NULL;

    uint8_t magic[8];
    if ( packet_id != NEXT_UPGRADE_REQUEST_PACKET )
    {
        memcpy( magic, server->current_magic, 8 );
    }
    else
    {
        memset( magic, 0, sizeof(magic) );
    }

    if ( next_encrypted_packets[packet_id] )
    {
        next_session_entry_t * session = next_session_manager_find_by_address( server->session_manager, to_address );

        if ( !session )
        {
            next_printf( NEXT_LOG_LEVEL_WARN, "server can't send encrypted packet to address. no session found" );
            return NEXT_ERROR;
        }

        sequence = &session->internal_send_sequence;
        send_key = session->send_key;
    }

    uint8_t from_address_data[32];
    uint8_t to_address_data[32];
    uint16_t from_address_port = 0;
    uint16_t to_address_port = 0;
    int from_address_bytes = 0;
    int to_address_bytes = 0;

    next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );

    // IMPORTANT: when the upgrade request packet is sent, the client doesn't know it's external address yet
    // so we must encode with a to address of zero bytes for the upgrade request packet
    if ( packet_id != NEXT_UPGRADE_REQUEST_PACKET )
    {
        next_address_data( to_address, to_address_data, &to_address_bytes, &to_address_port );
    }

    if ( next_write_packet( packet_id, packet_object, buffer, &packet_bytes, next_signed_packets, next_encrypted_packets, sequence, server->customer_private_key, send_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write internal packet with id %d", packet_id );
        return NEXT_ERROR;
    }

    next_assert( packet_bytes > 0 );
    next_assert( next_basic_packet_filter( buffer, packet_bytes ) );
    next_assert( next_advanced_packet_filter( buffer, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

    next_server_internal_send_packet_to_address( server, to_address, buffer, packet_bytes );

    return NEXT_OK;
}

inline int next_sequence_greater_than( uint8_t s1, uint8_t s2 )
{
    return ( ( s1 > s2 ) && ( s1 - s2 <= 128 ) ) ||
           ( ( s1 < s2 ) && ( s2 - s1  > 128 ) );
}

next_session_entry_t * next_server_internal_process_client_to_server_packet( next_server_internal_t * server, uint8_t packet_type, uint8_t * packet_data, int packet_bytes )
{
    next_assert( server );
    next_assert( packet_data );

    next_server_internal_verify_sentinels( server );

    if ( packet_bytes <= NEXT_HEADER_BYTES )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client to server packet. packet is too small to be valid" );
        return NULL;
    }

    uint64_t packet_sequence = 0;
    uint64_t packet_session_id = 0;
    uint8_t packet_session_version = 0;

    next_peek_header( &packet_sequence, &packet_session_id, &packet_session_version, packet_data, packet_bytes );

    next_session_entry_t * entry = next_session_manager_find_by_session_id( server->session_manager, packet_session_id );
    if ( !entry )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client to server packet. could not find session" );
        return NULL;
    }

    if ( !entry->has_pending_route && !entry->has_current_route && !entry->has_previous_route )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client to server packet. session has no route" );
        return NULL;
    }

    next_assert( packet_type == NEXT_CLIENT_TO_SERVER_PACKET || packet_type == NEXT_PING_PACKET );

    next_replay_protection_t * replay_protection = ( packet_type == NEXT_CLIENT_TO_SERVER_PACKET ) ? &entry->payload_replay_protection : &entry->special_replay_protection;

    if ( next_replay_protection_already_received( replay_protection, packet_sequence ) )
        return NULL;

    if ( entry->has_pending_route && next_read_header( packet_type, &packet_sequence, &packet_session_id, &packet_session_version, entry->pending_route_private_key, packet_data, packet_bytes ) == NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server promoted pending route for session %" PRIx64, entry->session_id );

        if ( entry->has_current_route )
        {
            entry->has_previous_route = true;
            entry->previous_route_send_address = entry->current_route_send_address;
            memcpy( entry->previous_route_private_key, entry->current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
        }

        entry->has_pending_route = false;
        entry->has_current_route = true;
        entry->current_route_session_version = entry->pending_route_session_version;
        entry->current_route_expire_timestamp = entry->pending_route_expire_timestamp;
        entry->current_route_expire_time = entry->pending_route_expire_time;
        entry->current_route_kbps_up = entry->pending_route_kbps_up;
        entry->current_route_kbps_down = entry->pending_route_kbps_down;
        entry->current_route_send_address = entry->pending_route_send_address;
        memcpy( entry->current_route_private_key, entry->pending_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );

        {
            next_platform_mutex_guard( &server->session_mutex );
            entry->mutex_envelope_kbps_up = entry->current_route_kbps_up;
            entry->mutex_envelope_kbps_down = entry->current_route_kbps_down;
            entry->mutex_send_over_network_next = true;
            entry->mutex_session_id = entry->session_id;
            entry->mutex_session_version = entry->current_route_session_version;
            entry->mutex_send_address = entry->current_route_send_address;
            memcpy( entry->mutex_private_key, entry->current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
        }
    }
    else
    {
        bool current_route_ok = false;
        bool previous_route_ok = false;

        if ( entry->has_current_route )
            current_route_ok = next_read_header( packet_type, &packet_sequence, &packet_session_id, &packet_session_version, entry->current_route_private_key, packet_data, packet_bytes ) == NEXT_OK;

        if ( entry->has_previous_route )
            previous_route_ok = next_read_header( packet_type, &packet_sequence, &packet_session_id, &packet_session_version, entry->previous_route_private_key, packet_data, packet_bytes ) == NEXT_OK;

        if ( !current_route_ok && !previous_route_ok )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client to server packet. did not verify" );
            return NULL;
        }
    }

    next_replay_protection_advance_sequence( replay_protection, packet_sequence );

    if ( packet_type == NEXT_CLIENT_TO_SERVER_PACKET )
    {
        next_packet_loss_tracker_packet_received( &entry->packet_loss_tracker, packet_sequence );
        next_out_of_order_tracker_packet_received( &entry->out_of_order_tracker, packet_sequence );
        next_jitter_tracker_packet_received( &entry->jitter_tracker, packet_sequence, next_platform_time() );
    }

    return entry;
}

void next_server_internal_update_route( next_server_internal_t * server )
{
    next_assert( server );

    next_server_internal_verify_sentinels( server );

    next_assert( !next_global_config.disable_network_next );

    if ( server->flushing )
        return;

    const double current_time = next_platform_time();

    const int max_index = server->session_manager->max_entry_index;

    for ( int i = 0; i <= max_index; ++i )
    {
        if ( server->session_manager->session_ids[i] == 0 )
            continue;

        next_session_entry_t * entry = &server->session_manager->entries[i];

        if ( entry->update_dirty && !entry->client_ping_timed_out && !entry->stats_fallback_to_direct && entry->update_last_send_time + NEXT_UPDATE_SEND_TIME <= current_time )
        {
            NextRouteUpdatePacket packet;
            memcpy( packet.upcoming_magic, server->upcoming_magic, 8 );
            memcpy( packet.current_magic, server->current_magic, 8 );
            memcpy( packet.previous_magic, server->previous_magic, 8 );
            packet.sequence = entry->update_sequence;
            packet.has_near_relays = entry->update_has_near_relays;
            if ( packet.has_near_relays )
            {
                packet.num_near_relays = entry->update_num_near_relays;
                memcpy( packet.near_relay_ids, entry->update_near_relay_ids, size_t(8) * entry->update_num_near_relays );
                memcpy( packet.near_relay_addresses, entry->update_near_relay_addresses, sizeof(next_address_t) * entry->update_num_near_relays );
                memcpy( packet.near_relay_ping_tokens, entry->update_near_relay_ping_tokens, NEXT_MAX_NEAR_RELAYS * NEXT_PING_TOKEN_BYTES );
                packet.near_relay_expire_timestamp = entry->update_near_relay_expire_timestamp;
            }
            packet.update_type = entry->update_type;
            packet.multipath = entry->multipath;
            packet.num_tokens = entry->update_num_tokens;
            if ( entry->update_type == NEXT_UPDATE_TYPE_ROUTE )
            {
                memcpy( packet.tokens, entry->update_tokens, NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES * size_t(entry->update_num_tokens) );
            }
            else if ( entry->update_type == NEXT_UPDATE_TYPE_CONTINUE )
            {
                memcpy( packet.tokens, entry->update_tokens, NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES * size_t(entry->update_num_tokens) );
            }
            packet.packets_lost_client_to_server = entry->stats_packets_lost_client_to_server;
            packet.packets_out_of_order_client_to_server = entry->stats_packets_out_of_order_client_to_server;
            packet.jitter_client_to_server = float( entry->stats_jitter_client_to_server );

            {
                next_platform_mutex_guard( &server->session_mutex );
                packet.packets_sent_server_to_client = entry->stats_packets_sent_server_to_client;
            }

            packet.has_debug = entry->has_debug;
            memcpy( packet.debug, entry->debug, NEXT_MAX_SESSION_DEBUG );

            next_server_internal_send_packet( server, &entry->address, NEXT_ROUTE_UPDATE_PACKET, &packet );

            entry->update_last_send_time = current_time;

            next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent route update packet to session %" PRIx64, entry->session_id );
        }
    }
}

void next_server_internal_update_pending_upgrades( next_server_internal_t * server )
{
    next_assert( server );

    next_server_internal_verify_sentinels( server );

    next_assert( !next_global_config.disable_network_next );

    if ( server->flushing )
        return;

    if ( server->state == NEXT_SERVER_STATE_DIRECT_ONLY )
        return;

    const double current_time = next_platform_time();

    const double packet_resend_time = 0.25;

    const int max_index = server->pending_session_manager->max_entry_index;

    for ( int i = 0; i <= max_index; ++i )
    {
        if ( server->pending_session_manager->addresses[i].type == NEXT_ADDRESS_NONE )
            continue;

        next_pending_session_entry_t * entry = &server->pending_session_manager->entries[i];

        if ( entry->upgrade_time + NEXT_UPGRADE_TIMEOUT <= current_time )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server upgrade request timed out for client %s", next_address_to_string( &entry->address, address_buffer ) );
            next_pending_session_manager_remove_at_index( server->pending_session_manager, i );
            next_server_notify_pending_session_timed_out_t * notify = (next_server_notify_pending_session_timed_out_t*) next_malloc( server->context, sizeof( next_server_notify_pending_session_timed_out_t ) );
            notify->type = NEXT_SERVER_NOTIFY_PENDING_SESSION_TIMED_OUT;
            notify->address = entry->address;
            notify->session_id = entry->session_id;
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_PENDING_SESSION_TIMED_OUT at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
                next_platform_mutex_guard( &server->notify_mutex );
                next_queue_push( server->notify_queue, notify );
            }
            continue;
        }

        if ( entry->last_packet_send_time + packet_resend_time <= current_time )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent upgrade request packet to client %s", next_address_to_string( &entry->address, address_buffer ) );

            entry->last_packet_send_time = current_time;

            NextUpgradeRequestPacket packet;
            packet.protocol_version = next_protocol_version();
            packet.session_id = entry->session_id;
            packet.client_address = entry->address;
            packet.server_address = server->server_address;
            memcpy( packet.server_kx_public_key, server->server_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
            memcpy( packet.upgrade_token, entry->upgrade_token, NEXT_UPGRADE_TOKEN_BYTES );
            memcpy( packet.upcoming_magic, server->upcoming_magic, 8 );
            memcpy( packet.current_magic, server->current_magic, 8 );
            memcpy( packet.previous_magic, server->previous_magic, 8 );

            next_server_internal_send_packet( server, &entry->address, NEXT_UPGRADE_REQUEST_PACKET, &packet );
        }
    }
}

void next_server_internal_update_sessions( next_server_internal_t * server )
{
    next_assert( server );

    next_server_internal_verify_sentinels( server );

    next_assert( !next_global_config.disable_network_next );

    if ( server->state == NEXT_SERVER_STATE_DIRECT_ONLY )
        return;

    const double current_time = next_platform_time();

    int index = 0;

    while ( index <= server->session_manager->max_entry_index )
    {
        if ( server->session_manager->session_ids[index] == 0 )
        {
            ++index;
            continue;
        }

        next_session_entry_t * entry = &server->session_manager->entries[index];

        if ( !entry->client_ping_timed_out &&
             entry->last_client_direct_ping + NEXT_SERVER_PING_TIMEOUT <= current_time &&
             entry->last_client_next_ping + NEXT_SERVER_PING_TIMEOUT <= current_time )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server client ping timed out for session %" PRIx64, entry->session_id );
            entry->client_ping_timed_out = true;
        }

        // IMPORTANT: Don't time out sessions during server flush. Otherwise the server flush might wait longer than necessary.
        if ( !server->flushing && entry->last_client_stats_update + NEXT_SERVER_SESSION_TIMEOUT <= current_time )
        {
            next_server_notify_session_timed_out_t * notify = (next_server_notify_session_timed_out_t*) next_malloc( server->context, sizeof( next_server_notify_session_timed_out_t ) );
            notify->type = NEXT_SERVER_NOTIFY_SESSION_TIMED_OUT;
            notify->address = entry->address;
            notify->session_id = entry->session_id;
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_SESSION_TIMED_OUT at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
                next_platform_mutex_guard( &server->notify_mutex );
                next_queue_push( server->notify_queue, notify );
            }

            {
                next_platform_mutex_guard( &server->session_mutex );
                next_session_manager_remove_at_index( server->session_manager, index );
            }
    
            continue;
        }

        if ( entry->has_current_route && entry->current_route_expire_time <= current_time )
        {
            // IMPORTANT: Only print this out as an error if it occurs *before* the client ping times out
            // otherwise we get red herring errors on regular client disconnect from server that make it
            // look like something is wrong when everything is fine...
            if ( !entry->client_ping_timed_out )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "server network next route expired for session %" PRIx64, entry->session_id );
            }

            entry->has_current_route = false;
            entry->has_previous_route = false;
            entry->update_dirty = false;
            entry->waiting_for_update_response = false;

            {
                next_platform_mutex_guard( &server->session_mutex );
                entry->mutex_send_over_network_next = false;
            }
        }

        index++;
    }
}

void next_server_internal_update_flush( next_server_internal_t * server )
{
    next_assert( !next_global_config.disable_network_next );

    if ( !server->flushing )
        return;

    if ( server->flushed )
        return;

    if ( next_global_config.disable_network_next || server->state != NEXT_SERVER_STATE_INITIALIZED || 
         ( server->num_flushed_session_updates == server->num_session_updates_to_flush && server->num_flushed_match_data == server->num_match_data_to_flush ) )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server internal flush completed" );
        
        server->flushed = true;

        next_server_notify_flush_finished_t * notify = (next_server_notify_flush_finished_t*) next_malloc( server->context, sizeof( next_server_notify_flush_finished_t ) );
        notify->type = NEXT_SERVER_NOTIFY_FLUSH_FINISHED;
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_FLUSH_FINISHED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify );
        }
    }
}

void next_server_internal_process_network_next_packet( next_server_internal_t * server, const next_address_t * from, uint8_t * packet_data, int begin, int end )
{
    next_assert( server );
    next_assert( from );
    next_assert( packet_data );
    next_assert( begin >= 0 );
    next_assert( end <= NEXT_MAX_PACKET_BYTES );

    next_server_internal_verify_sentinels( server );

    if ( next_global_config.disable_network_next )
        return;

    const int packet_id = packet_data[begin];

#if NEXT_ASSERT
    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_SPAM, "server received packet type %d from %s (%d bytes)", packet_id, next_address_to_string( from, address_buffer ), packet_bytes );
#endif // #if NEXT_ASSERT

    // run packet filters
    {
        if ( !next_basic_packet_filter( packet_data + begin, end - begin ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server basic packet filter dropped packet" );
            return;
        }

        uint8_t from_address_data[32];
        uint8_t to_address_data[32];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( from, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( &server->server_address, to_address_data, &to_address_bytes, &to_address_port );

        if ( packet_id != NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET &&
             packet_id != NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET &&
             packet_id != NEXT_BACKEND_SERVER_UPDATE_REQUEST_PACKET &&
             packet_id != NEXT_BACKEND_SERVER_UPDATE_RESPONSE_PACKET &&
             packet_id != NEXT_BACKEND_SESSION_UPDATE_RESPONSE_PACKET && 
             packet_id != NEXT_BACKEND_MATCH_DATA_RESPONSE_PACKET )
        {
            if ( !next_advanced_packet_filter( packet_data + begin, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, end - begin ) )
            {
                if ( !next_advanced_packet_filter( packet_data + begin, server->upcoming_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, end - begin ) )
                {
                    if ( !next_advanced_packet_filter( packet_data + begin, server->previous_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, end - begin ) )
                    {
                        next_printf( NEXT_LOG_LEVEL_DEBUG, "server advanced packet filter dropped packet" );
                        return;
                    }
                }
            }
        }
        else
        {
            uint8_t magic[8];
            memset( magic, 0, sizeof(magic) );
            if ( !next_advanced_packet_filter( packet_data + begin, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, end - begin ) )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server advanced packet filter dropped packet (backend)" );
                return;
            }
        }
    }

    begin += 16;
    end -= 2;

    if ( server->state == NEXT_SERVER_STATE_INITIALIZING )
    {
        // server init response

        if ( packet_id == NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET )
        {
            if ( server->state != NEXT_SERVER_STATE_INITIALIZING )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored init response packet from backend. server is not initializing" );
                return;
            }

            NextBackendServerInitResponsePacket packet;

            if ( next_read_backend_packet( packet_id, packet_data, begin, end, &packet, next_signed_packets, next_server_backend_public_key ) != packet_id )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored server init response packet from backend. packet failed to read" );
                return;
            }

            if ( packet.request_id != server->server_init_request_id )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored server init response packet from backend. request id mismatch (got %" PRIx64 ", expected %" PRIx64 ")", packet.request_id, server->server_init_request_id );
                return;
            }

            next_printf( NEXT_LOG_LEVEL_INFO, "server received init response from backend" );

            if ( packet.response != NEXT_SERVER_INIT_RESPONSE_OK )
            {
                switch ( packet.response )
                {
                    case NEXT_SERVER_INIT_RESPONSE_UNKNOWN_CUSTOMER:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend. unknown customer" );
                        return;

                    case NEXT_SERVER_INIT_RESPONSE_UNKNOWN_DATACENTER:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend. unknown datacenter" );
                        return;

                    case NEXT_SERVER_INIT_RESPONSE_SDK_VERSION_TOO_OLD:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend. sdk version too old" );
                        return;

                    case NEXT_SERVER_INIT_RESPONSE_SIGNATURE_CHECK_FAILED:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend. signature check failed" );
                        return;

                    case NEXT_SERVER_INIT_RESPONSE_CUSTOMER_NOT_ACTIVE:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend. customer not active" );
                        return;

                    case NEXT_SERVER_INIT_RESPONSE_DATACENTER_NOT_ENABLED:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend. datacenter not enabled" );
                        return;

                    default:
                        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to initialize with backend" );
                        return;
                }
            }

            next_printf( NEXT_LOG_LEVEL_INFO, "welcome to network next :)" );

            server->received_init_response = true;

            memcpy( server->upcoming_magic, packet.upcoming_magic, 8 );
            memcpy( server->current_magic, packet.current_magic, 8 );
            memcpy( server->previous_magic, packet.previous_magic, 8 );

            next_printf( NEXT_LOG_LEVEL_DEBUG, "server initial magic: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
                packet.upcoming_magic[0],
                packet.upcoming_magic[1],
                packet.upcoming_magic[2],
                packet.upcoming_magic[3],
                packet.upcoming_magic[4],
                packet.upcoming_magic[5],
                packet.upcoming_magic[6],
                packet.upcoming_magic[7],
                packet.current_magic[0],
                packet.current_magic[1],
                packet.current_magic[2],
                packet.current_magic[3],
                packet.current_magic[4],
                packet.current_magic[5],
                packet.current_magic[6],
                packet.current_magic[7],
                packet.previous_magic[0],
                packet.previous_magic[1],
                packet.previous_magic[2],
                packet.previous_magic[3],
                packet.previous_magic[4],
                packet.previous_magic[5],
                packet.previous_magic[6],
                packet.previous_magic[7] );

            next_server_notify_magic_updated_t * notify = (next_server_notify_magic_updated_t*) next_malloc( server->context, sizeof( next_server_notify_magic_updated_t ) );
            notify->type = NEXT_SERVER_NOTIFY_MAGIC_UPDATED;
            memcpy( notify->current_magic, server->current_magic, 8 );
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_MAGIC_UPDATED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
                next_platform_mutex_guard( &server->notify_mutex );
                next_queue_push( server->notify_queue, notify );
            }

            return;
        }
    }

    // don't process network next packets until the server is initialized

    if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored network next packet because it is not initialized" );
        return;
    }

    // direct packet

    if ( packet_id == NEXT_DIRECT_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing direct packet" );

        const int packet_bytes = end - begin;

        if ( packet_bytes <= 9 )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored direct packet from %s. packet is too small to be valid", next_address_to_string( from, address_buffer ) );
            return;
        }

        if ( packet_bytes > NEXT_MTU + 9 )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored direct packet from %s. packet is too large to be valid", next_address_to_string( from, address_buffer ) );
            return;
        }

        const uint8_t * p = packet_data + begin;

        uint8_t packet_session_sequence = next_read_uint8( &p );

        uint64_t packet_sequence = next_read_uint64( &p );

        next_session_entry_t * entry = next_session_manager_find_by_address( server->session_manager, from );
        if ( !entry )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored direct packet from %s. could not find session for address", next_address_to_string( from, address_buffer ) );
            return;
        }

        if ( packet_session_sequence != entry->client_open_session_sequence )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored direct packet from %s. session mismatch", next_address_to_string( from, address_buffer ) );
            return;
        }

        if ( next_replay_protection_already_received( &entry->payload_replay_protection, packet_sequence ) )
            return;

        next_replay_protection_advance_sequence( &entry->payload_replay_protection, packet_sequence );

        next_packet_loss_tracker_packet_received( &entry->packet_loss_tracker, packet_sequence );

        next_out_of_order_tracker_packet_received( &entry->out_of_order_tracker, packet_sequence );

        next_jitter_tracker_packet_received( &entry->jitter_tracker, packet_sequence, next_platform_time() );

        next_server_notify_packet_received_t * notify = (next_server_notify_packet_received_t*) next_malloc( server->context, sizeof( next_server_notify_packet_received_t ) );
        notify->type = NEXT_SERVER_NOTIFY_PACKET_RECEIVED;
        notify->from = *from;
        notify->packet_bytes = packet_bytes - 9;
        next_assert( notify->packet_bytes > 0 );
        next_assert( notify->packet_bytes <= NEXT_MTU );
        memcpy( notify->packet_data, packet_data + begin + 9, size_t(notify->packet_bytes) );
        {
#if NEXT_SPIKE_TRACKING
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_PACKET_RECEIVED at %s:%d - from = %s, packet_bytes = %d", __FILE__, __LINE__, next_address_to_string( &notify->from, address_buffer ), notify->packet_bytes );
#endif // #if NEXT_SPIKE_TRACKING                
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify );
        }

        return;
    }

    // backend server response

    if ( packet_id == NEXT_BACKEND_SERVER_UPDATE_RESPONSE_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing server update response packet" );

        NextBackendServerUpdateResponsePacket packet;

        if ( next_read_backend_packet( packet_id, packet_data, begin, end, &packet, next_signed_packets, next_server_backend_public_key ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored server update response packet from backend. packet failed to read" );
            return;
        }

        if ( packet.request_id != server->server_update_request_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored server update response packet from backend. request id does not match" );
            return;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server received server update response packet from backend" );

        server->server_update_request_id = 0;
        server->server_update_resend_time = 0.0;

        if ( memcmp( packet.upcoming_magic, server->upcoming_magic, 8 ) != 0 )
        {
            memcpy( server->upcoming_magic, packet.upcoming_magic, 8 );
            memcpy( server->current_magic, packet.current_magic, 8 );
            memcpy( server->previous_magic, packet.previous_magic, 8 );

            next_printf( NEXT_LOG_LEVEL_DEBUG, "server updated magic: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
                packet.upcoming_magic[0],
                packet.upcoming_magic[1],
                packet.upcoming_magic[2],
                packet.upcoming_magic[3],
                packet.upcoming_magic[4],
                packet.upcoming_magic[5],
                packet.upcoming_magic[6],
                packet.upcoming_magic[7],
                packet.current_magic[0],
                packet.current_magic[1],
                packet.current_magic[2],
                packet.current_magic[3],
                packet.current_magic[4],
                packet.current_magic[5],
                packet.current_magic[6],
                packet.current_magic[7],
                packet.previous_magic[0],
                packet.previous_magic[1],
                packet.previous_magic[2],
                packet.previous_magic[3],
                packet.previous_magic[4],
                packet.previous_magic[5],
                packet.previous_magic[6],
                packet.previous_magic[7] );

            next_server_notify_magic_updated_t * notify = (next_server_notify_magic_updated_t*) next_malloc( server->context, sizeof( next_server_notify_magic_updated_t ) );
            notify->type = NEXT_SERVER_NOTIFY_MAGIC_UPDATED;
            memcpy( notify->current_magic, server->current_magic, 8 );
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_MAGIC_UPDATED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
                next_platform_mutex_guard( &server->notify_mutex );
                next_queue_push( server->notify_queue, notify );
            }
        }
    }

    // backend session response

    if ( packet_id == NEXT_BACKEND_SESSION_UPDATE_RESPONSE_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing session update response packet" );

        NextBackendSessionUpdateResponsePacket packet;

        if ( next_read_backend_packet( packet_id, packet_data, begin, end, &packet, next_signed_packets, next_server_backend_public_key ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored session update response packet from backend. packet failed to read" );
            return;
        }

        next_session_entry_t * entry = next_session_manager_find_by_session_id( server->session_manager, packet.session_id );
        if ( !entry )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored session update response packet from backend. could not find session %" PRIx64, packet.session_id );
            return;
        }

        if ( !entry->waiting_for_update_response )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored session update response packet from backend. not waiting for session response" );
            return;
        }

        if ( packet.slice_number != entry->update_sequence - 1 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored session update response packet from backend. wrong sequence number" );
            return;
        }

        const char * update_type = "???";

        switch ( packet.response_type )
        {
            case NEXT_UPDATE_TYPE_DIRECT:    update_type = "direct route";     break;
            case NEXT_UPDATE_TYPE_ROUTE:     update_type = "next route";       break;
            case NEXT_UPDATE_TYPE_CONTINUE:  update_type = "continue route";   break;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server received session update response from backend for session %" PRIx64 " (%s)", entry->session_id, update_type );

        bool multipath = packet.multipath;

        if ( multipath && !entry->multipath )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server multipath enabled for session %" PRIx64, entry->session_id );
            entry->multipath = true;
            {
                next_platform_mutex_guard( &server->session_mutex );
                entry->mutex_multipath = true;
            }
        }

        entry->update_dirty = true;

        entry->update_type = (uint8_t) packet.response_type;

        entry->update_num_tokens = packet.num_tokens;

        if ( packet.response_type == NEXT_UPDATE_TYPE_ROUTE )
        {
            memcpy( entry->update_tokens, packet.tokens, NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES * size_t(packet.num_tokens) );
        }
        else if ( packet.response_type == NEXT_UPDATE_TYPE_CONTINUE )
        {
            memcpy( entry->update_tokens, packet.tokens, NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES * size_t(packet.num_tokens) );
        }

        entry->update_has_near_relays = packet.has_near_relays;
        if ( packet.has_near_relays )
        {
            entry->update_num_near_relays = packet.num_near_relays;
            memcpy( entry->update_near_relay_ids, packet.near_relay_ids, 8 * size_t(packet.num_near_relays) );
            memcpy( entry->update_near_relay_addresses, packet.near_relay_addresses, sizeof(next_address_t) * size_t(packet.num_near_relays) );
            memcpy( entry->update_near_relay_ping_tokens, packet.near_relay_ping_tokens, packet.num_near_relays * NEXT_PING_TOKEN_BYTES );
            entry->update_near_relay_expire_timestamp = packet.near_relay_expire_timestamp;
        }

        entry->update_last_send_time = -1000.0;

        entry->session_data_bytes = packet.session_data_bytes;
        memcpy( entry->session_data, packet.session_data, packet.session_data_bytes );
        memcpy( entry->session_data_signature, packet.session_data_signature, NEXT_CRYPTO_SIGN_BYTES );

        entry->waiting_for_update_response = false;

        if ( packet.response_type == NEXT_UPDATE_TYPE_DIRECT )
        {
            bool session_transitions_to_direct = false;
            {
                next_platform_mutex_guard( &server->session_mutex );
                if ( entry->mutex_send_over_network_next )
                {
                    entry->mutex_send_over_network_next = false;
                    session_transitions_to_direct = true;
                }
            }

            if ( session_transitions_to_direct )
            {
                entry->has_previous_route = entry->has_current_route;
                entry->has_current_route = false;
                entry->previous_route_send_address = entry->current_route_send_address;
                memcpy( entry->previous_route_private_key, entry->current_route_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
            }
        }

        entry->has_debug = packet.has_debug;
        memcpy( entry->debug, packet.debug, NEXT_MAX_SESSION_DEBUG );

        if ( entry->previous_session_events != 0 )
        {   
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server flushed session events %x to backend for session %" PRIx64 " at address %s", entry->previous_session_events, entry->session_id, next_address_to_string( from, address_buffer ));
            entry->previous_session_events = 0;
        }

        if ( entry->session_update_flush && entry->session_update_request_packet.client_ping_timed_out && packet.slice_number == entry->session_flush_update_sequence - 1 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server flushed session update for session %" PRIx64 " to backend", entry->session_id );
            entry->session_update_flush_finished = true;
            server->num_flushed_session_updates++;
        }

        return;
    }

    // match data response

    if ( packet_id == NEXT_BACKEND_MATCH_DATA_RESPONSE_PACKET)
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing match data response packet" );

        if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored session response packet from backend. server is not initialized" );
            return;
        }

        NextBackendMatchDataResponsePacket packet;
        memset( &packet, 0, sizeof(packet) );

        if ( next_read_backend_packet( packet_id, packet_data, begin, end, &packet, next_signed_packets, next_server_backend_public_key ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored match data response packet from backend. packet failed to read" );
            return;
        }

        next_session_entry_t * entry = next_session_manager_find_by_session_id( server->session_manager, packet.session_id );
        if ( !entry )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored match data response packet from backend. could not find session %" PRIx64, packet.session_id );
            return;
        }

        if ( !entry->waiting_for_match_data_response )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored match data response packet from backend. not waiting for match data response" );
            return;
        }

        entry->match_data_response_received = true;
        entry->waiting_for_match_data_response = false;

        if ( entry->match_data_flush )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server flushed match data for session %" PRIx64 " to backend", entry->session_id );
            entry->match_data_flush_finished = true;
            server->num_flushed_match_data++;
        }
        else
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server successfully recorded match data for session %" PRIx64 " with backend", packet.session_id );
        }

        return;
    }

    // upgrade response packet

    if ( packet_id == NEXT_UPGRADE_RESPONSE_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing upgrade response packet" );

        NextUpgradeResponsePacket packet;

        if ( next_read_packet( NEXT_UPGRADE_RESPONSE_PACKET, packet_data, begin, end, &packet, next_signed_packets, NULL, NULL, NULL, NULL, NULL ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response packet. did not read" );
            return;
        }

        NextUpgradeToken upgrade_token;

        // does the session already exist? if so we still need to reply with upgrade commit in case of server -> client packet loss

        bool upgraded = false;

        next_session_entry_t * existing_entry = next_session_manager_find_by_address( server->session_manager, from );

        if ( existing_entry )
        {
            if ( !upgrade_token.Read( packet.upgrade_token, existing_entry->ephemeral_private_key ) )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. could not decrypt upgrade token (existing entry)", next_address_to_string( from, address_buffer ) );
                return;
            }

            if ( upgrade_token.session_id != existing_entry->session_id )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. session id does not match existing entry", next_address_to_string( from, address_buffer ) );
                return;
            }

            if ( !next_address_equal( &upgrade_token.client_address, &existing_entry->address ) )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. client address does not match existing entry", next_address_to_string( from, address_buffer ) );
                return;
            }
        }
        else
        {
            // session does not exist yet. look up pending upgrade entry...

            next_pending_session_entry_t * pending_entry = next_pending_session_manager_find( server->pending_session_manager, from );
            if ( pending_entry == NULL )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. does not match any pending upgrade", next_address_to_string( from, address_buffer ) );
                return;
            }

            if ( !upgrade_token.Read( packet.upgrade_token, pending_entry->private_key ) )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. could not decrypt upgrade token", next_address_to_string( from, address_buffer ) );
                return;
            }

            if ( upgrade_token.session_id != pending_entry->session_id )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. session id does not match pending upgrade entry", next_address_to_string( from, address_buffer ) );
                return;
            }

            if ( !next_address_equal( &upgrade_token.client_address, &pending_entry->address ) )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response from %s. client address does not match pending upgrade entry", next_address_to_string( from, address_buffer ) );
                return;
            }

            uint8_t server_send_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
            uint8_t server_receive_key[NEXT_CRYPTO_KX_SESSIONKEYBYTES];
            if ( next_crypto_kx_server_session_keys( server_receive_key, server_send_key, server->server_kx_public_key, server->server_kx_private_key, packet.client_kx_public_key ) != 0 )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server could not generate session keys from client public key" );
                return;
            }

            // remove from pending upgrade

            next_pending_session_manager_remove_by_address( server->pending_session_manager, from );

            // add to established sessions

            next_session_entry_t * entry = NULL;
            {
                next_platform_mutex_guard( &server->session_mutex );
                // todo: probably need to bring across internal events
                entry = next_session_manager_add( server->session_manager, &pending_entry->address, pending_entry->session_id, pending_entry->private_key, pending_entry->upgrade_token );
            }
            if ( entry == NULL )
            {
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_ERROR, "server ignored upgrade response from %s. failed to add session", next_address_to_string( from, address_buffer ) );
                return;
            }

            memcpy( entry->send_key, server_send_key, NEXT_CRYPTO_KX_SESSIONKEYBYTES );
            memcpy( entry->receive_key, server_receive_key, NEXT_CRYPTO_KX_SESSIONKEYBYTES );
            memcpy( entry->client_route_public_key, packet.client_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );
            entry->last_client_stats_update = next_platform_time();
            entry->user_hash = pending_entry->user_hash;
            entry->client_open_session_sequence = packet.client_open_session_sequence;
            entry->stats_platform_id = packet.platform_id;
            entry->stats_connection_type = packet.connection_type;
            entry->last_upgraded_packet_receive_time = next_platform_time();

            // notify session upgraded

            next_server_notify_session_upgraded_t * notify = (next_server_notify_session_upgraded_t*) next_malloc( server->context, sizeof( next_server_notify_session_upgraded_t ) );
            notify->type = NEXT_SERVER_NOTIFY_SESSION_UPGRADED;
            notify->address = entry->address;
            notify->session_id = entry->session_id;
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_SESSION_UPGRADED at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
                next_platform_mutex_guard( &server->notify_mutex );
                next_queue_push( server->notify_queue, notify );
            }

            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server received upgrade response packet from client %s", next_address_to_string( from, address_buffer ) );

            upgraded = true;
        }

        if ( !next_address_equal( &upgrade_token.client_address, from ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response. client address does not match from address" );
            return;
        }

        if ( upgrade_token.expire_timestamp < uint64_t(next_platform_time()) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response. upgrade token expired" );
            return;
        }

        if ( !next_address_equal( &upgrade_token.client_address, from ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response. client address does not match from address" );
            return;
        }

        if ( !next_address_equal( &upgrade_token.server_address, &server->server_address ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored upgrade response. server address does not match" );
            return;
        }

        next_post_validate_packet( NEXT_UPGRADE_RESPONSE_PACKET, NULL, NULL, NULL );
        
        if ( !upgraded )
        {
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server received upgrade response packet from %s", next_address_to_string( from, address_buffer ) );
        }

        // reply with upgrade confirm

        NextUpgradeConfirmPacket response;
        response.upgrade_sequence = server->upgrade_sequence++;
        response.session_id = upgrade_token.session_id;
        response.server_address = server->server_address;
        memcpy( response.client_kx_public_key, packet.client_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );
        memcpy( response.server_kx_public_key, server->server_kx_public_key, NEXT_CRYPTO_KX_PUBLICKEYBYTES );

        if ( next_server_internal_send_packet( server, from, NEXT_UPGRADE_CONFIRM_PACKET, &response ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "server could not send upgrade confirm packet" );
            return;
        }

        char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent upgrade confirm packet to client %s", next_address_to_string( from, address_buffer ) );

        return;
    }

    // -------------------
    // PACKETS FROM RELAYS
    // -------------------

    if ( packet_id == NEXT_ROUTE_REQUEST_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing route request packet" );

        const int packet_bytes = end - begin;

        if ( packet_bytes != NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored route request packet. wrong size" );
            return;
        }

        uint8_t * buffer = packet_data + begin;
        next_route_token_t route_token;
        if ( next_read_encrypted_route_token( &buffer, &route_token, next_router_public_key, server->server_route_private_key ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored route request packet. bad route" );
            return;
        }

        next_session_entry_t * entry = next_session_manager_find_by_session_id( server->session_manager, route_token.session_id );
        if ( !entry )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored route request packet. could not find session %" PRIx64, route_token.session_id );
            return;
        }

        if ( entry->has_current_route && route_token.expire_timestamp < entry->current_route_expire_timestamp )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored route request packet. expire timestamp is older than current route" );
            return;
        }

        if ( entry->has_current_route && next_sequence_greater_than( entry->most_recent_session_version, route_token.session_version ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored route request packet. route is older than most recent session (%d vs. %d)", route_token.session_version, entry->most_recent_session_version );
            return;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server received route request packet from relay for session %" PRIx64, route_token.session_id );

        if ( next_sequence_greater_than( route_token.session_version, entry->pending_route_session_version ) )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server added pending route for session %" PRIx64, route_token.session_id );
            entry->has_pending_route = true;
            entry->pending_route_session_version = route_token.session_version;
            entry->pending_route_expire_timestamp = route_token.expire_timestamp;
            entry->pending_route_expire_time = entry->has_current_route ? ( entry->current_route_expire_time + NEXT_SLICE_SECONDS * 2 ) : ( next_platform_time() + NEXT_SLICE_SECONDS * 2 );
            entry->pending_route_kbps_up = route_token.kbps_up;
            entry->pending_route_kbps_down = route_token.kbps_down;
            entry->pending_route_send_address = *from;
            memcpy( entry->pending_route_private_key, route_token.private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
            entry->most_recent_session_version = route_token.session_version;
        }

        uint64_t session_send_sequence = entry->special_send_sequence++;

        uint8_t from_address_data[4];
        uint8_t to_address_data[4];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( from, to_address_data, &to_address_bytes, &to_address_port );

        uint8_t response_data[NEXT_MAX_PACKET_BYTES];

        int response_bytes = next_write_route_response_packet( response_data, session_send_sequence, entry->session_id, entry->pending_route_session_version, entry->pending_route_private_key, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

        next_assert( response_bytes > 0 );

        next_assert( next_basic_packet_filter( response_data, response_bytes ) );
        next_assert( next_advanced_packet_filter( response_data, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, response_bytes ) );

        next_server_internal_send_packet_to_address( server, from, response_data, response_bytes );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent route response packet to relay for session %" PRIx64, entry->session_id );

        return;
    }

    // continue request packet

    if ( packet_id == NEXT_CONTINUE_REQUEST_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing continue request packet" );

        const int packet_bytes = end - begin;

        if ( packet_bytes != NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored continue request packet. wrong size" );
            return;
        }

        uint8_t * buffer = packet_data + begin;
        next_continue_token_t continue_token;
        if ( next_read_encrypted_continue_token( &buffer, &continue_token, next_router_public_key, server->server_route_private_key ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored continue request packet from relay. bad token" );
            return;
        }

        next_session_entry_t * entry = next_session_manager_find_by_session_id( server->session_manager, continue_token.session_id );
        if ( !entry )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored continue request packet from relay. could not find session %" PRIx64, continue_token.session_id );
            return;
        }

        if ( !entry->has_current_route )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored continue request packet from relay. session has no route to continue" );
            return;
        }

        if ( continue_token.session_version != entry->current_route_session_version )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored continue request packet from relay. session version does not match" );
            return;
        }

        if ( continue_token.expire_timestamp < entry->current_route_expire_timestamp )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored continue request packet from relay. expire timestamp is older than current route" );
            return;
        }

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server received continue request packet from relay for session %" PRIx64, continue_token.session_id );

        entry->current_route_expire_timestamp = continue_token.expire_timestamp;
        entry->current_route_expire_time += NEXT_SLICE_SECONDS;
        entry->has_previous_route = false;

        uint64_t session_send_sequence = entry->special_send_sequence++;

        uint8_t from_address_data[4];
        uint8_t to_address_data[4];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( from, to_address_data, &to_address_bytes, &to_address_port );

        uint8_t response_data[NEXT_MAX_PACKET_BYTES];

        int response_bytes = next_write_continue_response_packet( response_data, session_send_sequence, entry->session_id, entry->current_route_session_version, entry->current_route_private_key, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

        next_assert( response_bytes > 0 );

        next_assert( next_basic_packet_filter( response_data, response_bytes ) );
        next_assert( next_advanced_packet_filter( response_data, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, response_bytes ) );

        next_server_internal_send_packet_to_address( server, from, response_data, response_bytes );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent continue response packet to relay for session %" PRIx64, entry->session_id );

        return;
    }

    // client to server packet

    if ( packet_id == NEXT_CLIENT_TO_SERVER_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing client to server packet" );

        const int packet_bytes = end - begin;

        if ( packet_bytes <= NEXT_HEADER_BYTES )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client to server packet. packet too small to be valid" );
            return;
        }

        next_session_entry_t * entry = next_server_internal_process_client_to_server_packet( server, packet_id, packet_data + begin, packet_bytes );
        if ( !entry )
        {
            // IMPORTANT: There is no need to log this case, because next_server_internal_process_client_to_server_packet already
            // logs all cases where it returns NULL to the debug log. Logging here duplicates the log and incorrectly prints
            // out an error when the packet has already been received on the direct path, when multipath is enabled.
            return;
        }

        next_server_notify_packet_received_t * notify = (next_server_notify_packet_received_t*) next_malloc( server->context, sizeof( next_server_notify_packet_received_t ) );
        notify->type = NEXT_SERVER_NOTIFY_PACKET_RECEIVED;
        notify->from = entry->address;
        notify->packet_bytes = packet_bytes - NEXT_HEADER_BYTES;
        next_assert( notify->packet_bytes > 0 );
        next_assert( notify->packet_bytes <= NEXT_MTU );
        memcpy( notify->packet_data, packet_data + begin + NEXT_HEADER_BYTES, size_t(notify->packet_bytes) );
        {
#if NEXT_SPIKE_TRACKING
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_PACKET_RECEIVED at %s:%d - from = %s, packet_bytes = %d", __FILE__, __LINE__, next_address_to_string( &notify->from, address_buffer ), notify->packet_bytes );
#endif // #if NEXT_SPIKE_TRACKING                
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify );
        }

        return;
    }

    // ping packet

    if ( packet_id == NEXT_PING_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing next ping packet" );

        const int packet_bytes = end - begin;

        if ( packet_bytes != NEXT_HEADER_BYTES + 8 )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored next ping packet. bad packet size" );
            return;
        }

        next_session_entry_t * entry = next_server_internal_process_client_to_server_packet( server, packet_id, packet_data + begin, packet_bytes );
        if ( !entry )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored next ping packet. did not verify" );
            return;
        }

        const uint8_t * p = packet_data + begin + NEXT_HEADER_BYTES;

        uint64_t ping_sequence = next_read_uint64( &p );

        entry->last_client_next_ping = next_platform_time();

        uint64_t send_sequence = entry->special_send_sequence++;

        uint8_t from_address_data[4];
        uint8_t to_address_data[4];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( from, to_address_data, &to_address_bytes, &to_address_port );

        uint8_t pong_packet_data[NEXT_MAX_PACKET_BYTES];

        int pong_packet_bytes = next_write_pong_packet( pong_packet_data, send_sequence, entry->session_id, entry->current_route_session_version, entry->current_route_private_key, ping_sequence, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

        next_assert( pong_packet_bytes > 0 );

        next_assert( next_basic_packet_filter( pong_packet_data, pong_packet_bytes ) );
        next_assert( next_advanced_packet_filter( pong_packet_data, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, pong_packet_bytes ) );

        next_server_internal_send_packet_to_address( server, from, pong_packet_data, pong_packet_bytes );

        return;
    }

    // ----------------------------------
    // ENCRYPTED CLIENT TO SERVER PACKETS
    // ----------------------------------

    next_session_entry_t * session = NULL;

    if ( next_encrypted_packets[packet_id] )
    {
        session = next_session_manager_find_by_address( server->session_manager, from );
        if ( !session )
        {
            next_printf( NEXT_LOG_LEVEL_SPAM, "server dropped encrypted packet because it couldn't find any session for it" );
            return;
        }

        session->last_upgraded_packet_receive_time = next_platform_time();
    }

    // direct ping packet

    if ( packet_id == NEXT_DIRECT_PING_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing direct ping packet" );

        next_assert( session );

        if ( session == NULL )
            return;

        uint64_t packet_sequence = 0;

        NextDirectPingPacket packet;
        if ( next_read_packet( NEXT_DIRECT_PING_PACKET, packet_data, begin, end, &packet, next_signed_packets, next_encrypted_packets, &packet_sequence, NULL, session->receive_key, &session->internal_replay_protection ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored direct ping packet. could not read" );
            return;
        }

        session->last_client_direct_ping = next_platform_time();

        next_post_validate_packet( NEXT_DIRECT_PING_PACKET, next_encrypted_packets, &packet_sequence, &session->internal_replay_protection );

        NextDirectPongPacket response;
        response.ping_sequence = packet.ping_sequence;

        if ( next_server_internal_send_packet( server, from, NEXT_DIRECT_PONG_PACKET, &response ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server could not send upgrade confirm packet" );
            return;
        }

        return;
    }

    // client stats packet

    if ( packet_id == NEXT_CLIENT_STATS_PACKET )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing client stats packet" );

        next_assert( session );

        if ( session == NULL )
            return;

        NextClientStatsPacket packet;

        uint64_t packet_sequence = 0;

        if ( next_read_packet( NEXT_CLIENT_STATS_PACKET, packet_data, begin, end, &packet, next_signed_packets, next_encrypted_packets, &packet_sequence, NULL, session->receive_key, &session->internal_replay_protection ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client stats packet. could not read" );
            return;
        }

        next_post_validate_packet( NEXT_CLIENT_STATS_PACKET, next_encrypted_packets, &packet_sequence, &session->internal_replay_protection );

        if ( packet_sequence > session->stats_sequence )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server received client stats packet for session %" PRIx64, session->session_id );

            if ( !session->stats_fallback_to_direct && packet.fallback_to_direct )
            {
                next_printf( NEXT_LOG_LEVEL_INFO, "server session fell back to direct %" PRIx64, session->session_id );
            }

            session->stats_sequence = packet_sequence;

            session->stats_reported = packet.reported;
            session->stats_multipath = packet.multipath;
            session->stats_fallback_to_direct = packet.fallback_to_direct;
            if ( packet.next_bandwidth_over_limit )
            {
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server session sees client over next bandwidth limit %" PRIx64, session->session_id );
                session->stats_client_bandwidth_over_limit = true;
            }

            session->stats_platform_id = packet.platform_id;
            session->stats_connection_type = packet.connection_type;
            session->stats_direct_kbps_up = packet.direct_kbps_up;
            session->stats_direct_kbps_down = packet.direct_kbps_down;
            session->stats_next_kbps_up = packet.next_kbps_up;
            session->stats_next_kbps_down = packet.next_kbps_down;
            session->stats_direct_rtt = packet.direct_rtt;
            session->stats_direct_jitter = packet.direct_jitter;
            session->stats_direct_packet_loss = packet.direct_packet_loss;
            session->stats_direct_max_packet_loss_seen = packet.direct_max_packet_loss_seen;
            session->stats_next = packet.next;
            session->stats_next_rtt = packet.next_rtt;
            session->stats_next_jitter = packet.next_jitter;
            session->stats_next_packet_loss = packet.next_packet_loss;
            session->stats_has_near_relay_pings = packet.num_near_relays > 0;
            session->stats_num_near_relays = packet.num_near_relays;
            for ( int i = 0; i < packet.num_near_relays; ++i )
            {
                session->stats_near_relay_ids[i] = packet.near_relay_ids[i];
                session->stats_near_relay_rtt[i] = packet.near_relay_rtt[i];
                session->stats_near_relay_jitter[i] = packet.near_relay_jitter[i];
                session->stats_near_relay_packet_loss[i] = packet.near_relay_packet_loss[i];
            }
            session->stats_packets_sent_client_to_server = packet.packets_sent_client_to_server;
            session->stats_packets_lost_server_to_client = packet.packets_lost_server_to_client;
            session->stats_jitter_server_to_client = packet.jitter_server_to_client;
            session->last_client_stats_update = next_platform_time();
        }

        return;
    }

    // route update ack packet

    if ( packet_id == NEXT_ROUTE_UPDATE_ACK_PACKET && session != NULL )
    {
        next_printf( NEXT_LOG_LEVEL_SPAM, "server processing route update ack packet" );

        NextRouteUpdateAckPacket packet;

        uint64_t packet_sequence = 0;

        if ( next_read_packet( NEXT_ROUTE_UPDATE_ACK_PACKET, packet_data, begin, end, &packet, next_signed_packets, next_encrypted_packets, &packet_sequence, NULL, session->receive_key, &session->internal_replay_protection ) != packet_id )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored client stats packet. could not read" );
            return;
        }

        if ( packet.sequence != session->update_sequence )
        {
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server ignored route update ack packet. wrong update sequence number" );
            return;
        }

        next_post_validate_packet( NEXT_ROUTE_UPDATE_ACK_PACKET, next_encrypted_packets, &packet_sequence, &session->internal_replay_protection );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server received route update ack from client for session %" PRIx64, session->session_id );

        if ( session->update_dirty )
        {
            session->update_dirty = false;
        }

        return;
    }
}

void next_server_internal_process_passthrough_packet( next_server_internal_t * server, const next_address_t * from, uint8_t * packet_data, int packet_bytes )
{
    next_assert( server );
    next_assert( from );
    next_assert( packet_data );

    next_printf( NEXT_LOG_LEVEL_SPAM, "server processing passthrough packet" );

    next_server_internal_verify_sentinels( server );

    if ( packet_bytes > 0 && packet_bytes <= NEXT_MAX_PACKET_BYTES - 1 )
    {
        if ( server->payload_receive_callback )
        {
            void * callback_data = server->payload_receive_callback_data;
            if ( server->payload_receive_callback( callback_data, from, packet_data, packet_bytes ) )
                return;
        }

        next_server_notify_packet_received_t * notify = (next_server_notify_packet_received_t*) next_malloc( server->context, sizeof( next_server_notify_packet_received_t ) );
        notify->type = NEXT_SERVER_NOTIFY_PACKET_RECEIVED;
        notify->from = *from;
        notify->packet_bytes = packet_bytes;
        next_assert( packet_bytes > 0 );
        next_assert( packet_bytes <= NEXT_MAX_PACKET_BYTES - 1 );
        memcpy( notify->packet_data, packet_data, size_t(packet_bytes) );
        {
#if NEXT_SPIKE_TRACKING
            char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_PACKET_RECEIVED at %s:%d - from = %s, packet_bytes = %d", __FILE__, __LINE__, next_address_to_string( &notify->from, address_buffer ), notify->packet_bytes );
#endif // #if NEXT_SPIKE_TRACKING                
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify );
        }
    }
}

void next_server_internal_block_and_receive_packet( next_server_internal_t * server )
{
    next_server_internal_verify_sentinels( server );

    uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

    next_assert( ( size_t(packet_data) % 4 ) == 0 );

    next_address_t from;

#if NEXT_SPIKE_TRACKING
    next_printf( NEXT_LOG_LEVEL_SPAM, "server calls next_platform_socket_receive_packet on internal thread" );
#endif // #if NEXT_SPIKE_TRACKING

    const int packet_bytes = next_platform_socket_receive_packet( server->socket, &from, packet_data, NEXT_MAX_PACKET_BYTES );

#if NEXT_SPIKE_TRACKING
    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_SPAM, "server next_platform_socket_receive_packet returns with a %d byte packet from %s", next_address_to_string( &from, address_buffer ) );
#endif // #if NEXT_SPIKE_TRACKING

    if ( packet_bytes == 0 )
        return;

    next_assert( packet_bytes > 0 );

    int begin = 0;
    int end = packet_bytes;

    if ( server->packet_receive_callback )
    {
        void * callback_data = server->packet_receive_callback_data;

        server->packet_receive_callback( callback_data, &from, packet_data, &begin, &end );

        next_assert( begin >= 0 );
        next_assert( end <= NEXT_MAX_PACKET_BYTES );

        if ( end - begin <= 0 )
            return;        
    }

#if NEXT_DEVELOPMENT
    if ( next_packet_loss && ( rand() % 10 ) == 0 )
         return;
#endif // #if NEXT_DEVELOPMENT

    const uint8_t packet_type = packet_data[begin];

    if ( packet_type != NEXT_PASSTHROUGH_PACKET )
    {
        next_server_internal_process_network_next_packet( server, &from, packet_data, begin, end );
    }
    else
    {
        begin += 1;
        next_server_internal_process_passthrough_packet( server, &from, packet_data + begin, end - begin );
    }
}

void next_server_internal_upgrade_session( next_server_internal_t * server, const next_address_t * address, uint64_t session_id, uint64_t user_hash )
{
    next_assert( server );
    next_assert( address );

    next_server_internal_verify_sentinels( server );

    if ( next_global_config.disable_network_next )
        return;

    if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
        return;

    next_assert( server->state == NEXT_SERVER_STATE_INITIALIZED || server->state == NEXT_SERVER_STATE_DIRECT_ONLY );

    if ( server->state == NEXT_SERVER_STATE_DIRECT_ONLY )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server cannot upgrade client. direct only mode" );
        return;
    }

    char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];

    next_printf( NEXT_LOG_LEVEL_DEBUG, "server upgrading client %s to session %" PRIx64, next_address_to_string( address, buffer ), session_id );

    NextUpgradeToken upgrade_token;

    upgrade_token.session_id = session_id;
    upgrade_token.expire_timestamp = uint64_t( next_platform_time() ) + 10;
    upgrade_token.client_address = *address;
    upgrade_token.server_address = server->server_address;

    unsigned char session_private_key[NEXT_CRYPTO_SECRETBOX_KEYBYTES];
    next_crypto_secretbox_keygen( session_private_key );

    uint8_t upgrade_token_data[NEXT_UPGRADE_TOKEN_BYTES];

    upgrade_token.Write( upgrade_token_data, session_private_key );

    next_pending_session_manager_remove_by_address( server->pending_session_manager, address );

    next_session_manager_remove_by_address( server->session_manager, address );

    next_pending_session_entry_t * entry = next_pending_session_manager_add( server->pending_session_manager, address, upgrade_token.session_id, session_private_key, upgrade_token_data, next_platform_time() );

    if ( entry == NULL )
    {
        next_assert( !"could not add pending session entry. this should never happen!" );
        return;
    }

    entry->user_hash = user_hash;
}

void next_server_internal_session_events( next_server_internal_t * server, const next_address_t * address, uint64_t session_events )
{
    next_assert( server );
    next_assert( address );

    next_server_internal_verify_sentinels( server );

    if ( next_global_config.disable_network_next )
        return;

    if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
        return;

    next_session_entry_t * entry = next_session_manager_find_by_address( server->session_manager, address );
    if ( !entry )
    {
        char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_printf( NEXT_LOG_LEVEL_DEBUG, "could not find session at address %s. not adding session event %x", next_address_to_string( address, buffer ), session_events );
        return;
    }

    entry->current_session_events |= session_events;
    char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_DEBUG, "server set session event %x for session %" PRIx64 " at address %s", session_events, entry->session_id, next_address_to_string( address, buffer ) );
}

void next_server_internal_match_data( next_server_internal_t * server, const next_address_t * address, uint64_t match_id, const double * match_values, int num_match_values )
{
    next_assert( server );
    next_assert( address );

    next_server_internal_verify_sentinels( server );

    if ( next_global_config.disable_network_next )
        return;

    if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
        return;

    next_session_entry_t * entry = next_session_manager_find_by_address( server->session_manager, address );    
    if ( !entry )
    {
        char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_printf( NEXT_LOG_LEVEL_DEBUG, "could not find session at address %s. server not sending match data", next_address_to_string( address, buffer ) );
        return;
    }

    if ( entry->has_match_data || entry->waiting_for_match_data_response || entry->match_data_response_received )
    {
        char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
        next_printf( NEXT_LOG_LEVEL_WARN, "server already sent match data for session %" PRIx64 " at address %s", entry->session_id, next_address_to_string( address, buffer ) );
        return;
    }

    entry->match_id = match_id;
    entry->num_match_values = num_match_values;
    for ( int i = 0; i < num_match_values; ++i )
    {
        entry->match_values[i] = match_values[i];
    }
    entry->has_match_data = true;

    char buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_printf( NEXT_LOG_LEVEL_DEBUG, "server adds match data for session %" PRIx64 " at address %s", entry->session_id, next_address_to_string( address, buffer ) );
}

void next_server_internal_flush_session_update( next_server_internal_t * server )
{
    next_assert( server );
    next_assert( server->session_manager );

    if ( next_global_config.disable_network_next )
        return;

    const int max_entry_index = server->session_manager->max_entry_index;

    for ( int i = 0; i <= max_entry_index; ++i )
    {
        if ( server->session_manager->session_ids[i] == 0 )
            continue;

        next_session_entry_t * session = &server->session_manager->entries[i];

        session->client_ping_timed_out = true;
        session->session_update_request_packet.client_ping_timed_out = true;

        // IMPORTANT: Make sure to only accept a backend session response for the next session update
        // sent out, not the current session update (if any is in flight). This way flush succeeds
        // even if it called in the middle of a session update in progress.
        session->session_flush_update_sequence = session->update_sequence + 1;
        session->session_update_flush = true;
        server->num_session_updates_to_flush++;
    }
}

void next_server_internal_flush_match_data( next_server_internal_t * server )
{
    next_assert( server );
    next_assert( server->session_manager );

    if ( next_global_config.disable_network_next )
        return;

    const int max_entry_index = server->session_manager->max_entry_index;

    for ( int i = 0; i <= max_entry_index; ++i )
    {
        if ( server->session_manager->session_ids[i] == 0 )
            continue;

        next_session_entry_t * session = &server->session_manager->entries[i];

        if ( ( !session->has_match_data ) || ( session->has_match_data && session->match_data_response_received ) )
            continue;

        session->match_data_flush = true;
        server->num_match_data_to_flush++;
    }
}

void next_server_internal_flush( next_server_internal_t * server )
{
    next_assert( server );
    next_assert( server->session_manager );

    next_server_internal_verify_sentinels( server );

    if ( next_global_config.disable_network_next )
    {
        server->flushing = true;
        server->flushed = true;
        return;
    }

    if ( server->flushing )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "server ignored flush. already flushed" );
        return;
    }

    server->flushing = true;

    next_server_internal_flush_session_update( server );

    next_server_internal_flush_match_data( server );

    next_printf( NEXT_LOG_LEVEL_DEBUG, "server flush started. %d session updates and %d match data to flush", server->num_session_updates_to_flush, server->num_match_data_to_flush );
}

void next_server_internal_pump_commands( next_server_internal_t * server )
{
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "next_server_internal_pump_commands" );
#endif // #if NEXT_SPIKE_TRACKING

    while ( true )
    {
        next_server_internal_verify_sentinels( server );

        void * entry = NULL;
        {
            next_platform_mutex_guard( &server->command_mutex );
            entry = next_queue_pop( server->command_queue );
        }

        if ( entry == NULL )
            break;

        next_server_command_t * command = (next_server_command_t*) entry;

        switch ( command->type )
        {
            case NEXT_SERVER_COMMAND_UPGRADE_SESSION:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_UPGRADE_SESSION" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_command_upgrade_session_t * upgrade_session = (next_server_command_upgrade_session_t*) command;
                next_server_internal_upgrade_session( server, &upgrade_session->address, upgrade_session->session_id, upgrade_session->user_hash );
            }
            break;

            case NEXT_SERVER_COMMAND_SESSION_EVENT:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_SESSION_EVENT" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_command_session_event_t * event = (next_server_command_session_event_t*) command;
                next_server_internal_session_events( server, &event->address, event->session_events );
            }
            break;

            case NEXT_SERVER_COMMAND_MATCH_DATA:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_MATCH_DATA" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_command_match_data_t * match_data = (next_server_command_match_data_t*) command;
                next_server_internal_match_data( server, &match_data->address, match_data->match_id, match_data->match_values, match_data->num_match_values );
            }
            break;

            case NEXT_SERVER_COMMAND_FLUSH:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_FLUSH" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_internal_flush( server );
            }
            break;

            case NEXT_SERVER_COMMAND_SET_PACKET_RECEIVE_CALLBACK:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_SET_PACKET_RECEIVE_CALLBACK" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_command_set_packet_receive_callback_t * cmd = (next_server_command_set_packet_receive_callback_t*) command;
                server->packet_receive_callback = cmd->callback;
                server->packet_receive_callback_data = cmd->callback_data;
            }
            break;

            case NEXT_SERVER_COMMAND_SET_SEND_PACKET_TO_ADDRESS_CALLBACK:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_SET_SEND_PACKET_TO_ADDRESS_CALLBACK" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_command_set_send_packet_to_address_callback_t * cmd = (next_server_command_set_send_packet_to_address_callback_t*) command;
                server->send_packet_to_address_callback = cmd->callback;
                server->send_packet_to_address_callback_data = cmd->callback_data;
            }
            break;

            case NEXT_SERVER_COMMAND_SET_PAYLOAD_RECEIVE_CALLBACK:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread receives NEXT_SERVER_COMMAND_SET_PAYLOAD_RECEIVE_CALLBACK" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_command_set_payload_receive_callback_t * cmd = (next_server_command_set_payload_receive_callback_t*) command;
                server->payload_receive_callback = cmd->callback;
                server->payload_receive_callback_data = cmd->callback_data;
            }
            break;

            default: break;
        }

        next_free( server->context, command );
    }
}

static void next_server_internal_resolve_hostname_thread_function( void * context )
{
    next_assert( context );

    double start_time = next_platform_time();

    next_server_internal_t * server = (next_server_internal_t*) context;

    const char * hostname = next_global_config.server_backend_hostname;
    const char * port = NEXT_SERVER_BACKEND_PORT;
    const char * override_port = next_platform_getenv( "NEXT_SERVER_BACKEND_PORT" );
    if ( !override_port )
    {
        override_port = next_platform_getenv( "NEXT_PORT" );
    }
    if ( override_port )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "override server backend port: '%s'", override_port );
        port = override_port;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "server resolving backend hostname '%s'", hostname );

    next_address_t address;

    bool success = false;

    // first try to parse the hostname directly as an address, this is a common case in testbeds and there's no reason to actually run a DNS resolve on it

    if ( next_address_parse( &address, hostname ) == NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "server backend hostname is an address" );
        next_assert( address.type == NEXT_ADDRESS_IPV4 || address.type == NEXT_ADDRESS_IPV6 );
        address.port = uint16_t( atoi(port) );
        success = true;
    }    
    else
    {
        // try to resolve the hostname, retry a few times if it doesn't succeed right away

        for ( int i = 0; i < 10; ++i )
        {
            if ( next_platform_hostname_resolve( hostname, port, &address ) == NEXT_OK )
            {
                next_assert( address.type == NEXT_ADDRESS_IPV4 || address.type == NEXT_ADDRESS_IPV6 );
                success = true;
                break;
            }
            else
            {
                next_printf( NEXT_LOG_LEVEL_WARN, "server failed to resolve hostname: '%s' (%d)", hostname, i );
                next_platform_sleep( 1.0 );
            }
        }
    }

    if ( !success )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to resolve backend hostname: %s", hostname );
        next_platform_mutex_guard( &server->resolve_hostname_mutex );
        server->resolve_hostname_finished = true;
        memset( &server->resolve_hostname_result, 0, sizeof(next_address_t) );
        return;
    }

#if NEXT_DEVELOPMENT
    if ( next_platform_getenv( "NEXT_FORCE_RESOLVE_HOSTNAME_TIMEOUT" ) )
    {
        next_platform_sleep( NEXT_SERVER_RESOLVE_HOSTNAME_TIMEOUT * 2 );
    }
#endif // #if NEXT_DEVELOPMENT

    if ( next_platform_time() - start_time > NEXT_SERVER_AUTODETECT_TIMEOUT )
    {
        // IMPORTANT: if we have timed out, don't grab the mutex or write results. 
        // our thread has been destroyed and if we are unlucky, the next_server_internal_t instance has as well.
        return;
    }

    next_platform_mutex_guard( &server->resolve_hostname_mutex );
    server->resolve_hostname_finished = true;
    server->resolve_hostname_result = address;
}

static bool next_server_internal_update_resolve_hostname( next_server_internal_t * server )
{
    next_assert( server );

    next_server_internal_verify_sentinels( server );

    next_assert( !next_global_config.disable_network_next );

    if ( !server->resolving_hostname )
        return true;

    bool finished = false;
    next_address_t result;
    memset( &result, 0, sizeof(next_address_t) );
    {
        next_platform_mutex_guard( &server->resolve_hostname_mutex );
        finished = server->resolve_hostname_finished;
        result = server->resolve_hostname_result;
    }

    if ( finished )
    {
        next_platform_thread_join( server->resolve_hostname_thread );
    }
    else
    {
        if ( next_platform_time() < server->resolve_hostname_start_time + NEXT_SERVER_RESOLVE_HOSTNAME_TIMEOUT )
        {
            // keep waiting
            return false;
        }
        else
        {
            // but don't wait forever...
            next_printf( NEXT_LOG_LEVEL_INFO, "resolve hostname timed out" );
        }
    }
    
    next_platform_thread_destroy( server->resolve_hostname_thread );
    
    server->resolve_hostname_thread = NULL;
    server->resolving_hostname = false;
    server->backend_address = result;

    char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];

    if ( result.type != NEXT_ADDRESS_NONE )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server resolved backend hostname to %s", next_address_to_string( &result, address_buffer ) );
    }
    else
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server failed to resolve backend hostname" );
        server->resolving_hostname = false;
    }

    return true;
}

static void next_server_internal_autodetect_thread_function( void * context )
{
    next_assert( context );

    double start_time = next_platform_time();

    next_server_internal_t * server = (next_server_internal_t*) context;

    bool autodetect_result = false;
    bool autodetect_actually_did_something = false;
    char autodetect_output[NEXT_MAX_DATACENTER_NAME_LENGTH];

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

    // autodetect datacenter is currently windows and linux only (mac is just for testing...)

    const char * autodetect_input = server->datacenter_name;
    
    char autodetect_address[NEXT_MAX_ADDRESS_STRING_LENGTH];
    next_address_t server_address_no_port = server->server_address;
    server_address_no_port.port = 0;
    next_address_to_string( &server_address_no_port, autodetect_address );

    if ( !next_global_config.disable_autodetect &&
         ( autodetect_input[0] == '\0' 
            ||
         ( autodetect_input[0] == 'c' &&
           autodetect_input[1] == 'l' &&
           autodetect_input[2] == 'o' &&
           autodetect_input[3] == 'u' &&
           autodetect_input[4] == 'd' &&
           autodetect_input[5] == '\0' ) 
            ||
         ( autodetect_input[0] == 'm' && 
           autodetect_input[1] == 'u' && 
           autodetect_input[2] == 'l' && 
           autodetect_input[3] == 't' && 
           autodetect_input[4] == 'i' && 
           autodetect_input[5] == 'p' && 
           autodetect_input[6] == 'l' && 
           autodetect_input[7] == 'a' && 
           autodetect_input[8] == 'y' && 
           autodetect_input[9] == '.' ) ) )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server attempting to autodetect datacenter" );

        autodetect_result = next_autodetect_datacenter( autodetect_input, autodetect_address, autodetect_output, sizeof(autodetect_output) );
        
        autodetect_actually_did_something = true;
    }

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#if NEXT_DEVELOPMENT
    if ( next_platform_getenv( "NEXT_FORCE_AUTODETECT_TIMEOUT" ) )
    {
        next_platform_sleep( NEXT_SERVER_AUTODETECT_TIMEOUT * 1.25 );
    }
#endif // #if NEXT_DEVELOPMENT

    if ( next_platform_time() - start_time > NEXT_SERVER_RESOLVE_HOSTNAME_TIMEOUT )
    {
        // IMPORTANT: if we have timed out, don't grab the mutex or write results. 
        // our thread has been destroyed and if we are unlucky, the next_server_internal_t instance is as well.
        return;
    }

    next_platform_mutex_guard( &server->autodetect_mutex );
    next_copy_string( server->autodetect_result, autodetect_output, NEXT_MAX_DATACENTER_NAME_LENGTH );
    server->autodetect_finished = true;
    server->autodetect_succeeded = autodetect_result;
    server->autodetect_actually_did_something = autodetect_actually_did_something;
}

static bool next_server_internal_update_autodetect( next_server_internal_t * server )
{
    next_assert( server );

    next_server_internal_verify_sentinels( server );

    next_assert( !next_global_config.disable_network_next );

    if ( server->resolving_hostname )    // IMPORTANT: wait until resolving hostname is finished, before autodetect complete!
        return true;

    if ( !server->autodetecting )
        return true;

    bool finished = false;
    {
        next_platform_mutex_guard( &server->autodetect_mutex );
        finished = server->autodetect_finished;
    }

    if ( finished )
    {
        next_platform_thread_join( server->autodetect_thread );
    }
    else
    {
        if ( next_platform_time() < server->autodetect_start_time + NEXT_SERVER_AUTODETECT_TIMEOUT )
        {
            // keep waiting
            return false;
        }
        else
        {
            // but don't wait forever...
            next_printf( NEXT_LOG_LEVEL_INFO, "autodetect timed out. sticking with '%s' [%" PRIx64 "]", server->datacenter_name, server->datacenter_id );
        }
    }
    
    next_platform_thread_destroy( server->autodetect_thread );
    
    server->autodetect_thread = NULL;
    server->autodetecting = false;

    if ( server->autodetect_actually_did_something )
    {
        if ( server->autodetect_succeeded )
        {
            memset( server->datacenter_name, 0, sizeof(server->datacenter_name) );
            next_copy_string( server->datacenter_name, server->autodetect_result, NEXT_MAX_DATACENTER_NAME_LENGTH );
            server->datacenter_id = next_datacenter_id( server->datacenter_name );
            next_printf( NEXT_LOG_LEVEL_INFO, "server autodetected datacenter '%s' [%" PRIx64 "]", server->datacenter_name, server->datacenter_id );
        }
        else
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "server autodetect datacenter failed. sticking with '%s' [%" PRIx64 "]", server->datacenter_name, server->datacenter_id );
        }
    }

    return true;
}

void next_server_internal_update_init( next_server_internal_t * server )
{
    next_server_internal_verify_sentinels( server );

    next_assert( server );

    next_assert( !next_global_config.disable_network_next );

    if ( server->state != NEXT_SERVER_STATE_INITIALIZING )
        return;

    // check for init timeout

    const double current_time = next_platform_time();

    if ( server->server_init_timeout_time <= current_time )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server init timed out. falling back to direct mode only :(" );

        server->state = NEXT_SERVER_STATE_DIRECT_ONLY;

        next_server_notify_ready_t * notify_ready = (next_server_notify_ready_t*) next_malloc( server->context, sizeof( next_server_notify_ready_t ) );
        notify_ready->type = NEXT_SERVER_NOTIFY_READY;
        next_copy_string( notify_ready->datacenter_name, server->datacenter_name, NEXT_MAX_DATACENTER_NAME_LENGTH );

        next_server_notify_direct_only_t * notify_direct_only = (next_server_notify_direct_only_t*) next_malloc( server->context, sizeof(next_server_notify_direct_only_t) );
        next_assert( notify_direct_only );
        notify_direct_only->type = NEXT_SERVER_NOTIFY_DIRECT_ONLY;

        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_DIRECT_ONLY and NEXT_SERVER_NOTIFY_READY at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING                
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify_direct_only );
            next_queue_push( server->notify_queue, notify_ready );
        }

        return;
    }

    // check for initializing -> initialized transition

    if ( server->resolve_hostname_finished && server->autodetect_finished && server->received_init_response )
    {
        next_assert( server->backend_address.type == NEXT_ADDRESS_IPV4 || server->backend_address.type == NEXT_ADDRESS_IPV6 );
        next_server_notify_ready_t * notify = (next_server_notify_ready_t*) next_malloc( server->context, sizeof( next_server_notify_ready_t ) );
        notify->type = NEXT_SERVER_NOTIFY_READY;
        next_copy_string( notify->datacenter_name, server->datacenter_name, NEXT_MAX_DATACENTER_NAME_LENGTH );
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_READY at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify );
        }
        server->state = NEXT_SERVER_STATE_INITIALIZED;
    }

    // wait until we have resolved the backend hostname

    if ( !server->resolve_hostname_finished )
        return;

    // wait until we have autodetected the datacenter

    if ( !server->autodetect_finished )
        return;

    // wait until the backend 

    // if we have started flushing, abort the init...

    if ( server->flushing )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server aborted init" );
        server->state = NEXT_SERVER_STATE_DIRECT_ONLY;
        next_server_notify_direct_only_t * notify_direct_only = (next_server_notify_direct_only_t*) next_malloc( server->context, sizeof(next_server_notify_direct_only_t) );
        next_assert( notify_direct_only );
        notify_direct_only->type = NEXT_SERVER_NOTIFY_DIRECT_ONLY;
        {
#if NEXT_SPIKE_TRACKING
            next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_DIRECT_ONLY at %s:%d", __FILE__, __FILE__ );
#endif // #if NEXT_SPIKE_TRACKING
            next_platform_mutex_guard( &server->notify_mutex );
            next_queue_push( server->notify_queue, notify_direct_only );
        }
        return;
    }

    // send init request packets repeatedly until we get a response or time out...

    if ( server->server_init_request_id != 0 && server->server_init_resend_time > current_time )
        return;

    while ( server->server_init_request_id == 0 )
    {
        server->server_init_request_id = next_random_uint64();
    }

    server->server_init_resend_time = current_time + 1.0;

    NextBackendServerInitRequestPacket packet;

    packet.request_id = server->server_init_request_id;
    packet.customer_id = server->customer_id;
    packet.datacenter_id = server->datacenter_id;
    next_copy_string( packet.datacenter_name, server->datacenter_name, NEXT_MAX_DATACENTER_NAME_LENGTH );
    packet.datacenter_name[NEXT_MAX_DATACENTER_NAME_LENGTH-1] = '\0';

    uint8_t magic[8];
    memset( magic, 0, sizeof(magic) );

    uint8_t from_address_data[32];
    uint8_t to_address_data[32];
    uint16_t from_address_port;
    uint16_t to_address_port;
    int from_address_bytes;
    int to_address_bytes;

    next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
    next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

    uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

    next_assert( ( size_t(packet_data) % 4 ) == 0 );

    int packet_bytes = 0;
    if ( next_write_backend_packet( NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET, &packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write server init request packet for backend" );
        return;
    }

    next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
    next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

    next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );

    next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent init request to backend" );
}

void next_server_internal_backend_update( next_server_internal_t * server )
{
    next_server_internal_verify_sentinels( server );

    next_assert( server );

    if ( next_global_config.disable_network_next )
        return;

    double current_time = next_platform_time();

    // don't do anything until we resolve the backend hostname

    if ( server->resolving_hostname )
        return;

    // tracker updates

    const int max_entry_index = server->session_manager->max_entry_index;

    for ( int i = 0; i <= max_entry_index; ++i )
    {
        if ( server->session_manager->session_ids[i] == 0 )
            continue;

        next_session_entry_t * session = &server->session_manager->entries[i];

        if ( session->stats_fallback_to_direct )
            continue;

        if ( session->next_tracker_update_time <= current_time )
        {
            const int packets_lost = next_packet_loss_tracker_update( &session->packet_loss_tracker );
            session->stats_packets_lost_client_to_server += packets_lost;
            session->stats_packets_out_of_order_client_to_server = session->out_of_order_tracker.num_out_of_order_packets;
            session->stats_jitter_client_to_server = session->jitter_tracker.jitter * 1000.0;
            session->next_tracker_update_time = current_time + NEXT_SECONDS_BETWEEN_PACKET_LOSS_UPDATES;
        }
    }

    if ( server->state != NEXT_SERVER_STATE_INITIALIZED )
        return;

    // server update

    bool first_server_update = server->server_update_first;

    if ( server->state != NEXT_SERVER_STATE_DIRECT_ONLY && server->server_update_last_time + NEXT_SECONDS_BETWEEN_SERVER_UPDATES <= current_time )
    {
        if ( server->server_update_request_id != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_INFO, "server update response timed out. falling back to direct mode only :(" );
            server->state = NEXT_SERVER_STATE_DIRECT_ONLY;
            next_server_notify_direct_only_t * notify_direct_only = (next_server_notify_direct_only_t*) next_malloc( server->context, sizeof(next_server_notify_direct_only_t) );
            next_assert( notify_direct_only );
            notify_direct_only->type = NEXT_SERVER_NOTIFY_DIRECT_ONLY;
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server internal thread queued up NEXT_SERVER_NOTIFY_DIRECT_ONLY at %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
                next_platform_mutex_guard( &server->notify_mutex );
                next_queue_push( server->notify_queue, notify_direct_only );
            }
            return;
        }

        while ( server->server_update_request_id == 0 )
        {
            server->server_update_request_id = next_random_uint64();
        }

        server->server_update_resend_time = current_time + 1.0;
        server->server_update_num_sessions = next_session_manager_num_entries( server->session_manager );

        NextBackendServerUpdateRequestPacket packet;

        packet.request_id = server->server_update_request_id;
        packet.customer_id = server->customer_id;
        packet.datacenter_id = server->datacenter_id;
        packet.match_id = server->match_id;
        packet.num_sessions = server->server_update_num_sessions;
        packet.server_address = server->server_address;

        uint8_t magic[8];
        memset( magic, 0, sizeof(magic) );

        uint8_t from_address_data[32];
        uint8_t to_address_data[32];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

        uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

        next_assert( ( size_t(packet_data) % 4 ) == 0 );

        int packet_bytes = 0;
        if ( next_write_backend_packet( NEXT_BACKEND_SERVER_UPDATE_REQUEST_PACKET, &packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write server update request packet for backend" );
            return;
        }

        next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
        next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

        next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );

        server->server_update_last_time = current_time;

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent server update packet to backend (%d sessions)", packet.num_sessions );

        server->server_update_first = false;
    }

    if ( first_server_update )
        return;

    // server update resend

    if ( server->server_update_request_id && server->server_update_resend_time <= current_time )
    {
        NextBackendServerUpdateRequestPacket packet;

        packet.request_id = server->server_update_request_id;
        packet.customer_id = server->customer_id;
        packet.datacenter_id = server->datacenter_id;
        packet.num_sessions = server->server_update_num_sessions;
        packet.server_address = server->server_address;

        uint8_t magic[8];
        memset( magic, 0, sizeof(magic) );

        uint8_t from_address_data[32];
        uint8_t to_address_data[32];
        uint16_t from_address_port;
        uint16_t to_address_port;
        int from_address_bytes;
        int to_address_bytes;

        next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
        next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

        uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

        next_assert( ( size_t(packet_data) % 4 ) == 0 );

        int packet_bytes = 0;
        if ( next_write_backend_packet( NEXT_BACKEND_SERVER_UPDATE_REQUEST_PACKET, &packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write server update packet for backend" );
            return;
        }

        next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
        next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

        next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );

        next_printf( NEXT_LOG_LEVEL_DEBUG, "server resent server update packet to backend", packet.num_sessions );

        server->server_update_resend_time = current_time + 1.0;
    }

    // session updates

    for ( int i = 0; i <= max_entry_index; ++i )
    {
        if ( server->session_manager->session_ids[i] == 0 )
            continue;

        next_session_entry_t * session = &server->session_manager->entries[i];

        if ( !session->session_update_timed_out && ( ( session->next_session_update_time >= 0.0 && session->next_session_update_time <= current_time ) || ( session->session_update_flush && !session->session_update_flush_finished && !session->waiting_for_update_response ) ) )
        {
            NextBackendSessionUpdateRequestPacket packet;

            packet.Reset();

            packet.customer_id = server->customer_id;
            packet.datacenter_id = server->datacenter_id;
            packet.session_id = session->session_id;
            packet.slice_number = session->update_sequence++;
            packet.platform_id = session->stats_platform_id;
            packet.user_hash = session->user_hash;
            session->previous_session_events = session->current_session_events;
            session->current_session_events = 0;
            packet.session_events = session->previous_session_events;
            // todo: packet.internal_events
            packet.reported = session->stats_reported;
            packet.fallback_to_direct = session->stats_fallback_to_direct;
            packet.client_bandwidth_over_limit = session->stats_client_bandwidth_over_limit;
            packet.server_bandwidth_over_limit = session->stats_server_bandwidth_over_limit;
            packet.client_ping_timed_out = session->client_ping_timed_out;
            packet.connection_type = session->stats_connection_type;
            packet.direct_kbps_up = session->stats_direct_kbps_up;
            packet.direct_kbps_down = session->stats_direct_kbps_down;
            packet.next_kbps_up = session->stats_next_kbps_up;
            packet.next_kbps_down = session->stats_next_kbps_down;
            packet.packets_sent_client_to_server = session->stats_packets_sent_client_to_server;
            {
                next_platform_mutex_guard( &server->session_mutex );
                packet.packets_sent_server_to_client = session->stats_packets_sent_server_to_client;
            }

            // IMPORTANT: hold near relay stats for the rest of the session
            if ( session->num_held_near_relays == 0 && session->stats_num_near_relays != 0 )
            {
                session->num_held_near_relays = session->stats_num_near_relays;
                for ( int j = 0; j < session->stats_num_near_relays; j++ )
                {
                    session->held_near_relay_ids[j] = session->stats_near_relay_ids[j];    
                    session->held_near_relay_rtt[j] = session->stats_near_relay_rtt[j];
                    session->held_near_relay_jitter[j] = session->stats_near_relay_jitter[j];
                    session->held_near_relay_packet_loss[j] = session->stats_near_relay_packet_loss[j];    
                }
            }

            packet.packets_lost_client_to_server = session->stats_packets_lost_client_to_server;
            packet.packets_lost_server_to_client = session->stats_packets_lost_server_to_client;
            packet.packets_out_of_order_client_to_server = session->stats_packets_out_of_order_client_to_server;
            packet.packets_out_of_order_server_to_client = session->stats_packets_out_of_order_server_to_client;
            packet.jitter_client_to_server = session->stats_jitter_client_to_server;
            packet.jitter_server_to_client = session->stats_jitter_server_to_client;
            packet.next = session->stats_next;
            packet.next_rtt = session->stats_next_rtt;
            packet.next_jitter = session->stats_next_jitter;
            packet.next_packet_loss = session->stats_next_packet_loss;
            packet.direct_rtt = session->stats_direct_rtt;
            packet.direct_jitter = session->stats_direct_jitter;
            packet.direct_packet_loss = session->stats_direct_packet_loss;
            packet.direct_max_packet_loss_seen = session->stats_direct_max_packet_loss_seen;
            packet.has_near_relay_pings = session->num_held_near_relays != 0;
            packet.num_near_relays = session->num_held_near_relays;
            for ( int j = 0; j < packet.num_near_relays; ++j )
            {
                packet.near_relay_ids[j] = session->held_near_relay_ids[j];
                packet.near_relay_rtt[j] = session->held_near_relay_rtt[j];
                packet.near_relay_jitter[j] = session->held_near_relay_jitter[j];
                packet.near_relay_packet_loss[j] = session->held_near_relay_packet_loss[j];
            }
            packet.client_address = session->address;
            packet.server_address = server->server_address;
            memcpy( packet.client_route_public_key, session->client_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );
            memcpy( packet.server_route_public_key, server->server_route_public_key, NEXT_CRYPTO_BOX_PUBLICKEYBYTES );

            next_assert( session->session_data_bytes >= 0 );
            next_assert( session->session_data_bytes <= NEXT_MAX_SESSION_DATA_BYTES );
            packet.session_data_bytes = session->session_data_bytes;
            memcpy( packet.session_data, session->session_data, session->session_data_bytes );
            memcpy( packet.session_data_signature, session->session_data_signature, NEXT_CRYPTO_SIGN_BYTES );

            session->session_update_request_packet = packet;

            uint8_t magic[8];
            memset( magic, 0, sizeof(magic) );

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

            next_assert( ( size_t(packet_data) % 4 ) == 0 );

            int packet_bytes = 0;
            if ( next_write_backend_packet( NEXT_BACKEND_SESSION_UPDATE_REQUEST_PACKET, &packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write server init request packet for backend" );
                return;
            }

            next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
            next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

            next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );

            next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent session update packet to backend for session %" PRIx64, session->session_id );

            if ( session->next_session_update_time == 0.0 )
            {
                session->next_session_update_time = current_time + NEXT_SECONDS_BETWEEN_SESSION_UPDATES;
            }
            else
            {
                session->next_session_update_time += NEXT_SECONDS_BETWEEN_SESSION_UPDATES;
            }

            session->stats_client_bandwidth_over_limit = false;
            session->stats_server_bandwidth_over_limit = false;

            session->next_session_resend_time = current_time + NEXT_SESSION_UPDATE_RESEND_TIME;

            session->waiting_for_update_response = true;
        }

        if ( session->waiting_for_update_response && session->next_session_resend_time <= current_time )
        {
            session->session_update_request_packet.retry_number++;

            next_printf( NEXT_LOG_LEVEL_DEBUG, "server resent session update packet to backend for session %" PRIx64 " (%d)", session->session_id, session->session_update_request_packet.retry_number );

            uint8_t magic[8];
            memset( magic, 0, sizeof(magic) );

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

            next_assert( ( size_t(packet_data) % 4 ) == 0 );

            int packet_bytes = 0;
            if ( next_write_backend_packet( NEXT_BACKEND_SESSION_UPDATE_REQUEST_PACKET, &session->session_update_request_packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write server init request packet for backend" );
                return;
            }

            next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
            next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

            next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );

            session->next_session_resend_time += NEXT_SESSION_UPDATE_RESEND_TIME;
        }

        if ( !session->session_update_timed_out && session->waiting_for_update_response && session->next_session_update_time - NEXT_SECONDS_BETWEEN_SESSION_UPDATES + NEXT_SESSION_UPDATE_TIMEOUT <= current_time )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "server timed out waiting for backend response for session %" PRIx64, session->session_id );
            session->waiting_for_update_response = false;
            session->next_session_update_time = -1.0;
            session->session_update_timed_out = true;

            // IMPORTANT: Send packets direct from now on for this session
            {
                next_platform_mutex_guard( &server->session_mutex );
                session->mutex_send_over_network_next = false;
            }
        }
    }

    // match data

    for ( int i = 0; i <= max_entry_index; ++i )
    {
        if ( server->session_manager->session_ids[i] == 0 )
            continue;

        next_session_entry_t * session = &server->session_manager->entries[i];

        if ( !session->has_match_data || session->match_data_response_received )
            continue;

        if ( ( session->next_match_data_resend_time == 0.0 && !session->waiting_for_match_data_response) || ( session->match_data_flush && !session->waiting_for_match_data_response ) )
        {
            NextBackendMatchDataRequestPacket packet;
            
            packet.Reset();
            
            packet.customer_id = server->customer_id;
            packet.datacenter_id = server->datacenter_id;
            packet.server_address = server->server_address;
            packet.user_hash = session->user_hash;
            packet.session_id = session->session_id;
            packet.match_id = session->match_id;
            packet.num_match_values = session->num_match_values;
            next_assert( packet.num_match_values <= NEXT_MAX_MATCH_VALUES );
            for ( int j = 0; j < session->num_match_values; ++j )
            {
                packet.match_values[j] = session->match_values[j];
            }

            session->match_data_request_packet = packet;

            uint8_t magic[8];
            memset( magic, 0, sizeof(magic) );

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

            next_assert( ( size_t(packet_data) % 4 ) == 0 );

            int packet_bytes = 0;
            if ( next_write_backend_packet( NEXT_BACKEND_MATCH_DATA_REQUEST_PACKET, &session->match_data_request_packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write match data request packet for backend" );
                return;
            }

            next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
            next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

            next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );
            
            next_printf( NEXT_LOG_LEVEL_DEBUG, "server sent match data packet to backend for session %" PRIx64, session->session_id );

            session->next_match_data_resend_time = ( session->match_data_flush ) ? current_time + NEXT_MATCH_DATA_FLUSH_RESEND_TIME : current_time + NEXT_MATCH_DATA_RESEND_TIME;

            session->waiting_for_match_data_response = true;
        }

        if ( session->waiting_for_match_data_response && session->next_match_data_resend_time <= current_time )
        {
            session->match_data_request_packet.retry_number++;

            next_printf( NEXT_LOG_LEVEL_DEBUG, "server resent match data packet to backend for session %" PRIx64 " (%d)", session->session_id, session->match_data_request_packet.retry_number );

            uint8_t magic[8];
            memset( magic, 0, sizeof(magic) );

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( &server->server_address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &server->backend_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t packet_data[NEXT_MAX_PACKET_BYTES];

            next_assert( ( size_t(packet_data) % 4 ) == 0 );

            int packet_bytes = 0;
            if ( next_write_backend_packet( NEXT_BACKEND_MATCH_DATA_REQUEST_PACKET, &session->match_data_request_packet, packet_data, &packet_bytes, next_signed_packets, server->customer_private_key, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port ) != NEXT_OK )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "server failed to write match data request packet for backend" );
                return;
            }

            next_assert( next_basic_packet_filter( packet_data, packet_bytes ) );
            next_assert( next_advanced_packet_filter( packet_data, magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, packet_bytes ) );

            next_server_internal_send_packet_to_backend( server, packet_data, packet_bytes );

            session->next_match_data_resend_time += ( session->match_data_flush && !session->match_data_flush_finished ) ? NEXT_MATCH_DATA_FLUSH_RESEND_TIME : NEXT_MATCH_DATA_RESEND_TIME;
        }
    }
}

static void next_server_update_internal( next_server_internal_t * server )
{
    next_assert( !next_global_config.disable_network_next );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_server_internal_update_flush( server );

    next_server_internal_update_resolve_hostname( server );

    next_server_internal_update_autodetect( server );

    next_server_internal_update_init( server );

    next_server_internal_update_pending_upgrades( server );

    next_server_internal_update_route( server );

    next_server_internal_update_sessions( server );

    next_server_internal_backend_update( server );

    next_server_internal_pump_commands( server );

#if NEXT_SPIKE_TRACKING

    double finish_time = next_platform_time();

    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_server_update_internal spike %.2f milliseconds", ( finish_time - start_time ) * 1000.0 );
    }

#endif // #if NEXT_SPIKE_TRACKING
}

static void next_server_internal_thread_function( void * context )
{
    next_assert( context );

    next_server_internal_t * server = (next_server_internal_t*) context;

    double last_update_time = next_platform_time();

    while ( !server->quit )
    {
        next_server_internal_block_and_receive_packet( server );

        if ( !next_global_config.disable_network_next && next_platform_time() >= last_update_time + 0.1 )
        {
            next_server_update_internal( server );

            last_update_time = next_platform_time();
        }
    }
}

// ---------------------------------------------------------------

struct next_server_t
{
    NEXT_DECLARE_SENTINEL(0)

    void * context;
    next_server_internal_t * internal;
    next_platform_thread_t * thread;
    next_proxy_session_manager_t * pending_session_manager;
    next_proxy_session_manager_t * session_manager;
    next_address_t address;
    uint16_t bound_port;
    bool ready;
    char datacenter_name[NEXT_MAX_DATACENTER_NAME_LENGTH];
    bool flushing;
    bool flushed;
    bool direct_only;

    NEXT_DECLARE_SENTINEL(1)

    uint8_t current_magic[8];

    NEXT_DECLARE_SENTINEL(2)

    void (*packet_received_callback)( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes );
    int (*send_packet_to_address_callback)( void * data, const next_address_t * address, const uint8_t * packet_data, int packet_bytes );
    void * send_packet_to_address_callback_data;

    NEXT_DECLARE_SENTINEL(3)
};

void next_server_initialize_sentinels( next_server_t * server )
{
    (void) server;
    next_assert( server );
    NEXT_INITIALIZE_SENTINEL( server, 0 )
    NEXT_INITIALIZE_SENTINEL( server, 1 )
    NEXT_INITIALIZE_SENTINEL( server, 2 )
    NEXT_INITIALIZE_SENTINEL( server, 3 )
}

void next_server_verify_sentinels( next_server_t * server )
{
    (void) server;
    next_assert( server );
    NEXT_VERIFY_SENTINEL( server, 0 )
    NEXT_VERIFY_SENTINEL( server, 1 )
    NEXT_VERIFY_SENTINEL( server, 2 )
    NEXT_VERIFY_SENTINEL( server, 3 )
    if ( server->session_manager )
        next_proxy_session_manager_verify_sentinels( server->session_manager );
    if ( server->pending_session_manager )
        next_proxy_session_manager_verify_sentinels( server->pending_session_manager );
}

void next_server_destroy( next_server_t * server );

next_server_t * next_server_create( void * context, const char * server_address, const char * bind_address, const char * datacenter, void (*packet_received_callback)( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes ) )
{
    next_assert( server_address );
    next_assert( bind_address );
    next_assert( packet_received_callback );

    next_server_t * server = (next_server_t*) next_malloc( context, sizeof(next_server_t) );
    if ( !server )
        return NULL;

    memset( server, 0, sizeof( next_server_t) );

    next_server_initialize_sentinels( server );

    server->context = context;

    server->internal = next_server_internal_create( context, server_address, bind_address, datacenter );
    if ( !server->internal )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create internal server" );
        next_server_destroy( server );
        return NULL;
    }

    server->address = server->internal->server_address;
    server->bound_port = server->internal->server_address.port;

    server->thread = next_platform_thread_create( server->context, next_server_internal_thread_function, server->internal );
    if ( !server->thread )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create server thread" );
        next_server_destroy( server );
        return NULL;
    }

    // todo
    /*
    if ( next_platform_thread_high_priority( server->thread ) )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server increased thread priority" );
    }
    */

    server->pending_session_manager = next_proxy_session_manager_create( context, NEXT_INITIAL_PENDING_SESSION_SIZE );
    if ( server->pending_session_manager == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create pending session manager (proxy)" );
        next_server_destroy( server );
        return NULL;
    }

    server->session_manager = next_proxy_session_manager_create( context, NEXT_INITIAL_SESSION_SIZE );
    if ( server->session_manager == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server could not create session manager (proxy)" );
        next_server_destroy( server );
        return NULL;
    }

    server->context = context;
    server->packet_received_callback = packet_received_callback;

    next_server_verify_sentinels( server );

    return server;
}

uint16_t next_server_port( next_server_t * server )
{
    next_server_verify_sentinels( server );

    return server->bound_port;
}

const next_address_t * next_server_address( next_server_t * server )
{
    next_server_verify_sentinels( server );

    return &server->address;
}

void next_server_destroy( next_server_t * server )
{
    next_server_verify_sentinels( server );

    if ( server->pending_session_manager )
    {
        next_proxy_session_manager_destroy( server->pending_session_manager );
    }

    if ( server->session_manager )
    {
        next_proxy_session_manager_destroy( server->session_manager );
    }

    if ( server->thread )
    {
        next_server_internal_quit( server->internal );
        next_platform_thread_join( server->thread );
        next_platform_thread_destroy( server->thread );
    }

    if ( server->internal )
    {
        next_server_internal_destroy( server->internal );
    }

    next_clear_and_free( server->context, server, sizeof(next_server_t) );
}

void next_server_update( next_server_t * server )
{
    next_server_verify_sentinels( server );

#if NEXT_SPIKE_TRACKING
    next_printf( NEXT_LOG_LEVEL_SPAM, "next_server_update" );
#endif // #if NEXT_SPIKE_TRACKING

    while ( true )
    {
        void * queue_entry = NULL;
        {
            next_platform_mutex_guard( &server->internal->notify_mutex );
            queue_entry = next_queue_pop( server->internal->notify_queue );
        }

        if ( queue_entry == NULL )
            break;

        next_server_notify_t * notify = (next_server_notify_t*) queue_entry;

        switch ( notify->type )
        {
            case NEXT_SERVER_NOTIFY_PACKET_RECEIVED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_PACKET_RECEIVED" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_notify_packet_received_t * packet_received = (next_server_notify_packet_received_t*) notify;
                next_assert( packet_received->packet_data );
                next_assert( packet_received->packet_bytes > 0 );
                next_assert( packet_received->packet_bytes <= NEXT_MAX_PACKET_BYTES - 1 );
#if NEXT_SPIKE_TRACKING
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_SPAM, "server calling packet received callback: from = %s, packet_bytes = %d", next_address_to_string( &packet_received->from, address_buffer ), packet_received->packet_bytes );
#endif // #if NEXT_SPIKE_TRACKING
                server->packet_received_callback( server, server->context, &packet_received->from, packet_received->packet_data, packet_received->packet_bytes );
            }
            break;

            case NEXT_SERVER_NOTIFY_SESSION_UPGRADED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_SESSION_UPDATED" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_notify_session_upgraded_t * session_upgraded = (next_server_notify_session_upgraded_t*) notify;
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_INFO, "server upgraded client %s to session %" PRIx64, next_address_to_string( &session_upgraded->address, address_buffer ), session_upgraded->session_id );
                next_proxy_session_entry_t * proxy_entry = next_proxy_session_manager_find( server->pending_session_manager, &session_upgraded->address );
                if ( proxy_entry && proxy_entry->session_id == session_upgraded->session_id )
                {
                    next_proxy_session_manager_remove_by_address( server->session_manager, &session_upgraded->address );
                    next_proxy_session_manager_remove_by_address( server->pending_session_manager, &session_upgraded->address );
                    proxy_entry = next_proxy_session_manager_add( server->session_manager, &session_upgraded->address, session_upgraded->session_id );
                }
            }
            break;

            case NEXT_SERVER_NOTIFY_PENDING_SESSION_TIMED_OUT:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_PENDING_SESSION_TIMED_OUT" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_notify_pending_session_timed_out_t * pending_session_timed_out = (next_server_notify_pending_session_timed_out_t*) notify;
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_DEBUG, "server timed out pending upgrade of client %s to session %" PRIx64, next_address_to_string( &pending_session_timed_out->address, address_buffer ), pending_session_timed_out->session_id );
                next_proxy_session_entry_t * pending_entry = next_proxy_session_manager_find( server->pending_session_manager, &pending_session_timed_out->address );
                if ( pending_entry && pending_entry->session_id == pending_session_timed_out->session_id )
                {
                    next_proxy_session_manager_remove_by_address( server->pending_session_manager, &pending_session_timed_out->address );
                    next_proxy_session_manager_remove_by_address( server->session_manager, &pending_session_timed_out->address );
                }
            }
            break;

            case NEXT_SERVER_NOTIFY_SESSION_TIMED_OUT:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_SESSION_TIMED_OUT" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_notify_session_timed_out_t * session_timed_out = (next_server_notify_session_timed_out_t*) notify;
                char address_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH];
                next_printf( NEXT_LOG_LEVEL_INFO, "server timed out client %s from session %" PRIx64, next_address_to_string( &session_timed_out->address, address_buffer ), session_timed_out->session_id );
                next_proxy_session_entry_t * proxy_session_entry = next_proxy_session_manager_find( server->session_manager, &session_timed_out->address );
                if ( proxy_session_entry && proxy_session_entry->session_id == session_timed_out->session_id )
                {
                    next_proxy_session_manager_remove_by_address( server->session_manager, &session_timed_out->address );
                }
            }
            break;

            case NEXT_SERVER_NOTIFY_READY:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_READY" );
#endif // #if NEXT_SPIKE_TRACKING
                next_server_notify_ready_t * ready = (next_server_notify_ready_t*) notify;
                next_copy_string( server->datacenter_name, ready->datacenter_name, NEXT_MAX_DATACENTER_NAME_LENGTH );
                server->ready = true;
                next_printf( NEXT_LOG_LEVEL_INFO, "server datacenter is '%s'", ready->datacenter_name );
                next_printf( NEXT_LOG_LEVEL_INFO, "server is ready to receive client connections" );
            }
            break;

            case NEXT_SERVER_NOTIFY_FLUSH_FINISHED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_FLUSH_FINISHED" );
#endif // #if NEXT_SPIKE_TRACKING
                server->flushed = true;
            }
            break;

            case NEXT_SERVER_NOTIFY_MAGIC_UPDATED:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_MAGIC_UPDATED" );
#endif // #if NEXT_SPIKE_TRACKING

                next_server_notify_magic_updated_t * magic_updated = (next_server_notify_magic_updated_t*) notify;

                memcpy( server->current_magic, magic_updated->current_magic, 8 );

                next_printf( NEXT_LOG_LEVEL_DEBUG, "server current magic: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
                    server->current_magic[0],
                    server->current_magic[1],
                    server->current_magic[2],
                    server->current_magic[3],
                    server->current_magic[4],
                    server->current_magic[5],
                    server->current_magic[6],
                    server->current_magic[7] );
            }
            break;

            case NEXT_SERVER_NOTIFY_DIRECT_ONLY:
            {
#if NEXT_SPIKE_TRACKING
                next_printf( NEXT_LOG_LEVEL_SPAM, "server received NEXT_SERVER_NOTIFY_DIRECT_ONLY" );
#endif // #if NEXT_SPIKE_TRACKING
                server->direct_only = true;
            }
            break;

            default: break;
        }

        next_free( server->context, queue_entry );
    }
}

uint64_t next_generate_session_id()
{
    uint64_t session_id = 0;
    while ( session_id == 0 )
    {
        next_crypto_random_bytes( (uint8_t*) &session_id, 8 );
    }
    return session_id;
}

uint64_t next_server_upgrade_session( next_server_t * server, const next_address_t * address, const char * user_id )
{
    next_server_verify_sentinels( server );

    next_assert( server->internal );

    // send upgrade session command to internal server

    next_server_command_upgrade_session_t * command = (next_server_command_upgrade_session_t*) next_malloc( server->context, sizeof( next_server_command_upgrade_session_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server upgrade session failed. could not create upgrade session command" );
        return 0;
    }

    uint64_t session_id = next_generate_session_id();

    uint64_t user_hash = ( user_id != NULL ) ? next_hash_string( user_id ) : 0;

    command->type = NEXT_SERVER_COMMAND_UPGRADE_SESSION;
    command->address = *address;
    command->user_hash = user_hash;
    command->session_id = session_id;

    {
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_UPGRADE_SESSION from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }

    // remove any existing entry for this address. latest upgrade takes precedence

    next_proxy_session_manager_remove_by_address( server->session_manager, address );
    next_proxy_session_manager_remove_by_address( server->pending_session_manager, address );

    // add a new pending session entry for this address

    next_proxy_session_entry_t * entry = next_proxy_session_manager_add( server->pending_session_manager, address, session_id );

    if ( entry == NULL )
    {
        next_assert( !"could not add pending session entry. this should never happen!" );
        return 0;
    }

    return session_id;
}

bool next_server_session_upgraded( next_server_t * server, const next_address_t * address )
{
    next_server_verify_sentinels( server );

    next_assert( server->internal );

    next_proxy_session_entry_t * pending_entry = next_proxy_session_manager_find( server->pending_session_manager, address );
    if ( pending_entry != NULL )
        return true;

    next_proxy_session_entry_t * entry = next_proxy_session_manager_find( server->session_manager, address );
    if ( entry != NULL )
        return true;

    return false;
}

void next_server_send_packet_to_address( next_server_t * server, const next_address_t * address, const uint8_t * packet_data, int packet_bytes )
{
    next_server_verify_sentinels( server );

    next_assert( address );
    next_assert( address->type != NEXT_ADDRESS_NONE );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

    if ( server->send_packet_to_address_callback )
    {
        void * callback_data = server->send_packet_to_address_callback_data;
        if ( server->send_packet_to_address_callback( callback_data, address, packet_data, packet_bytes ) != 0 )
            return;
    }

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( server->internal->socket, address, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING
}

void next_server_send_packet( next_server_t * server, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes )
{
    next_server_verify_sentinels( server );

    next_assert( to_address );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

    if ( next_global_config.disable_network_next )
    {
        next_server_send_packet_direct( server, to_address, packet_data, packet_bytes );
        return;
    }

    if ( packet_bytes > NEXT_MAX_PACKET_BYTES - 1 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server can't send packet because packet size is too large" );
        return;
    }

    next_proxy_session_entry_t * entry = next_proxy_session_manager_find( server->session_manager, to_address );

    bool send_over_network_next = false;
    bool send_upgraded_direct = false;

    if ( entry && packet_bytes <= NEXT_MTU )
    {
        bool multipath = false;
        int envelope_kbps_down = 0;
        uint8_t open_session_sequence = 0;
        uint64_t send_sequence = 0;
        uint64_t session_id = 0;
        uint8_t session_version = 0;
        next_address_t session_address;
        double last_upgraded_packet_receive_time = 0.0;
        uint8_t session_private_key[NEXT_CRYPTO_BOX_SECRETKEYBYTES];

        next_session_entry_t * internal_entry = NULL;
        {
            next_platform_mutex_guard( &server->internal->session_mutex );
            internal_entry = next_session_manager_find_by_address( server->internal->session_manager, to_address );
            if ( internal_entry )
            {
                last_upgraded_packet_receive_time = internal_entry->last_upgraded_packet_receive_time;
            }
        }

        // IMPORTANT: If we haven't received any upgraded packets in the last second send passthrough packets.
        // This makes reconnect robust when a client reconnects using the same port number.
        if ( !internal_entry || last_upgraded_packet_receive_time + 1.0 < next_platform_time() )
        {
            next_server_send_packet_direct( server, to_address, packet_data, packet_bytes );
            return;
        }

        {
            next_platform_mutex_guard( &server->internal->session_mutex );
            multipath = internal_entry->mutex_multipath;
            envelope_kbps_down = internal_entry->mutex_envelope_kbps_down;
            send_over_network_next = internal_entry->mutex_send_over_network_next;
            send_upgraded_direct = !send_over_network_next;
            send_sequence = internal_entry->mutex_payload_send_sequence++;
            open_session_sequence = internal_entry->client_open_session_sequence;
            session_id = internal_entry->mutex_session_id;
            session_version = internal_entry->mutex_session_version;
            session_address = internal_entry->mutex_send_address;
            memcpy( session_private_key, internal_entry->mutex_private_key, NEXT_CRYPTO_BOX_SECRETKEYBYTES );
            internal_entry->stats_packets_sent_server_to_client++;
        }

        if ( multipath )
        {
            send_upgraded_direct = true;
        }

        if ( send_over_network_next )
        {
            const int wire_packet_bits = next_wire_packet_bits( packet_bytes );

            bool over_budget = next_bandwidth_limiter_add_packet( &entry->send_bandwidth, next_platform_time(), envelope_kbps_down, wire_packet_bits );

            if ( over_budget )
            {
                next_printf( NEXT_LOG_LEVEL_WARN, "server exceeded bandwidth budget for session %" PRIx64 " (%d kbps)", session_id, envelope_kbps_down );
                {
                    next_platform_mutex_guard( &server->internal->session_mutex );
                    internal_entry->stats_server_bandwidth_over_limit = true;
                }
                send_over_network_next = false;
                if ( !multipath )
                {
                    send_upgraded_direct = true;
                }
            }
        }

        if ( send_over_network_next )
        {
            // send over network next

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port;
            uint16_t to_address_port;
            int from_address_bytes;
            int to_address_bytes;

            next_address_data( &server->address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( &session_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t next_packet_data[NEXT_MAX_PACKET_BYTES];

            int next_packet_bytes = next_write_server_to_client_packet( next_packet_data, send_sequence, session_id, session_version, session_private_key, packet_data, packet_bytes, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

            next_assert( next_packet_bytes > 0 );

            next_assert( next_basic_packet_filter( next_packet_data, next_packet_bytes ) );
            next_assert( next_advanced_packet_filter( next_packet_data, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, next_packet_bytes ) );

            next_server_send_packet_to_address( server, &session_address, next_packet_data, next_packet_bytes );
        }

        if ( send_upgraded_direct )
        {
            // direct packet

            uint8_t from_address_data[32];
            uint8_t to_address_data[32];
            uint16_t from_address_port = 0;
            uint16_t to_address_port = 0;
            int from_address_bytes = 0;
            int to_address_bytes = 0;

            next_address_data( &server->address, from_address_data, &from_address_bytes, &from_address_port );
            next_address_data( to_address, to_address_data, &to_address_bytes, &to_address_port );

            uint8_t direct_packet_data[NEXT_MAX_PACKET_BYTES];

            int direct_packet_bytes = next_write_direct_packet( direct_packet_data, open_session_sequence, send_sequence, packet_data, packet_bytes, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port );

            next_assert( direct_packet_bytes >= 27 );
            next_assert( direct_packet_bytes <= NEXT_MTU + 27 );
            next_assert( direct_packet_data[0] == NEXT_DIRECT_PACKET );

            next_assert( next_basic_packet_filter( direct_packet_data, direct_packet_bytes ) );
            next_assert( next_advanced_packet_filter( direct_packet_data, server->current_magic, from_address_data, from_address_bytes, from_address_port, to_address_data, to_address_bytes, to_address_port, direct_packet_bytes ) );

            next_server_send_packet_to_address( server, to_address, direct_packet_data, direct_packet_bytes );
        }
    }
    else
    {
        // passthrough packet

        next_server_send_packet_direct( server, to_address, packet_data, packet_bytes );
    }
}

void next_server_send_packet_direct( next_server_t * server, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes )
{
    next_server_verify_sentinels( server );

    next_assert( to_address );
    next_assert( to_address->type != NEXT_ADDRESS_NONE );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

    if ( packet_bytes > NEXT_MAX_PACKET_BYTES - 1 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server can't send packet because packet size is too large\n" );
        return;
    }

    uint8_t buffer[NEXT_MAX_PACKET_BYTES];
    buffer[0] = NEXT_PASSTHROUGH_PACKET;
    memcpy( buffer + 1, packet_data, packet_bytes );
    next_server_send_packet_to_address( server, to_address, buffer, packet_bytes + 1 );
}

void next_server_send_packet_raw( struct next_server_t * server, const struct next_address_t * to_address, const uint8_t * packet_data, int packet_bytes )
{
    next_server_verify_sentinels( server );

    next_assert( to_address );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

#if NEXT_SPIKE_TRACKING
    double start_time = next_platform_time();
#endif // #if NEXT_SPIKE_TRACKING

    next_platform_socket_send_packet( server->internal->socket, to_address, packet_data, packet_bytes );

#if NEXT_SPIKE_TRACKING
    double finish_time = next_platform_time();
    if ( finish_time - start_time > 0.001 )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "next_platform_socket_send_packet spiked %.2f milliseconds at %s:%d", ( finish_time - start_time ) * 1000.0, __FILE__, __LINE__ );
    }
#endif // #if NEXT_SPIKE_TRACKING
}

bool next_server_stats( next_server_t * server, const next_address_t * address, next_server_stats_t * stats )
{
    next_assert( server );
    next_assert( address );
    next_assert( stats );

    next_platform_mutex_guard( &server->internal->session_mutex );

    next_session_entry_t * entry = next_session_manager_find_by_address( server->internal->session_manager, address );
    if ( !entry )
        return false;

    stats->session_id = entry->session_id;
    stats->user_hash = entry->user_hash;
    stats->platform_id = entry->stats_platform_id;
    stats->connection_type = entry->stats_connection_type;
    stats->next = entry->stats_next;
    stats->multipath = entry->stats_multipath;
    stats->reported = entry->stats_reported;
    stats->fallback_to_direct = entry->stats_fallback_to_direct;
    stats->direct_rtt = entry->stats_direct_rtt;
    stats->direct_jitter = entry->stats_direct_jitter;
    stats->direct_packet_loss = entry->stats_direct_packet_loss;
    stats->direct_max_packet_loss_seen = entry->stats_direct_max_packet_loss_seen;
    stats->next_rtt = entry->stats_next_rtt;
    stats->next_jitter = entry->stats_next_jitter;
    stats->next_packet_loss = entry->stats_next_packet_loss;
    stats->direct_kbps_up = entry->stats_direct_kbps_up;
    stats->direct_kbps_down = entry->stats_direct_kbps_down;
    stats->next_kbps_up = entry->stats_next_kbps_up;
    stats->next_kbps_down = entry->stats_next_kbps_down;
    stats->packets_sent_client_to_server = entry->stats_packets_sent_client_to_server;
    stats->packets_sent_server_to_client = entry->stats_packets_sent_server_to_client;
    stats->packets_lost_client_to_server = entry->stats_packets_lost_client_to_server;
    stats->packets_lost_server_to_client = entry->stats_packets_lost_server_to_client;
    stats->packets_out_of_order_client_to_server = entry->stats_packets_out_of_order_client_to_server;
    stats->packets_out_of_order_server_to_client = entry->stats_packets_out_of_order_server_to_client;
    stats->jitter_client_to_server = entry->stats_jitter_client_to_server;
    stats->jitter_server_to_client = entry->stats_jitter_server_to_client;

    return true;
}

bool next_server_ready( next_server_t * server ) 
{
    next_server_verify_sentinels( server );
    return ( next_global_config.disable_network_next || server->ready ) ? true : false;
}

const char * next_server_datacenter( next_server_t * server )
{
    next_server_verify_sentinels( server );

    return server->datacenter_name;
}

void next_server_session_event( struct next_server_t * server, const struct next_address_t * address, uint64_t session_events )
{
    next_assert( server );
    next_assert( address );
    next_assert( server->internal );

    if ( server->flushing )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "ignoring session event. server is flushed" );
        return;
    }
    
    // send session event command to internal server

    next_server_command_session_event_t * command = (next_server_command_session_event_t*) next_malloc( server->context, sizeof( next_server_command_session_event_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "session event failed. could not create session event command" );
        return;
    }

    command->type = NEXT_SERVER_COMMAND_SESSION_EVENT;
    command->address = *address;
    command->session_events = session_events;

    {    
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_SERVER_EVENT from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }
}

void next_server_match( struct next_server_t * server, const struct next_address_t * address, const char * match_id, const double * match_values, int num_match_values )
{
    next_server_verify_sentinels( server );

    next_assert( server );
    next_assert( address );
    next_assert( server->internal );
    next_assert( num_match_values >= 0 );
    next_assert( num_match_values <= NEXT_MAX_MATCH_VALUES );

    if ( server->flushing )
    {
        next_printf( NEXT_LOG_LEVEL_WARN, "ignoring server match. server is flushed" );
        return;
    }

    // send match data command to internal server

    next_server_command_match_data_t * command = (next_server_command_match_data_t*) next_malloc( server->context, sizeof( next_server_command_match_data_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server match data failed. could not create match data command" );
        return;
    }

    command->type = NEXT_SERVER_COMMAND_MATCH_DATA;
    command->address = *address;
    command->match_id = next_hash_string( match_id );
    memset( command->match_values, 0, sizeof(command->match_values) );
    for ( int i = 0; i < num_match_values; ++i )
    {
        command->match_values[i] = match_values[i];
    }
    command->num_match_values = num_match_values;

    {
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_MATCH_DATA from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }
}

void next_server_flush( struct next_server_t * server )
{
    next_assert( server );

    if ( next_global_config.disable_network_next == true )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring server flush. network next is disabled" );
        return;
    }

    if ( server->flushing )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring server flush. server is already flushed" );
        return;
    }

    // send flush command to internal server

    next_server_command_flush_t * command = (next_server_command_flush_t*) next_malloc( server->context, sizeof( next_server_command_flush_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server flush failed. could not create server flush command" );
        return;
    }

    command->type = NEXT_SERVER_COMMAND_FLUSH;

    {    
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_FLUSH from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }

    server->flushing = true;

    next_printf( NEXT_LOG_LEVEL_INFO, "server flush started" );

    double flush_timeout = next_platform_time() + NEXT_SERVER_FLUSH_TIMEOUT;

    while ( !server->flushed && next_platform_time() < flush_timeout )
    {
        next_server_update( server );
        
        next_platform_sleep( 0.1 );
    }

    if ( next_platform_time() > flush_timeout )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server flush timed out :(" );
    }
    else
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "server flush finished" );    
    }
}

void next_server_set_packet_receive_callback( struct next_server_t * server, void (*callback) ( void * data, next_address_t * from, uint8_t * packet_data, int * begin, int * end ), void * callback_data )
{
    next_assert( server );

    next_server_command_set_packet_receive_callback_t * command = (next_server_command_set_packet_receive_callback_t*) next_malloc( server->context, sizeof( next_server_command_set_packet_receive_callback_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server set packet receive callback failed. could not create command" );
        return;
    }

    command->type = NEXT_SERVER_COMMAND_SET_PACKET_RECEIVE_CALLBACK;
    command->callback = callback;
    command->callback_data = callback_data;

    {    
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_SET_PACKET_RECEIVE_CALLBACK from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }
}

void next_server_set_send_packet_to_address_callback( struct next_server_t * server, int (*callback) ( void * data, const next_address_t * from, const uint8_t * packet_data, int packet_bytes ), void * callback_data )
{
    next_assert( server );

    server->send_packet_to_address_callback = callback;
    server->send_packet_to_address_callback_data = callback_data;

    next_server_command_set_send_packet_to_address_callback_t * command = (next_server_command_set_send_packet_to_address_callback_t*) next_malloc( server->context, sizeof( next_server_command_set_send_packet_to_address_callback_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server set send packet to address callback failed. could not create command" );
        return;
    }

    command->type = NEXT_SERVER_COMMAND_SET_SEND_PACKET_TO_ADDRESS_CALLBACK;
    command->callback = callback;
    command->callback_data = callback_data;

    {    
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_SEND_PACKET_TO_ADDRESS_CALLBACK from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }
}

void next_server_set_payload_receive_callback( struct next_server_t * server, int (*callback) ( void * data, const next_address_t * client_address, const uint8_t * payload_data, int payload_bytes ), void * callback_data )
{
    next_assert( server );

    next_server_command_set_payload_receive_callback_t * command = (next_server_command_set_payload_receive_callback_t*) next_malloc( server->context, sizeof( next_server_command_set_payload_receive_callback_t ) );
    if ( !command )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "server set payload receive callback failed. could not create command" );
        return;
    }

    command->type = NEXT_SERVER_COMMAND_SET_PAYLOAD_RECEIVE_CALLBACK;
    command->callback = callback;
    command->callback_data = callback_data;

    {    
#if NEXT_SPIKE_TRACKING
        next_printf( NEXT_LOG_LEVEL_SPAM, "server queues up NEXT_SERVER_COMMAND_SEND_PACKET_TO_ADDRESS_CALLBACK from %s:%d", __FILE__, __LINE__ );
#endif // #if NEXT_SPIKE_TRACKING
        next_platform_mutex_guard( &server->internal->command_mutex );
        next_queue_push( server->internal->command_queue, command );
    }
}

bool next_server_direct_only( struct next_server_t * server )
{
    next_assert( server );
    return server->direct_only;
}

#ifdef _MSC_VER
#pragma warning(pop)
#endif
