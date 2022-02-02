/*
    Network Next SDK. Copyright © 2017 - 2022 Network Next, Inc.

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

#include "next.h"
#include "../source/next_crypto.h" // todo
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
#include <inttypes.h>

#if NEXT_EXPERIMENTAL

#if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS
#define strncpy_s strncpy
#endif // #if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS

const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";
const char * customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";

static volatile int quit = 0;

extern int next_base64_encode_data( const uint8_t * input, size_t input_length, char * output, size_t output_size );

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

int main()
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );
    
    // initialize network next. this reads in the base64 private/public customer keypair

    next_config_t config;
    next_default_config( &config );
    strncpy_s( config.ping_backend_hostname, "127.0.0.1", sizeof(config.ping_backend_hostname) - 1 );
    strncpy_s( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );
    strncpy_s( config.customer_private_key, customer_private_key, sizeof(config.customer_private_key) - 1 );

    if ( next_init( NULL, &config ) != NEXT_OK )
    {
        printf( "error: could not initialize network next\n" );
        return 1;
    }

    // grab customer info from the SDK post init

    const uint64_t customer_id = next_customer_id();
    const uint8_t * customer_public_key = next_customer_public_key();
    const uint8_t * customer_private_key = next_customer_private_key();

    // generate the ping tokens. this should be done on your backend because it uses your private key

    const int NumTokens = 3;

    const char * datacenter_names[NumTokens] = { "linode.fremont", "vultr.chicago", "google.iowa.1" };

    const char * user_id = "12345";

    next_address_t client_address;
    next_address_parse( &client_address, "127.0.0.1" );   // IMPORTANT: change this to the public IP address of the client

    int ping_token_bytes[NumTokens];
    memset( ping_token_bytes, 0, sizeof(ping_token_bytes) );
    
    uint8_t ping_token_data[NumTokens][NEXT_MAX_PING_TOKEN_BYTES];

    for ( int i = 0; i < NumTokens; i++ )
    {
        next_generate_ping_token( customer_id, customer_private_key, &client_address, datacenter_names[i], user_id, ping_token_data[i], &ping_token_bytes[i] );
    }

    // make sure the ping tokens validate. this checks that the signature on the ping token can be verified with your public key

    for ( int i = 0; i < NumTokens; ++i )
    {
        if ( !next_validate_ping_token( customer_id, customer_public_key, &client_address, ping_token_data[i], ping_token_bytes[i] ) )
        {
            printf( "error: ping token %d did not validate\n", i );
            return 1;
        }

        next_printf( NEXT_LOG_LEVEL_INFO, "ping token %d validated", i );
    }

    // run pings. this should be done on your client

    const uint8_t * token_data_array[NumTokens];

    for ( int i = 0; i < NumTokens; ++i )
    {
        token_data_array[i] = ping_token_data[i];
    }

    next_ping_t * ping = next_ping_create( NULL, "0.0.0.0:0", token_data_array, ping_token_bytes, NumTokens );
    if ( !ping )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not create ping" );
        return 1;
    }

    while ( !quit )
    {
        next_ping_update( ping );

        const int ping_state = next_ping_state( ping );

        if ( ping_state == NEXT_PING_STATE_ERROR || ping_state == NEXT_PING_STATE_FINISHED )
            break;
        
        next_sleep( 1.0 / 60.0 );
    }

    // print out ping summary

    const int ping_state = next_ping_state( ping );

    if ( ping_state != NEXT_PING_STATE_ERROR )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "ping results" );

        // todo
    }
    else
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "ping was not successful" );
    }

    next_ping_destroy( ping );

    next_term();
    
    return 0;
}

#else // #if NEXT_EXPERIMENTAL

int main()
{
    return 0;
}

#endif // #if NEXT_EXPERIMENTAL