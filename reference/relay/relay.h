/*
    Network Next Relay.
    Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

#include <stdint.h>

#ifndef RELAY_H
#define RELAY_H

#define RELAY_OK                                                   0
#define RELAY_ERROR                                               -1

#define RELAY_ADDRESS_NONE                                         0
#define RELAY_ADDRESS_IPV4                                         1
#define RELAY_ADDRESS_IPV6                                         2

#define RELAY_MAX_ADDRESS_STRING_LENGTH                          256
#define RELAY_MAX_VERSION_STRING_LENGTH                           32

#define RELAY_PLATFORM_WINDOWS                                     1
#define RELAY_PLATFORM_MAC                                         2
#define RELAY_PLATFORM_LINUX                                       3

#if defined(_WIN32)
    #define RELAY_PLATFORM RELAY_PLATFORM_WINDOWS
#elif defined(__APPLE__)
    #define RELAY_PLATFORM RELAY_PLATFORM_MAC
#else
    #define RELAY_PLATFORM RELAY_PLATFORM_LINUX
#endif

#if !defined ( RELAY_LITTLE_ENDIAN ) && !defined( RELAY_BIG_ENDIAN )

  #ifdef __BYTE_ORDER__
    #if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
      #define RELAY_LITTLE_ENDIAN 1
    #elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
      #define RELAY_BIG_ENDIAN 1
    #else
      #error Unknown machine endianess detected. Please define RELAY_LITTLE_ENDIAN or RELAY_BIG_ENDIAN.
    #endif // __BYTE_ORDER__

  // Detect with GLIBC's endian.h
  #elif defined(__GLIBC__)
    #include <endian.h>
    #if (__BYTE_ORDER == __LITTLE_ENDIAN)
      #define RELAY_LITTLE_ENDIAN 1
    #elif (__BYTE_ORDER == __BIG_ENDIAN)
      #define RELAY_BIG_ENDIAN 1
    #else
      #error Unknown machine endianess detected. Please define RELAY_LITTLE_ENDIAN or RELAY_BIG_ENDIAN.
    #endif // __BYTE_ORDER

  // Detect with _LITTLE_ENDIAN and _BIG_ENDIAN macro
  #elif defined(_LITTLE_ENDIAN) && !defined(_BIG_ENDIAN)
    #define RELAY_LITTLE_ENDIAN 1
  #elif defined(_BIG_ENDIAN) && !defined(_LITTLE_ENDIAN)
    #define RELAY_BIG_ENDIAN 1

  // Detect with architecture macros
  #elif    defined(__sparc)     || defined(__sparc__)                           \
        || defined(_POWER)      || defined(__powerpc__)                         \
        || defined(__ppc__)     || defined(__hpux)      || defined(__hppa)      \
        || defined(_MIPSEB)     || defined(_POWER)      || defined(__s390__)
    #define RELAY_BIG_ENDIAN 1
  #elif    defined(__i386__)    || defined(__alpha__)   || defined(__ia64)      \
        || defined(__ia64__)    || defined(_M_IX86)     || defined(_M_IA64)     \
        || defined(_M_ALPHA)    || defined(__amd64)     || defined(__amd64__)   \
        || defined(_M_AMD64)    || defined(__x86_64)    || defined(__x86_64__)  \
        || defined(_M_X64)      || defined(__bfin__)
    #define RELAY_LITTLE_ENDIAN 1
  #elif defined(_MSC_VER) && defined(_M_ARM)
    #define RELAY_LITTLE_ENDIAN 1
  #else
    #error Unknown machine endianess detected. Please define RELAY_LITTLE_ENDIAN or RELAY_BIG_ENDIAN.
  #endif

#endif

#if RELAY_PLATFORM == RELAY_PLATFORM_WINDOWS
#include "relay_windows.h"
#elif RELAY_PLATFORM == RELAY_PLATFORM_MAC
#include "relay_mac.h"
#elif RELAY_PLATFORM == RELAY_PLATFORM_LINUX
#include "relay_linux.h"
#endif

struct relay_address_t
{
    union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
    uint16_t port;
    uint8_t type;
};

void relay_printf( const char * format, ... );

int relay_address_parse( relay_address_t * address, const char * address_string_in );

const char * relay_address_to_string( const relay_address_t * address, char * buffer );

int relay_address_equal( const relay_address_t * a, const relay_address_t * b );

#endif // #ifndef RELAY_H
