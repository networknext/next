/*
    Network Next XDP Relay
*/

#include "relay_ping.h"
#include "relay_queue.h"
#include "relay_messages.h"
#include "relay_platform.h"
#include "relay_encoding.h"
#include "relay_ping_history.h"
#include "relay_manager.h"
#include "relay_shared.h"
#include "relay_config.h"
#include "relay_main.h"
#include "relay_bpf.h"
#include "relay_endian.h"

#include <stdlib.h>
#include <sodium.h>
#include <inttypes.h>

// --------------------------------------------------------------------------------------------------------------------------------------------------

int ping_init( struct ping_t * ping, struct config_t * config, struct main_t * main, struct bpf_t * bpf )
{
    struct relay_platform_socket_t * ping_socket = relay_platform_socket_create( 0, config->relay_port, RELAY_PLATFORM_SOCKET_BLOCKING, 0.1f, 100 * 1024, 100 * 1024 );
    if ( ping_socket == NULL )
    {
        printf( "\ncould not create ping socket\n\n" );
        return RELAY_ERROR;
    }

    ping->socket = ping_socket;
    ping->relay_port = config->relay_port;
    ping->relay_public_address = config->relay_public_address;
    ping->relay_internal_address = config->relay_internal_address;
    ping->relay_manager = relay_manager_create();
    ping->control_queue = main->control_queue;
    ping->control_mutex = main->control_mutex;
    ping->stats_queue = main->stats_queue;
    ping->stats_mutex = main->stats_mutex;
#ifdef COMPILE_WITH_BPF
    ping->relay_map_fd = bpf->relay_map_fd;
#endif // #ifdef COMPILE_WITH_BPF

    assert( ping->control_queue );
    assert( ping->control_mutex );
    assert( ping->stats_queue );
    assert( ping->stats_mutex );

    return RELAY_OK;
}

int ping_start_thread( struct ping_t * ping )
{
    printf( "Starting ping thread\n" );

    ping->thread = relay_platform_thread_create( ping_thread_function, ping );
    if ( !ping->thread )
    {
        printf( "\nerror: could not create ping thread\n\n" );
        return RELAY_ERROR;
    }    

    return RELAY_OK;
}

void ping_join_thread( struct ping_t * ping )
{
    if ( ping->thread )
    {
        printf( "Waiting for ping thread\n" );

        relay_platform_thread_join( ping->thread );
    }
}

void ping_shutdown( struct ping_t * ping )
{
    if ( ping->thread )
    {
        relay_platform_thread_destroy( ping->thread );
    }

    if ( ping->socket )
    {
        relay_platform_socket_destroy( ping->socket );
    }    

    if ( ping->relay_manager )
    {
        relay_manager_destroy( ping->relay_manager );
    }
}

// --------------------------------------------------------------------------------------------------------------------------------------------------

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

// --------------------------------------------------------------------------------------------------------------------------------------------------

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

// --------------------------------------------------------------------------------------------------------------------------------------------------

extern bool quit;

