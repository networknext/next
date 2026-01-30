/*
    Network Next XDP Relay
*/

#include "relay_platform.h"

#include <time.h>
#include <sodium.h>
#include <unistd.h>
#include <pthread.h>
#include <arpa/inet.h>
#include <fcntl.h>
#include <string.h>

// -----------------------------------------------------------------------------------------------------------------------------------------------

static double time_start;

int relay_platform_init()
{
    struct timespec ts;
    clock_gettime( CLOCK_MONOTONIC_RAW, &ts );
    time_start = ts.tv_sec + ( (double) ( ts.tv_nsec ) ) / 1000000000.0;
    int result = sodium_init();
    (void) result;
    return RELAY_OK;
}

double relay_platform_time()
{
    struct timespec ts;
    clock_gettime( CLOCK_MONOTONIC_RAW, &ts );
    double current = ts.tv_sec + ( (double) ( ts.tv_nsec ) ) / 1000000000.0;
    return current - time_start;
}

void relay_platform_sleep( double time )
{
    usleep( (int) ( time * 1000000 ) );
}

// -----------------------------------------------------------------------------------------------------------------------------------------------

void relay_platform_random_bytes( uint8_t * buffer, int bytes )
{
    randombytes_buf( buffer, bytes );
}

// -----------------------------------------------------------------------------------------------------------------------------------------------

int relay_platform_parse_address( char * address_string, uint32_t * address, uint16_t * port )
{
    assert( address_string );
    assert( address );
    assert( port );

    *port = 0;

    int address_string_length = (int) strlen( address_string );

    int base_index = address_string_length - 1;

    for ( int i = 0; i < 6; ++i )
    {
        const int index = base_index - i;
        if ( index < 0 )
            break;
        if ( address_string[index] == ':' )
        {
            *port = (uint16_t)( atoi( &address_string[index + 1] ) );
            address_string[index] = '\0';
        }
    }

    if ( inet_pton( AF_INET, address_string, address ) != 1 ) 
        return RELAY_ERROR;

    *address = htonl( *address );

    return RELAY_OK;
}

// -----------------------------------------------------------------------------------------------------------------------------------------------

struct relay_platform_socket_t * relay_platform_socket_create( uint32_t address, uint16_t port, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size )
{
    struct relay_platform_socket_t * s = (struct relay_platform_socket_t*) malloc( sizeof( struct relay_platform_socket_t ) );

    assert( s );

    // create socket

    s->type = socket_type;

    s->handle = socket( AF_INET, SOCK_DGRAM, IPPROTO_UDP );

    if ( s->handle < 0 )
    {
        printf( "error: failed to create socket\n" );
        return NULL;
    }

    // increase socket send and receive buffer sizes

    if ( setsockopt( s->handle, SOL_SOCKET, SO_SNDBUF, (char*)( &send_buffer_size ), sizeof( int ) ) != 0 )
    {
        printf( "failed to set socket send buffer size to %d\n", send_buffer_size );
        relay_platform_socket_destroy( s );
        return NULL;
    }

    if ( setsockopt( s->handle, SOL_SOCKET, SO_RCVBUF, (char*)( &receive_buffer_size ), sizeof( int ) ) != 0 )
    {
        printf( "failed to set socket receive buffer size to %d\n", receive_buffer_size );
        relay_platform_socket_destroy( s );
        return NULL;
    }

    // bind to port

    struct sockaddr_in socket_address = {0};
    socket_address.sin_family = AF_INET;
    socket_address.sin_addr.s_addr = htonl( address );
    socket_address.sin_port = htons( port );
    if ( bind( s->handle, (struct sockaddr*) &socket_address, sizeof( socket_address ) ) < 0 )
    {
        printf( "failed to bind socket\n" );
        relay_platform_socket_destroy( s );
        return NULL;
    }

    // set don't fragment bit

    int val = IP_PMTUDISC_DO;

    setsockopt( s->handle, IPPROTO_IP, IP_MTU_DISCOVER, &val, sizeof(val) );

    // set non-blocking io and receive timeout

    if ( socket_type == RELAY_PLATFORM_SOCKET_NON_BLOCKING )
    {
        if ( fcntl( s->handle, F_SETFL, O_NONBLOCK, 1 ) == -1 )
        {
            printf( "failed to set socket to non-blocking\n" );
            relay_platform_socket_destroy( s );
            return NULL;
        }
    }
    else if ( timeout_seconds > 0.0f )
    {
        // set receive timeout
        struct timeval tv;
        tv.tv_sec = 0;
        tv.tv_usec = (int) ( timeout_seconds * 1000000.0f );
        if ( setsockopt( s->handle, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof( tv ) ) < 0 )
        {
            printf( "failed to set socket receive timeout\n" );
            relay_platform_socket_destroy( s );
            return NULL;
        }
    }
    else
    {
        // socket is blocking with no timeout
    }

