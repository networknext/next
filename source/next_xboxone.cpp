/*
    Network Next SDK. Copyright © 2017 - 2020 Network Next, Inc.

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

// IMPORTANT: you must compile this file with /ZW to get windows runtime components

#include "next_xboxone.h"

#if NEXT_PLATFORM == NEXT_PLATFORM_XBOX_ONE

#include <sodium.h>

#define NOMINMAX
#define _WINSOCK_DEPRECATED_NO_WARNINGS

#pragma pack(push, 8)
#include <windows.h>
#include <winsock2.h>
#include <ws2tcpip.h>
#include <ws2ipdef.h>
#include <malloc.h>
#include <time.h>
#include <bcrypt.h> // random
#pragma pack(pop)

#pragma comment( lib, "WS2_32.lib" )

#ifdef SetPort
#undef SetPort
#endif // #ifdef SetPort

NEXT_PACK_PUSH()

extern void * next_malloc( void * context, size_t bytes );

extern void next_free( void * context, void * p );

static BCRYPT_ALG_HANDLE bcrypt_algorithm_provider;

static int connection_type = NEXT_CONNECTION_TYPE_UNKNOWN;

// threads

next_platform_thread_t * next_platform_thread_create( void * context, next_platform_thread_func_t * fn, void * arg )
{
    next_platform_thread_t * thread = (next_platform_thread_t *) next_malloc( context, sizeof( next_platform_thread_t ) );
    
    next_assert( thread );

    thread->context = context;

    thread->handle = CreateThread
    (
        NULL, // default security attributes
        0, // use default stack size  
        fn, // thread function name
        arg, // argument to thread function 
        0, // use default creation flags 
        NULL // returns the thread identifier 
    );

    if ( thread->handle == NULL )
    {
        next_free( context, thread );
        return NULL;
    }

    return thread;
}

void next_platform_thread_join( next_platform_thread_t * thread )
{
    next_assert( thread );
    WaitForSingleObject( thread->handle, INFINITE );
}

void next_platform_thread_destroy( next_platform_thread_t * thread )
{
    next_assert( thread );
    CloseHandle( thread->handle );
    next_free( thread->context, thread );
}

void next_platform_thread_high_priority( next_platform_thread_t * thread )
{
    // todo
    (void)thread;
}

int next_platform_mutex_create( next_platform_mutex_t * mutex )
{
	next_assert( mutex );
	memset( mutex, 0, sizeof(next_platform_mutex_t) );
    if ( !InitializeCriticalSectionAndSpinCount( &mutex->handle, 0xFF ) )
    {
		return NEXT_ERROR;
    }
	mutex->ok = true;
	return NEXT_OK;
}

void next_platform_mutex_acquire( next_platform_mutex_t * mutex )
{
    next_assert( mutex );
	next_assert( mutex->ok );
    EnterCriticalSection( &mutex->handle );
}

void next_platform_mutex_release( next_platform_mutex_t * mutex )
{
    next_assert( mutex );
	next_assert( mutex->ok );
    LeaveCriticalSection( &mutex->handle );
}

void next_platform_mutex_destroy( next_platform_mutex_t * mutex )
{
    next_assert( mutex );
	if ( mutex->ok )
	{
		DeleteCriticalSection( &mutex->handle );
		memset( mutex, 0, sizeof(next_platform_mutex_t) );
	}
}

// time

void next_platform_sleep( double time )
{
    const int milliseconds = (int) ( time * 1000 );
    Sleep( milliseconds );
}

static int timer_initialized = 0;
static LARGE_INTEGER timer_frequency;
static LARGE_INTEGER timer_start;

static const char * next_randombytes_implementation_name()
{
    return "switch";
}

static uint32_t next_randombytes_random()
{
	uint32_t random = 0;
    bool success = BCRYPT_SUCCESS( BCryptGenRandom( bcrypt_algorithm_provider, (uint8_t *)( &random ), ULONG( sizeof( random ) ), 0 ) );
    (void) success;
    next_assert( success );
    return random;
}

static void next_randombytes_stir()
{
}

static uint32_t next_randombytes_uniform( const uint32_t upper_bound )
{
    uint32_t mask = upper_bound - 1;

    mask |= mask >> 1;
    mask |= mask >> 2;
    mask |= mask >> 4;
    mask |= mask >> 8; // mask is smallest ((power of 2) - 1) > upper_bound

    uint32_t result;
    do
    {
        result = mask & next_randombytes_random();  // 16-bit random number
    } while ( result >= upper_bound );
    return result;
}

static void next_randombytes_buf( void * const buf, const size_t size )
{
    bool success = BCRYPT_SUCCESS( BCryptGenRandom( bcrypt_algorithm_provider, (uint8_t *)( buf ), ULONG( size ), 0 ) );
    (void) success;
    next_assert( success );
}

static int next_randombytes_close()
{
    return 0;
}

static randombytes_implementation next_random_implementation =
{
    &next_randombytes_implementation_name,
    &next_randombytes_random,
    &next_randombytes_stir,
    &next_randombytes_uniform,
    &next_randombytes_buf,
    &next_randombytes_close,
};

int next_platform_init()
{
    if ( randombytes_set_implementation( &next_random_implementation ) != 0 )
        return NEXT_ERROR;

    QueryPerformanceFrequency( &timer_frequency );
    QueryPerformanceCounter( &timer_start );

    WSADATA WsaData;
    if ( WSAStartup( MAKEWORD(2,2), &WsaData ) != NO_ERROR )
    {
        return NEXT_ERROR;
    }

    if ( !BCRYPT_SUCCESS( BCryptOpenAlgorithmProvider( &bcrypt_algorithm_provider, BCRYPT_RNG_ALGORITHM, NULL, 0 ) ) )
    {
        return NEXT_ERROR;
    }

    connection_type = NEXT_CONNECTION_TYPE_UNKNOWN;
	try
	{
		auto profile = Windows::Networking::Connectivity::NetworkInformation::GetInternetConnectionProfile();
		if ( profile )
		{
			switch ( profile->NetworkAdapter->IanaInterfaceType )
			{
				case 6:
					connection_type = NEXT_CONNECTION_TYPE_WIRED;
					break;
				case 71:
					connection_type = NEXT_CONNECTION_TYPE_WIFI;
					break;
				case 237:
				case 243:
				case 244:
					connection_type = NEXT_CONNECTION_TYPE_CELLULAR;
					break;
			}
		}
	}
	catch ( Platform::Exception^ )
	{
	}
    return NEXT_OK;
}

double next_platform_time()
{
    LARGE_INTEGER now;
    QueryPerformanceCounter( &now );
    return ( (double) ( now.QuadPart - timer_start.QuadPart ) ) / ( (double) ( timer_frequency.QuadPart ) );
}

void next_platform_term()
{
    WSACleanup();
    BCryptCloseAlgorithmProvider( bcrypt_algorithm_provider, 0 );
}

const char * next_platform_getenv( const char * )
{
    return NULL; // not supported
}

// sockets

uint16_t next_platform_ntohs( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

uint16_t next_platform_htons( uint16_t in )
{
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
}

int next_platform_hostname_resolve( const char * hostname, const char * port, next_address_t * address )
{
    addrinfo hints;
    memset( &hints, 0, sizeof(hints) );
    addrinfo * result;
    if ( getaddrinfo( hostname, port, &hints, &result ) == 0 )
    {
        if ( result )
        {
            if ( result->ai_addr->sa_family == AF_INET6 )
            {
                sockaddr_in6 * addr_ipv6 = (sockaddr_in6 *)( result->ai_addr );
                address->type = NEXT_ADDRESS_IPV6;
                for ( int i = 0; i < 8; ++i )
                {
                    address->data.ipv6[i] = next_platform_ntohs( ( (uint16_t*) &addr_ipv6->sin6_addr ) [i] );
                }
                address->port = next_platform_ntohs( addr_ipv6->sin6_port );
                freeaddrinfo( result );
                return NEXT_OK;
            }
            else if ( result->ai_addr->sa_family == AF_INET )
            {
                sockaddr_in * addr_ipv4 = (sockaddr_in *)( result->ai_addr );
                address->type = NEXT_ADDRESS_IPV4;
                address->data.ipv4[0] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0x000000FF ) );
                address->data.ipv4[1] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0x0000FF00 ) >> 8 );
                address->data.ipv4[2] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0x00FF0000 ) >> 16 );
                address->data.ipv4[3] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0xFF000000 ) >> 24 );
                address->port = next_platform_ntohs( addr_ipv4->sin_port );
                freeaddrinfo( result );
                return NEXT_OK;
            }
            else
            {
                next_assert( 0 );
                freeaddrinfo( result );
                return NEXT_ERROR;
            }
        }
    }

    return NEXT_ERROR;
}

int next_platform_inet_pton4( const char * address_string, uint32_t * address_out )
{
    #if WINVER <= 0x0502
        sockaddr_in sockaddr4;
        wchar_t w_buffer[NEXT_MAX_ADDRESS_STRING_LENGTH + NEXT_ADDRESS_BUFFER_SAFETY*2] = { 0 };
        MultiByteToWideChar( CP_UTF8, 0, address_string, strlen( address_string ), w_buffer, sizeof( w_buffer ) / sizeof( w_buffer[0] ) );
        int addr_size = int( sizeof( sockaddr4 ) );
        bool success = WSAStringToAddress( w_buffer, AF_INET, NULL, LPSOCKADDR( &sockaddr4 ), &addr_size ) == 0;
        *address_out = sockaddr4.sin_addr.s_addr;
        return success ? NEXT_OK : NEXT_ERROR;
    #else
        sockaddr_in sockaddr4;
        bool success = inet_pton( AF_INET, address_string, &sockaddr4.sin_addr ) == 1;
        *address_out = sockaddr4.sin_addr.s_addr;
        return success ? NEXT_OK : NEXT_ERROR;
    #endif
}

// address_out should be a uint16_t[8]
int next_platform_inet_pton6( const char * address_string, uint16_t * address_out )
{
    #if WINVER <= 0x0502
        (void) address_string;
        (void) address_out;
        return NEXT_ERROR;
    #else
        return inet_pton( AF_INET6, address_string, address_out ) == 1 ? NEXT_OK : NEXT_ERROR;
    #endif
}

// address should be a uint16_t[8]
int next_platform_inet_ntop6( const uint16_t * address, char * address_string, size_t address_string_size )
{
    #if WINVER <= 0x0502
        (void) address_string;
        (void) address;
        (void) address_string_size;
        return NEXT_ERROR;
    #else
        return inet_ntop( AF_INET6, (void*)address, address_string, address_string_size ) == NULL ? NEXT_ERROR : NEXT_OK;
    #endif
}

void next_platform_socket_destroy( next_platform_socket_t * );

next_platform_socket_t * next_platform_socket_create( void * context, next_address_t * address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size, bool enable_packet_tagging )
{
    next_assert( address );
    next_assert( address->type != NEXT_ADDRESS_NONE );

    next_platform_socket_t * s = (next_platform_socket_t *) next_malloc( context, sizeof( next_platform_socket_t ) );

    next_assert( s );

    s->context = context;

    // create socket

    s->handle = socket( ( address->type == NEXT_ADDRESS_IPV6 ) ? AF_INET6 : AF_INET, SOCK_DGRAM, IPPROTO_UDP );

    if ( s->handle == INVALID_SOCKET )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create socket" );
        next_free( context, s );
        return NULL;
    }

    // force IPv6 only if necessary

    if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        int yes = 1;
        if ( setsockopt( s->handle, IPPROTO_IPV6, IPV6_V6ONLY, (char*)( &yes ), sizeof( yes ) ) != 0 )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "failed to set socket ipv6 only" );
            next_platform_socket_destroy( s );
            return NULL;
        }
    }

    // increase socket send and receive buffer sizes

    if ( setsockopt( s->handle, SOL_SOCKET, SO_SNDBUF, (char*)( &send_buffer_size ), sizeof( int ) ) != 0 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to set socket send buffer size" );
        next_platform_socket_destroy( s );
        return NULL;
    }

    if ( setsockopt( s->handle, SOL_SOCKET, SO_RCVBUF, (char*)( &receive_buffer_size ), sizeof( int ) ) != 0 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to set socket receive buffer size" );
        next_platform_socket_destroy( s );
        return NULL;
    }

    // bind to port

    if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        sockaddr_in6 socket_address;
        memset( &socket_address, 0, sizeof( sockaddr_in6 ) );
        socket_address.sin6_family = AF_INET6;
        for ( int i = 0; i < 8; ++i )
        {
            ( (uint16_t*) &socket_address.sin6_addr ) [i] = next_platform_htons( address->data.ipv6[i] );
        }
        socket_address.sin6_port = next_platform_htons( address->port );

        if ( bind( s->handle, (sockaddr*) &socket_address, sizeof( socket_address ) ) < 0 )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "failed to bind socket (ipv6)" );
            next_platform_socket_destroy( s );
            return NULL;
        }
    }
    else
    {
        sockaddr_in socket_address;
        memset( &socket_address, 0, sizeof( socket_address ) );
        socket_address.sin_family = AF_INET;
        socket_address.sin_addr.s_addr = ( ( (uint32_t) address->data.ipv4[0] ) )      | 
                                         ( ( (uint32_t) address->data.ipv4[1] ) << 8 )  | 
                                         ( ( (uint32_t) address->data.ipv4[2] ) << 16 ) | 
                                         ( ( (uint32_t) address->data.ipv4[3] ) << 24 );
        socket_address.sin_port = next_platform_htons( address->port );

        if ( bind( s->handle, (sockaddr*) &socket_address, sizeof( socket_address ) ) < 0 )
        {
            next_printf( NEXT_LOG_LEVEL_ERROR, "failed to bind socket (ipv4)" );
            next_platform_socket_destroy( s );
            return NULL;
        }
    }

    // if bound to port 0 find the actual port we got

    if ( address->port == 0 )
    {
        if ( address->type == NEXT_ADDRESS_IPV6 )
        {
            sockaddr_in6 sin;
            socklen_t len = sizeof( sin );
            if ( getsockname( s->handle, (sockaddr*)( &sin ), &len ) == -1 )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "failed to get socket port (ipv6)" );
                next_platform_socket_destroy( s );
                return NULL;
            }
            address->port = next_platform_ntohs( sin.sin6_port );
        }
        else
        {
            sockaddr_in sin;
            socklen_t len = sizeof( sin );
            if ( getsockname( s->handle, (sockaddr*)( &sin ), &len ) == -1 )
            {
                next_printf( NEXT_LOG_LEVEL_ERROR, "failed to get socket port (ipv4)" );
                next_platform_socket_destroy( s );
                return NULL;
            }
            address->port = next_platform_ntohs( sin.sin_port );
        }
    }

    // set non-blocking io

    if ( socket_type == NEXT_PLATFORM_SOCKET_NON_BLOCKING )
    {
        DWORD nonBlocking = 1;
        if ( ioctlsocket( s->handle, FIONBIO, &nonBlocking ) != 0 )
        {
            next_platform_socket_destroy( s );
            return NULL;
        }
    }
    else if ( timeout_seconds > 0.0f )
    {
        // set receive timeout
        DWORD tv = DWORD( timeout_seconds * 1000.0f );
        if ( setsockopt( s->handle, SOL_SOCKET, SO_RCVTIMEO, (const char *)( &tv ), sizeof( tv ) ) < 0 )
            return NULL;
    }
    else
    {
        // timeout <= 0, socket is blocking with no timeout
    }

    return s;
}

void next_platform_socket_destroy( next_platform_socket_t * socket )
{
    next_assert( socket );

    if ( socket->handle != 0 )
    {
        closesocket( socket->handle );
        socket->handle = 0;
    }

    memset( socket, 0, sizeof( *socket ) );

    next_free( socket->context, socket );
}

void next_platform_socket_send_packet( next_platform_socket_t * socket, const next_address_t * to, const void * packet_data, int packet_bytes )
{
    next_assert( socket );
    next_assert( to );
    next_assert( to->type == NEXT_ADDRESS_IPV6 || to->type == NEXT_ADDRESS_IPV4 );
    next_assert( packet_data );
    next_assert( packet_bytes > 0 );

    if ( to->type == NEXT_ADDRESS_IPV6 )
    {
        sockaddr_in6 socket_address;
        memset( &socket_address, 0, sizeof( socket_address ) );
        socket_address.sin6_family = AF_INET6;
        for ( int i = 0; i < 8; ++i )
        {
            ( (uint16_t*) &socket_address.sin6_addr ) [i] = next_platform_htons( to->data.ipv6[i] );
        }
        socket_address.sin6_port = next_platform_htons( to->port );
        sendto( socket->handle, (char*)( packet_data ), packet_bytes, 0, (sockaddr*)( &socket_address ), sizeof( sockaddr_in6 ) );
    }
    else if ( to->type == NEXT_ADDRESS_IPV4 )
    {
        sockaddr_in socket_address;
        memset( &socket_address, 0, sizeof( socket_address ) );
        socket_address.sin_family = AF_INET;
        socket_address.sin_addr.s_addr = ( ( (uint32_t) to->data.ipv4[0] ) )        | 
                                         ( ( (uint32_t) to->data.ipv4[1] ) << 8 )   | 
                                         ( ( (uint32_t) to->data.ipv4[2] ) << 16 )  | 
                                         ( ( (uint32_t) to->data.ipv4[3] ) << 24 );
        socket_address.sin_port = next_platform_htons( to->port );
        sendto( socket->handle, (const char*)( packet_data ), packet_bytes, 0, (sockaddr*)( &socket_address ), sizeof( sockaddr_in ) );
    }
}

int next_platform_socket_receive_packet( next_platform_socket_t * socket, next_address_t * from, void * packet_data, int max_packet_size )
{
    next_assert( socket );
    next_assert( from );
    next_assert( packet_data );
    next_assert( max_packet_size > 0 );

    typedef int socklen_t;
    
    sockaddr_storage sockaddr_from;
    socklen_t from_length = sizeof( sockaddr_from );

    int result = recvfrom( socket->handle, (char*) packet_data, max_packet_size, 0, (sockaddr*) &sockaddr_from, &from_length );

    if ( result == SOCKET_ERROR )
    {
        int error = WSAGetLastError();

        if ( error == WSAEWOULDBLOCK || error == WSAETIMEDOUT || error == WSAECONNRESET )
            return 0;

        next_printf( NEXT_LOG_LEVEL_DEBUG, "recvfrom failed with error %d", error );

        return 0;
    }

    if ( sockaddr_from.ss_family == AF_INET6 )
    {
        sockaddr_in6 * addr_ipv6 = (sockaddr_in6*) &sockaddr_from;
        from->type = NEXT_ADDRESS_IPV6;
        for ( int i = 0; i < 8; ++i )
        {
            from->data.ipv6[i] = next_platform_ntohs( ( (uint16_t*) &addr_ipv6->sin6_addr ) [i] );
        }
        from->port = next_platform_ntohs( addr_ipv6->sin6_port );
    }
    else if ( sockaddr_from.ss_family == AF_INET )
    {
        sockaddr_in * addr_ipv4 = (sockaddr_in*) &sockaddr_from;
        from->type = NEXT_ADDRESS_IPV4;
        from->data.ipv4[0] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0x000000FF ) );
        from->data.ipv4[1] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0x0000FF00 ) >> 8 );
        from->data.ipv4[2] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0x00FF0000 ) >> 16 );
        from->data.ipv4[3] = (uint8_t) ( ( addr_ipv4->sin_addr.s_addr & 0xFF000000 ) >> 24 );
        from->port = next_platform_ntohs( addr_ipv4->sin_port );
    }
    else
    {
        next_assert( 0 );
        return 0;
    }
  
    next_assert( result >= 0 );

    return result;
}

int next_platform_connection_type()
{
    return connection_type;
}

int next_platform_id()
{
    return NEXT_PLATFORM_XBOX_ONE;
}

NEXT_PACK_POP()

#else // #if NEXT_PLATFORM == NEXT_PLATFORM_XBOX_ONE

int next_xbox_one_dummy_symbol = 0;

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_XBOX_ONE
