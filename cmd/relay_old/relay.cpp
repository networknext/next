/*
    Network Next: $(NEXT_VERSION_FULL)
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#include "relay.h"
#include "relay_internal.h"
#include <stdlib.h>
#include <memory.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>
#include <math.h>
#include <time.h>
#include "compat.hpp"

// -------------------------------------------------------------

static void * (*next_alloc_function)( size_t );
static void * (*next_realloc_function)( void*, size_t );
static void (*next_free_function)( void* );

volatile bool quit = false;

void next_set_allocator( void * (*alloc_function)(size_t), void * (*realloc_function)(void*, size_t), void (*free_function)(void*) )
{
    next_assert( alloc_function );
    next_assert( realloc_function );
    next_assert( free_function );
    next_alloc_function = alloc_function;
    next_realloc_function = realloc_function;
    next_free_function = free_function;
}

void * next_realloc( void * p, size_t bytes )
{
    next_assert( next_realloc_function );
    return next_realloc_function( p, bytes );
}

void * next_alloc( size_t bytes )
{
    next_assert( next_alloc_function );
    return next_alloc_function( bytes );
}

void next_free( void * p )
{
    next_assert( next_alloc_function );
    return next_free_function( p );
}

int next_init()
{
    if ( next_alloc_function == NULL )
    {
        next_alloc_function = malloc;
    }

    if ( next_realloc_function == NULL )
    {
        next_realloc_function = realloc;
    }

    if ( next_free_function == NULL )
    {
        next_free_function = free;
    }

    if ( !next_internal_init() )
    {
        return NEXT_ERROR;
    }

    if ( sodium_init() == -1 )
    {
        return NEXT_ERROR;
    }

    return NEXT_OK;
}

void next_term()
{
    next_internal_term();
}

#include <sodium.h>

#include "miniz.h"
#include <sparsehash/dense_hash_map>
#include <signal.h>
#include "concurrentqueue.h"
#include <unistd.h>
#include <fcntl.h>
#include "rapidjson/filereadstream.h"

#include <sys/types.h>
#include <sys/stat.h>
#include <ifaddrs.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <time.h>

// allow about 10 kbps for server pings and pongs
#define NEXT_SERVER_PING_PONG_BITS ( NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS( NEXT_HEADER_BYTES + NEXT_PACKET_V2_PING_PONG_BYTES ) )
#define NEXT_SERVER_PING_PONG_KBPS ( ( 13 * NEXT_SERVER_PING_PONG_BITS ) / 1000 )

// allow about 10 kbps for client-relay pings (it's 6 kbps for pongs)
#define NEXT_CLIENT_RELAY_PING_BITS ( NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS( BYTES_V3_PING ) )
#define NEXT_CLIENT_RELAY_PING_KBPS ( ( 13 * NEXT_CLIENT_RELAY_PING_BITS ) / 1000 )

// allow about 10 kbps for relay pings (it's 6 kbps for pongs)
#define NEXT_RELAY_PING_BITS ( NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS( BYTES_V3_PING ) )
#define NEXT_RELAY_PING_KBPS ( ( 13 * NEXT_RELAY_PING_BITS ) / 1000 )

// allow about 76 kbps for management traffic up
#define NEXT_MAX_ROUTE_REQUEST_BITS ( NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS( NEXT_MAX_ROUTE_REQUEST_BYTES ) )
#define NEXT_MANAGEMENT_UP_KBPS ( ( 13 * NEXT_MAX_ROUTE_REQUEST_BITS ) / 1000 )

// allow about 12 kbps for management traffic down
#define NEXT_ROUTE_RESPONSE_BITS ( NEXT_FLOW_BANDWIDTH_BYTES_TO_BITS( NEXT_HEADER_BYTES + NEXT_ENCRYPTED_SERVER_TOKEN_BYTES  ) )
#define NEXT_MANAGEMENT_DOWN_KBPS ( ( 13 * NEXT_ROUTE_RESPONSE_BITS ) / 1000 )

#define MAX_ENVS 3

#define INTERVAL_UPDATE ( int64_t( 10.0 * NEXT_ONE_SECOND_NS ) )
#define INTERVAL_TRAFFIC_STATS ( int64_t( 10.0 * NEXT_ONE_SECOND_NS ) )
#define INTERVAL_PING_TARGETS ( int64_t( 0.1 * NEXT_ONE_SECOND_NS ) )
#define MANAGE_PEER_HISTORY_ENTRY_COUNT 20
#define MANAGE_PEER_HISTORY_PACKET_LOSS_ENTRY_COUNT 110
#define FLOW_THREAD_MAX 64
#define DEFAULT_PORT 40000
#define FLOW_TIMEOUT 10

#define NEXT_SIGNATURE_KEY_BYTES 64

#define NEXT_PACKET_TYPE_V3_RELAY_PING 5
#define NEXT_PACKET_TYPE_V3_RELAY_PONG 6

#define NEXT_PING_KEY_BYTES 32

#define RELAY_NAME_MAX 128

#define BAD_CPU_INDEX ( -1 )

#define SHUTDOWN_ACK_SECONDS 30.0f
#define SHUTDOWN_MAX_SECONDS 60.0f

#define USAGE_MAX_SAMPLES 32
#define SPEED_MB          1000000.0
#define SPEED_HUNDRED_MB     ( SPEED_MB  * 100  )
#define SPEED_GIG            ( SPEED_MB  * 1000 )
#define SPEED_TEN_GIG        ( SPEED_GIG * 10   )
#define DEFAULT_BITS_PER_SEC ( SPEED_GIG )

#define ARRAY_SIZE(array) ( sizeof( array ) / sizeof( array[0] ) )

#define ASSERT_CONCAT_(a, b) a##b
#define ASSERT_CONCAT(a, b) ASSERT_CONCAT_(a, b)
#define ASSERT_ON_COMPILE(e) \
enum { ASSERT_CONCAT(assert_line_, __LINE__) = 1/(!!(e)) }

#define NEXT_LOW_LEVEL_HEADER_BYTES ( 14 + 20 + 8 + 4 )

// public key used to verify the master's UDP packets
uint8_t NEXT_MASTER_UDP_SIGN_KEY[] =
{
    0x60,0x45,0x96,0x52,0x4f,0x1c,0x00,0xda,
    0x35,0x1b,0x6c,0x17,0x8b,0xa8,0xaa,0xac,
    0xb4,0x8c,0x76,0xb1,0x72,0xa6,0xfa,0x7f,
    0x52,0x28,0xd8,0x6d,0x9e,0x2b,0x91,0xec
};

// we seal our UDP packets up with the master's public key before sending them
uint8_t NEXT_MASTER_UDP_SEAL_KEY[] =
{
    0x77,0x9f,0xf2,0xeb,0x45,0xfb,0xe8,0x25,
    0x7a,0xf3,0x78,0xf9,0x26,0x22,0x29,0xc0,
    0xa8,0xd0,0x66,0x92,0x8b,0xf9,0x47,0xcc,
    0x8b,0x93,0x62,0xbe,0xb3,0x88,0xf9,0x6f
};

// padding to prevent UDP amplification attacks
uint8_t NEXT_MASTER_INIT_KEY[] =
{
    0x91, 0x45, 0x19, 0x24, 0xec, 0xb0, 0x8b, 0xd7, 0xe4, 0xe6, 0xb4, 0x4d, 0xd4, 0x21, 0xab, 0x21,
    0x10, 0xc1, 0xf5, 0xfb, 0x02, 0x26, 0x14, 0xd7, 0x78, 0xb6, 0x2c, 0x83, 0x41, 0x3d, 0x53, 0x07,
    0x9a, 0x0c, 0x32, 0xb1, 0xf3, 0x65, 0x82, 0x6e, 0x2a, 0xe8, 0x33, 0xc3, 0xd2, 0x2b, 0x69, 0x7b,
    0x38, 0x79, 0x39, 0x20, 0x7a, 0xc7, 0x03, 0xc8, 0xab, 0xae, 0x9c, 0x94, 0xf1, 0xac, 0xf0, 0xe3,
    0x74, 0xbc, 0x3c, 0xc0, 0x45, 0xeb, 0xfc, 0x81, 0x63, 0xcb, 0xe6, 0xd1, 0x94, 0x2b, 0x90, 0x1f,
    0xc7, 0x96, 0xb8, 0x83, 0xe8, 0xf6, 0x2f, 0x2c, 0x23, 0x0b, 0x23, 0x0d, 0xaf, 0x7c, 0x26, 0xee,
    0x2c, 0x4a, 0xee, 0x46, 0xa0, 0xf6, 0xb8, 0xe4, 0x34, 0xfa, 0xc8, 0x7b, 0x9a, 0x52, 0x06, 0xd8,
    0x35, 0x0c, 0x2f, 0x53, 0x2b, 0xab, 0x41, 0x17, 0x04, 0xfb, 0x87, 0xb6, 0xeb, 0xa3, 0x2d, 0xf5,
    0x57, 0xac, 0x22, 0xb0, 0x49, 0x0f, 0x96, 0xdf, 0xda, 0xa6, 0x7a, 0x97, 0x0c, 0x47, 0x18, 0x61,
    0x45, 0xda, 0x3e, 0x23, 0x3d, 0x58, 0xe4, 0xe1, 0x5e, 0x2a, 0x27, 0x51, 0xc1, 0xc0, 0x93, 0x1b,
    0x4f, 0x8b, 0x98, 0x7c, 0x13, 0x71, 0xdf, 0xbb, 0x97, 0x3a, 0x11, 0xd3, 0x3b, 0x84, 0xba, 0x31,
    0xc3, 0x21, 0x53, 0xb9, 0x37, 0x9d, 0x1e, 0x19, 0x94, 0xf3, 0x44, 0x3b, 0x34, 0x88, 0x52, 0x99,
    0x9b, 0x02, 0x22, 0x73, 0x04, 0x0c, 0xf1, 0xcb, 0xe5, 0xad, 0x6f, 0xa2, 0xee, 0x7d, 0xa9, 0xce,
    0xdd, 0x4d, 0x16, 0xae, 0x59, 0x2c, 0x05, 0xa1, 0x68, 0xde, 0xa4, 0x24, 0x31, 0xcd, 0x6d, 0xeb,
    0x94, 0xc2, 0xef, 0x7a, 0xa2, 0xee, 0x6d, 0xc1, 0x96, 0x2c, 0x45, 0xf9, 0x00, 0xf3, 0x69, 0x8e,
    0x83, 0x55, 0x31, 0xd2, 0x14, 0xb2, 0x47, 0xee, 0x10, 0x08, 0x60, 0xe4, 0x41, 0x6d, 0xb2, 0x10,
};

// -------------------------------------------------------------

#define METRICS_RATE 30.0f
enum
{
    RELAY_STATS_INVALID_INGRESS_BYTES = 0,
    RELAY_STATS_PAID_EGRESS_BYTES,
    RELAY_STATS_MANAGEMENT_EGRESS_BYTES,
    RELAY_STATS_MEASUREMENT_EGRESS_BYTES,
    RELAY_COUNTER_ROUTE_REQUEST_PACKETS_FORWARDED,
    RELAY_COUNTER_ROUTE_REQUEST_PACKET_DECRYPT_TOKEN_FAILED,
    RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED,
    RELAY_COUNTER_FLOW_CREATE,
    RELAY_COUNTER_FLOW_UPDATE,
    RELAY_COUNTER_ROUTE_REQUEST_PACKETS_RECEIVED,
    RELAY_STATS_MANAGEMENT_INGRESS_BYTES,
    RELAY_COUNTER_ROUTE_RESPONSE_PACKET_BAD_SIZE,
    RELAY_COUNTER_ROUTE_RESPONSE_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_ROUTE_RESPONSE_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_ROUTE_RESPONSE_PACKETS_FORWARDED,
    RELAY_COUNTER_ROUTE_RESPONSE_PACKETS_RECEIVED,
    RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_BAD_SIZE,
    RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_CLIENT_TO_SERVER_PACKETS_FORWARDED,
    RELAY_COUNTER_CLIENT_TO_SERVER_PACKETS_RECEIVED,
    RELAY_STATS_PAID_INGRESS_BYTES,
    RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_BAD_SIZE,
    RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_SERVER_TO_CLIENT_PACKETS_FORWARDED,
    RELAY_COUNTER_SERVER_TO_CLIENT_PACKETS_RECEIVED,
    RELAY_COUNTER_CONTINUE_REQUEST_PACKETS_FORWARDED,
    RELAY_COUNTER_CONTINUE_REQUEST_PACKET_BAD_SIZE,
    RELAY_COUNTER_CONTINUE_REQUEST_PACKET_DECRYPT_TOKEN_FAILED,
    RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED,
    RELAY_COUNTER_CONTINUE_REQUEST_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_FLOW_CONTINUE,
    RELAY_COUNTER_CONTINUE_REQUEST_PACKETS_RECEIVED,
    RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_BAD_SIZE,
    RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_CONTINUE_RESPONSE_PACKETS_FORWARDED,
    RELAY_COUNTER_CONTINUE_RESPONSE_PACKETS_RECEIVED,

    // todo: remove migrate concept. not used in v3
    RELAY_COUNTER_MIGRATE_PACKET_BAD_SIZE,
    RELAY_COUNTER_MIGRATE_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_MIGRATE_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_MIGRATE_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_MIGRATE_PACKETS_RECEIVED,
    RELAY_COUNTER_MIGRATE_PACKETS_FORWARDED,
    RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_BAD_SIZE,
    RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_MIGRATE_RESPONSE_PACKETS_FORWARDED,
    RELAY_COUNTER_MIGRATE_RESPONSE_PACKETS_RECEIVED,

    // todo: remove destroy concept. not used in v3.
    RELAY_COUNTER_DESTROY_PACKET_BAD_SIZE,
    RELAY_COUNTER_DESTROY_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_DESTROY_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_DESTROY_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_DESTROY_PACKETS_FORWARDED,
    RELAY_COUNTER_DESTROY_PACKETS_RECEIVED,

    // todo: remove flow timout. not used in v3.
    RELAY_COUNTER_FLOW_TIMEOUT,

    RELAY_COUNTER_FLOW_EXPIRE,
    RELAY_COUNTER_RELAY_PING_PACKETS_SENT,
    RELAY_COUNTER_RELAY_PING_PACKET_BAD_SIZE,
    RELAY_COUNTER_RELAY_PING_PACKET_NO_PING_TARGET,
    RELAY_COUNTER_RELAY_PING_PACKET_COULD_NOT_READ,
    RELAY_COUNTER_RELAY_PING_PACKET_COULD_NOT_WRITE_PONG_DATA,
    RELAY_COUNTER_RELAY_PING_PACKETS_RECEIVED,

    // todo: remove. measurement relay concept no longer is used.
    RELAY_STATS_MEASUREMENT_INGRESS_BYTES,

    RELAY_COUNTER_RELAY_PONG_PACKETS_SENT,
    RELAY_COUNTER_RELAY_PONG_PACKET_BAD_SIZE,
    RELAY_COUNTER_RELAY_PONG_PACKET_NO_PING_TARGET,
    RELAY_COUNTER_RELAY_PONG_PACKET_COULD_NOT_READ,
    RELAY_COUNTER_RELAY_PONG_PACKETS_RECEIVED,
    RELAY_COUNTER_PING_PACKET_BAD_SIZE,
    RELAY_COUNTER_CLIENT_PING_PACKETS_RECEIVED,
    RELAY_COUNTER_NEXT_SERVER_PING_PACKET_BAD_SIZE,
    RELAY_COUNTER_NEXT_SERVER_PING_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_NEXT_SERVER_PING_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_NEXT_SERVER_PING_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_NEXT_SERVER_PING_PACKETS_FORWARDED,
    RELAY_COUNTER_NEXT_SERVER_PING_PACKETS_RECEIVED,
    RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_BAD_SIZE,
    RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_NO_FLOW_ENTRY,
    RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_VERIFY_HEADER_FAILED,
    RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_ALREADY_RECEIVED,
    RELAY_COUNTER_NEXT_SERVER_PONG_PACKETS_FORWARDED,
    RELAY_COUNTER_NEXT_SERVER_PONG_PACKETS_RECEIVED,
    RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED,
    RELAY_COUNTER_POST_RELAY_STATS_BAD_HTTP_STATUS_CODE,
    RELAY_COUNTER_PING_PACKET_INVALID_SIGNATURE,
    RELAY_COUNTER_PING_PACKET_EXPIRED,
    RELAY_COUNTER_CLIENT_TO_SERVER_FLOW_EXCEED_BANDWIDTH,
    RELAY_COUNTER_SERVER_TO_CLIENT_FLOW_EXCEED_BANDWIDTH,
    RELAY_COUNTER_NEAR_EXCEED_BANDWIDTH,
    RELAY_COUNTER_RELAY_PING_EXCEED_BANDWIDTH,
    RELAY_COUNTER_CLIENT_TO_SERVER_MGMT_EXCEED_BANDWIDTH,
    RELAY_COUNTER_SERVER_TO_CLIENT_MGMT_EXCEED_BANDWIDTH,

    /* add new count metrics above here */
    COUNT_METRIC_MAX,
};

