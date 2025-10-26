/*
    Network Next XDP Relay
*/

#ifndef RELAY_MAIN_H
#define RELAY_MAIN_H

#include "relay.h"
#include "relay_messages.h"

struct main_t
{
    void * curl;
    uint64_t start_time;
    uint64_t current_timestamp;
    const char * relay_backend_url;
    uint8_t * update_response_memory;
    uint32_t relay_public_address;
    uint32_t relay_internal_address;
    uint16_t relay_port;
    bool initialized;
    bool shutting_down;
    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    uint8_t relay_secret_key[RELAY_SECRET_KEY_BYTES];
    uint8_t relay_backend_public_key[RELAY_BACKEND_PUBLIC_KEY_BYTES];
    uint64_t pings_sent;
    uint64_t bytes_sent;
    int stats_fd;
    int state_fd;
    int config_fd;
    int session_map_fd;
    int whitelist_map_fd;
    struct relay_set relay_ping_set;
    struct relay_queue_t * control_queue;
    struct relay_platform_mutex_t * control_mutex;
    struct relay_queue_t * stats_queue;
    struct relay_platform_mutex_t * stats_mutex;
    struct relay_ping_stats_t ping_stats;
    uint64_t last_stats_packets_sent;
    uint64_t last_stats_packets_received;
    uint64_t last_stats_bytes_sent;
    uint64_t last_stats_bytes_received;
    uint64_t last_stats_client_pings_received;
    uint64_t last_stats_server_pings_received;
    uint64_t last_stats_relay_pings_received;
    double last_stats_time;
};

struct config_t;
struct bpf_t;

int main_init( struct main_t * main, struct config_t * config, struct bpf_t * bpf );

int main_run( struct main_t * main );

void main_shutdown( struct main_t * main );

#endif // #ifndef RELAY_MAIN_H
