/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_CLIENT_H
#define NEXT_CLIENT_H

#include "next_address.h"

struct next_client_t;

void next_client_initialize_sentinels( next_client_t * client );

void next_client_verify_sentinels( next_client_t * client );

next_client_t * next_client_create( void * context, const char * bind_address, void (*packet_received_callback)( next_client_t * client, void * context, const struct next_address_t * from, const uint8_t * packet_data, int packet_bytes ) );

void next_client_destroy( next_client_t * client );

void next_client_open_session( next_client_t * client, const char * server_address_string );

bool next_client_is_session_open( next_client_t * client );

int next_client_state( next_client_t * client );

void next_client_close_session( next_client_t * client );

void next_client_update( next_client_t * client );

bool next_client_ready( next_client_t * client );

bool next_client_fallback_to_direct( struct next_client_t * client );

void next_client_send_packet( next_client_t * client, const uint8_t * packet_data, int packet_bytes );

void next_client_send_packet_direct( next_client_t * client, const uint8_t * packet_data, int packet_bytes );

void next_client_send_packet_raw( next_client_t * client, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes );

uint16_t next_client_port( next_client_t * client );

uint64_t next_client_session_id( next_client_t * client );

uint64_t next_client_server_id( next_client_t * client );

void next_client_report_session( next_client_t * client );

const next_client_stats_t * next_client_stats( next_client_t * client );

const next_address_t * next_client_server_address( next_client_t * client );

void next_client_counters( next_client_t * client, uint64_t * counters );

void next_client_game_stats( next_client_t * client, float game_rtt, float game_jitter, float game_packet_loss );

#endif // #ifndef NEXT_CLIENT_H
