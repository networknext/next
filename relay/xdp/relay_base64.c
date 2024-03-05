/*
    Network Next XDP Relay
*/

#include "relay_base64.h"
#include <string.h>

static const int base64_table_decode[256] =
{
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0, 62, 63, 62, 62, 63, 52, 53, 54, 55,
    56, 57, 58, 59, 60, 61,  0,  0,  0,  0,  0,  0,  0,  0,  1,  2,  3,  4,  5,  6,
    7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,  0,
    0,  0,  0,  63,  0, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
};

int relay_base64_decode_data( const char * input, uint8_t * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

    size_t input_length = strlen( input );
    int pad = input_length > 0 && ( input_length % 4 || input[input_length - 1] == '=' );
    size_t L = ( ( input_length + 3 ) / 4 - pad ) * 4;
    size_t output_length = L / 4 * 3 + pad;

    if ( output_length > output_size )
    {
        return 0;
    }

    for ( size_t i = 0, j = 0; i < L; i += 4 )
    {
        int n = base64_table_decode[ (int) ( input[i] )] << 18 | base64_table_decode[ (int) ( input[i + 1] ) ] << 12 | base64_table_decode[ (int) ( input[i + 2] ) ] << 6 | base64_table_decode[ (int) ( input[i + 3] ) ];
        output[j++] = (uint8_t) ( n >> 16 );
        output[j++] = (uint8_t) ( n >> 8 & 0xFF );
        output[j++] = (uint8_t) ( n & 0xFF );
    }

    if ( pad )
    {
        int n = base64_table_decode[ (int) ( input[L] ) ] << 18 | base64_table_decode[ (int) ( input[L + 1] ) ] << 12;
        output[output_length - 1] = ( uint8_t) ( n >> 16 );

        if ( input_length > L + 2 && input[L + 2] != '=' )
        {
            n |= base64_table_decode[ (int) ( input[L + 2] ) ] << 6;
            output_length += 1;
            if ( output_length > output_size )
            {
                return 0;
            }
            output[output_length - 1] = (uint8_t) ( n >> 8 & 0xFF );
        }
    }

    return (int) output_length;
}

int relay_base64_decode_string( const char * input, char * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

    int output_length = relay_base64_decode_data( input, (uint8_t *)( output ), output_size - 1 );
    if ( output_length < 0 )
    {
        return 0;
    }

    output[output_length] = '\0';

    return output_length;
}
