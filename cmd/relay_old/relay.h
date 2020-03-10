/*
    Network Next: $(NEXT_VERSION_FULL)
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#ifndef NEXT_H
#define NEXT_H

#include <stdint.h>
#include <stddef.h>

#define NEXT_VERSION_FULL              "$(NEXT_VERSION_FULL)"
#define NEXT_VERSION_MAJOR            "$(NEXT_VERSION_MAJOR)"
#define NEXT_VERSION_MINOR            "$(NEXT_VERSION_MINOR)"
#define NEXT_VERSION_GITHUB          "$(NEXT_VERSION_GITHUB)"

#define NEXT_OK                                             0
#define NEXT_ERROR                                         -1

// This value is intentionally the SDK's NEXT_MTU, plus an
// additional allowance of 200 bytes. The SDK checks the payload
// length against NEXT_MTU, but this doesn't include the header
// bytes the packet will have attached to it when the relays
// process it. Therefore, we add a generous buffer to make sure
// the relays *NEVER* silently truncate packets sent by the SDK.
#define NEXT_MTU                                 (1300 + 200)

#define NEXT_PUBLIC_KEY_BYTES                              32
#define NEXT_PRIVATE_KEY_BYTES                             32

#define NEXT_LOG_LEVEL_NONE                                 0
#define NEXT_LOG_LEVEL_ERROR                                1
#define NEXT_LOG_LEVEL_WARN                                 2
#define NEXT_LOG_LEVEL_INFO                                 3
#define NEXT_LOG_LEVEL_DEBUG                                4

#define NEXT_ADDRESS_NONE                                   0
#define NEXT_ADDRESS_IPV4                                   1
#define NEXT_ADDRESS_IPV6                                   2

#define NEXT_MAX_ADDRESS_STRING_LENGTH                    256

#define NEXT_MAX_NEAR_RELAYS                               10
#define NEXT_MAX_RELAY_HOPS                                 5
#define NEXT_MAX_FLOW_TOKENS      ( NEXT_MAX_RELAY_HOPS + 2 )

#define NEXT_RELAY_PING_SAFETY                             10
#define NEXT_SEC_TO_NS                           1000000000.0

#if defined(_WIN32)
#define NOMINMAX
#endif

#if defined( NEXT_SHARED )
    #if defined(_WIN32)
        #ifdef NEXT_EXPORT
            #define NEXT_EXPORT_FUNC extern "C" __declspec(dllexport)
        #else
            #define NEXT_EXPORT_FUNC extern "C" __declspec(dllimport)
        #endif
    #else
        #define NEXT_EXPORT_FUNC extern "C"
    #endif
#else
    #define NEXT_EXPORT_FUNC extern
#endif

// -----------------------------------------

NEXT_EXPORT_FUNC int next_init();

NEXT_EXPORT_FUNC void next_term();

// -----------------------------------------

struct next_address_t
{
    union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
    uint16_t port;
    uint8_t type;
    next_address_t() {}
    bool operator==( const next_address_t & other ) const;
};

NEXT_EXPORT_FUNC int next_address_parse( next_address_t * address, const char * address_string_in );

NEXT_EXPORT_FUNC char * next_address_to_string( const next_address_t * address, char * buffer );

NEXT_EXPORT_FUNC int next_address_equal( const next_address_t * a, const next_address_t * b );

NEXT_EXPORT_FUNC int next_ip_equal( const next_address_t * a, const next_address_t * b );

// -----------------------------------------

NEXT_EXPORT_FUNC int64_t next_time();      // nanoseconds

NEXT_EXPORT_FUNC void next_sleep( int64_t time_nanoseconds );

NEXT_EXPORT_FUNC void next_set_log_level( int level );

NEXT_EXPORT_FUNC void next_set_assert_function( void (*function)( const char *, const char *, const char * file, int line ) );

NEXT_EXPORT_FUNC void next_set_print_function( void (*function)( int level, const char *, ... ) );

NEXT_EXPORT_FUNC void next_set_allocator( void * (*alloc_function)(size_t), void * (*realloc_function)(void*,size_t), void (*free_function)(void*) );

NEXT_EXPORT_FUNC void next_printf( int level, const char * format, ... );

NEXT_EXPORT_FUNC void next_generate_keypair( uint8_t * public_key, uint8_t * private_key );

NEXT_EXPORT_FUNC uint64_t next_relay_id( const char * );

// -----------------------------------------

#endif // #ifndef NEXT_H
