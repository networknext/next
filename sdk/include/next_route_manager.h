/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_ROUTE_MANAGER_H
#define NEXT_ROUTE_MANAGER_H

#include "next.h"

struct next_route_manager_t;

void next_route_manager_initialize_sentinels( next_route_manager_t * route_manager );

void next_route_manager_verify_sentinels( next_route_manager_t * route_manager );

next_route_manager_t * next_route_manager_create( void * context );

void next_route_manager_destroy( next_route_manager_t * route_manager );

void next_route_manager_reset( next_route_manager_t * route_manager );

void next_route_manager_fallback_to_direct( next_route_manager_t * route_manager, uint32_t flags );

void next_route_manager_direct_route( next_route_manager_t * route_manager, bool quiet );

void next_route_manager_begin_next_route( next_route_manager_t * route_manager, int num_tokens, uint8_t * tokens, const uint8_t * client_secret_key, const uint8_t * magic, const next_address_t * client_external_address );

void next_route_manager_continue_next_route( next_route_manager_t * route_manager, int num_tokens, uint8_t * tokens, const uint8_t * client_secret_key, const uint8_t * magic, const next_address_t * client_external_address );

void next_route_manager_update( next_route_manager_t * route_manager, int update_type, int num_tokens, uint8_t * tokens, const uint8_t * client_secret_key, const uint8_t * magic, const next_address_t * client_external_address );

bool next_route_manager_has_network_next_route( next_route_manager_t * route_manager );

uint64_t next_route_manager_next_send_sequence( next_route_manager_t * route_manager );

bool next_route_manager_prepare_send_packet( next_route_manager_t * route_manager, uint64_t sequence, next_address_t * to, const uint8_t * payload_data, int payload_bytes, uint8_t * packet_data, int * packet_bytes, const uint8_t * magic, const next_address_t * client_external_address );

bool next_route_manager_process_server_to_client_packet( next_route_manager_t * route_manager, uint8_t packet_type, uint8_t * packet_data, int packet_bytes, uint64_t * payload_sequence );

void next_route_manager_check_for_timeouts( next_route_manager_t * route_manager );

bool next_route_manager_send_route_request( next_route_manager_t * route_manager, next_address_t * to, uint8_t * packet_data, int * packet_bytes );

bool next_route_manager_send_continue_request( next_route_manager_t * route_manager, next_address_t * to, uint8_t * packet_data, int * packet_bytes );

void next_route_manager_get_pending_route_data( next_route_manager_t * route_manager, bool * fallback_to_direct, bool * pending_route, uint64_t * pending_route_session_id, uint8_t * pending_route_session_version, uint8_t * pending_route_private_key );

void next_route_manager_confirm_pending_route( next_route_manager_t * route_manager, int * route_kbps_up, int * route_kbps_down );

void next_route_manager_get_current_route_data( next_route_manager_t * route_manager, bool * fallback_to_direct, bool * pending_route, bool * pending_continue, uint64_t * pending_route_session_id, uint8_t * pending_route_session_version, uint8_t * current_route_private_key );

void next_route_manager_confirm_continue_route( next_route_manager_t * route_manager );

bool next_route_manager_get_fallback_to_direct( next_route_manager_t * route_manager );

uint32_t next_route_manager_get_flags( next_route_manager_t * route_manager );

void next_route_manager_get_next_route_data( next_route_manager_t * route_manager, uint64_t * session_id, uint8_t * session_version, next_address_t * to, uint8_t * private_key );

#endif // #ifndef NEXT_ROUTE_MANAGER_H
