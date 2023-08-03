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

#ifndef NEXT_CLIENT_H
#define NEXT_CLIENT_H

#include "next_address.h"

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

struct next_client_internal_t;

void next_client_internal_initialize_sentinels( next_client_internal_t * client );

void next_client_internal_verify_sentinels( next_client_internal_t * client );

void next_client_internal_destroy( next_client_internal_t * client );

next_client_internal_t * next_client_internal_create( void * context, const char * bind_address_string );

void next_client_internal_destroy( next_client_internal_t * client );

int next_client_internal_send_packet_to_server( next_client_internal_t * client, uint8_t packet_id, void * packet_object );

void next_client_internal_process_network_next_packet( next_client_internal_t * client, const next_address_t * from, uint8_t * packet_data, int packet_bytes, double packet_receive_time );

void next_client_internal_process_passthrough_packet( next_client_internal_t * client, const next_address_t * from, uint8_t * packet_data, int packet_bytes );

void next_client_internal_block_and_receive_packet( next_client_internal_t * client );

bool next_client_internal_pump_commands( next_client_internal_t * client );

void next_client_internal_update_stats( next_client_internal_t * client );

void next_client_internal_update_direct_pings( next_client_internal_t * client );

void next_client_internal_update_next_pings( next_client_internal_t * client );

void next_client_internal_send_pings_to_near_relays( next_client_internal_t * client );

void next_client_internal_update_fallback_to_direct( next_client_internal_t * client );

void next_client_internal_update_route_manager( next_client_internal_t * client );

void next_client_internal_update_upgrade_response( next_client_internal_t * client );

void next_client_internal_update( next_client_internal_t * client );

// ---------------------------------------------------------------

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

const next_client_stats_t * next_client_stats( next_client_t * client );

const next_address_t * next_client_server_address( next_client_t * client );

void next_client_counters( next_client_t * client, uint64_t * counters );

// ---------------------------------------------------------------

#endif // #ifndef NEXT_H