// todo: clean up metrics to match changes above ----^
const char * metric_count_to_string[] =
{
    "relay.stats.invalid_ingress_bytes",
    "relay.stats.paid_egress_bytes",
    "relay.stats.management_egress_bytes",
    "relay.stats.measurement_egress_bytes",
    "relay.counter.route_request_packets_forwarded",
    "relay.counter.route_request_packet_decrypt_token_failed",
    "relay.counter.route_request_packet_token_expired",
    "relay.counter.flow_create",
    "relay.counter.flow_update",
    "relay.counter.route_request_packets_received",
    "relay.stats.management_ingress_bytes",
    "relay.counter.route_response_packet_bad_size",
    "relay.counter.route_response_packet_no_flow_entry",
    "relay.counter.route_response_packet_verify_header_failed",
    "relay.counter.route_response_packet_already_received",
    "relay.counter.route_response_packets_forwarded",
    "relay.counter.route_response_packets_received",
    "relay.counter.client_to_server_packet_bad_size",
    "relay.counter.client_to_server_packet_no_flow_entry",
    "relay.counter.client_to_server_packet_verify_header_failed",
    "relay.counter.client_to_server_packet_already_received",
    "relay.counter.client_to_server_packets_forwarded",
    "relay.counter.client_to_server_packets_received",
    "relay.stats.paid_ingress_bytes",
    "relay.counter.server_to_client_packet_bad_size",
    "relay.counter.server_to_client_packet_no_flow_entry",
    "relay.counter.server_to_client_packet_verify_header_failed",
    "relay.counter.server_to_client_packet_already_received",
    "relay.counter.server_to_client_packets_forwarded",
    "relay.counter.server_to_client_packets_received",
    "relay.counter.continue_request_packets_forwarded",
    "relay.counter.continue_request_packet_bad_size",
    "relay.counter.continue_request_packet_decrypt_token_failed",
    "relay.counter.continue_request_packet_token_expired",
    "relay.counter.continue_request_packet_no_flow_entry",
    "relay.counter.flow_continue",
    "relay.counter.continue_request_packets_received",
    "relay.counter.continue_response_packet_bad_size",
    "relay.counter.continue_response_packet_no_flow_entry",
    "relay.counter.continue_response_packet_verify_header_failed",
    "relay.counter.continue_response_packet_already_received",
    "relay.counter.continue_response_packets_forwarded",
    "relay.counter.continue_response_packets_received",
    "relay.counter.migrate_packet_bad_size",
    "relay.counter.migrate_packet_no_flow_entry",
    "relay.counter.migrate_packet_verify_header_failed",
    "relay.counter.migrate_packet_already_received",
    "relay.counter.migrate_packets_received",
    "relay.counter.migrate_packets_forwarded",
    "relay.counter.migrate_response_packet_bad_size",
    "relay.counter.migrate_response_packet_no_flow_entry",
    "relay.counter.migrate_response_packet_verify_header_failed",
    "relay.counter.migrate_response_packet_already_received",
    "relay.counter.migrate_response_packets_forwarded",
    "relay.counter.migrate_response_packets_received",
    "relay.counter.destroy_packet_bad_size",
    "relay.counter.destroy_packet_no_flow_entry",
    "relay.counter.destroy_packet_verify_header_failed",
    "relay.counter.destroy_packet_already_received",
    "relay.counter.destroy_packets_forwarded",
    "relay.counter.destroy_packets_received",
    "relay.counter.flow_timeout",
    "relay.counter.flow_expire",
    "relay.counter.relay_ping_packets_sent",
    "relay.counter.relay_ping_packet_bad_size",
    "relay.counter.relay_ping_packet_no_ping_target",
    "relay.counter.relay_ping_packet_could_not_read",
    "relay.counter.relay_ping_packet_could_not_write_pong_data",
    "relay.counter.relay_ping_packets_received",
    "relay.stats.measurement_ingress_bytes",
    "relay.counter.relay_pong_packets_sent",
    "relay.counter.relay_pong_packet_bad_size",
    "relay.counter.relay_pong_packet_no_ping_target",
    "relay.counter.relay_pong_packet_could_not_read",
    "relay.counter.relay_pong_packets_received",
    "relay.counter.ping_packet_bad_size",
    "relay.counter.client_ping_packets_received",
    "relay.counter.next_server_ping_packet_bad_size",
    "relay.counter.next_server_ping_packet_no_flow_entry",
    "relay.counter.next_server_ping_packet_verify_header_failed",
    "relay.counter.next_server_ping_packet_already_received",
    "relay.counter.next_server_ping_packets_forwarded",
    "relay.counter.next_server_ping_packets_received",
    "relay.counter.next_server_pong_packet_bad_size",
    "relay.counter.next_server_pong_packet_no_flow_entry",
    "relay.counter.next_server_pong_packet_verify_header_failed",
    "relay.counter.next_server_pong_packet_already_received",
    "relay.counter.next_server_pong_packets_forwarded",
    "relay.counter.next_server_pong_packets_received",
    "relay.counter.update_ping_targets_read_response_json_failed",
    "relay.counter.post_relay_stats_bad_http_status_code",
    "relay.counter.ping_packet_invalid_signature",
    "relay.counter.ping_packet_expired",
    "relay.counter.client.to.server.flow.exceed.bandwidth",
    "relay.counter.server.to.client.flow.exceed.bandwidth",
    "relay.counter.near.exceed.bandwidth",
    "relay.counter.relay.ping.exceed.bandwidth",
    "relay.counter.client.to.server.mgmt.exceed.bandwidth",
    "relay.counter.server.to.client.mgmt.exceed.bandwidth",

    /* add new metrics above here */
    "count.metric.max",
};

ASSERT_ON_COMPILE( ARRAY_SIZE( metric_count_to_string ) == COUNT_METRIC_MAX+1 );

template<typename key_t, typename value_t> using dense_hash_map = google::dense_hash_map<key_t, value_t>;
template<typename T> using queue_t = moodycamel::ConcurrentQueue<T>;
typedef moodycamel::ConsumerToken queue_consumer_token_t;
typedef moodycamel::ProducerToken queue_producer_token_t;

// const uint8_t FLOW_STATE_ACTIVE = 0;
// todo: remove migrated/destroyed state. not used in v3.
const uint8_t FLOW_STATE_MIGRATED = 1;
const uint8_t FLOW_STATE_DESTROYED = 2;

uint64_t next_clean_sequence( uint64_t sequence )
{
    uint64_t mask = ~( (1ULL<<63) | (1ULL<<62) );
    return sequence & mask;
}

struct flow_entry_t
{
    uint64_t expire_timestamp;
    uint32_t state;
    uint32_t kbps_up;
    uint32_t kbps_down;
    next_address_t address_prev;
    next_address_t address_next;
    uint8_t symmetric_key[NEXT_PRIVATE_KEY_BYTES];
    next_flow_bandwidth_t flow_bandwidth_client_to_server;
    next_flow_bandwidth_t flow_bandwidth_server_to_client;
    next_flow_bandwidth_t flow_bandwidth_mgmt_client_to_server;
    next_flow_bandwidth_t flow_bandwidth_mgmt_server_to_client;
    next_replay_protection_t replay_protection_server_to_client_payload;
    next_replay_protection_t replay_protection_client_to_server_payload;
    next_replay_protection_t replay_protection_server_to_client_special;
    next_replay_protection_t replay_protection_client_to_server_special;
};

struct flow_hash_t
{
    uint64_t id;
    uint8_t version;

    flow_hash_t()
        : id(), version()
    {
    }

    flow_hash_t( uint64_t id, uint8_t version )
        : id( id ), version( version )
    {
    }

    bool operator==( const flow_hash_t & other ) const
    {
        return id == other.id && version == other.version;
    }
};

#define LE_UINT64_TO_BYTE(_u, _b)        \
do {                                     \
    ((_b)[0]) = ((uint8_t)(_u));         \
    ((_b)[1]) = ((uint8_t)((_u) >> 8));  \
    ((_b)[2]) = ((uint8_t)((_u) >> 16)); \
    ((_b)[3]) = ((uint8_t)((_u) >> 24)); \
    ((_b)[4]) = ((uint8_t)((_u) >> 32)); \
    ((_b)[5]) = ((uint8_t)((_u) >> 40)); \
    ((_b)[6]) = ((uint8_t)((_u) >> 48)); \
    ((_b)[7]) = ((uint8_t)((_u) >> 56)); \
} while(0)

namespace std
{
    #ifdef __linux__
    namespace tr1 {
    #endif

    template <> struct hash<flow_hash_t>
    {
        std::size_t operator()( const flow_hash_t& h ) const
        {
            int key_len = sizeof( uint64_t ) + sizeof( uint8_t );
            uint8_t key[key_len];
            LE_UINT64_TO_BYTE( h.id, key );
            key[sizeof( uint64_t )] = h.version;

            uint32_t hash = 0;
            for( int i = 0; i < key_len; ++i )
            {
                hash += key[i];
                hash += ( hash << 10 );
                hash ^= ( hash >> 6 );
            }
            hash += ( hash << 3 );
            hash ^= ( hash >> 11 );
            hash += ( hash << 15 );
            return hash;
        }
    };

    template <> struct hash<next_address_t>
    {
        std::size_t operator()( const next_address_t& h ) const
        {
            const int key_len = sizeof( uint8_t ) + sizeof( uint16_t ) + sizeof( uint16_t ) * 8;
            uint8_t key[key_len];
            key[0] = h.type;
            if ( h.type == NEXT_ADDRESS_IPV4 )
            {
                key[1] = h.port;
                key[2] = h.port >> 8;

                memcpy( &key[3], h.data.ipv4, sizeof( h.data.ipv4 ) );
                const int index = 3 + int( sizeof( h.data.ipv4 ) );
                memset( &key[index], 0, key_len - index );
            }
            else if ( h.type == NEXT_ADDRESS_IPV6 )
            {
                key[1] = h.port;
                key[2] = h.port >> 8;

                memcpy( &key[3], h.data.ipv6, sizeof( h.data.ipv6 ) );
            }
            else
            {
                memset( &key[3], 0, key_len - 3 );
            }

            uint32_t hash = 0;
            for( int i = 0; i < key_len; ++i )
            {
                hash += key[i];
                hash += ( hash << 10 );
                hash ^= ( hash >> 6 );
            }
            hash += ( hash << 3 );
            hash ^= ( hash >> 11 );
            hash += ( hash << 15 );
            return hash;
        }
    };

    #ifdef __linux__
    }
    #endif
}

typedef dense_hash_map<flow_hash_t, flow_entry_t> flow_map_t;

enum
{
    // main thread -> flow thread
    MSG_FLOW_QUIT,
    MSG_FLOW_CLEAN_SHUTDOWN,

    // flow thread -> flow thread
    MSG_FLOW_CREATE,
    MSG_FLOW_MIGRATE,           // todo: not used anymore by v3
    MSG_FLOW_DESTROY,           // todo: not used anymore by v3

    // manage thread -> flow thread
    MSG_FLOW_TIMEOUT,           // todo: remove timeout. not used anymore by v3
    MSG_FLOW_EXPIRED,           // todo: this is now the normal way flows are destroyed
};

struct msg_flow
{
    union
    {
        struct
        {
            uint64_t id;
            uint8_t version;
        } flow_timeout_expire;
        struct
        {
            uint64_t id;
            uint8_t version;
        } flow_migrate_destroy;
        struct
        {
            next_flow_token_t token;
            next_address_t address_prev;
        } flow_create;
    };
    uint8_t type;

    msg_flow() {}
};

const int BYTES_V2_CLIENT_PING_PONG = 1 + 8 + 8;

const int BYTES_V3_PING = 1 + 8 + 8 + NEXT_PING_MAC_BYTES + 8;
const int BYTES_V3_PONG = 1 + 8 + 8;

enum
{
    // main thread -> manage thread
    MSG_MANAGE_QUIT,
    MSG_MANAGE_CLEAN_SHUTDOWN,

    // flow thread -> manage thread
    MSG_MANAGE_FLOW_UPDATE,                 // todo: possibly remove? we don't update flows anymore, we continue them.
    MSG_MANAGE_FLOW_MIGRATE_DESTROY,        // todo: remove. not used for anything in v3
    MSG_MANAGE_RELAY_PING_INCOMING,
    MSG_MANAGE_RELAY_PONG_INCOMING,
    MSG_MANAGE_MASTER_PACKET_INCOMING,

    // manage thread -> flow thread
    MSG_MANAGE_RELAY_PING_OUTGOING,
    MSG_MANAGE_RELAY_PONG_OUTGOING,
    MSG_MANAGE_MASTER_PACKET_OUTGOING,
};

struct msg_manage
{
    union
    {
        struct
        {
            uint64_t expire_timestamp;
            uint64_t id;
            uint8_t version;
        } flow_update;

        struct
        {
            uint64_t id;
            uint8_t version;
        } flow_packet;

        struct
        {
            uint64_t id;
            uint8_t version;
        } flow_migrate_destroy;

        struct
        {
            uint64_t id;
            uint64_t sequence;
            next_address_t address;
        } relay_ping_incoming;

        struct
        {
            uint64_t sequence;
            uint8_t token[NEXT_PING_TOKEN_BYTES];
            next_address_t address;
        } relay_ping_outgoing;

        struct
        {
            uint64_t id;
            uint64_t sequence;
            next_address_t address;
        } relay_pong;

        struct
        {
            // IMPORTANT: packet buffer is allocated by flow thread and freed by management thread
            uint8_t * packet_data;
            int packet_bytes;
            next_address_t address;
        } master_packet_incoming;

        struct
        {
            // IMPORTANT: packet buffer is allocated by management thread and freed by flow thread
            uint8_t * packet_data;
            int packet_bytes;
            next_address_t address;
        } master_packet_outgoing;
    };
    uint8_t type;

    msg_manage() {}
};

struct flow_thread_context_t
{
    queue_t<msg_flow> queue;
    next_thread_t thread;
};

struct address_string_t
{
    char address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];

    address_string_t()
    {
        memset( address_string, 0, sizeof( address_string ) );
    }

    address_string_t( const char address[NEXT_MAX_ADDRESS_STRING_LENGTH] )
    {
        strncpy( address_string, address, sizeof( address_string ) );
    }

    bool operator==( const address_string_t & other ) const
    {
        return strncmp ( address_string, other.address_string, NEXT_MAX_ADDRESS_STRING_LENGTH ) == 0;
    }
};

struct bandwidth_insert_t
{
    uint64_t flow_id;
    next_flow_bandwidth_t bandwidth;
};

typedef dense_hash_map<uint64_t, next_flow_bandwidth_t> bandwidth_map_t;

struct manage_thread_context_t
{
    next_thread_t thread;
};

struct global_t
{
    bool dev;
    int cpu_cores[FLOW_THREAD_MAX + 1];
    int cpu_count;
    int bind_port;

    flow_thread_context_t flow_thread_contexts[FLOW_THREAD_MAX];
    int flow_thread_count;

    manage_thread_context_t manage_thread_context;
    queue_t<msg_manage> manage_queue_in;
    queue_t<msg_manage> manage_queue_out;

    std::atomic<uint64_t> bytes_per_sec_paid_tx;
    std::atomic<uint64_t> bytes_per_sec_paid_rx;
    std::atomic<uint64_t> bytes_per_sec_management_tx;
    std::atomic<uint64_t> bytes_per_sec_management_rx;
    std::atomic<uint64_t> bytes_per_sec_measurement_tx;
    std::atomic<uint64_t> bytes_per_sec_measurement_rx;
    std::atomic<uint64_t> bytes_per_sec_invalid_rx;

    std::atomic<uint64_t> master_timestamp; // one second resolution

    uint8_t relay_private_key[NEXT_PRIVATE_KEY_BYTES];
    uint8_t relay_public_key[NEXT_PUBLIC_KEY_BYTES];
    uint8_t relay_ping_key[NEXT_PING_KEY_BYTES];
    uint8_t router_public_key[NEXT_PUBLIC_KEY_BYTES];

    const char * backend_hostname;
};
global_t global;

struct manage_relay_t
{
    uint64_t id;
    uint64_t group_id;
    double speed;
    uint8_t update_key[NEXT_SIGNATURE_KEY_BYTES];
    char name[RELAY_NAME_MAX];
    char group_name[RELAY_NAME_MAX];
    next_address_t address;
};

// todo: we should try to remove the fragmentation support here, if we can. is it really needed? probably not.

#define MASTER_FRAGMENT_SIZE 1024
#define MASTER_FRAGMENT_MAX   255

struct master_fragment_t
{
    uint8_t data[MASTER_FRAGMENT_SIZE];
    uint16_t length;
    bool received;
};

struct master_request_t
{
    uint64_t id;
    master_fragment_t fragments[MASTER_FRAGMENT_MAX];
    uint8_t fragment_total;
    uint8_t type;
};

#define MASTER_TOKEN_BYTES ( NEXT_ADDRESS_BYTES + 32 )

struct master_token_t
{
    next_address_t address;
    uint8_t hmac[32];
};

struct master_init_data_t
{
    uint64_t timestamp; // in nanosecond resolution
    int64_t requested;
    int64_t received;
    master_token_t token;
};

struct manage_peer_history_entry_t
{
    uint64_t sequence;
    int64_t time_ping_sent;
    int64_t time_pong_received;
};

struct manage_peer_history_t
{
    manage_peer_history_entry_t entries_packet_loss[MANAGE_PEER_HISTORY_PACKET_LOSS_ENTRY_COUNT];
    manage_peer_history_entry_t entries[MANAGE_PEER_HISTORY_ENTRY_COUNT];
    int64_t time_sent_of_latest_pong_received;
    uint64_t sequence_current;
    int index_current;
    int index_current_packet_loss_calc;
};

struct manage_peer_t
{
    uint64_t relay_id;
    uint64_t group_id;
    manage_peer_history_t history;
    next_address_t address;
    next_flow_bandwidth_t ping_bandwidth;
    uint8_t ping_token[NEXT_PING_TOKEN_BYTES];
    bool dirty;
};

void manage_peer_init( manage_peer_t * peer )
{
    memset( peer, 0, sizeof( *peer ) );
    next_flow_bandwidth_reset( &peer->ping_bandwidth );
}

typedef dense_hash_map<next_address_t, manage_peer_t> manage_peer_map_t;
typedef dense_hash_map<next_address_t, bool> manage_ping_map_t;

struct manage_environment_t
{
    manage_peer_map_t peers;
    manage_relay_t relay;
    master_request_t master_request;
    master_request_t init_request;
    master_init_data_t init_data;
    resolver_t master;
    next_json_document_t relay_data_json;
    bool valid;
    int idx;
    bool http_success;
};

struct manage_t
{
    manage_environment_t envs[MAX_ENVS];
    int env_count;
};
manage_t manage;
struct metrics_t
{
    std::atomic<std::uint64_t> count_metrics[COUNT_METRIC_MAX];
};
metrics_t metrics;

void relay_printf( int level, const char * format, ... )
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    const char * level_str = next_log_level_str( level );

    time_t ltime = ::time( NULL );
    struct tm local;
    localtime_r( &ltime, &local );
    char timestamp_string[32];
    strftime( timestamp_string, sizeof( timestamp_string ), "%Y-%m-%d %H:%M:%S", &local );

    printf( "%s %s: %s\n", timestamp_string, level_str, buffer );
    va_end( args );
    fflush( stdout );
}

int64_t relay_time()
{
    timespec ts;
    clock_gettime( CLOCK_MONOTONIC_RAW, &ts );
    return ( ts.tv_sec * NEXT_ONE_SECOND_NS ) + ts.tv_nsec;
}

void relay_flow_log( int level, uint64_t flow_id, uint8_t flow_version, const char * format, ... )
{
    va_list args;
    va_start( args, format );
    char buffer[1024];
    vsnprintf( buffer, sizeof( buffer ), format, args );
    va_end( args );
    relay_printf( level, "%" PRIx64 ".%hhu: %s", flow_id, flow_version, buffer );
}

#define metric_count(_metric, _value)                                           \
do                                                                              \
{                                                                               \
    ASSERT_ON_COMPILE( (_metric) < COUNT_METRIC_MAX );                          \
    metrics.count_metrics[(_metric)] += (_value);                               \
} while (0);

#define metric_raw( name, value ) ((void)0)

static void metrics_post()
{
    for ( int i = 0; i < COUNT_METRIC_MAX; ++i )
    {
        uint64_t count = metrics.count_metrics[i];
        if ( count > 0 )
        {
            metrics.count_metrics[i] -= count;
            metric_raw( metric_count_to_string[i], count );
        }
    }
}

flow_entry_t * flow_get( flow_map_t * flows, uint64_t flow_id, uint8_t flow_version )
{
    flow_map_t::iterator i = flows->find( flow_hash_t( flow_id, flow_version ) );
    if ( i == flows->end() )
    {
        return NULL;
    }
    else
    {
        return &i->second;
    }
}

void flow_producer_tokens_init( next_vector_t<queue_producer_token_t> * flow_producer_tokens )
{
    flow_producer_tokens->resize( global.flow_thread_count );
    for ( int i = 0; i < global.flow_thread_count; i++ )
    {
        new ( &(*flow_producer_tokens)[i] ) queue_producer_token_t( global.flow_thread_contexts[i].queue );
    }
}

void flow_producer_tokens_cleanup( next_vector_t<queue_producer_token_t> * flow_producer_tokens )
{
    for ( int i = 0; i < flow_producer_tokens->length; i++ )
    {
        (*flow_producer_tokens)[i].~queue_producer_token_t();
    }
    flow_producer_tokens->length = 0;
}

