/*
    Network Next SDK. Copyright © 2017 - 2021 Network Next, Inc.

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
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
#include <inttypes.h>

#if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS
#define strncpy_s strncpy
#endif // #if NEXT_PLATFORM != NEXT_PLATFORM_WINDOWS

const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";
const char * customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";

static volatile int quit = 0;

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

    // generate the ping token. this should be done on the server because it uses your private key

    const char * datacenter_name = "linode.fremont";

    const char * user_id = "12345";

    next_address_t client_address;
    next_address_parse( &client_address, "127.0.0.1" );   // set to the public IP address pings are sent from

    int ping_token_bytes = 0;
    uint8_t ping_token_data[NEXT_MAX_PING_TOKEN_BYTES];
    next_generate_ping_token( customer_id, customer_private_key, &client_address, datacenter_name, user_id, ping_token_data, &ping_token_bytes );

    // make sure the ping token validates. this checks that the signature on the ping token can be verified with your public key

    if ( !next_validate_ping_token( customer_id, customer_public_key, &client_address, ping_token_data, ping_token_bytes ) )
    {
        printf( "error: ping token did not validate\n" );
        return 1;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "ping token validated" );

    // run pings

    const int num_ping_tokens = 1;
    const uint8_t * token_data_array[] = { ping_token_data };
    const int token_bytes_array[] = { ping_token_bytes };

    next_ping_t * ping = next_ping_create( NULL, token_data_array, token_bytes_array, num_ping_tokens );
    if ( !ping )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not create ping" );
        return 1;
    }

    double start_time = next_time();

    while ( !quit )
    {
        if ( next_time() - start_time > NEXT_PING_DURATION )
            break;

        next_ping_update( ping );
        
        next_sleep( 1.0 );
    }

    // todo: print out ping summary

    next_ping_destroy( ping );

    next_term();
    
    return 0;
}
