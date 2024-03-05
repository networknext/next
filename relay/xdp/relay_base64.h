/*
    Network Next XDP Relay
*/

#ifndef RELAY_BASE64_H
#define RELAY_BASE64_H

#include "relay.h"

int relay_base64_decode_data( const char * input, uint8_t * output, size_t output_size );

int relay_base64_decode_string( const char * input, char * output, size_t output_size );

#endif // #ifndef RELAY_BASE64_H