flow_entry_t * flow_create( flow_map_t * flows, next_flow_token_t * token, next_address_t * from, uint8_t thread_id = 0, next_vector_t<queue_producer_token_t> * flow_producer_tokens = NULL )
{
    flow_hash_t hash( token->flow_id, token->flow_version );
    flow_map_t::iterator i = flows->find( hash );
    if ( i == flows->end() )
    {
        // flow doesn't exist
        // insert flow entry

        flow_entry_t e;
        memset( &e, 0, sizeof( e ) );
        e.kbps_up = uint32_t( token->kbps_up );
        e.kbps_down = uint32_t( token->kbps_down );
        e.address_prev = *from;
        e.address_next = token->next_address;
        memcpy( e.symmetric_key, token->private_key, sizeof( e.symmetric_key ) );

        next_replay_protection_reset( &e.replay_protection_server_to_client_payload );
        next_replay_protection_reset( &e.replay_protection_client_to_server_payload );
        next_replay_protection_reset( &e.replay_protection_server_to_client_special );
        next_replay_protection_reset( &e.replay_protection_client_to_server_special );
        next_flow_bandwidth_reset( &e.flow_bandwidth_client_to_server );
        next_flow_bandwidth_reset( &e.flow_bandwidth_server_to_client );
        next_flow_bandwidth_reset( &e.flow_bandwidth_mgmt_client_to_server );
        next_flow_bandwidth_reset( &e.flow_bandwidth_mgmt_server_to_client );

        auto inserted = flows->insert( flow_map_t::value_type( hash, e ) );
        i = inserted.first;

        if ( flow_producer_tokens )
        {
            // we're the first thread to create this flow
            // notify the other flow threads they need to do the same

            msg_flow msg;
            msg.type = MSG_FLOW_CREATE;
            msg.flow_create.token = *token;
            msg.flow_create.address_prev = *from;

            for ( int i = 0; i < global.flow_thread_count; i++ )
            {
                if ( i != thread_id )
                {
                    global.flow_thread_contexts[i].queue.enqueue( (*flow_producer_tokens)[i], msg );
                }
            }
        }
    }
    return &i->second;
}

enum
{
    FLOW_DIRECTION_SERVER_TO_CLIENT,
    FLOW_DIRECTION_CLIENT_TO_SERVER,
};

