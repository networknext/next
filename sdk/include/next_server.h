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

#ifndef NEXT_SERVER_H
#define NEXT_SERVER_H

#include "next.h"
#include "next_config.h"
#include "next_address.h"
#include "next_session_manager.h"

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

struct next_server_internal_t;

void next_server_internal_initialize_sentinels( next_server_internal_t * server );

void next_server_internal_verify_sentinels( next_server_internal_t * server );

void next_server_internal_resolve_hostname( next_server_internal_t * server );

void next_server_internal_autodetect( next_server_internal_t * server );

void next_server_internal_initialize( next_server_internal_t * server );

void next_server_internal_destroy( next_server_internal_t * server );

next_server_internal_t * next_server_internal_create( void * context, const char * server_address_string, const char * bind_address_string, const char * datacenter_string );

void next_server_internal_destroy( next_server_internal_t * server );

void next_server_internal_quit( next_server_internal_t * server );

void next_server_internal_send_packet_to_address( next_server_internal_t * server, const next_address_t * address, const uint8_t * packet_data, int packet_bytes );

void next_server_internal_send_packet_to_backend( next_server_internal_t * server, const uint8_t * packet_data, int packet_bytes );

int next_server_internal_send_packet( next_server_internal_t * server, const next_address_t * to_address, uint8_t packet_id, void * packet_object );

next_session_entry_t * next_server_internal_process_client_to_server_packet( next_server_internal_t * server, uint8_t packet_type, uint8_t * packet_data, int packet_bytes );

void next_server_internal_update_route( next_server_internal_t * server );

void next_server_internal_update_pending_upgrades( next_server_internal_t * server );

void next_server_internal_update_sessions( next_server_internal_t * server );

void next_server_internal_update_flush( next_server_internal_t * server );

void next_server_internal_process_network_next_packet( next_server_internal_t * server, const next_address_t * from, uint8_t * packet_data, int begin, int end );

void next_server_internal_process_passthrough_packet( next_server_internal_t * server, const next_address_t * from, uint8_t * packet_data, int packet_bytes );

void next_server_internal_block_and_receive_packet( next_server_internal_t * server );

void next_server_internal_upgrade_session( next_server_internal_t * server, const next_address_t * address, uint64_t session_id, uint64_t user_hash );

void next_server_internal_session_events( next_server_internal_t * server, const next_address_t * address, uint64_t session_events );

void next_server_internal_match_data( next_server_internal_t * server, const next_address_t * address, uint64_t match_id, const double * match_values, int num_match_values );

void next_server_internal_flush_session_update( next_server_internal_t * server );

void next_server_internal_flush_match_data( next_server_internal_t * server );

void next_server_internal_flush( next_server_internal_t * server );

void next_server_internal_pump_commands( next_server_internal_t * server );

void next_server_internal_update_init( next_server_internal_t * server );

void next_server_internal_backend_update( next_server_internal_t * server );

// ---------------------------------------------------------------

struct next_server_t;

void next_server_initialize_sentinels( next_server_t * server );

void next_server_verify_sentinels( next_server_t * server );

next_server_t * next_server_create( void * context, const char * server_address, const char * bind_address, const char * datacenter, void (*packet_received_callback)( next_server_t * server, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes ) );

uint16_t next_server_port( next_server_t * server );

const next_address_t * next_server_address( next_server_t * server );

void next_server_destroy( next_server_t * server );

void next_server_update( next_server_t * server );

uint64_t next_server_upgrade_session( next_server_t * server, const next_address_t * address, const char * user_id );

bool next_server_session_upgraded( next_server_t * server, const next_address_t * address );

void next_server_send_packet_to_address( next_server_t * server, const next_address_t * address, const uint8_t * packet_data, int packet_bytes );

void next_server_send_packet( next_server_t * server, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes );

void next_server_send_packet_direct( next_server_t * server, const next_address_t * to_address, const uint8_t * packet_data, int packet_bytes );

void next_server_send_packet_raw( struct next_server_t * server, const struct next_address_t * to_address, const uint8_t * packet_data, int packet_bytes );

bool next_server_stats( next_server_t * server, const next_address_t * address, next_server_stats_t * stats );

bool next_server_ready( next_server_t * server );

const char * next_server_datacenter( next_server_t * server );

void next_server_session_event( struct next_server_t * server, const struct next_address_t * address, uint64_t session_events );

void next_server_match( struct next_server_t * server, const struct next_address_t * address, const char * match_id, const double * match_values, int num_match_values );

void next_server_flush( struct next_server_t * server );

void next_server_set_packet_receive_callback( struct next_server_t * server, void (*callback) ( void * data, next_address_t * from, uint8_t * packet_data, int * begin, int * end ), void * callback_data );

void next_server_set_send_packet_to_address_callback( struct next_server_t * server, int (*callback) ( void * data, const next_address_t * from, const uint8_t * packet_data, int packet_bytes ), void * callback_data );

void next_server_set_payload_receive_callback( struct next_server_t * server, int (*callback) ( void * data, const next_address_t * client_address, const uint8_t * payload_data, int payload_bytes ), void * callback_data );

bool next_server_direct_only( struct next_server_t * server );

// ---------------------------------------------------------------

#endif // #ifndef NEXT_H
