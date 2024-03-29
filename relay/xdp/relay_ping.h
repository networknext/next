/*
    Network Next XDP Relay
*/

#ifndef RELAY_PING_H
#define RELAY_PING_H

#include "relay.h"

// --------------------------------------------------------------------------------------------------------------

struct relay_ping_stats_t
{
    int num_relays;
    uint64_t relay_ids[MAX_RELAYS];
    float relay_rtt[MAX_RELAYS];
    float relay_jitter[MAX_RELAYS];
    float relay_packet_loss[MAX_RELAYS];
};

// --------------------------------------------------------------------------------------------------------------

struct relay_ping_history_stats_t
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
    struct relay_ping_history_entry_t entries[RELAY_PING_HISTORY_SIZE];
};

void relay_ping_history_clear( struct relay_ping_history_t * history );

uint64_t relay_ping_history_ping_sent( struct relay_ping_history_t * history, double time );

void relay_ping_history_pong_received( struct relay_ping_history_t * history, uint64_t sequence, double time );

void relay_ping_history_get_stats( const struct relay_ping_history_t * history, double start, double end, struct relay_ping_history_stats_t * stats, double ping_safety );

// --------------------------------------------------------------------------------------------------------------

struct ping_t
{
    int relay_port;
    uint32_t relay_public_address;
    uint32_t relay_internal_address;
    struct relay_platform_socket_t * socket;
    struct relay_manager_t * relay_manager;
    uint64_t current_timestamp;
    uint8_t current_magic[8];
    bool has_ping_key;
    uint8_t ping_key[RELAY_PING_KEY_BYTES];
    uint64_t pings_sent;
    uint64_t bytes_sent;
    struct relay_queue_t * control_queue;
    struct relay_platform_mutex_t * control_mutex;
    struct relay_queue_t * stats_queue;
    struct relay_platform_mutex_t * stats_mutex;
    int relay_map_fd;
    struct relay_platform_thread_t * thread;
};

struct config_t;
struct main_t;
struct bpf_t;

int ping_init( struct ping_t * ping, struct config_t * config, struct main_t * main, struct bpf_t * bpf );

void ping_shutdown( struct ping_t * ping );

int ping_start_thread( struct ping_t * ping );

void ping_join_thread( struct ping_t * ping );

void * ping_thread_function( void * context );

// --------------------------------------------------------------------------------------------------------------

#endif // #ifndef RELAY_PING_H