void * ping_thread_function( void * context )
{
    struct ping_t * ping = (struct ping_t*) context;

    uint8_t packet_data[RELAY_MAX_PACKET_BYTES];

    double last_update_time = 0.0;

    double last_ping_stats_time = 0.0;

    while ( !quit )
    {
        uint32_t from_address = 0;
        uint16_t from_port = 0;

        int packet_bytes = relay_platform_socket_receive_packet( ping->socket, &from_address, &from_port, packet_data, sizeof(packet_data) );

        double current_time = relay_platform_time();

        // process relay pong packets immediately

        if ( packet_bytes == 18 + 8 && packet_data[0] == RELAY_PONG_PACKET )
        {
            const uint8_t * p = packet_data + 18;
            uint64_t sequence = relay_read_uint64( &p );
            relay_manager_process_pong( ping->relay_manager, from_address, from_port, sequence );
        }

        // run update logic ~100 times per-second

        if ( last_update_time + 0.01 <= current_time )
        {
            last_update_time = current_time;

            // process control messages

            while ( !quit )
            {
                relay_platform_mutex_acquire( ping->control_mutex );
                struct relay_control_message * message = (struct relay_control_message*) relay_queue_pop( ping->control_queue );
                relay_platform_mutex_release( ping->control_mutex );
     
                if ( !message )
                    break;

                ping->current_timestamp = message->current_timestamp;
                ping->has_ping_key = true;
                memcpy( ping->ping_key, message->ping_key, RELAY_PING_KEY_BYTES );
                memcpy( ping->current_magic, message->current_magic, 8 );

                if ( message->new_relays.num_relays > 0 )
                {
                    printf( "-------------------------------------------------------\n" );
                    for ( int i = 0; i < message->new_relays.num_relays; i++ )
                    {
#ifdef COMPILE_WITH_BPF
                        __u64 key = ( ( (__u64)relay_htonl(message->new_relays.address[i]) ) << 32 ) | relay_htons(message->new_relays.port[i]);
                        __u64 value = 1;
                        if ( bpf_map_update_elem( ping->relay_map_fd, &key, &value, BPF_NOEXIST ) == 0 )
#endif // #ifdef COMPILE_WITH_BPF
                        {
                            printf( "new relay %d.%d.%d.%d:%d\n", 
                                ((uint8_t*)&message->new_relays.address[i])[3], 
                                ((uint8_t*)&message->new_relays.address[i])[2], 
                                ((uint8_t*)&message->new_relays.address[i])[1], 
                                ((uint8_t*)&message->new_relays.address[i])[0], 
                                message->new_relays.port[i] );
                        }
                    }
                    printf( "-------------------------------------------------------\n" );

                    fflush( stdout );
                }

                if ( message->delete_relays.num_relays > 0 )
                {
                    printf( "-------------------------------------------------------\n" );
                    for ( int i = 0; i < message->delete_relays.num_relays; i++ )
                    {
#ifdef COMPILE_WITH_BPF
                        __u64 key = ( ( (__u64)relay_htonl(message->delete_relays.address[i]) ) << 32 ) | relay_htons(message->delete_relays.port[i]);
                        __u64 value = 1;
                        if ( bpf_map_delete_elem( ping->relay_map_fd, &key ) == 0 )
#endif // #ifdef COMPILE_WITH_BPF
                        {
                            printf( "delete relay %d.%d.%d.%d:%d\n", 
                                ((uint8_t*)&message->delete_relays.address[i])[3], 
                                ((uint8_t*)&message->delete_relays.address[i])[2], 
                                ((uint8_t*)&message->delete_relays.address[i])[1], 
                                ((uint8_t*)&message->delete_relays.address[i])[0], 
                                message->delete_relays.port[i] );
                        }
                    }
                    printf( "-------------------------------------------------------\n" );

                    fflush( stdout );
                }

                relay_manager_update( ping->relay_manager, &message->new_relays, &message->delete_relays );

                free( message );
            }

            // send ping packets

            if ( ping->has_ping_key )
            {
                uint64_t expire_timestamp = ping->current_timestamp + 30;

                for ( int i = 0; i < ping->relay_manager->num_relays; ++i )
                {
                    if ( ping->relay_manager->relay_last_ping_time[i] + RELAY_PING_TIME <= current_time )
                    {
                        // send relay ping packet

                        struct ping_token_data token_data;

                        token_data.source_address = ping->relay_manager->relay_internal[i] ? relay_htonl( ping->relay_internal_address ) : relay_htonl( ping->relay_public_address );
                        token_data.source_port = relay_htons( ping->relay_port );
                        token_data.dest_address = relay_htonl( ping->relay_manager->relay_addresses[i] );
                        token_data.dest_port = relay_htons( ping->relay_manager->relay_ports[i] );
                        token_data.expire_timestamp = expire_timestamp;

                        memcpy( token_data.ping_key, ping->ping_key, RELAY_PING_KEY_BYTES );

                        uint8_t ping_token[RELAY_PING_TOKEN_BYTES];

                        crypto_hash_sha256( ping_token, (const unsigned char*) &token_data, sizeof(struct ping_token_data) );

                        uint8_t packet_data[256];

                        packet_data[0] = RELAY_PING_PACKET;

                        uint8_t * a = packet_data + 1;
                        uint8_t * b = packet_data + 3;
                        uint8_t * p = packet_data + 18;

                        uint64_t sequence = relay_ping_history_ping_sent( ping->relay_manager->relay_ping_history[i], current_time );

                        relay_write_uint64( &p, sequence );
                        relay_write_uint64( &p, expire_timestamp );
                        relay_write_uint8( &p, ping->relay_manager->relay_internal[i] );
                        relay_write_bytes( &p, ping_token, RELAY_PING_TOKEN_BYTES );

                        int packet_length = p - packet_data;

                        uint8_t to_address_data[4];
                        uint8_t from_address_data[4];

                        relay_address_data( relay_htonl( ping->relay_manager->relay_addresses[i] ), to_address_data );

                        if ( !ping->relay_manager->relay_internal[i] )
                        {
                            relay_address_data( relay_htonl( ping->relay_public_address ), from_address_data );
                        }
                        else
                        {
                            relay_address_data( relay_htonl( ping->relay_internal_address ), from_address_data );
                        }

                        relay_generate_pittle( a, from_address_data, to_address_data, packet_length );
                        relay_generate_chonkle( b, ping->current_magic, from_address_data, to_address_data, packet_length );

                        relay_platform_socket_send_packet( ping->socket, ping->relay_manager->relay_addresses[i], ping->relay_manager->relay_ports[i], packet_data, packet_length );

                        ping->relay_manager->relay_last_ping_time[i] = current_time;

                        ping->bytes_sent += 8 + 20 + 18 + 1 + 8 + 8 + RELAY_PING_TOKEN_BYTES;
                        ping->pings_sent ++;
                    }
                }
            }

            // post ping stats to main thread

            if ( last_ping_stats_time + 0.1 <= current_time )
            {
                last_ping_stats_time = current_time;

                struct relay_stats_message * message = (struct relay_stats_message*) malloc( sizeof(struct relay_stats_message) );

                relay_manager_get_ping_stats( ping->relay_manager, &message->ping_stats );

                message->pings_sent = ping->pings_sent;
                message->bytes_sent = ping->bytes_sent;

                relay_platform_mutex_acquire( ping->stats_mutex );
                relay_queue_push( ping->stats_queue, message );
                relay_platform_mutex_release( ping->stats_mutex );
            }  
        }
    }

    return NULL;
}