flow_entry_t * flow_verify( flow_map_t * flows, uint8_t * packet_data, int packet_bytes, int direction, uint64_t * out_flow_id = NULL, uint8_t * out_flow_version = NULL )
{
    uint8_t type;
    uint64_t sequence;
    uint64_t flow_id;
    uint8_t flow_version;
    uint8_t flow_flags;
    if ( next_peek_header( &type, &sequence, &flow_id, &flow_version, &flow_flags, packet_data, packet_bytes ) != NEXT_OK )
    {
        return NULL;
    }

    if ( out_flow_id )
    {
        *out_flow_id = flow_id;
    }
    if ( out_flow_version )
    {
        *out_flow_version = flow_version;
    }

    flow_entry_t * flow = flow_get( flows, flow_id, flow_version );

    if ( !flow )
    {
        switch ( packet_data[0] )
        {
            case NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE:
            {
                metric_count( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE:
            {
                metric_count( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER:
            case NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER:
            {
                metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT:
            case NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT:
            {
                metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING:
            case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING:
            {
                metric_count( RELAY_COUNTER_NEXT_SERVER_PING_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG:
            case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG:
            {
                metric_count( RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_MIGRATE:
            case NEXT_PACKET_TYPE_V3_MIGRATE:
            {
                metric_count( RELAY_COUNTER_MIGRATE_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_DESTROY:
            case NEXT_PACKET_TYPE_V3_DESTROY:
            {
                metric_count( RELAY_COUNTER_DESTROY_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE:
            {
                metric_count( RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_NO_FLOW_ENTRY, 1 );
                break;
            }
            default:
            {
                break;
            }
        }
        metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
        return NULL;
    }

    next_replay_protection_t * replay_protection;
    if ( packet_data[0] == NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER || packet_data[0] == NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT )
    {
        if ( direction == FLOW_DIRECTION_SERVER_TO_CLIENT )
        {
            replay_protection = &flow->replay_protection_server_to_client_payload;
        }
        else
        {
            replay_protection = &flow->replay_protection_client_to_server_payload;
        }
    }
    else
    {
        if ( direction == FLOW_DIRECTION_SERVER_TO_CLIENT )
        {
            replay_protection = &flow->replay_protection_server_to_client_special;
        }
        else
        {
            replay_protection = &flow->replay_protection_client_to_server_special;
        }
    }

    uint64_t clean_sequence = next_clean_sequence( sequence );

    if ( next_replay_protection_already_received( replay_protection, clean_sequence ) )
    {
        switch ( packet_data[0] )
        {
            case NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE:
            {
            	metric_count( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE:
            {
            	metric_count( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER:
            case NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER:
            {
            	metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT:
            case NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT:
            {
            	metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING:
            case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING:
            {
            	metric_count( RELAY_COUNTER_NEXT_SERVER_PING_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG:
            case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG:
            {
            	metric_count( RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_MIGRATE:
            case NEXT_PACKET_TYPE_V3_MIGRATE:
            {
            	metric_count( RELAY_COUNTER_MIGRATE_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_DESTROY:
            case NEXT_PACKET_TYPE_V3_DESTROY:
            {
            	metric_count( RELAY_COUNTER_DESTROY_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE:
            {
            	metric_count( RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_ALREADY_RECEIVED, 1 );
                break;
            }
            default:
            {
                break;
            }
        }
        metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
        return NULL;
    }

    if ( next_read_header( &type, &sequence, &flow_id, &flow_version, &flow_flags, flow->symmetric_key, packet_data, packet_bytes ) != NEXT_OK )
    {
        switch ( packet_data[0] )
        {
            case NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE:
            {
            	metric_count( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE:
            {
            	metric_count( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER:
            case NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER:
            {
            	metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT:
            case NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT:
            {
            	metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING:
            case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING:
            {
            	metric_count( RELAY_COUNTER_NEXT_SERVER_PING_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG:
            case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG:
            {
	            metric_count( RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_MIGRATE:
            case NEXT_PACKET_TYPE_V3_MIGRATE:
            {
            	metric_count( RELAY_COUNTER_MIGRATE_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_DESTROY:
            case NEXT_PACKET_TYPE_V3_DESTROY:
            {
            	metric_count( RELAY_COUNTER_DESTROY_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            case NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE:
            case NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE:
            {
            	metric_count( RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_VERIFY_HEADER_FAILED, 1 );
                break;
            }
            default:
            {
                break;
            }
        }
        metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
        return NULL;
    }

    next_replay_protection_advance_sequence( replay_protection, clean_sequence );

    // int64_t time = relay_time();

    switch ( packet_data[0] )
    {
        case NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER:
        case NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER:
        case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING:
        case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING:
        {
            /*
            if ( next_flow_bandwidth_over_budget ( &flow->flow_bandwidth_client_to_server, time, NEXT_ONE_SECOND_NS, flow->kbps_up + NEXT_SERVER_PING_PONG_KBPS, packet_bytes ) )
            {
                metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_FLOW_EXCEED_BANDWIDTH, 1 );
                return NULL;
            }
            */
        }
        break;

        case NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT:
        case NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT:
        case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG:
        case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG:
        {
            /*
            if ( next_flow_bandwidth_over_budget ( &flow->flow_bandwidth_server_to_client, time, NEXT_ONE_SECOND_NS, flow->kbps_down + NEXT_SERVER_PING_PONG_KBPS, packet_bytes ) )
            {
                metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_FLOW_EXCEED_BANDWIDTH, 1 );
                return NULL;
            }
            */
        }
        break;

        case NEXT_PACKET_TYPE_V2_CONTINUE_REQUEST:
        case NEXT_PACKET_TYPE_V2_MIGRATE:
        case NEXT_PACKET_TYPE_V3_MIGRATE:
        case NEXT_PACKET_TYPE_V2_DESTROY:
        case NEXT_PACKET_TYPE_V3_DESTROY:
        {
            /*
            if ( next_flow_bandwidth_over_budget ( &flow->flow_bandwidth_mgmt_client_to_server, time, NEXT_ONE_SECOND_NS, NEXT_MANAGEMENT_UP_KBPS, packet_bytes ) )
            {
                metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_MGMT_EXCEED_BANDWIDTH, 1 );
                return NULL;
            }
            */
        }
        break;

        case NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE:
        case NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE:
        case NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE:
        case NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE:
        case NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE:
        case NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE:
        {
            /*
            if ( next_flow_bandwidth_over_budget ( &flow->flow_bandwidth_mgmt_server_to_client, time, NEXT_ONE_SECOND_NS, NEXT_MANAGEMENT_DOWN_KBPS, packet_bytes ) )
            {
                metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_MGMT_EXCEED_BANDWIDTH, 1 );
                return NULL;
            }
            */
        }
        break;

        default:
            break;
    }

    return flow;
}

int relay_receive_packet( next_socket_t * socket, next_address_t * from, uint8_t * packet_data, int max_packet_size )
{
    int packet_bytes = next_socket_receive_packet( socket, from, packet_data, max_packet_size );
    return packet_bytes;
}

void relay_send_packet( next_socket_t * socket, next_address_t * to, uint8_t * packet_data, int packet_bytes )
{
    uint8_t packet_type = packet_data[0];
    if ( packet_type == NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER
        || packet_type == NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER
        || packet_type == NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT
        || packet_type == NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT )
    {
        metric_count( RELAY_STATS_PAID_EGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_paid_tx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
    }
    else if ( packet_type == NEXT_PACKET_TYPE_V2_ROUTE_REQUEST
        || packet_type == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE
        || packet_type == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE
        || packet_type == NEXT_PACKET_TYPE_V2_CONTINUE_REQUEST
        || packet_type == NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE
        || packet_type == NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE
        // todo: remove migrate, migrate response and destroy packet processing. not used in v3.
        || packet_type == NEXT_PACKET_TYPE_V2_MIGRATE
        || packet_type == NEXT_PACKET_TYPE_V3_MIGRATE
        || packet_type == NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE
        || packet_type == NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE
        || packet_type == NEXT_PACKET_TYPE_V2_DESTROY
        || packet_type == NEXT_PACKET_TYPE_V3_DESTROY )
    {
        metric_count( RELAY_STATS_MANAGEMENT_EGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_management_tx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
    }
    else
    {
        metric_count( RELAY_STATS_MEASUREMENT_EGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_measurement_tx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
    }

    next_socket_send_packet( socket, to, packet_data, packet_bytes );
}

void flow_relay_ping_send( next_socket_t * socket, next_address_t * address, uint8_t * ping_token, uint64_t sequence )
{
    uint8_t packet_data[BYTES_V3_PING];
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_PACKET_TYPE_V3_RELAY_PING );
    next_write_bytes( &p, ping_token, NEXT_PING_TOKEN_BYTES );
    next_write_uint64( &p, sequence );
    relay_send_packet( socket, address, packet_data, sizeof( packet_data ) );
}

void flow_relay_pong_send( next_socket_t * socket, next_address_t * address, uint64_t flow_id, uint64_t sequence )
{
    uint8_t packet_data[BYTES_V3_PONG];
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_PACKET_TYPE_V3_RELAY_PONG );
    next_write_uint64( &p, flow_id );
    next_write_uint64( &p, sequence );
    relay_send_packet( socket, address, packet_data, sizeof( packet_data ) );
}

// todo: remove flow timeout. no longer needed in v3.
void flow_timeout_update( queue_producer_token_t * manage_producer_token, uint64_t flow_id, uint8_t flow_version, uint64_t expire_timestamp )
{
    msg_manage msg;
    msg.type = MSG_MANAGE_FLOW_UPDATE;
    msg.flow_update.id = flow_id;
    msg.flow_update.version = flow_version;
    msg.flow_update.expire_timestamp = expire_timestamp;
    global.manage_queue_out.enqueue( *manage_producer_token, msg );
}

uint8_t flow_thread_id( flow_thread_context_t * context )
{
    return uint8_t( context - &global.flow_thread_contexts[0] );
}

enum
{
    FLOW_THREAD_STATE_STOP,
    FLOW_THREAD_STATE_CLEAN_SHUTDOWN,
    FLOW_THREAD_STATE_RUN,
};

// todo: remove migrate/destroy concept

// first flow thread to hear about this flow getting migrated/destroyed calls this
void flow_migrate_destroy(
    flow_entry_t * flow,
    int state,
    uint64_t flow_id,
    uint8_t flow_version,
    queue_producer_token_t * manage_producer_token,
    uint8_t thread_id,
    next_vector_t<queue_producer_token_t> * flow_producer_tokens )
{
    if ( flow->state != uint32_t(state) )
    {
        flow->state = state;

        // notify the manage thread
        {
            msg_manage msg;
            msg.type = MSG_MANAGE_FLOW_MIGRATE_DESTROY;
            msg.flow_migrate_destroy.id = flow_id;
            msg.flow_migrate_destroy.version = flow_version;
            global.manage_queue_in.enqueue( *manage_producer_token, msg );
        }

        // notify the other flow threads
        {
            msg_flow msg;

            switch ( state )
            {
                case FLOW_STATE_DESTROYED:
                {
                    msg.type = MSG_FLOW_DESTROY;
                    break;
                }
                case FLOW_STATE_MIGRATED:
                {
                    msg.type = MSG_FLOW_MIGRATE;
                    break;
                }
                default:
                {
                    next_assert( false );
                    break;
                }
            }

            msg.flow_migrate_destroy.id = flow_id;
            msg.flow_migrate_destroy.version = flow_version;

            for ( int i = 0; i < global.flow_thread_count; i++ )
            {
                if ( i != thread_id )
                {
                    global.flow_thread_contexts[i].queue.enqueue( (*flow_producer_tokens)[i], msg );
                }
            }
        }
    }
}

// other flow threads who receive the migrate/destroy message call this function
void flow_migrate_destroy( flow_entry_t * flow, int state )
{
    flow->state = state;
}

// todo: remove old remove the old pings. not used in v3.
int flow_ping_read( uint8_t * packet_data, int packet_bytes, uint64_t timestamp, uint64_t * id, uint64_t * sequence )
{
    if ( packet_bytes != BYTES_V3_PING )
    {
        metric_count( RELAY_COUNTER_PING_PACKET_BAD_SIZE, 1 );
        metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
        return NEXT_ERROR;
    }

    uint8_t * p = &packet_data[1];
    uint64_t expire_timestamp = next_read_uint64( &p );

    if ( timestamp > expire_timestamp )
    {
        metric_count( RELAY_COUNTER_PING_PACKET_EXPIRED, 1 );
        metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
        global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
        return NEXT_ERROR;
    }

    *id = next_read_uint64( &p );
    relay_printf( NEXT_LOG_LEVEL_DEBUG, "ping received, id = %lu", *id );
    uint8_t * ping_mac = p;

    // if ( crypto_auth_verify( ping_mac, &packet_data[1], 8 + 8, global.relay_ping_key ) != 0 )
    // {
    //   std::cout << "failed to verify ping packet";
    //     metric_count( RELAY_COUNTER_PING_PACKET_INVALID_SIGNATURE, 1 );
    //     metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
    //     global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
    //     return NEXT_ERROR;
    // }

    p += NEXT_PING_MAC_BYTES;
    *sequence = next_read_uint64( &p );

    return NEXT_OK;
}

int flow_near_bandwidth_check( bandwidth_map_t * near_bandwidth, next_vector_t<bandwidth_insert_t> * near_bandwidth_inserts, uint64_t flow_id, int64_t time, int packet_bytes )
{
    (void) time;
    (void) packet_bytes;

    bandwidth_map_t::iterator i = near_bandwidth->find( flow_id );

    // next_flow_bandwidth_t * bandwidth;

    if ( i == near_bandwidth->end() )
    {
        // bandwidth entry doesn't exist
        // insert it

        bandwidth_insert_t * insert = near_bandwidth_inserts->add();
        insert->flow_id = flow_id;
        next_flow_bandwidth_reset( &insert->bandwidth );
        // bandwidth = &insert->bandwidth;
    }
    /*
    else
    {
        bandwidth = &i->second;
    }
    */

    /*
    if ( next_flow_bandwidth_over_budget( bandwidth, time, NEXT_ONE_SECOND_NS, NEXT_CLIENT_RELAY_PING_KBPS, packet_bytes ) )
    {
        metric_count( RELAY_COUNTER_NEAR_EXCEED_BANDWIDTH, 1 );
        return NEXT_ERROR;
    }
    */

    return NEXT_OK;
}

next_thread_return_t NEXT_THREAD_FUNC flow_thread( void * param )
{
    flow_thread_context_t * context = (flow_thread_context_t *)( param );

    uint8_t thread_id = flow_thread_id( context );

    next_socket_t socket;
    {
        next_address_t bind_address;
        memset( &bind_address, 0, sizeof(bind_address) );
        bind_address.type = NEXT_ADDRESS_IPV4;
        bind_address.port = global.bind_port;
        if ( next_socket_create( &socket, &bind_address, 0, NEXT_SOCKET_SNDBUF_SIZE, NEXT_SOCKET_RCVBUF_SIZE, NEXT_SOCKET_FLAG_REUSEPORT ) != NEXT_SOCKET_ERROR_NONE )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to bind to socket" );
            exit( 1 );
        }
    }

    bandwidth_map_t near_bandwidth;
    near_bandwidth.set_empty_key( 0 );
    near_bandwidth.set_deleted_key( uint64_t( -1 ) );
    bandwidth_map_t::iterator near_bandwidth_iterator = near_bandwidth.end();

    next_vector_t<bandwidth_insert_t> near_bandwidth_inserts;

    flow_map_t flows;
    flows.set_empty_key( flow_hash_t( 0, 0 ) );
    flows.set_deleted_key( flow_hash_t( uint64_t( -1 ), uint8_t( -1 ) ) );

    queue_consumer_token_t flow_consumer_token( context->queue );
    queue_consumer_token_t manage_consumer_token( global.manage_queue_out );
    queue_producer_token_t manage_producer_token( global.manage_queue_in );

    next_vector_t<queue_producer_token_t> flow_producer_tokens;
    flow_producer_tokens_init( &flow_producer_tokens );

    int state = FLOW_THREAD_STATE_RUN;
    while ( true )
    {
        if ( state == FLOW_THREAD_STATE_STOP || ( state == FLOW_THREAD_STATE_CLEAN_SHUTDOWN && flows.size() == 0 ) )
            break;

        int64_t time = relay_time();
        uint64_t timestamp = global.master_timestamp;

        // read packets
        {
            int packet_count = 0;

            next_address_t from;
            uint8_t packet_data[NEXT_MTU];
            int packet_bytes;
            const int PACKET_BULK = 32;
            while ( packet_count < PACKET_BULK && ( packet_bytes = relay_receive_packet( &socket, &from, packet_data, sizeof( packet_data ) ) ) > 0 )
            {
                switch ( packet_data[0] )
                {
                    // todo: this is the only client <-> relay ping actually used by v3
                    case NEXT_PACKET_TYPE_V3_SDK_CLIENT_RELAY_PING:
                    {
                        if ( packet_bytes != 1 + 8 + 8 + 8 + 8 )
                        {
                            metric_count( RELAY_COUNTER_PING_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        packet_data[0] = NEXT_PACKET_TYPE_V3_SDK_CLIENT_RELAY_PONG;
                        relay_send_packet( &socket, &from, packet_data, packet_bytes - 16 );
                        metric_count( RELAY_COUNTER_CLIENT_PING_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, packet_bytes );
                        global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                    }
                    break;

                    // todo: remove. this is not used by v3.
                    case NEXT_PACKET_TYPE_V2_CLIENT_RELAY_PING:
                    {
                        if ( packet_bytes != BYTES_V2_CLIENT_PING_PONG )
                        {
                            metric_count( RELAY_COUNTER_PING_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        packet_data[0] = NEXT_PACKET_TYPE_V2_CLIENT_RELAY_PONG;
                        relay_send_packet( &socket, &from, packet_data, packet_bytes );
                        metric_count( RELAY_COUNTER_CLIENT_PING_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, packet_bytes );
                        global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                    }
                    break;

                    case NEXT_PACKET_TYPE_V3_RELAY_PING:
                    {
                        uint64_t relay_id;
                        uint64_t sequence;
                        if ( flow_ping_read( packet_data, packet_bytes, timestamp, &relay_id, &sequence ) == NEXT_OK )
                        {
                            msg_manage msg_out;
                            msg_out.type = MSG_MANAGE_RELAY_PING_INCOMING;
                            msg_out.relay_ping_incoming.address = from;
                            msg_out.relay_ping_incoming.sequence = sequence;
                            msg_out.relay_ping_incoming.id = relay_id;
                            global.manage_queue_in.enqueue( manage_producer_token, msg_out );
                        }
                        else
                        {
                            metric_count( RELAY_COUNTER_RELAY_PING_PACKET_COULD_NOT_READ, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }
                    }
                    break;

                    case NEXT_PACKET_TYPE_V3_RELAY_PONG:
                    {
                        if ( packet_bytes != BYTES_V3_PONG )
                        {
                            metric_count( RELAY_COUNTER_RELAY_PONG_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        uint8_t * p = &packet_data[1];

                        msg_manage msg_out;
                        msg_out.type = MSG_MANAGE_RELAY_PONG_INCOMING;
                        msg_out.relay_pong.address = from;
                        msg_out.relay_pong.id = next_read_uint64( &p );
                        msg_out.relay_pong.sequence = next_read_uint64( &p );
                        global.manage_queue_in.enqueue( manage_producer_token, msg_out );
                    }
                    break;

                    // todo: remove. this is not used by v3.
                    case NEXT_PACKET_TYPE_V3_CLIENT_RELAY_PING:
                    {
                        uint64_t flow_id;
                        uint64_t sequence;
                        if ( flow_ping_read( packet_data, packet_bytes, timestamp, &flow_id, &sequence ) != NEXT_OK )
                            break;

                        metric_count( RELAY_COUNTER_CLIENT_PING_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, BYTES_V3_PONG );
                        global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;

                        if ( flow_near_bandwidth_check( &near_bandwidth, &near_bandwidth_inserts, flow_id, time, packet_bytes ) != NEXT_OK )
                            break;

                        uint8_t * p = packet_data;
                        next_write_uint8( &p, NEXT_PACKET_TYPE_V3_CLIENT_RELAY_PONG );
                        next_write_uint64( &p, flow_id );
                        next_write_uint64( &p, sequence );
                        next_assert( p - packet_data == BYTES_V3_PONG );
                        relay_send_packet( &socket, &from, packet_data, BYTES_V3_PONG );
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_ROUTE_REQUEST:
                    {
                        if ( packet_bytes < int( 1 + NEXT_ENCRYPTED_FLOW_TOKEN_BYTES * 2 ) )
                        {
                            relay_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring route request. bad packet size (%d)", packet_bytes );
                            break;
                        }

                        uint8_t * p = &packet_data[1];

                        next_flow_token_t token;
                        if ( next_read_encrypted_flow_token( &p, &token, NEXT_KEY_MASTER, global.relay_private_key ) != NEXT_OK )
                        {
                            relay_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring route request. could not read flow token" );
                            metric_count( RELAY_COUNTER_ROUTE_REQUEST_PACKET_DECRYPT_TOKEN_FAILED, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        if ( token.expire_timestamp < timestamp )
                        {
                            relay_flow_log( NEXT_LOG_LEVEL_DEBUG, token.flow_id, token.flow_version, "ignoring route request. expired (%lu vs. %lu)", token.expire_timestamp, timestamp );
                            metric_count( RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        flow_entry_t * flow = flow_create( &flows, &token, &from, thread_id, &flow_producer_tokens );

                        /*
                        if ( next_flow_bandwidth_over_budget ( &flow->flow_bandwidth_mgmt_client_to_server, time, NEXT_ONE_SECOND_NS, NEXT_MANAGEMENT_UP_KBPS, packet_bytes ) )
                        {
                            relay_flow_log( NEXT_LOG_LEVEL_DEBUG, token.flow_id, token.flow_version, "ignoring route request. exeeded rate limit" );
                            metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_MGMT_EXCEED_BANDWIDTH, 1 );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }
                        */

                        flow_timeout_update( &manage_producer_token, token.flow_id, token.flow_version, token.expire_timestamp );

                        // todo: remove concept of flow flags entirely.
                        token.flow_flags |= NEXT_FLOW_FLAG_FLOW_CREATE;

                        metric_count( RELAY_COUNTER_FLOW_CREATE, 1 );

                        relay_flow_log( NEXT_LOG_LEVEL_INFO, token.flow_id, token.flow_version, "flow created" );

                        packet_data[NEXT_ENCRYPTED_FLOW_TOKEN_BYTES] = NEXT_PACKET_TYPE_V2_ROUTE_REQUEST;
                        relay_send_packet( &socket, &flow->address_next, &packet_data[NEXT_ENCRYPTED_FLOW_TOKEN_BYTES], packet_bytes - NEXT_ENCRYPTED_FLOW_TOKEN_BYTES );

                        metric_count( RELAY_COUNTER_ROUTE_REQUEST_PACKETS_FORWARDED, 1 );
                        metric_count( RELAY_STATS_MANAGEMENT_EGRESS_BYTES, packet_bytes );

                        metric_count( RELAY_COUNTER_ROUTE_REQUEST_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MANAGEMENT_INGRESS_BYTES, packet_bytes );

                        global.bytes_per_sec_management_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_CONTINUE_REQUEST:
                    {
                        if ( packet_bytes < int( 1 + NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES * 2 ) )
                        {
                            relay_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring continue request. bad packet size: %d", packet_bytes );
                            metric_count( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        uint8_t * p = &packet_data[1];

                        next_continue_token_t token;
                        if ( next_read_encrypted_continue_token( &p, &token, NEXT_KEY_MASTER, global.relay_private_key ) != NEXT_OK )
                        {
                            relay_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring continue request. could not decrypt continue token" );
                            metric_count( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_DECRYPT_TOKEN_FAILED, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        if ( token.expire_timestamp < timestamp )
                        {
                            relay_flow_log( NEXT_LOG_LEVEL_DEBUG, token.flow_id, token.flow_version, "ignoring continue request. expired (%lu vs. %lu)", token.expire_timestamp, timestamp );
                            metric_count( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        flow_entry_t * flow = flow_get( &flows, token.flow_id, token.flow_version );
                        if ( !flow )
                        {
                            relay_flow_log( NEXT_LOG_LEVEL_DEBUG, token.flow_id, token.flow_version, "ignoring continue request. could not find flow" );
                            metric_count( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_NO_FLOW_ENTRY, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }
                        flow_timeout_update( &manage_producer_token, token.flow_id, token.flow_version, token.expire_timestamp );

                        packet_data[NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES] = NEXT_PACKET_TYPE_V2_CONTINUE_REQUEST;
                        relay_send_packet( &socket, &flow->address_next, &packet_data[NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES], packet_bytes - NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES );
                        metric_count( RELAY_COUNTER_CONTINUE_REQUEST_PACKETS_FORWARDED, 1 );
                        metric_count( RELAY_STATS_MANAGEMENT_EGRESS_BYTES, packet_bytes );

                        relay_flow_log( NEXT_LOG_LEVEL_INFO, token.flow_id, token.flow_version, "flow continued" );

                        metric_count( RELAY_COUNTER_FLOW_CONTINUE, 1 );
                        metric_count( RELAY_COUNTER_CONTINUE_REQUEST_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MANAGEMENT_INGRESS_BYTES, packet_bytes );
                        global.bytes_per_sec_management_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE:
                    case NEXT_PACKET_TYPE_V2_CONTINUE_RESPONSE:
                    // todo: remove the v3 route response. paradoxically, not used by v3.
                    case NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE:
                    case NEXT_PACKET_TYPE_V3_CONTINUE_RESPONSE:
                    {
                        if ( packet_bytes != NEXT_HEADER_BYTES && packet_bytes != NEXT_HEADER_BYTES + NEXT_ENCRYPTED_SERVER_TOKEN_BYTES )
                        {
                            relay_printf( NEXT_LOG_LEVEL_DEBUG, "ignoring %s response. bad packet size %d", ( packet_data[0] == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE || packet_data[0] == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE ) ? "route" : "continue", packet_bytes );
                            if ( packet_data[0] == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE
                                || packet_data[0] == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE )
                            {
                                metric_count( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_BAD_SIZE, 1 );
                            }
                            else
                            {
                                metric_count( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_BAD_SIZE, 1 );
                            }
                            break;
                        }

                        uint64_t flow_id;
                        uint8_t flow_version;
                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_SERVER_TO_CLIENT, &flow_id, &flow_version );
                        if ( !flow )
                        {
                            relay_flow_log( NEXT_LOG_LEVEL_DEBUG, flow_id, flow_version, "ignoring %s response. could not find flow", ( packet_data[0] == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE || packet_data[0] == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE ) ? "route" : "continue" );
                            break;
                        }

                        relay_send_packet( &socket, &flow->address_prev, packet_data, packet_bytes );
                        if ( packet_data[0] == NEXT_PACKET_TYPE_V2_ROUTE_RESPONSE
                            || packet_data[0] == NEXT_PACKET_TYPE_V3_ROUTE_RESPONSE )
                        {
                            metric_count( RELAY_COUNTER_ROUTE_RESPONSE_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_ROUTE_RESPONSE_PACKETS_RECEIVED, 1 );
                        }
                        else
                        {
                            metric_count( RELAY_COUNTER_CONTINUE_RESPONSE_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_CONTINUE_RESPONSE_PACKETS_RECEIVED, 1 );
                        }
                        metric_count( RELAY_STATS_MANAGEMENT_INGRESS_BYTES, packet_bytes );
                        global.bytes_per_sec_management_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_CLIENT_TO_SERVER:
                    // todo: remove v3 packet. not actually used by v3 SDK
                    case NEXT_PACKET_TYPE_V3_CLIENT_TO_SERVER:
                    {
                        if ( packet_bytes <= NEXT_HEADER_BYTES )
                        {
                            metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }
                        uint64_t flow_id;
                        uint8_t flow_version;
                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_CLIENT_TO_SERVER, &flow_id, &flow_version );
                        if ( flow )
                        {
                            relay_send_packet( &socket, &flow->address_next, packet_data, packet_bytes );
                            metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_CLIENT_TO_SERVER_PACKETS_RECEIVED, 1 );
                            metric_count( RELAY_STATS_PAID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_paid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                        }
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_SERVER_TO_CLIENT:
                    // todo: remove v3 packet. not actually used by v3.
                    case NEXT_PACKET_TYPE_V3_SERVER_TO_CLIENT:
                    {
                        if ( packet_bytes <= NEXT_HEADER_BYTES )
                        {
                            metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }
                        uint64_t flow_id;
                        uint8_t flow_version;
                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_SERVER_TO_CLIENT, &flow_id, &flow_version );
                        if ( flow )
                        {
                            relay_send_packet( &socket, &flow->address_prev, packet_data, packet_bytes );
                            metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_SERVER_TO_CLIENT_PACKETS_RECEIVED, 1 );
                            metric_count( RELAY_STATS_PAID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_paid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                        }
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PING:
                    // todo: remove v3 next server ping
                    case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PING:
                    {
                        uint64_t flow_id;
                        uint8_t flow_version;

                        if ( packet_bytes != NEXT_HEADER_BYTES + 8 &&
                             packet_bytes != NEXT_HEADER_BYTES + NEXT_PACKET_V2_PING_PONG_BYTES )   // todo: bug in v2 SDK, preserve this until v2 goes away
                        {
                            metric_count( RELAY_COUNTER_NEXT_SERVER_PING_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_CLIENT_TO_SERVER, &flow_id, &flow_version );
                        if ( flow )
                        {
                            relay_send_packet( &socket, &flow->address_next, packet_data, packet_bytes );
                            metric_count( RELAY_COUNTER_NEXT_SERVER_PING_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_NEXT_SERVER_PING_PACKETS_RECEIVED, 1 );
                            metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                        }
                    }
                    break;

                    case NEXT_PACKET_TYPE_V2_NEXT_SERVER_PONG:
                    // todo: remove v3 next server pong.
                    case NEXT_PACKET_TYPE_V3_NEXT_SERVER_PONG:
                    {
                        uint64_t flow_id;
                        uint8_t flow_version;

                        if ( packet_bytes != NEXT_HEADER_BYTES + 8 &&
                             packet_bytes != NEXT_HEADER_BYTES + NEXT_PACKET_V2_PING_PONG_BYTES )   // todo: bug in v2 SDK, preserve this until v2 goes away
                        {
                            metric_count( RELAY_COUNTER_NEXT_SERVER_PONG_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_SERVER_TO_CLIENT, &flow_id, &flow_version);
                        if ( flow )
                        {
                            relay_send_packet( &socket, &flow->address_prev, packet_data, packet_bytes );
                            metric_count( RELAY_COUNTER_NEXT_SERVER_PONG_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_NEXT_SERVER_PONG_PACKETS_RECEIVED, 1 );
                            metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                        }
                    }
                    break;

                    // todo: remove when v2 goes away
                    case NEXT_PACKET_TYPE_V2_MIGRATE:
                    case NEXT_PACKET_TYPE_V3_MIGRATE:
                    case NEXT_PACKET_TYPE_V2_DESTROY:
                    case NEXT_PACKET_TYPE_V3_DESTROY:
                    {
                        if ( packet_bytes != NEXT_HEADER_BYTES )
                        {
                            if ( packet_data[0] == NEXT_PACKET_TYPE_V2_MIGRATE
                                || packet_data[0] == NEXT_PACKET_TYPE_V3_MIGRATE )
                            {
                                metric_count( RELAY_COUNTER_MIGRATE_PACKET_BAD_SIZE, 1 );
                                metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            }
                            else
                            {
                                metric_count( RELAY_COUNTER_DESTROY_PACKET_BAD_SIZE, 1 );
                                metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            }
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        uint64_t flow_id;
                        uint8_t flow_version;
                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_CLIENT_TO_SERVER, &flow_id, &flow_version );
                        if ( flow )
                        {
                            if ( packet_data[0] == NEXT_PACKET_TYPE_V2_MIGRATE
                                || packet_data[0] == NEXT_PACKET_TYPE_V3_MIGRATE )
                            {
                                flow_migrate_destroy( flow, FLOW_STATE_MIGRATED, flow_id, flow_version, &manage_producer_token, thread_id, &flow_producer_tokens );
                                metric_count( RELAY_COUNTER_MIGRATE_PACKETS_RECEIVED, 1 );
                                metric_count( RELAY_COUNTER_MIGRATE_PACKETS_FORWARDED, 1 );
                            }
                            else
                            {
                                flow_migrate_destroy( flow, FLOW_STATE_DESTROYED, flow_id, flow_version, &manage_producer_token, thread_id, &flow_producer_tokens );
                                metric_count( RELAY_COUNTER_DESTROY_PACKETS_RECEIVED, 1 );
                                metric_count( RELAY_COUNTER_DESTROY_PACKETS_FORWARDED, 1 );
                            }
                            metric_count( RELAY_STATS_MANAGEMENT_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_management_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            relay_send_packet( &socket, &flow->address_next, packet_data, packet_bytes );
                        }
                    }
                    break;

                    // todo: remove when v2 goes away
                    case NEXT_PACKET_TYPE_V2_MIGRATE_RESPONSE:
                    case NEXT_PACKET_TYPE_V3_MIGRATE_RESPONSE:
                    {
                        if ( packet_bytes != NEXT_HEADER_BYTES )
                        {
                            metric_count( RELAY_COUNTER_MIGRATE_RESPONSE_PACKET_BAD_SIZE, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                            break;
                        }

                        flow_entry_t * flow = flow_verify( &flows, packet_data, packet_bytes, FLOW_DIRECTION_SERVER_TO_CLIENT );
                        if ( flow )
                        {
                            relay_send_packet( &socket, &flow->address_prev, packet_data, packet_bytes );

                            metric_count( RELAY_COUNTER_MIGRATE_RESPONSE_PACKETS_FORWARDED, 1 );
                            metric_count( RELAY_COUNTER_MIGRATE_RESPONSE_PACKETS_RECEIVED, 1 );
                            metric_count( RELAY_STATS_MANAGEMENT_INGRESS_BYTES, packet_bytes );
                            global.bytes_per_sec_management_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                        }
                    }
                    break;

                    case NEXT_PACKET_TYPE_V4_MASTER_RELAY_CONFIG_RESPONSE:
                    case NEXT_PACKET_TYPE_V4_MASTER_RELAY_RESPONSE:
                    case NEXT_PACKET_TYPE_V4_MASTER_INIT_RESPONSE:
                    {
                        // todo: validate master packet before passing it to management thread
                        if ( packet_bytes < 1 || packet_bytes > NEXT_MTU + 64 )
                            break;

                        msg_manage msg;
                        msg.type = MSG_MANAGE_MASTER_PACKET_INCOMING;

                        // IMPORTANT: packet buffer is allocated by flow thread and freed by management thread
                        uint8_t * buffer = (uint8_t *)( next_alloc( packet_bytes ) );

                        memcpy( buffer, packet_data, size_t( packet_bytes ) );
                        msg.master_packet_incoming.packet_data = buffer;
                        msg.master_packet_incoming.packet_bytes = packet_bytes;
                        msg.master_packet_incoming.address = from;
                        global.manage_queue_in.enqueue( manage_producer_token, msg );
                        break;
                    }

                    default:
                    {
                        metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, packet_bytes );
                        global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + packet_bytes;
                    }
                    break;

                }
                packet_count++;
            }
        }

        // timeout near bandwidth counters
        {
            // scan counters
            if ( near_bandwidth_iterator != near_bandwidth.end() )
            {
                const int BANDWIDTH_SCAN_BULK = 64; // only scan N bandwidth entries at a time
                for ( int i = 0; i < BANDWIDTH_SCAN_BULK; i++ )
                {
                    next_flow_bandwidth_t * bandwidth = &near_bandwidth_iterator->second;
                    if ( time - bandwidth->last_bandwidth_check_timestamp > 10 * NEXT_ONE_SECOND_NS )
                    {
                        near_bandwidth.erase( near_bandwidth_iterator );
                    }

                    near_bandwidth_iterator++;

                    if ( near_bandwidth_iterator == near_bandwidth.end() )
                    {
                        break;
                    }
                }
            }

            // once scan finishes, apply bandwidth inserts, then start over
            if ( near_bandwidth_iterator == near_bandwidth.end() )
            {
                for ( int j = 0; j < near_bandwidth_inserts.length; j++)
                {
                    bandwidth_insert_t * insert = &near_bandwidth_inserts[j];
                    near_bandwidth[insert->flow_id] = insert->bandwidth;
                }
                near_bandwidth_inserts.length = 0;
                near_bandwidth_iterator = near_bandwidth.begin();
            }
        }

        // read flow messages
        {
            const int MSG_BULK = 128;
            msg_flow msgs[MSG_BULK];
            int msg_count = context->queue.try_dequeue_bulk( flow_consumer_token, msgs, MSG_BULK );
            for ( int i = 0; i < msg_count; i++ )
            {
                msg_flow * msg = &msgs[i];
                switch ( msg->type )
                {
                    case MSG_FLOW_QUIT:
                    {
                        state = FLOW_THREAD_STATE_STOP;
                        break;
                    }
                    case MSG_FLOW_CLEAN_SHUTDOWN:
                    {
                        state = FLOW_THREAD_STATE_CLEAN_SHUTDOWN;
                        break;
                    }
                    case MSG_FLOW_CREATE:
                    {
                        flow_create( &flows, &msg->flow_create.token, &msg->flow_create.address_prev );
                        break;
                    }
                    // todo: remove migrate
                    case MSG_FLOW_MIGRATE:
                    {
                        flow_entry_t * flow = flow_get( &flows, msg->flow_migrate_destroy.id, msg->flow_migrate_destroy.version );
                        flow_migrate_destroy( flow, FLOW_STATE_MIGRATED );
                        break;
                    }
                    // todo: remove destroy
                    case MSG_FLOW_DESTROY:
                    {
                        flow_entry_t * flow = flow_get( &flows, msg->flow_migrate_destroy.id, msg->flow_migrate_destroy.version );
                        flow_migrate_destroy( flow, FLOW_STATE_DESTROYED );
                        break;
                    }
                    // todo: remove timeout
                    case MSG_FLOW_TIMEOUT:
                    // todo: IMPORTANT: keep expired
                    case MSG_FLOW_EXPIRED:
                    {
                        flow_map_t::iterator i = flows.find( flow_hash_t( msg->flow_timeout_expire.id, msg->flow_timeout_expire.version ) );
                        flows.erase( i );
                        break;
                    }
                    default:
                    {
                        relay_printf( NEXT_LOG_LEVEL_ERROR, "bad flow message type: %hhu", msg->type );
                        next_assert( false );
                        break;
                    }
                }
            }
        }

        // read management messages
        {
            const int MSG_BULK = 128;
            msg_manage msgs[MSG_BULK];
            int msg_count = global.manage_queue_out.try_dequeue_bulk( manage_consumer_token, msgs, MSG_BULK );
            for ( int i = 0; i < msg_count; i++ )
            {
                msg_manage * msg = &msgs[i];
                switch ( msg->type )
                {
                    case MSG_MANAGE_RELAY_PING_OUTGOING:
                    {
                        flow_relay_ping_send( &socket, &msg->relay_ping_outgoing.address, msg->relay_ping_outgoing.token, msg->relay_ping_outgoing.sequence );
                        break;
                    }
                    case MSG_MANAGE_RELAY_PONG_OUTGOING:
                    {
                        relay_printf( NEXT_LOG_LEVEL_DEBUG, "sending pong, id = %lu, sequence = %lu\n", msg->relay_pong.id, msg->relay_pong.sequence );
                        flow_relay_pong_send( &socket, &msg->relay_pong.address, msg->relay_pong.id, msg->relay_pong.sequence );
                        break;
                    }
                    case MSG_MANAGE_MASTER_PACKET_OUTGOING:
                    {
                        relay_send_packet( &socket, &msg->master_packet_outgoing.address, msg->master_packet_outgoing.packet_data, msg->master_packet_outgoing.packet_bytes );
                        // IMPORTANT: packet buffer is allocated by management thread and freed by flow thread
                        next_free( msg->master_packet_outgoing.packet_data );
                        break;
                    }
                    default:
                    {
                        relay_printf( NEXT_LOG_LEVEL_ERROR, "bad manage to flow message type: %hhu", msg->type );
                        next_assert( false );
                        break;
                    }
                }
            }
        }

        if ( global.dev )
        {
            next_sleep( int64_t( 0.002 * double( NEXT_ONE_SECOND_NS ) ) );
        }
    }

    flow_producer_tokens_cleanup( &flow_producer_tokens );

    next_socket_destroy( &socket );

    relay_printf( NEXT_LOG_LEVEL_INFO, "shutdown flow thread" );

    return 0;
}

const uint8_t MANAGE_TIMEOUT_STATE_ACTIVE = 0;
const uint8_t MANAGE_TIMEOUT_STATE_MIGRATED_DESTROYED = 1;

struct manage_timeout_t
{
    uint64_t expire_timestamp;
    uint8_t state;
};
typedef dense_hash_map<flow_hash_t, manage_timeout_t> manage_timeout_map_t;

manage_timeout_t * manage_timeout_get( manage_timeout_map_t * timeouts, uint64_t flow_id, uint8_t flow_version )
{
    manage_timeout_map_t::iterator i = timeouts->find( flow_hash_t( flow_id, flow_version ) );
    if ( i == timeouts->end() )
    {
        return NULL;
    }
    else
    {
        return &i->second;
    }
}

void manage_timeout_update( manage_timeout_map_t * timeouts, msg_manage * msg )
{
    flow_hash_t hash( msg->flow_update.id, msg->flow_update.version );
    manage_timeout_map_t::iterator i = timeouts->find( hash );
    if ( i == timeouts->end() )
    {
        // timeout doesn't exist insert it
        manage_timeout_t timeout;
        timeout.expire_timestamp = msg->flow_update.expire_timestamp;
        timeout.state = MANAGE_TIMEOUT_STATE_ACTIVE;
        (*timeouts)[hash] = timeout;
    }
    else
    {
        manage_timeout_t * timeout = &i->second;
        timeout->expire_timestamp = msg->flow_update.expire_timestamp;
    }
}

struct manage_stats_t
{
    uint64_t bytes_paid_tx;
    uint64_t bytes_paid_rx;
    uint64_t bytes_management_tx;
    uint64_t bytes_management_rx;
    uint64_t bytes_measurement_tx;
    uint64_t bytes_measurement_rx;
    uint64_t bytes_invalid_rx;
    double usage_samples[USAGE_MAX_SAMPLES];
    int usage_samples_index;

    manage_stats_t()
    {
        memset( this, 0, sizeof(manage_stats_t) );
    }
};

uint64_t manage_peer_history_insert( manage_peer_history_t * history, int64_t time )
{
    manage_peer_history_entry_t * entry = &history->entries[history->index_current];
    entry->sequence = history->sequence_current;
    entry->time_ping_sent = time;
    entry->time_pong_received = 0;
    history->index_current = ( history->index_current + 1 ) % MANAGE_PEER_HISTORY_ENTRY_COUNT;

    manage_peer_history_entry_t * entry_packet_loss = &history->entries_packet_loss[history->index_current_packet_loss_calc];
    entry_packet_loss->sequence = history->sequence_current;
    entry_packet_loss->time_ping_sent = time;
    entry_packet_loss->time_pong_received = 0;
    history->index_current_packet_loss_calc = ( history->index_current_packet_loss_calc + 1 ) % MANAGE_PEER_HISTORY_PACKET_LOSS_ENTRY_COUNT;

    history->sequence_current++;
    return entry->sequence;
}

void manage_peer_history_pong_received( manage_peer_history_t * history, uint64_t sequence, int64_t time )
{
    for ( int i = 0; i < MANAGE_PEER_HISTORY_ENTRY_COUNT; i++ )
    {
        manage_peer_history_entry_t * entry = &history->entries[i];
        if ( entry->time_ping_sent > 0 && entry->sequence == sequence )
        {
            entry->time_pong_received = time;
            break;
        }
    }
    for ( int i = 0; i < MANAGE_PEER_HISTORY_PACKET_LOSS_ENTRY_COUNT; i++ )
    {
        manage_peer_history_entry_t * entry = &history->entries_packet_loss[i];
        if ( entry->time_ping_sent > 0 && entry->sequence == sequence )
        {
            history->time_sent_of_latest_pong_received = entry->time_ping_sent;
            entry->time_pong_received = time;
            break;
        }
    }
}

void manage_peer_history_stats( const manage_peer_history_t * history, int64_t start, int64_t end, next_route_stats_t * stats )
{
    assert( history );
    assert( stats );
    assert( start < end );

    double start_sec = start / NEXT_SEC_TO_NS;
    double end_sec = end / NEXT_SEC_TO_NS;

    stats->rtt = -1.0f;
    stats->jitter = -1.0f;
    stats->packet_loss = -1.0f;

    // calculate packet loss

    int packet_count_sent_new_packet_loss = 0;
    int packet_count_received_new_packet_loss = 0;

    //new packet loss calculation
    for ( int i = 0; i < MANAGE_PEER_HISTORY_PACKET_LOSS_ENTRY_COUNT; i++ )
    {
        const manage_peer_history_entry_t * entry = &history->entries_packet_loss[i];
        //include pings that were sent anytime up to the time of the most recent ping that has received a pong
        if ( entry->time_ping_sent <= history->time_sent_of_latest_pong_received )
        {
            if ( entry->time_pong_received > entry->time_ping_sent )
            {
                // pong received
                packet_count_sent_new_packet_loss++;
                packet_count_received_new_packet_loss++;
            }
            else if ( entry->time_ping_sent > 0 )
            {
                packet_count_sent_new_packet_loss++;
            }
        }
    }

    if ( packet_count_sent_new_packet_loss > 0 && packet_count_received_new_packet_loss > 0)
    {
        stats->packet_loss = 100.0f * ( 1.0f - ( float( packet_count_received_new_packet_loss ) / float( packet_count_sent_new_packet_loss ) ) );
    }

    // calculate mean RTT

    double mean_rtt = 0.0;
    int num_pings = 0;
    int num_pongs = 0;

    for ( int i = 0; i < MANAGE_PEER_HISTORY_ENTRY_COUNT; i++ )
    {
        const manage_peer_history_entry_t * entry = &history->entries[i];

        double time_ping_sent_sec = entry->time_ping_sent / NEXT_SEC_TO_NS;
        double time_pong_received_sec = entry->time_pong_received / NEXT_SEC_TO_NS;

        if ( time_ping_sent_sec >= start_sec && time_ping_sent_sec <= end_sec )
        {
            if ( time_pong_received_sec > time_ping_sent_sec )
            {
                mean_rtt += 1000.0 * ( time_pong_received_sec - time_ping_sent_sec );
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

    for ( int i = 0; i < MANAGE_PEER_HISTORY_ENTRY_COUNT; i++ )
    {
        const manage_peer_history_entry_t * entry = &history->entries[i];

        double time_ping_sent_sec = entry->time_ping_sent / NEXT_SEC_TO_NS;
        double time_pong_received_sec = entry->time_pong_received / NEXT_SEC_TO_NS;

        if ( time_ping_sent_sec >= start_sec && time_ping_sent_sec <= end_sec )
        {
            if ( time_pong_received_sec > time_ping_sent_sec )
            {
                // pong received
                double rtt = 1000.0 * ( time_pong_received_sec - time_ping_sent_sec );
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

double manage_get_usage_from_samples( manage_stats_t * stats )
{
    double max_usage = 0.0;
    for ( int i = 0; i < USAGE_MAX_SAMPLES; i++ )
    {
        if ( stats->usage_samples[i] > max_usage )
        {
            max_usage = stats->usage_samples[i];
        }
    }
    return max_usage;
}

void manage_update_callback( manage_environment_t * env, int status, next_json_document_t * doc )
{
    if ( status != 200 )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "relay update post failed: http status: %d", status );
        metric_count( RELAY_COUNTER_POST_RELAY_STATS_BAD_HTTP_STATUS_CODE, 1 );
        return;
    }

    if ( !doc->HasMember( "PingTargets" ) )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to get ping targets. missing \"PingTargets\" JSON" );
        metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
        return;
    }

    const next_json_value_t& ping_targets = (*doc)["PingTargets"];

    if ( !ping_targets.IsArray() )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to get ping targets. missing \"PingTargets\" JSON" );
        metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
        return;
    }

    for ( next_json_value_t::ConstValueIterator i = ping_targets.Begin(); i != ping_targets.End(); i++ )
    {
        if ( !(*i).HasMember( "Address" ) )
        {
            relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing address entry for ping target" );
            metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
            return;
        }

        if ( !(*i).HasMember( "Id" ) )
        {
            relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing id entry for ping target" );
            metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
            return;
        }

        if ( !(*i).HasMember( "Group" ) )
        {
            relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing group entry for ping target" );
            metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
            return;
        }

        if ( !(*i).HasMember( "PingToken" ) )
        {
            relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing ping token entry for ping target" );
            metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
            return;
        }

        next_address_t address;
        {
            const char * address_base64 = (*i)["Address"].GetString();

            char address_string[NEXT_MAX_ADDRESS_STRING_LENGTH] = {};
            if ( next_base64_decode_string( address_base64, address_string, sizeof( address_string ) ) <= 0 )
            {
                relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to base64 decode ping target address: %s", address_base64 );
                metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                return;
            }

            if ( next_address_parse( &address, address_string ) != NEXT_OK )
            {
                relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to parse ping target address: '%s'", address_string );
                metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                return;
            }
        }

        uint8_t ping_token[NEXT_PING_TOKEN_BYTES];
        {
            const char * ping_token_base64 = (*i)["PingToken"].GetString();

            if ( next_base64_decode_data( ping_token_base64, ping_token, sizeof( ping_token ) ) <= 0 )
            {
                relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to base64 decode ping token: %s", ping_token_base64 );
                metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                return;
            }
        }

        // upsert
        manage_peer_t * peer;
        {
            manage_peer_map_t::iterator i = env->peers.find( address );
            if ( i == env->peers.end() )
            {
                manage_peer_t p;
                manage_peer_init( &p );
                auto inserted = env->peers.insert( manage_peer_map_t::value_type( address, p ) );
                i = inserted.first;
            }

            peer = &i->second;
        }

        peer->relay_id = (*i)["Id"].GetUint64();
        peer->address = address;
        memcpy( peer->ping_token, ping_token, sizeof( ping_token ) );
        peer->group_id = (*i)["Group"].GetUint64();
        peer->dirty = true;
    }

    // remove peers that are not marked dirty
    for ( manage_peer_map_t::iterator i = env->peers.begin(); i != env->peers.end(); i++ )
    {
        manage_peer_t * peer = &i->second;
        if ( peer->dirty )
        {
            peer->dirty = false;
        }
        else
        {
            env->peers.erase( i );
        }
    }

    env->peers.resize( 0 ); // compact hash table
}

static bool manage_check_for_shutdown_ack( manage_peer_map_t * peers, next_address_t * address )
{
    manage_peer_map_t::iterator i = peers->find( *address );
    if ( i != peers->end() )
    {
        return false;
    }
    return true;
}

bool manage_should_ping( manage_environment_t * env, manage_peer_t * peer )
{
    return peer->relay_id != env->relay.id;
}

static void * next_zalloc( void *, size_t count, size_t size )
{
    return next_alloc( count * size );
}

static void next_zfree( void *, void * p )
{
    next_free( p );
}

static int manage_master_packet_read_complete( uint8_t * packet_data, int packet_bytes, next_json_document_t * doc )
{
    const int MAX_PAYLOAD = 2 * MASTER_FRAGMENT_SIZE * MASTER_FRAGMENT_MAX;
    uint8_t * buffer = (uint8_t *)( next_alloc( MAX_PAYLOAD + 1 ) ); // include space for null terminator

    z_stream z;
    z.zalloc = next_zalloc;
    z.zfree = next_zfree;
    z.opaque = NULL;
    z.next_in = (Bytef*)( &packet_data[0] );
    z.avail_in = packet_bytes;
    z.next_out = (Bytef*)( buffer );
    z.avail_out = MAX_PAYLOAD;

    int result = inflateInit( &z );
    if ( result != Z_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to decompress master UDP packet: inflateInit failed" );
        next_free( buffer );
        return NEXT_ERROR;
    }

    result = inflate( &z, Z_NO_FLUSH );
    if ( result != Z_STREAM_END )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to decompress master UDP packet: inflate failed" );
        next_free( buffer );
        return NEXT_ERROR;
    }

    result = inflateEnd( &z );
    if ( result != Z_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to decompress master UDP packet: inflateEnd failed" );
        next_free( buffer );
        return NEXT_ERROR;
    }

    int bytes = int( MAX_PAYLOAD - z.avail_out );
    if ( bytes == 0 )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to decompress master UDP packet: not enough buffer space" );
        next_free( buffer );
        return NEXT_ERROR;
    }

    buffer[bytes] = '\0';

    doc->Parse( (char *)( buffer ) );

    next_free( buffer );
    return NEXT_OK;
}

int manage_master_packet_read(
    uint8_t * packet_data,
    int packet_bytes,
    master_request_t * request,
    int * status_code,
    next_json_document_t * doc )
{
    // 1 byte packet type
    // 64 byte signature
    // <signed>
    //     8 byte GUID
    //     1 byte fragment index
    //     1 byte fragment count
    //     2 byte status code
    //     <zipped>
    //         JSON string
    //     </zipped>
    // </signed>

    int zip_start = int( 1 + crypto_sign_BYTES + sizeof(uint64_t) + sizeof(uint16_t) + sizeof(uint16_t) );

    if ( packet_bytes < zip_start || packet_bytes > zip_start + MASTER_FRAGMENT_SIZE )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "invalid master UDP packet. expected between %d and %d bytes, got %d", zip_start, zip_start + MASTER_FRAGMENT_SIZE, packet_bytes );
        return NEXT_ERROR;
    }

    if ( crypto_sign_verify_detached(
        &packet_data[1],
        &packet_data[1 + crypto_sign_BYTES],
        packet_bytes - ( 1 + crypto_sign_BYTES ),
        NEXT_MASTER_UDP_SIGN_KEY ) != 0 )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "invalid master UDP packet. bad cryptographic signature." );
        return NEXT_ERROR;
    }

    uint8_t * p = &packet_data[1 + crypto_sign_BYTES];
    uint64_t packet_id = next_read_uint64( &p );
    if ( packet_id != request->id )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "discarding unexpected master UDP packet, expected ID %ld, got %ld", request->id, packet_id );
        return NEXT_ERROR;
    }

    uint8_t fragment_index = next_read_uint8( &p );
    uint8_t fragment_total = next_read_uint8( &p );
    uint16_t status = next_read_uint16( &p );

    if ( fragment_total == 0 )
    {
        next_printf( NEXT_LOG_LEVEL_DEBUG, "invalid master fragment count (%hhu), discarding packet", fragment_total );
        return NEXT_ERROR;
    }

    if ( fragment_index >= fragment_total )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "invalid master fragment index (%hhu/%hhu), discarding packet", fragment_index + 1, fragment_total );
        return NEXT_ERROR;
    }

    if ( request->fragment_total == 0 )
    {
        request->type = packet_data[0];
        request->fragment_total = fragment_total;
    }

    if ( packet_data[0] != request->type )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "expected packet type %hhu, got %hhu, discarding packet", request->type, packet_data[0] );
        return NEXT_ERROR;
    }

    if ( fragment_total != request->fragment_total )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "expected %hhu fragments, got fragment %hhu/%hhu, discarding packet", request->fragment_total, fragment_index + 1, fragment_total );
        return NEXT_ERROR;
    }

    if ( request->fragments[fragment_index].received )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "already received master fragment %hhu/%hhu, ignoring packet", fragment_index + 1, fragment_total );
        return NEXT_ERROR;
    }

    // save this fragment
    {
        master_fragment_t * fragment = &request->fragments[fragment_index];
        fragment->length = uint16_t( packet_bytes - zip_start );
        memcpy( fragment->data, &packet_data[zip_start], fragment->length );
        fragment->received = true;
    }

    // check received fragments

    int complete_bytes = 0;

    for ( int i = 0; i < request->fragment_total; i++ )
    {
        master_fragment_t * fragment = &request->fragments[i];
        if ( fragment->received )
        {
            complete_bytes += fragment->length;
        }
        else
        {
            return NEXT_ERROR; // not all fragments have been received yet
        }
    }

    // all fragments have been received

    *status_code = int( status );

    request->id = 0;

    uint8_t * complete_buffer = (uint8_t *)( alloca( complete_bytes ) );
    int bytes = 0;
    for ( int i = 0; i < request->fragment_total; i++ )
    {
        master_fragment_t * fragment = &request->fragments[i];
        memcpy( &complete_buffer[bytes], fragment->data, fragment->length );
        bytes += fragment->length;
    }
    next_assert( bytes == complete_bytes );
    return manage_master_packet_read_complete( complete_buffer, complete_bytes, doc );
}

static int manage_master_packet_send_fragment( uint8_t packet_type,
                                               uint64_t id,
                                               int fragment_index,
                                               int fragment_total,
                                               uint8_t * packet_data,
                                               int packet_bytes,
                                               queue_producer_token_t * producer_token,
                                               master_token_t * master_token,
                                               const next_address_t * address )
{
    next_assert( address );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );
    next_assert( fragment_total > 0 && fragment_index >= 0 && fragment_index < fragment_total && fragment_total <= MASTER_FRAGMENT_MAX );
    next_assert( master_token );

    // 1 byte packet type
    // <encrypted>
    //     <master token>
    //         19 byte IP address
    //         8 byte timestamp
    //         32 byte MAC
    //     </master token>
    //     8 byte GUID
    //     1 byte fragment index
    //     1 byte fragment count
    //     <zipped>
    //         JSON string
    //     </zipped>
    // </encrypted>
    // 64 byte MAC (handled automatically by sodium)

    int header_bytes = 1 + MASTER_TOKEN_BYTES + sizeof( uint64_t ) + 2;

    int total_bytes = header_bytes + packet_bytes + crypto_box_SEALBYTES;

    uint8_t * buffer = (uint8_t *)( alloca( total_bytes - 1 ) );

    uint8_t * p = buffer;
    next_write_address( &p, &master_token->address );
    next_write_bytes( &p, master_token->hmac, sizeof( master_token->hmac ) );
    next_write_uint64( &p, id );
    next_write_uint8( &p, uint8_t( fragment_index ) );
    next_write_uint8( &p, uint8_t( fragment_total ) );
    memcpy( p, packet_data, packet_bytes );

    // IMPORTANT: packet buffer is allocated by management thread and freed by flow thread
    uint8_t * final_packet = (uint8_t *)( next_alloc( total_bytes ) );
    final_packet[0] = packet_type;

    if ( crypto_box_seal( &final_packet[1], buffer, header_bytes - 1 + packet_bytes, NEXT_MASTER_UDP_SEAL_KEY ) != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to crypto seal master UDP packet" );
        return NEXT_ERROR;
    }

    msg_manage msg;
    msg.type = MSG_MANAGE_MASTER_PACKET_OUTGOING;
    msg.master_packet_outgoing.address = *address;
    msg.master_packet_outgoing.packet_data = final_packet;
    msg.master_packet_outgoing.packet_bytes = total_bytes;
    global.manage_queue_out.enqueue( *producer_token, msg );

    return NEXT_OK;
}

int manage_master_packet_send(
    queue_producer_token_t * producer_token,
    const next_address_t * master_address,
    master_token_t * master_token,
    uint8_t packet_type,
    master_request_t * request,
    uint8_t * packet_data,
    int packet_bytes )
{
    if ( master_address->type == NEXT_ADDRESS_NONE )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "can't send master UDP packet: address has not resolved yet" );
        return NEXT_ERROR;
    }

    memset( request, 0, sizeof( *request ) );
    next_random_bytes( (uint8_t *)( &request->id ), sizeof(request->id) );

    int compressed_bytes_available = packet_bytes + 32;
    uint8_t * compressed_buffer = (uint8_t *)( next_alloc( compressed_bytes_available ) );

    z_stream z;
    z.zalloc = next_zalloc;
    z.zfree = next_zfree;
    z.opaque = NULL;
    z.next_out = (Bytef*)( compressed_buffer );
    z.avail_out = compressed_bytes_available;
    z.next_in = (Bytef*)( packet_data );
    z.avail_in = packet_bytes;

    int result = deflateInit( &z, Z_DEFAULT_COMPRESSION );
    if ( result != Z_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to compress master UDP packet: deflateInit failed" );
        next_free( compressed_buffer );
        return NEXT_ERROR;
    }

    result = deflate( &z, Z_FINISH );

    if ( result != Z_STREAM_END || z.avail_in > 0 )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to compress master UDP packet: deflate failed" );
        next_free( compressed_buffer );
        return NEXT_ERROR;
    }

    result = deflateEnd(&z);
    if ( result != Z_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to compress master UDP packet: deflateEnd failed" );
        next_free( compressed_buffer );
        return NEXT_ERROR;
    }

    int compressed_bytes = compressed_bytes_available - z.avail_out;

    int fragment_total = compressed_bytes / MASTER_FRAGMENT_SIZE;
    if ( compressed_bytes % MASTER_FRAGMENT_SIZE != 0 )
    {
        fragment_total += 1;
    }

    if ( fragment_total > MASTER_FRAGMENT_MAX )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "%d byte master packet is too large even for %d fragments!", compressed_bytes, MASTER_FRAGMENT_MAX );
        next_free( compressed_buffer );
        return NEXT_ERROR;
    }

    for ( int i = 0; i < fragment_total; i++ )
    {
        int fragment_bytes;
        if ( i == fragment_total - 1 )
        {
            // last fragment
            fragment_bytes = compressed_bytes - ( ( fragment_total - 1 ) * MASTER_FRAGMENT_SIZE );
        }
        else
        {
            fragment_bytes = MASTER_FRAGMENT_SIZE;
        }

        if ( manage_master_packet_send_fragment(
            packet_type,
            request->id,
            i,
            fragment_total,
            &compressed_buffer[i * MASTER_FRAGMENT_SIZE],
            fragment_bytes,
            producer_token,
            master_token,
            master_address
            ) != NEXT_OK )
        {
            next_free( compressed_buffer );
            return NEXT_ERROR;
        }
    }

    next_free( compressed_buffer );
    return NEXT_OK;
}

int manage_init_callback( next_json_document_t * document, master_init_data_t * init_data, int env_index )
{
	int64_t now = relay_time();

	if ( now - init_data->requested > 1 * NEXT_ONE_SECOND_NS )
		return NEXT_ERROR; // request took too long to complete; timestamp would be inaccurate

    if ( !document->HasMember( "Timestamp" ) )
    	return NEXT_ERROR;

    const next_json_value_t & timestamp = (*document)["Timestamp"];
    if ( timestamp.GetType() != rapidjson::kNumberType )
    	return NEXT_ERROR;

    if ( !document->HasMember( "Token" ) )
        return NEXT_ERROR;

    const next_json_value_t & token = (*document)["Token"];
    if ( token.GetType() != rapidjson::kStringType )
        return NEXT_ERROR;

    uint8_t master_token_bytes[MASTER_TOKEN_BYTES];
    int bytes = next_base64_decode_data( token.GetString(), master_token_bytes, sizeof( master_token_bytes ) );
    if ( bytes != sizeof( master_token_bytes ) )
        return NEXT_ERROR;

    uint8_t * p = master_token_bytes;
    next_read_address( &p, &init_data->token.address );
    next_read_bytes( &p, init_data->token.hmac, sizeof( init_data->token.hmac ) );

    init_data->received = now;

    // timestamp from master is millisecond resolution; convert it to nanoseconds and compensate for RTT
    init_data->timestamp = ( timestamp.GetUint64() * ( NEXT_ONE_SECOND_NS / 1000ULL ) ) + ( ( init_data->received - init_data->requested ) / 2ULL );

	relay_printf( NEXT_LOG_LEVEL_INFO, "processed init data from master (env %d)...", env_index );

    return NEXT_OK;
}

void manage_sign_request( manage_environment_t * env, next_json_document_t * doc )
{
    next_json_allocator_t& allocator = doc->GetAllocator();

    // timestamp and signature

    uint64_t timestamp = global.master_timestamp;
    next_json_value_t value;
    value.SetUint64( timestamp );
    doc->AddMember( "Timestamp", value, allocator );

    char string[1024];

    uint8_t signature[crypto_sign_BYTES];
    crypto_sign_detached( signature, NULL, (unsigned char *)( &timestamp ), sizeof( uint64_t ), env->relay.update_key );
    int signature_size = next_base64_encode_data( signature, sizeof( signature ), string, sizeof( string ) );
    value.SetString( string, next_json_size_t( signature_size ), allocator );
    doc->AddMember( "Signature", value, allocator );
}

void manage_config_callback( manage_environment_t * env, int status_code, next_json_document_t * doc )
{
    if ( status_code != 200 )
    {
        relay_printf( NEXT_LOG_LEVEL_DEBUG, "relay config returned status code %d", status_code );
        return;
    }

    // received relay config
    strncpy( env->relay.group_name, "", sizeof( env->relay.group_name ) );
    if ( doc->HasMember( "Group" ) )
    {
        strncpy( env->relay.group_name, (*doc)["Group"].GetString(), sizeof( env->relay.group_name ) );
    }
    env->relay.group_id = next_relay_id( env->relay.group_name );

    env->valid = true;
}

void manage_master_packet_handle( manage_environment_t * env, msg_manage * msg )
{
    global.bytes_per_sec_management_rx += NEXT_LOW_LEVEL_HEADER_BYTES + msg->master_packet_incoming.packet_bytes;

    if ( msg->master_packet_incoming.packet_bytes < 1 )
    	return;

    int packet_type = msg->master_packet_incoming.packet_data[0];

    master_request_t * request = ( packet_type == NEXT_PACKET_TYPE_V4_MASTER_INIT_RESPONSE ) ? &env->init_request : &env->master_request;

    int status_code;
    next_json_document_t doc;
    int result = manage_master_packet_read(
        msg->master_packet_incoming.packet_data,
        msg->master_packet_incoming.packet_bytes,
        request,
        &status_code,
        &doc );

    // IMPORTANT: packet buffer is allocated by flow thread and freed by management thread
    next_free( msg->master_packet_incoming.packet_data );

    if ( result != NEXT_OK )
        return;

    if ( packet_type == NEXT_PACKET_TYPE_V4_MASTER_RELAY_CONFIG_RESPONSE )
    {
        manage_config_callback( env, status_code, &doc );
    }
    else if ( packet_type == NEXT_PACKET_TYPE_V4_MASTER_RELAY_RESPONSE )
    {
        // received report response
        manage_update_callback( env, status_code, &doc );
    }
    else if ( packet_type == NEXT_PACKET_TYPE_V4_MASTER_INIT_RESPONSE )
    {
        manage_init_callback( &doc, &env->init_data, env->idx );
    }
}

manage_environment_t * manage_env_for_master( next_address_t * address )
{
    for ( int i = 0; i < manage.env_count; i++ )
    {
        manage_environment_t * env = &manage.envs[i];
        if ( next_ip_equal( address, resolver_address( &env->master ) ) )
        {
            return env;
        }
    }
    return NULL;
}

void manage_master_packets_read( queue_consumer_token_t * consumer_token )
{
    const int MSG_BULK = 256;
    msg_manage msgs[MSG_BULK];
    int msg_count;
    while ( ( msg_count = global.manage_queue_in.try_dequeue_bulk( *consumer_token, msgs, MSG_BULK ) ) > 0 )
    {
        for ( int i = 0; i < msg_count; i++ )
        {
            msg_manage * msg = &msgs[i];
            if ( msg->type == MSG_MANAGE_MASTER_PACKET_INCOMING )
            {
                manage_environment_t * env = manage_env_for_master( &msg->master_packet_incoming.address );
                if ( env )
                {
                    manage_master_packet_handle( env, msg );
                }
            }
        }
    }
}

// the first environment in the envs list that contains a peer relay at a certain IP address
// is the one where we keep track of ping history for that peer. If the same peer appears in
// other environments, those copies will not have any ping history. Therefore, when reporting
// ping stats for that peer to the other environments, we have to pull the history from the
// first environment.
manage_peer_t * manage_deduplicate_relay( next_address_t * address )
{
    // return peer from first environment that has this relay
    for ( int i = 0; i < manage.env_count; i++ )
    {
        manage_environment_t * env = &manage.envs[i];

        manage_peer_map_t::iterator j = env->peers.find( *address );

        if ( j != env->peers.end() )
        {
            manage_peer_t * peer = &j->second;
            return peer;
        }
    }
    return NULL;
}

// calculate current master timestamp (one second resolution)
// based on init data received from master
// usage: master_timestamp( init_data, relay_time() )
uint64_t master_timestamp( master_init_data_t * init_data, int64_t current_time )
{
    return ( init_data->timestamp + ( current_time - init_data->received ) ) / NEXT_ONE_SECOND_NS;
}

next_thread_return_t NEXT_THREAD_FUNC manage_thread( void * )
{
    next_vector_t<msg_manage> timeout_updates;
    manage_timeout_map_t timeouts;
    timeouts.set_empty_key( flow_hash_t( 0, 0 ) );
    timeouts.set_deleted_key( flow_hash_t( uint64_t( -1 ), uint8_t( -1 ) ) );

    manage_timeout_map_t::iterator timeout_iterator = timeouts.end();

    bool clean_shutdown_started = false;
    bool clean_shutdown_ended = false;
    int64_t clean_shutdown_start_time = 0;
    int64_t clean_shutdown_end_time = 0;
    int64_t clean_shutdown_countdown_tick_time = 0;
    bool clean_shutdown_ack = false;
    int clean_shutdown_countdown = 0;

    manage_stats_t stats;

    // ping map determines which relays have been pinged (to avoid double-pinging relays that appear in multiple environments)
    manage_ping_map_t ping_map;
    {
        next_address_t address;
        address.type = NEXT_ADDRESS_NONE;
        address.data.ipv6[0] = 0;
        address.data.ipv6[1] = 0;
        address.data.ipv6[2] = 0;
        address.data.ipv6[3] = 0;
        address.data.ipv6[4] = 0;
        address.data.ipv6[5] = 0;
        address.data.ipv6[6] = 0;
        address.data.ipv6[7] = 0;
        address.port = 0;
        ping_map.set_empty_key( address );
        address.port = 1;
        ping_map.set_deleted_key( address );
    }

    queue_consumer_token_t consumer_token( global.manage_queue_in );
    queue_producer_token_t producer_token( global.manage_queue_out );

    next_vector_t<queue_producer_token_t> flow_producer_tokens;
    flow_producer_tokens_init( &flow_producer_tokens );

    int64_t update_last = relay_time();
    int64_t ping_targets_last = update_last;
    int64_t traffic_stats_last = update_last;
    int64_t metrics_post_last = update_last;
    bool run = true;

    for ( int i = 0; i < manage.env_count; i++ )
    {
      break;
        if ( quit )
            break;

        manage_environment_t * env = &manage.envs[i];

        relay_printf( NEXT_LOG_LEVEL_INFO, "http requesting init data (env %d)...", i );

        int64_t init_request_first = relay_time();

        while (!quit)
        {
            int64_t now = relay_time();
            if ( now - init_request_first > 60 * NEXT_ONE_SECOND_NS )
            {
                relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to post init data (env %d)...", i );
                exit( 1 );
            }

            if ( now - env->init_data.requested > 1 * NEXT_ONE_SECOND_NS )
            {
                env->init_data.requested = now;
                std::string resp;
                char addr_buff[NEXT_ADDRESS_BYTES + NEXT_ADDRESS_BUFFER_SAFETY] = {};
                next_address_to_string(&env->relay.address, addr_buff);
                auto request_success = compat::next_curl_init(global.backend_hostname, addr_buff, global.bind_port, global.router_public_key, global.relay_private_key, resp);
                env->init_data.received = relay_time();

                json::JSON doc;
                if (doc.parse(resp)) {
                    if (!doc.memberExists("Timestamp")) {
                      printf("timestamp not sent from relay backend\n");
                    } else {
                        env->init_data.timestamp = doc.get<uint64_t>("Timestamp");
                        env->valid = true;
                        env->http_success = true;
                        break;
                    }
                }
            }
        }
    }

    // get master init data
    for ( int i = 0; i < manage.env_count; i++ )
    {
        if ( quit )
            break;

        manage_environment_t * env = &manage.envs[i];

        if ( env->http_success )
        {
            break;
        }

        relay_printf( NEXT_LOG_LEVEL_INFO, "requesting init data (env %d)...", i );

	    int64_t init_request_first = relay_time();

        // todo
        env->init_request.id = 0;

        while ( !quit )
        {
            int64_t now = relay_time();
            if ( now - init_request_first > 60 * NEXT_ONE_SECOND_NS )
            {
                relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to get init data (env %d)...", i );
                exit( 1 );
            }

            if ( now - env->init_data.requested > 1 * NEXT_ONE_SECOND_NS )
            {
		        relay_printf( NEXT_LOG_LEVEL_INFO, "requesting init data (env %d)...", i );
                env->init_data.requested = now;
                manage_master_packet_send
                (
                    &producer_token,
                    resolver_address( &env->master ),
                    &env->init_data.token,
                    NEXT_PACKET_TYPE_V4_MASTER_INIT_REQUEST,
                    &env->init_request,
                    NEXT_MASTER_INIT_KEY,
                    sizeof( NEXT_MASTER_INIT_KEY )
                );
            }

            manage_master_packets_read( &consumer_token );

            resolver_update( &env->master );

            if ( !env->init_request.id ) // request was completed
                break;
        }

        if ( env->init_data.token.address.type == NEXT_ADDRESS_NONE )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to get init data (env %d)...", i );
            exit( 1 );
        }
    }

    // get config
    for ( int i = 0; i < manage.env_count; i++ )
    {
        if ( quit )
            break;

        manage_environment_t * env = &manage.envs[i];

        if ( env->http_success )
        {
            break;
        }

        int64_t first_config_request = relay_time();
        int64_t last_config_request = -1000;

        while ( !quit )
        {
            int64_t time = relay_time();

            global.master_timestamp = master_timestamp( &env->init_data, time );

            if ( time - first_config_request > 60 * NEXT_ONE_SECOND_NS )
            {
                relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to pull relay config (env %d)...", i );
                exit( 1 );
            }

            if ( time - last_config_request > 1 * NEXT_ONE_SECOND_NS )
            {
                last_config_request = time;

                relay_printf( NEXT_LOG_LEVEL_INFO, "requesting relay config (env %d)...", i );

                next_json_document_t doc;
                doc.SetObject();

                next_json_value_t value;
                value.SetUint64( env->relay.id );
                doc.AddMember( "RelayId", value, doc.GetAllocator() );

                manage_sign_request( env, &doc );

                next_json_string_buffer_t request_buffer;
                next_json_writer_t writer( request_buffer );
                doc.Accept( writer );

                manage_master_packet_send
                (
                    &producer_token,
                    resolver_address( &env->master ),
                    &env->init_data.token,
                    NEXT_PACKET_TYPE_V4_MASTER_RELAY_CONFIG_REQUEST,
                    &env->master_request,
                    (uint8_t *)( request_buffer.GetString() ),
                    request_buffer.GetSize()
                );
            }

            manage_master_packets_read( &consumer_token );

            resolver_update( &env->master );

            if ( !env->master_request.id && env->valid ) // request was completed and successful
                break;
        }
    }

    // setup relay data JSON
    for ( int i = 0; i < manage.env_count; i++ )
    {
        manage_environment_t * env = &manage.envs[i];

        env->relay_data_json.SetObject();
        next_json_allocator_t& allocator = env->relay_data_json.GetAllocator();
        next_json_value_t value;
        char string[1024];

        // id
        value.SetUint64( env->relay.id );
        env->relay_data_json.AddMember( "Id", value, allocator );

        // public key
        int public_key_size = next_base64_encode_data( global.relay_public_key, sizeof( global.relay_public_key ), string, sizeof( string ) );
        value.SetString( string, next_json_size_t( public_key_size ), allocator );
        env->relay_data_json.AddMember( "PublicKey", value, allocator );

        // ping key
        int ping_key_size = next_base64_encode_data( global.relay_ping_key, sizeof( global.relay_ping_key ), string, sizeof( string ) );
        value.SetString( string, next_json_size_t( ping_key_size ), allocator );
        env->relay_data_json.AddMember( "PingKey", value, allocator );

        // group
        value.SetString( env->relay.group_name, next_json_size_t( strlen( env->relay.group_name ) ), allocator );
        env->relay_data_json.AddMember( "Group", value, allocator );

        // shutdown
        value.SetBool( false );
        env->relay_data_json.AddMember( "Shutdown", value, allocator );
    }

    // started!
    for ( int i = 0; i < manage.env_count; i++ )
    {
        manage_environment_t * env = &manage.envs[i];
        char relay_address_string[NEXT_MAX_ADDRESS_STRING_LENGTH];
        env->relay.address.port = global.bind_port;
        next_address_to_string( &env->relay.address, relay_address_string );
        if ( !quit )
        {
            relay_printf( NEXT_LOG_LEVEL_INFO, "started relay: %s (%s)", env->relay.name, relay_address_string );
        }
    }

    while ( run )
    {
        int64_t time = relay_time();

        global.master_timestamp = master_timestamp( &manage.envs[0].init_data, time );

        uint64_t timestamp = global.master_timestamp;

        // read messages
        const int MSG_BULK = 256;
        msg_manage msgs[MSG_BULK];
        int msg_count;
        while ( ( msg_count = global.manage_queue_in.try_dequeue_bulk( consumer_token, msgs, MSG_BULK ) ) > 0 )
        {
            for ( int i = 0; i < msg_count; i++ )
            {
                msg_manage * msg = &msgs[i];
                switch ( msg->type )
                {
                    case MSG_MANAGE_RELAY_PING_INCOMING:
                    {
                        // check if any environment has a relay with this IP and ID
                        bool match = false;
                        // bool bandwidth_over_budget = false;
                        relay_printf( NEXT_LOG_LEVEL_DEBUG, "ping incoming, id = %lu\n", msg->relay_ping_incoming.id );
                        for ( int i = 0; i < manage.env_count; i++ )
                        {
                            manage_environment_t * env = &manage.envs[i];
                            manage_peer_map_t::iterator j = env->peers.find( msg->relay_ping_incoming.address );
                            manage_peer_t * peer = j == env->peers.end() ? NULL : &j->second;

                            if ( !peer )
                                continue;

                            relay_printf( NEXT_LOG_LEVEL_DEBUG, "checking against id = %lu", peer->relay_id );

                            if ( msg->relay_ping_incoming.id != peer->relay_id )
                                continue;

                            match = true;

                            /*
                            if ( next_flow_bandwidth_over_budget( &peer->ping_bandwidth, time, NEXT_ONE_SECOND_NS, NEXT_RELAY_PING_KBPS, BYTES_V3_PING ) )
                            {
                                bandwidth_over_budget = true;
                            }
                            */

                            break;
                        }

                        if ( !match )
                        {
                            metric_count( RELAY_COUNTER_RELAY_PING_PACKET_NO_PING_TARGET, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, BYTES_V3_PING );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + BYTES_V3_PING;
                            break;
                        }

                        metric_count( RELAY_COUNTER_RELAY_PING_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, BYTES_V3_PING );
                        global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + BYTES_V3_PING;

                        /*
                        if ( bandwidth_over_budget )
                        {
                            metric_count( RELAY_COUNTER_RELAY_PING_EXCEED_BANDWIDTH, 1 );
                            break;
                        }
                        */

                        msg_manage msg_pong;
                        msg_pong.type = MSG_MANAGE_RELAY_PONG_OUTGOING;
                        msg_pong.relay_pong.address = msg->relay_ping_incoming.address;
                        msg_pong.relay_pong.id = msg->relay_ping_incoming.id;
                        msg_pong.relay_pong.sequence = msg->relay_ping_incoming.sequence;
                        global.manage_queue_out.enqueue( producer_token, msg_pong );
                        break;
                    }
                    case MSG_MANAGE_RELAY_PONG_INCOMING:
                    {
                        // check if any of our environments are expecting a pong from this address and with this relay ID
                        bool match = false;
                        for ( int i = 0; i < manage.env_count; i++ )
                        {
                            manage_environment_t * env = &manage.envs[i];
                            manage_peer_map_t::iterator j = env->peers.find( msg->relay_pong.address );

                            if ( j == env->peers.end() )
                                continue;

                            manage_peer_t * peer = &j->second;

                            if ( peer && msg->relay_pong.id == env->relay.id )
                            {
                                match = true;
                                break;
                            }
                        }

                        if ( !match )
                        {
                            metric_count( RELAY_COUNTER_RELAY_PONG_PACKET_NO_PING_TARGET, 1 );
                            metric_count( RELAY_STATS_INVALID_INGRESS_BYTES, BYTES_V3_PONG );
                            global.bytes_per_sec_invalid_rx += NEXT_LOW_LEVEL_HEADER_BYTES + BYTES_V3_PONG;
                            break;
                        }

                        // count the pong in the first environment that has this relay
                        manage_peer_t * peer = manage_deduplicate_relay( &msg->relay_pong.address );
                        manage_peer_history_pong_received( &peer->history, msg->relay_pong.sequence, time );

                        metric_count( RELAY_COUNTER_RELAY_PONG_PACKETS_RECEIVED, 1 );
                        metric_count( RELAY_STATS_MEASUREMENT_INGRESS_BYTES, BYTES_V3_PONG );
                        global.bytes_per_sec_measurement_rx += NEXT_LOW_LEVEL_HEADER_BYTES + BYTES_V3_PONG;

                        break;
                    }
                    case MSG_MANAGE_FLOW_UPDATE:
                    {
                        timeout_updates.add( *msg );
                        break;
                    }
                    case MSG_MANAGE_FLOW_MIGRATE_DESTROY:
                    {
                        manage_timeout_t * timeout = manage_timeout_get( &timeouts, msg->flow_packet.id, msg->flow_packet.version );
                        if ( timeout )
                        {
                            timeout->state = MANAGE_TIMEOUT_STATE_MIGRATED_DESTROYED;
                        }
                        break;
                    }
                    case MSG_MANAGE_CLEAN_SHUTDOWN:
                    {
                        if ( !clean_shutdown_started )
                        {
                            clean_shutdown_start_time = relay_time();
                            clean_shutdown_end_time = SHUTDOWN_MAX_SECONDS * NEXT_ONE_SECOND_NS;
                            clean_shutdown_countdown_tick_time = clean_shutdown_start_time;
                            clean_shutdown_ack = false;
                            clean_shutdown_countdown = int( clean_shutdown_end_time / NEXT_ONE_SECOND_NS );
                            clean_shutdown_started = true;
                            clean_shutdown_ended = false;
                            relay_printf( NEXT_LOG_LEVEL_INFO, "starting relay shutdown" );
                        }
                        break;
                    }
                    case MSG_MANAGE_MASTER_PACKET_INCOMING:
                    {
                        manage_environment_t * env = manage_env_for_master( &msg->master_packet_incoming.address );
                        if ( env )
                        {
                            manage_master_packet_handle( env, msg );
                        }
                        break;
                    }
                    case MSG_MANAGE_QUIT:
                    {
                        run = false;
                        break;
                    }
                    default:
                    {
                        relay_printf( NEXT_LOG_LEVEL_ERROR, "bad manage message type: %hhu", msg->type );
                        next_assert( false );
                        break;
                    }
                }
            }
        }

        // clean shutdown
        for ( int i = 0; i < manage.env_count; i++ )
        {
            manage.envs[i].relay_data_json["Shutdown"].SetBool( clean_shutdown_started );
        }
        if ( clean_shutdown_started )
        {
            if ( !clean_shutdown_ended )
            {
                int64_t current_time = relay_time();
                if ( current_time - clean_shutdown_countdown_tick_time > 1 * NEXT_ONE_SECOND_NS )
                {
                    relay_printf( NEXT_LOG_LEVEL_INFO, "countdown: %d", clean_shutdown_countdown );
                    clean_shutdown_countdown_tick_time = current_time;
                    clean_shutdown_countdown--;
                }

                if ( current_time - clean_shutdown_start_time > clean_shutdown_end_time )
                {
                    clean_shutdown_ended = true;
                    relay_printf( NEXT_LOG_LEVEL_INFO, "*** shutdown ***" );

                    // notify flow threads
                    {
                        msg_flow msg;
                        msg.type = MSG_FLOW_CLEAN_SHUTDOWN;
                        for ( int i = 0; i < global.flow_thread_count; i++ )
                        {
                            global.flow_thread_contexts[i].queue.enqueue( msg );
                        }
                    }
                }
            }

            if ( !clean_shutdown_ack )
            {
                if ( ( clean_shutdown_ack = manage_check_for_shutdown_ack( &manage.envs[0].peers, &manage.envs[0].relay.address ) ) )
                {
                    clean_shutdown_start_time = relay_time();
                    clean_shutdown_end_time = SHUTDOWN_ACK_SECONDS * NEXT_ONE_SECOND_NS;
                    clean_shutdown_countdown = int( clean_shutdown_end_time / NEXT_ONE_SECOND_NS );
                    relay_printf( NEXT_LOG_LEVEL_INFO, "received shutdown ack from master, shutting down in %d seconds", int( SHUTDOWN_ACK_SECONDS ) );
                }
            }
        }

        // timeout code
        {
            // scan timeouts
            if ( timeout_iterator != timeouts.end() )
            {
                const int TIMEOUT_SCAN_BULK = 1024; // only scan N timeouts at a time
                const int TIMEOUT_SEND_BULK = 64; // only send N timeouts at a time
                msg_flow msgs[TIMEOUT_SEND_BULK];
                int msg_index = 0;
                for ( int i = 0; i < TIMEOUT_SCAN_BULK && msg_index < TIMEOUT_SEND_BULK; i++ )
                {
                    manage_timeout_t * timeout = &timeout_iterator->second;
                    if ( timestamp > timeout->expire_timestamp )
                    {
                        // expired
                        const flow_hash_t * hash = &timeout_iterator->first;

                        if ( timeout->state == MANAGE_TIMEOUT_STATE_ACTIVE )
                        {
                            relay_flow_log( NEXT_LOG_LEVEL_INFO, hash->id, hash->version, "flow expired" );
                            metric_count( RELAY_COUNTER_FLOW_EXPIRE, 1 );
                        }

                        msg_flow * msg = &msgs[msg_index];
                        msg->type = MSG_FLOW_EXPIRED;
                        msg->flow_timeout_expire.id = hash->id;
                        msg->flow_timeout_expire.version = hash->version;
                        msg_index++;

                        timeouts.erase( timeout_iterator );
                    }

                    timeout_iterator++;

                    if ( timeout_iterator == timeouts.end() )
                    {
                        break;
                    }
                }

                if ( msg_index > 0 )
                {
                    // send timeouts to flow threads
                    for ( int i = 0; i < global.flow_thread_count; i++ )
                    {
                        global.flow_thread_contexts[i].queue.enqueue_bulk( flow_producer_tokens[i], msgs, msg_index );
                    }
                }
            }

            // once scan finishes, apply timeout updates, then start over
            if ( timeout_iterator == timeouts.end() )
            {
                for ( int j = 0; j < timeout_updates.length; j++)
                {
                    manage_timeout_update( &timeouts, &timeout_updates[j] );
                }
                timeout_updates.length = 0;
                timeout_iterator = timeouts.begin();
            }
        }

        // sync time periodically
        for ( int i = 0; i < manage.env_count; i++ )
        {
            manage_environment_t * env = &manage.envs[i];
            if ( env->http_success )
            {
                break;
            }

            if ( time - env->init_data.requested > 5 * 60 * NEXT_ONE_SECOND_NS )
            {
            	relay_printf( NEXT_LOG_LEVEL_INFO, "requesting init data for time sync (env %d)...", i );

                env->init_data.requested = time;
                manage_master_packet_send
                (
                    &producer_token,
                    resolver_address( &env->master ),
                    &env->init_data.token,
                    NEXT_PACKET_TYPE_V4_MASTER_INIT_REQUEST,
                    &env->init_request,
                    NEXT_MASTER_INIT_KEY,
                    sizeof( NEXT_MASTER_INIT_KEY )
                );
            }
        }

        // update
        if ( time - update_last > INTERVAL_UPDATE )
        {
            update_last = time;

            // usage
            {
                uint64_t bytes_per_sec_paid_tx = global.bytes_per_sec_paid_tx;
                uint64_t bytes_per_sec_paid_rx = global.bytes_per_sec_paid_rx;
                uint64_t bytes_per_sec_management_tx = global.bytes_per_sec_management_tx;
                uint64_t bytes_per_sec_management_rx = global.bytes_per_sec_management_rx;
                uint64_t bytes_per_sec_measurement_tx = global.bytes_per_sec_measurement_tx;
                uint64_t bytes_per_sec_measurement_rx = global.bytes_per_sec_measurement_rx;
                uint64_t bytes_per_sec_invalid_rx = global.bytes_per_sec_invalid_rx;

                global.bytes_per_sec_paid_tx -= bytes_per_sec_paid_tx;
                global.bytes_per_sec_paid_rx -= bytes_per_sec_paid_rx;
                global.bytes_per_sec_management_tx -= bytes_per_sec_management_tx;
                global.bytes_per_sec_management_rx -= bytes_per_sec_management_rx;
                global.bytes_per_sec_measurement_tx -= bytes_per_sec_measurement_tx;
                global.bytes_per_sec_measurement_rx -= bytes_per_sec_measurement_rx;
                global.bytes_per_sec_invalid_rx -= bytes_per_sec_invalid_rx;

                stats.bytes_paid_tx += bytes_per_sec_paid_tx;
                stats.bytes_paid_rx += bytes_per_sec_paid_rx;
                stats.bytes_management_tx += bytes_per_sec_management_tx;
                stats.bytes_management_rx += bytes_per_sec_management_rx;
                stats.bytes_measurement_tx += bytes_per_sec_measurement_tx;
                stats.bytes_measurement_rx += bytes_per_sec_measurement_rx;
                stats.bytes_invalid_rx += bytes_per_sec_invalid_rx;

                uint64_t bytes_per_sec_total_tx = bytes_per_sec_paid_tx
                    + bytes_per_sec_management_tx
                    + bytes_per_sec_measurement_tx;
                uint64_t bytes_per_sec_total_rx =
                    bytes_per_sec_paid_rx
                    + bytes_per_sec_management_rx
                    + bytes_per_sec_measurement_rx
                    + bytes_per_sec_invalid_rx;

                double usage = (( 100.0 * 8.0 * ( bytes_per_sec_total_tx + bytes_per_sec_total_rx ) ) / manage.envs[0].relay.speed)/INTERVAL_UPDATE;
                stats.usage_samples[stats.usage_samples_index] = usage;
                stats.usage_samples_index = ( stats.usage_samples_index + 1 ) % USAGE_MAX_SAMPLES;
            }

            next_json_document_t traffic_stats;

            // traffic stats
            if ( time - traffic_stats_last > INTERVAL_TRAFFIC_STATS )
            {
                traffic_stats_last = time;

                traffic_stats.SetObject();

                next_json_allocator_t& allocator = traffic_stats.GetAllocator();

                next_json_value_t value;
                value.SetUint64( stats.bytes_paid_tx );
                traffic_stats.AddMember( "BytesPaidTx", value, allocator );
                value.SetUint64( stats.bytes_paid_rx );
                traffic_stats.AddMember( "BytesPaidRx", value, allocator );
                value.SetUint64( stats.bytes_management_tx );
                traffic_stats.AddMember( "BytesManagementTx", value, allocator );
                value.SetUint64( stats.bytes_management_rx );
                traffic_stats.AddMember( "BytesManagementRx", value, allocator );
                value.SetUint64( stats.bytes_measurement_tx );
                traffic_stats.AddMember( "BytesMeasurementTx", value, allocator );
                value.SetUint64( stats.bytes_measurement_rx );
                traffic_stats.AddMember( "BytesMeasurementRx", value, allocator );
                value.SetUint64( stats.bytes_invalid_rx );
                traffic_stats.AddMember( "BytesInvalidRx", value, allocator );

                stats.bytes_paid_tx = 0;
                stats.bytes_paid_rx = 0;
                stats.bytes_management_tx = 0;
                stats.bytes_management_rx = 0;
                stats.bytes_measurement_tx = 0;
                stats.bytes_measurement_rx = 0;
                stats.bytes_invalid_rx = 0;

                uint64_t flow_count = uint64_t( timeouts.size() );
                value.SetUint64( flow_count );
                traffic_stats.AddMember( "SessionCount", value, allocator );
            }

            for ( int i = 0; i < manage.env_count; i++ )
            {
                manage_environment_t * env = &manage.envs[i];

                // collect and post stats
                next_json_document_t relay_stats_json;
                relay_stats_json.SetObject();
                next_json_allocator_t& allocator = relay_stats_json.GetAllocator();
                next_json_value_t stats_json;

                next_json_value_t u;
                u.SetDouble( manage_get_usage_from_samples( &stats ) );
                relay_stats_json.AddMember( "Usage", u, allocator );

                if ( traffic_stats.IsObject() )
                {
                    next_json_value_t t;
                    t.CopyFrom( traffic_stats, allocator );
                    relay_stats_json.AddMember( "TrafficStats", t, allocator );
                }

                stats_json.SetArray();
                for ( manage_peer_map_t::iterator i = env->peers.begin(); i != env->peers.end(); i++ )
                {
                    manage_peer_t * peer = &i->second;
                    if ( manage_should_ping( env, peer ) )
                    {
                        // get stats from first environment this relay appears in
                        manage_peer_t * deduplicated_peer = manage_deduplicate_relay( &peer->address );

                        next_route_stats_t stats;
                        manage_peer_history_stats( &deduplicated_peer->history, time - ( 2 * NEXT_ONE_SECOND_NS ), time, &stats );

                        if ( stats.rtt >= 0.0f )
                        {
                            next_json_value_t entry_json;
                            entry_json.SetObject();

                            next_json_value_t value;

                            // here we use the original peer, not the deduplicated peer
                            // this is because a relay may have different IDs in different environments.
                            // we want the ping stats from the deduplicated peer (the one in the first environment in which it appears),
                            // but we want the relay ID from the peer inside the current environment (manage.envs[i]).
                            value.SetUint64( peer->relay_id );
                            entry_json.AddMember( "RelayId", value, allocator );

                            value.SetDouble( stats.rtt );
                            entry_json.AddMember( "RTT", value, allocator );

                            value.SetDouble( stats.jitter );
                            entry_json.AddMember( "Jitter", value, allocator );

                            value.SetDouble( stats.packet_loss );
                            entry_json.AddMember( "PacketLoss", value, allocator );

                            stats_json.PushBack( entry_json, allocator );
                        }
                    }
                }

                if ( stats_json.IsArray() && stats_json.Size() > 0 )
                {
                    relay_stats_json.AddMember( "PingStats", stats_json, allocator );
                }
                else
                {
                    next_json_value_t null_json;
                    relay_stats_json.AddMember( "PingStats", null_json, allocator );
                }


                next_json_value_t metadata_copy;
                metadata_copy.CopyFrom( env->relay_data_json, allocator );
                relay_stats_json.AddMember( "Metadata", metadata_copy, allocator );

                manage_sign_request( env, &relay_stats_json );

                next_json_string_buffer_t request_buffer;
                next_json_writer_t writer( request_buffer );
                relay_stats_json.Accept( writer );
                if ( manage_master_packet_send(
                        &producer_token,
                        resolver_address( &env->master ),
                        &env->init_data.token,
                        NEXT_PACKET_TYPE_V4_MASTER_RELAY_REPORT,
                        &env->master_request,
                        (uint8_t *)( request_buffer.GetString() ),
                        request_buffer.GetSize()
                    ) != NEXT_OK )
                {
                    relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to upload relay report to master" );
                }

                resolver_update( &env->master );

                /******** Begin Relay Http Compat *********/
                json::JSON respDoc;
                respDoc.parse(request_buffer.GetString());
                relay_printf( NEXT_LOG_LEVEL_DEBUG, "Update json: %s", respDoc.toPrettyString().c_str() );
                char addr_buff[NEXT_ADDRESS_BYTES + NEXT_ADDRESS_BUFFER_SAFETY] = {};
                next_address_to_string( &env->relay.address, addr_buff);
                if (false && compat::next_curl_update(global.backend_hostname, request_buffer.GetString(), addr_buff, global.bind_port, env->relay.name, respDoc)) {
                    if (respDoc.memberExists("ping_data")) {
                        auto member = respDoc.get<rapidjson::Value*>("ping_data");
                          if (member->IsArray()) {
                            for(rapidjson::Value::ConstValueIterator i = member->Begin(); i != member->End(); i++ ) {
                                if ( !(*i).HasMember( "relay_address" ) )
                                {
                                    relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing address entry for ping target" );
                                    metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                                    continue;
                                }

                                if ( !(*i).HasMember( "relay_id" ) )
                                {
                                    relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing id entry for ping target" );
                                    metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                                    continue;
                                }

                                if ( !(*i).HasMember( "ping_info" ) )
                                {
                                    relay_printf( NEXT_LOG_LEVEL_DEBUG, "missing ping info entry for ping target" );
                                    metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                                    continue;
                                }

                                std::string addr((*i)["relay_address"].GetString());

                                next_address_t address;
                                if ( next_address_parse( &address, addr.c_str() ) != NEXT_OK )
                                {
                                    relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to parse ping target address" );
                                    metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                                    continue;
                                }

                                uint8_t ping_token[NEXT_PING_TOKEN_BYTES];
                                {
                                    const char * ping_token_base64 = (*i)["ping_info"].GetString();

                                    if ( next_base64_decode_data( ping_token_base64, ping_token, sizeof( ping_token ) ) <= 0 )
                                    {
                                        relay_printf( NEXT_LOG_LEVEL_DEBUG, "failed to base64 decode ping token: %s", ping_token_base64 );
                                        metric_count( RELAY_COUNTER_UPDATE_PING_TARGETS_READ_RESPONSE_JSON_FAILED, 1 );
                                        return;
                                    }
                                }

                                // upsert
                                manage_peer_t * peer;
                                {
                                    manage_peer_map_t::iterator i = env->peers.find( address );
                                    if ( i == env->peers.end() )
                                    {
                                        manage_peer_t p;
                                        manage_peer_init( &p );
                                        auto inserted = env->peers.insert( manage_peer_map_t::value_type( address, p ) );
                                        i = inserted.first;
                                    }

                                    peer = &i->second;
                                }

                                peer->relay_id = (*i)["relay_id"].GetUint64();
                                peer->address = address;
                                memcpy( peer->ping_token, ping_token, sizeof( ping_token ) ); // ping token is | ping timeout | relay id | hash of two |
                                // ? peer->group_id = (*i)["Group"].GetUint64();; // what is this?
                                peer->dirty = true;
                            }
                        }
                    }

                    // remove peers that are not marked dirty
                    for ( manage_peer_map_t::iterator i = env->peers.begin(); i != env->peers.end(); i++ )
                    {
                        manage_peer_t * peer = &i->second;
                        if ( peer->dirty )
                        {
                            peer->dirty = false;
                        }
                        else
                        {
                            env->peers.erase( i );
                        }
                    }

                    env->peers.resize( 0 ); // compact hash table
                }
                /******** End Relay Http Compat *********/
            }
        }
        if ( time - ping_targets_last > INTERVAL_PING_TARGETS )
        {
            ping_targets_last = time;

            const int PING_BULK = 64;
            msg_manage pings[PING_BULK];
            int ping_index = 0;

            ping_map.clear();

            for ( int i = 0; i < manage.env_count; i++ )
            {
                manage_environment_t * env = &manage.envs[i];
                for ( manage_peer_map_t::iterator j = env->peers.begin(); j != env->peers.end(); j++ )
                {
                    if ( ping_index == PING_BULK )
                    {
                        global.manage_queue_out.enqueue_bulk( producer_token, pings, PING_BULK );
                        ping_index = 0;
                    }

                    manage_peer_t * peer = &j->second;

                    if ( manage_should_ping( env, peer ) && ping_map.find( peer->address ) == ping_map.end() )
                    {
                        // ping only in the first environment that has this relay
                        uint64_t sequence = manage_peer_history_insert( &peer->history, time );

                        msg_manage * msg_ping = &pings[ping_index];
                        msg_ping->type = MSG_MANAGE_RELAY_PING_OUTGOING;
                        msg_ping->relay_ping_outgoing.address = peer->address;
                        msg_ping->relay_ping_outgoing.sequence = sequence;
                        memcpy( msg_ping->relay_ping_outgoing.token, peer->ping_token, sizeof( peer->ping_token ) );

                        ping_map[peer->address] = true;

                        ping_index++;
                    }
                }
            }

            global.manage_queue_out.enqueue_bulk( producer_token, pings, ping_index );
        }

        // metrics
        if ( time - metrics_post_last > METRICS_RATE )
        {
            metric_raw( "relay.stats.usage.percent", manage_get_usage_from_samples( &stats ) );
            metrics_post_last = time;
            metrics_post();
        }

        if ( global.dev )
        {
            next_sleep( int64_t( 0.002 * double( NEXT_ONE_SECOND_NS ) ) );
        }
    }

    flow_producer_tokens_cleanup( &flow_producer_tokens );

    relay_printf( NEXT_LOG_LEVEL_INFO, "shutdown management thread" );

    return 0;
}

void interrupt_handler( int signal )
{
    // notify manage thread

    if ( signal == SIGHUP )
    {
        relay_printf( NEXT_LOG_LEVEL_INFO, "starting clean shutdown" );

        msg_manage msg;
        msg.type = MSG_MANAGE_CLEAN_SHUTDOWN;
        global.manage_queue_in.enqueue( msg );
    }
    else
    {
        msg_flow msg;
        msg.type = MSG_FLOW_QUIT;
        for ( int i = 0; i < global.flow_thread_count; i++ )
        {
            global.flow_thread_contexts[i].queue.enqueue( msg );
        }
    }

    quit = true;
}

int env_get_cores( void )
{
    for ( uint32_t i = 0; i < ARRAY_SIZE( global.cpu_cores ); ++i )
    {
        global.cpu_cores[i] = BAD_CPU_INDEX;
    }

#ifdef __linux__

    int num_cpu = sysconf( _SC_NPROCESSORS_ONLN );

    cpu_set_t cpuset;
    CPU_ZERO( &cpuset );
    if ( sched_getaffinity( getpid(), sizeof( cpuset ), &cpuset ) != 0 )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to get process affinity" );
        return NEXT_ERROR;
    }

    int cpu_available = 0;

    int flow_thread_max = FLOW_THREAD_MAX;

    for ( int i = 0; i < num_cpu; ++i )
    {
        if ( CPU_ISSET( i, &cpuset ) )
        {
            if ( cpu_available < flow_thread_max )
            {
                global.cpu_cores[cpu_available] = i;
                cpu_available++;
            }
        }
    }

#else // #ifdef __linux__

    int cpu_available = 4;

#endif // #ifdef __linux__

    char *relay_use_n_minus_two_cores = getenv("RELAY_USE_N_MINUS_2_CORES");
    if ( relay_use_n_minus_two_cores && strcmp(relay_use_n_minus_two_cores, "yes") == 0 )
    {
        // stackpath folks requested we relinquish 2 of the available cpu cores
        cpu_available = cpu_available - 2;
    }

    if ( global.dev )
    {
        if ( cpu_available >= 2 )
        {
            cpu_available = 2;
        }
    }

    if ( cpu_available < 1 )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "no cpus available to pin" );
        return NEXT_ERROR;
    }

    global.cpu_count = cpu_available;

    return NEXT_OK;
}

#ifdef __linux__
int thread_set_affinity( next_thread_t * thread, int core_id )
{
    cpu_set_t cpuset;
    CPU_ZERO( &cpuset );
    CPU_SET( core_id, &cpuset );
    return pthread_setaffinity_np( thread->handle, sizeof( cpu_set_t ), &cpuset ) == 0 ? NEXT_OK : NEXT_ERROR;
}
#endif // #ifdef __linux__

int thread_cancel( next_thread_t * thread )
{
    return pthread_cancel( thread->handle ) == 0 ? NEXT_OK : NEXT_ERROR;
}

int manage_thread_init( void )
{
    manage_thread_context_t * manage = &global.manage_thread_context;

    if ( next_thread_create( &manage->thread, manage_thread, NULL ) != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to create manage thread" );
        return NEXT_ERROR;
    }

#ifdef __linux__
    // first available core is always manage thread
    if ( thread_set_affinity( &manage->thread, global.cpu_cores[0] ) != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to set flow thread cpu affinity" );
        return NEXT_ERROR;
    }
#endif // #ifdef __linux__

    return NEXT_OK;
}

int flow_thread_init( void )
{
    global.flow_thread_count = global.cpu_count - 1;
    if ( global.flow_thread_count == 0 )
    {
        global.flow_thread_count = 1;
    }

    for ( int i = 0; i < global.flow_thread_count; i++ )
    {
        flow_thread_context_t * context = &global.flow_thread_contexts[i];

        if ( next_thread_create( &context->thread, flow_thread, context ) != NEXT_OK )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to create flow thread" );
            return NEXT_ERROR;
        }

#ifdef __linux__
        int cpu_index = ( 1 + i ) % global.cpu_count;
        int core = global.cpu_cores[cpu_index];
        if ( thread_set_affinity( &context->thread, core ) != NEXT_OK )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to set flow thread cpu affinity" );
            return NEXT_ERROR;
        }
#endif // #ifdef __linux__
    }

    relay_printf( NEXT_LOG_LEVEL_INFO, "created %d flow threads on %d cpus", global.flow_thread_count, global.cpu_count );

    return NEXT_OK;
}

static int parse_relay_speed( const char * input, double * output )
{
    if ( !input )
        return NEXT_ERROR;

    if ( strcmp( input, "100M" ) == 0 )
    {
        *output = SPEED_HUNDRED_MB;
        return NEXT_OK;
    }
    else if ( strcmp( input, "1G" ) == 0 )
    {
        *output = SPEED_GIG;
        return NEXT_OK;
    }
    else if ( strcmp( input, "10G" ) == 0 )
    {
        *output = SPEED_TEN_GIG;
        return NEXT_OK;
    }

    return NEXT_ERROR;
}

const char * relay_env_var( const char * name, int index )
{
    const int max_key = 128;
    char key[max_key] = {0};
    if ( index == 0 )
    {
        snprintf( key, max_key - 1, "%s", name );
        const char * value = getenv( key );
        if ( value && value[0] != '\0' )
        {
            return value;
        }
    }

    snprintf( key, max_key - 1, "%s_%d", name, index );

    return getenv( key );
}

int main( int, char ** )
{
    relay_printf( NEXT_LOG_LEVEL_INFO, "starting relay" );

    if ( next_init( ) != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to initialize next lib" );
        exit( 1 );
    }

    next_set_log_level( NEXT_LOG_LEVEL_INFO );

    next_set_print_function( relay_printf );

    {
        const char * dev = getenv( "RELAYDEV" );
        if ( dev && strcmp( dev, "1" ) == 0 )
        {
            relay_printf( NEXT_LOG_LEVEL_INFO, "relay is in dev mode" );
            global.dev = true;
        }
    }

    next_generate_keypair( global.relay_public_key, global.relay_private_key );
    {
        const char * private_key_base64 = relay_env_var( "RELAYPRIVATEKEY", 0 );
        const char * public_key_base64 = relay_env_var( "RELAYPUBLICKEY", 0 );
        const char * router_public_key_base64 = relay_env_var ( "RELAYROUTERPUBLICKEY", 0 );
        if ( private_key_base64 && public_key_base64 && router_public_key_base64 && private_key_base64[0] != '\0' && public_key_base64[0] != '\0' && router_public_key_base64[0] != '\0' )
        {
            if ( next_base64_decode_data( private_key_base64, global.relay_private_key, sizeof(global.relay_private_key) ) != sizeof(global.relay_private_key) )
            {
                relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to base64 decode RELAYPRIVATEKEY env var" );
                exit( 1 );
            }
            if ( next_base64_decode_data( public_key_base64, global.relay_public_key, sizeof(global.relay_public_key) ) != sizeof(global.relay_public_key) )
            {
                relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to base64 decode RELAYPUBLICKEY env var" );
                exit( 1 );
            }
            relay_printf( NEXT_LOG_LEVEL_INFO, "relay public key is %s", public_key_base64 );
            if ( next_base64_decode_data( router_public_key_base64, global.router_public_key, sizeof(global.router_public_key) ) != sizeof(global.router_public_key) )
            {
                relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to base64 decode RELAYROUTERPUBLICKEY env var" );
                exit( 1 );
            }
            relay_printf( NEXT_LOG_LEVEL_INFO, "router public key is %s", router_public_key_base64);
        }
    }

    crypto_auth_keygen( global.relay_ping_key );

    const char * relay_backend = relay_env_var( "RELAYBACKENDHOSTNAME", 0 );
    if ( !relay_backend || relay_backend[0] == '\0' ) {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to get the RELAYBACKENDHOSTNAME env var" );
        exit( 1 );
    }
    global.backend_hostname = relay_backend;

    for ( int i = 0; i < MAX_ENVS; i++ )
    {
        manage_environment_t * env = &manage.envs[i];
        env->idx = i;

        next_address_t address;
        address.type = NEXT_ADDRESS_NONE;
        address.data.ipv6[0] = 0;
        address.data.ipv6[1] = 0;
        address.data.ipv6[2] = 0;
        address.data.ipv6[3] = 0;
        address.data.ipv6[4] = 0;
        address.data.ipv6[5] = 0;
        address.data.ipv6[6] = 0;
        address.data.ipv6[7] = 0;
        address.port = 0;
        env->peers.set_empty_key( address );
        address.port = 1;
        env->peers.set_deleted_key( address );

        const char * relay_name = relay_env_var( "RELAYNAME", i );
        if ( !relay_name || relay_name[0] == '\0' )
            continue;

        next_printf( NEXT_LOG_LEVEL_INFO, "relay name is %s (%d)", relay_name, i );

        manage.env_count++;

        strncpy( env->relay.name, relay_name, sizeof( env->relay.name ) );

        env->relay.id = next_relay_id( env->relay.name );

        const char * master = relay_env_var( "RELAYMASTER", i );
        const char * update_key_base64 = relay_env_var( "RELAYUPDATEKEY", i );
        const char * relay_address = relay_env_var( "RELAYADDRESS", i );
        const char * relay_port = relay_env_var( "RELAYPORT", i );
        const char * relay_speed = relay_env_var( "RELAYSPEED", i );

        if ( master )
        {
            resolver_init( &env->master, master );
        }
        else
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "missing master address env variable" );
            exit( 1 );
        }

        if ( !relay_address || next_address_parse( &env->relay.address, relay_address ) != NEXT_OK )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "error: bad relay address '%s'", relay_address );
            exit( 1 );
        }

        if ( relay_port != NULL )
        {
            global.bind_port = atoi( relay_port );
        }

        relay_printf( NEXT_LOG_LEVEL_INFO, "relay binding to port %d", global.bind_port );

        if ( !update_key_base64 || next_base64_decode_data( update_key_base64, env->relay.update_key, sizeof( env->relay.update_key ) ) <= 0 )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to base64 decode update key from config" );
            exit( 1 );
        }

        // relay speed is optional

        env->relay.speed = DEFAULT_BITS_PER_SEC;
        if ( relay_speed && parse_relay_speed( relay_speed, &env->relay.speed ) != NEXT_OK )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "nic speed '%s' for relay '%s' is not correct", relay_speed, env->relay.name );
            exit( 1 );
        }

        // resolve master address

        int64_t resolve_start = relay_time();
        while ( resolver_address( &env->master )->type == NEXT_ADDRESS_NONE && relay_time() - resolve_start < 30 * NEXT_ONE_SECOND_NS )
        {
            resolver_update( &env->master );
            next_sleep( int64_t( 0.1 * double( NEXT_ONE_SECOND_NS ) ) );
        }

        if ( resolver_address( &env->master )->type == NEXT_ADDRESS_NONE )
        {
            relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to resolve master address '%s'", resolver_address_string( &env->master ) );
            exit( 1 );
        }
    }

    for ( int i = 0; i < manage.env_count; i++ )
    {
        manage_environment_t * env = &manage.envs[i];
        char address[256];
        next_address_to_string( resolver_address( &env->master ), address );
        relay_printf( NEXT_LOG_LEVEL_INFO, "master address: %s, master ip: %s", resolver_address_string( &env->master ), address );
    }

    if ( env_get_cores() != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to get available cores" );
        return 1;
    }

    if ( manage_thread_init() != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to create manage thread" );
        return 1;
    }

    if ( flow_thread_init() != NEXT_OK )
    {
        relay_printf( NEXT_LOG_LEVEL_ERROR, "failed to create flow threads" );
        return 1;
    }

    signal( SIGINT, interrupt_handler );
    signal( SIGTERM, interrupt_handler );
    signal( SIGHUP, interrupt_handler );

    for ( int i = 0; i < global.flow_thread_count; i++ )
    {
        flow_thread_context_t * context = &global.flow_thread_contexts[i];
        next_thread_join( &context->thread );
    }

    {
        msg_manage msg;
        msg.type = MSG_MANAGE_QUIT;
        global.manage_queue_in.enqueue( msg );
    }
    next_thread_join( &global.manage_thread_context.thread );

    next_term();

    return 0;
}