    return s;
}

void relay_platform_socket_destroy( struct relay_platform_socket_t * socket )
{
    assert( socket );
    if ( socket->handle != 0 )
    {
        close( socket->handle );
    }
    free( socket );
}

void relay_platform_socket_send_packet( struct relay_platform_socket_t * socket, uint32_t to_address, uint16_t to_port, const void * packet_data, int packet_bytes )
{
    assert( socket );
    assert( packet_data );
    assert( packet_bytes > 0 );

    struct sockaddr_in socket_address = {0};
    socket_address.sin_family = AF_INET;
    socket_address.sin_addr.s_addr = htonl( to_address );
    socket_address.sin_port = htons( to_port );

    sendto( socket->handle, (const char*)( packet_data ), packet_bytes, 0, (struct sockaddr*)( &socket_address ), sizeof(struct sockaddr_in) );
}

int relay_platform_socket_receive_packet( struct relay_platform_socket_t * socket, uint32_t * from_address, uint16_t * from_port, void * packet_data, int max_packet_size )
{
    assert( socket );
    assert( from_address );
    assert( from_port );
    assert( packet_data );
    assert( max_packet_size > 0 );

    struct sockaddr_storage sockaddr_from = {0};

    socklen_t from_length = sizeof( sockaddr_from );

    int result = (int) recvfrom( socket->handle, (char*) packet_data, max_packet_size, socket->type == RELAY_PLATFORM_SOCKET_NON_BLOCKING ? MSG_DONTWAIT : 0, (struct sockaddr*) &sockaddr_from, &from_length );
    if ( result <= 0 )
        return 0;

    if ( sockaddr_from.ss_family == AF_INET )
    {
        struct sockaddr_in * addr_ipv4 = (struct sockaddr_in*) &sockaddr_from;
        *from_address = ntohl( addr_ipv4->sin_addr.s_addr );
        *from_port = ntohs( addr_ipv4->sin_port );
        return result;
    }

    return 0;
}

// -----------------------------------------------------------------------------------------------------------------------------------------------

struct relay_platform_thread_t * relay_platform_thread_create( relay_platform_thread_func_t * thread_function, void * arg )
{
    struct relay_platform_thread_t * thread = (struct relay_platform_thread_t*) malloc( sizeof( struct relay_platform_thread_t) );

    assert( thread );

    if ( pthread_create( &thread->handle, NULL, thread_function, arg ) != 0 )
    {
        free( thread );
        return NULL;
    }

    return thread;
}

void relay_platform_thread_join( struct relay_platform_thread_t * thread )
{
    assert( thread );
    pthread_join( thread->handle, NULL );
}

void relay_platform_thread_destroy( struct relay_platform_thread_t * thread )
{
    assert( thread );
    free( thread );
}

// -----------------------------------------------------------------------------------------------------------------------------------------------

struct relay_platform_mutex_t * relay_platform_mutex_create()
{
    struct relay_platform_mutex_t * mutex = (struct relay_platform_mutex_t*) malloc( sizeof(struct relay_platform_mutex_t) ); assert( mutex );

    assert( mutex );

    pthread_mutexattr_t attr;
    pthread_mutexattr_init(&attr);
    pthread_mutexattr_settype( &attr, 0 );
    int result = pthread_mutex_init( &mutex->handle, &attr );
    pthread_mutexattr_destroy( &attr );

    if ( result != 0 )
    {
        free( mutex );
        return NULL;
    }

    return mutex;
}

void relay_platform_mutex_acquire( struct relay_platform_mutex_t * mutex )
{
    assert( mutex );
    pthread_mutex_lock( &mutex->handle );
}

void relay_platform_mutex_release( struct relay_platform_mutex_t * mutex )
{
    assert( mutex );
    pthread_mutex_unlock( &mutex->handle );
}

void relay_platform_mutex_destroy( struct relay_platform_mutex_t * mutex )
{
    assert( mutex );
    pthread_mutex_destroy( &mutex->handle );
    free( mutex );
}

// -----------------------------------------------------------------------------------------------------------------------------------------------
