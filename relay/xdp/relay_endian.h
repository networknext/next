/*
    Network Next XDP Relay
*/

#ifndef RELAY_ENDIAN_H
#define RELAY_ENDIAN_H

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

#endif // #if RELAY_ENDIAN

inline uint64_t bswap( uint32_t value )
{
#ifdef __GNUC__
    return __builtin_bswap32( value );
#else // #ifdef __GNUC__
    uint32_t output;
    output  = ( value & 0xFF000000 ) >> 24;
    output |= ( value & 0x00FF0000 ) >> 8;
    output |= ( value & 0x0000FF00 ) << 8;
    output |= ( value & 0x000000FF ) << 24;
#endif // #ifdef __GNUC__
}

uint16_t relay_ntohs( uint16_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
#endif // #if RELAY_BIG_ENDIAN
}

uint16_t relay_htons( uint16_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return (uint16_t)( ( ( in << 8 ) & 0xFF00 ) | ( ( in >> 8 ) & 0x00FF ) );
#endif // #if RELAY_BIG_ENDIAN
}

inline uint32_t relay_ntohl( uint32_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return bswap( in );
#endif // #if RELAY_BIG_ENDIAN
}

inline uint32_t relay_htonl( uint32_t in )
{
#if RELAY_BIG_ENDIAN
    return in;
#else // #if RELAY_BIG_ENDIAN
    return bswap( in );
#endif // #if RELAY_BIG_ENDIAN
}
