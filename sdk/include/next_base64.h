/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_BASE64_H
#define NEXT_BASE64_H

#include "next.h"

int next_base64_encode_data( const uint8_t * input, size_t input_length, char * output, size_t output_size );

int next_base64_decode_data( const char * input, uint8_t * output, size_t output_size );

int next_base64_encode_string( const char * input, char * output, size_t output_size );

int next_base64_decode_string( const char * input, char * output, size_t output_size );

#endif // #if NEXT_BASE64_H
