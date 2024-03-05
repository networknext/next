/*
    Network Next XDP Relay
*/

#ifndef RELAY_PLATFORM_H
#define RELAY_PLATFORM_H

#include "relay.h"
#include <pthread.h>

// -----------------------------------------------------------------------------------------------------------------------------------------------

#define RELAY_PLATFORM_SOCKET_NON_BLOCKING       0
#define RELAY_PLATFORM_SOCKET_BLOCKING           1

typedef int relay_platform_socket_handle_t;

struct relay_platform_socket_t
{
    int type;
    relay_platform_socket_handle_t handle;
};

struct relay_platform_thread_t
{
    pthread_t handle;
};

typedef void* (relay_platform_thread_func_t)(void*);

struct relay_platform_mutex_t
{
    pthread_mutex_t handle;
};

// -----------------------------------------------------------------------------------------------------------------------------------------------

int relay_platform_init();

double relay_platform_time();

void relay_platform_sleep( double time );

// -----------------------------------------------------------------------------------------------------------------------------------------------

void relay_platform_random_bytes( uint8_t * buffer, int bytes );

// -----------------------------------------------------------------------------------------------------------------------------------------------

int relay_platform_parse_address( char * address_string, uint32_t * address, uint16_t * port );

// -----------------------------------------------------------------------------------------------------------------------------------------------

struct relay_platform_socket_t * relay_platform_socket_create( uint32_t address, uint16_t port, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size );

void relay_platform_socket_destroy( struct relay_platform_socket_t * socket );

void relay_platform_socket_send_packet( struct relay_platform_socket_t * socket, uint32_t to_address, uint16_t to_port, const void * packet_data, int packet_bytes );

int relay_platform_socket_receive_packet( struct relay_platform_socket_t * socket, uint32_t * from_address, uint16_t * from_port, void * packet_data, int max_packet_size );

// -----------------------------------------------------------------------------------------------------------------------------------------------

struct relay_platform_thread_t * relay_platform_thread_create( relay_platform_thread_func_t * thread_function, void * arg );

void relay_platform_thread_join( struct relay_platform_thread_t * thread );

void relay_platform_thread_destroy( struct relay_platform_thread_t * thread );

bool relay_platform_thread_set_high_priority( struct relay_platform_thread_t * thread );

// -----------------------------------------------------------------------------------------------------------------------------------------------

struct relay_platform_mutex_t * relay_platform_mutex_create();

void relay_platform_mutex_acquire( struct relay_platform_mutex_t * mutex );

void relay_platform_mutex_release( struct relay_platform_mutex_t * mutex );

void relay_platform_mutex_destroy( struct relay_platform_mutex_t * mutex );

// -----------------------------------------------------------------------------------------------------------------------------------------------

#endif // #ifndef RELAY_PLATFORM_H
