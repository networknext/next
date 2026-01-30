/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_ADDRESS_H
#define NEXT_ADDRESS_H

#include "next.h"

#include <stdint.h>

#ifndef NEXT_ADDRESS_ALREADY_DEFINED
#define NEXT_ADDRESS_ALREADY_DEFINED
struct next_address_t
{
    union { uint32_t ip; uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
    uint16_t port;
    uint8_t type;
};
#endif 

NEXT_EXPORT_FUNC int next_address_parse( struct next_address_t * address, const char * address_string );

NEXT_EXPORT_FUNC const char * next_address_to_string( const struct next_address_t * address, char * buffer );

NEXT_EXPORT_FUNC const char * next_address_to_string_without_port( const struct next_address_t * address, char * buffer );

NEXT_EXPORT_FUNC bool next_address_equal( const struct next_address_t * a, const struct next_address_t * b );

NEXT_EXPORT_FUNC void next_address_anonymize( struct next_address_t * address );

NEXT_EXPORT_FUNC bool next_address_is_ipv4_in_ipv6( struct next_address_t * address );

NEXT_EXPORT_FUNC void next_address_convert_ipv6_to_ipv4( struct next_address_t * address );

NEXT_EXPORT_FUNC void next_address_convert_ipv4_to_ipv6( struct next_address_t * address );

#endif // #ifndef NEXT_ADDRESS_H
