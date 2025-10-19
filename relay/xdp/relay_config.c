/*
    Network Next XDP Relay (userspace)
*/

#include "relay_config.h"
#include "relay_platform.h"
#include "relay_base64.h"

#include <linux/if_ether.h>

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sodium.h>

int read_config( struct config_t * config )
{
    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_name = getenv( "RELAY_NAME" );
    if ( !relay_name )
    {
        printf( "\nerror: RELAY_NAME not set\n\n" );
        return RELAY_ERROR;
    }

    printf( "Relay name is '%s'\n", relay_name );

    strncpy( config->relay_name, relay_name, sizeof(config->relay_name) - 1 );

    // -----------------------------------------------------------------------------------------------------------------------------

    char * relay_public_address_env = getenv( "RELAY_PUBLIC_ADDRESS" );
    if ( !relay_public_address_env )
    {
        printf( "\nerror: RELAY_PUBLIC_ADDRESS not set\n\n" );
        return RELAY_ERROR;
    }

    if ( relay_platform_parse_address( relay_public_address_env, &config->relay_public_address, &config->relay_port  ) != RELAY_OK )
    {
        printf( "\nerror: invalid relay public address '%s'\n\n", relay_public_address_env );
        return RELAY_ERROR;
    }

    printf( "Relay port is %d\n", config->relay_port );

    printf( "Relay public address is %d.%d.%d.%d\n", 
        ((uint8_t*)&config->relay_public_address)[3], 
        ((uint8_t*)&config->relay_public_address)[2], 
        ((uint8_t*)&config->relay_public_address)[1], 
        ((uint8_t*)&config->relay_public_address)[0] 
    );

    // -----------------------------------------------------------------------------------------------------------------------------

    char * relay_internal_address_env = getenv( "RELAY_INTERNAL_ADDRESS" );
    if ( relay_internal_address_env )
    {
        uint16_t dummy_port = 0;
        if ( relay_platform_parse_address( relay_internal_address_env, &config->relay_internal_address, &dummy_port  ) != RELAY_OK )
        {
            printf( "\nerror: invalid relay internal address '%s'\n\n", relay_internal_address_env );
            return RELAY_ERROR;
        }

        printf( "Relay internal address is %d.%d.%d.%d\n", 
            ((uint8_t*)&config->relay_internal_address)[3], 
            ((uint8_t*)&config->relay_internal_address)[2], 
            ((uint8_t*)&config->relay_internal_address)[1], 
            ((uint8_t*)&config->relay_internal_address)[0]
        );
    }
    else
    {
        config->relay_internal_address = config->relay_public_address;
    }

    if ( config->relay_internal_address == 0 )
    {
        config->relay_internal_address = config->relay_public_address;
    }

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_public_key_env = getenv( "RELAY_PUBLIC_KEY" );
    if ( !relay_public_key_env )
    {
        printf( "\nerror: RELAY_PUBLIC_KEY not set\n\n" );
        return 1;
    }

    if ( relay_base64_decode_data( relay_public_key_env, config->relay_public_key, RELAY_PUBLIC_KEY_BYTES ) != RELAY_PUBLIC_KEY_BYTES )
    {
        printf( "\nerror: invalid relay public key\n\n" );
        return 1;
    }

    printf( "Relay public key is %s\n", relay_public_key_env );

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_private_key_env = getenv( "RELAY_PRIVATE_KEY" );
    if ( !relay_private_key_env )
    {
        printf( "\nerror: RELAY_PRIVATE_KEY not set\n\n" );
        return 1;
    }

    if ( relay_base64_decode_data( relay_private_key_env, config->relay_private_key, RELAY_PRIVATE_KEY_BYTES ) != RELAY_PRIVATE_KEY_BYTES )
    {
        printf( "\nerror: invalid relay private key\n\n" );
        return 1;
    }

    printf( "Relay private key is %s\n", relay_private_key_env );

    // -----------------------------------------------------------------------------------------------------------------------------

    const char * relay_backend_public_key_env = getenv( "RELAY_BACKEND_PUBLIC_KEY" );
    if ( !relay_backend_public_key_env )
    {
        printf( "\nerror: RELAY_BACKEND_PUBLIC_KEY not set\n\n" );
        return 1;
    }

    if ( relay_base64_decode_data( relay_backend_public_key_env, config->relay_backend_public_key, RELAY_BACKEND_PUBLIC_KEY_BYTES ) != RELAY_BACKEND_PUBLIC_KEY_BYTES )
    {
        printf( "\nerror: invalid relay backend public key\n\n" );
        return 1;
    }

    printf( "Relay backend public key is %s\n", relay_backend_public_key_env );

    // -----------------------------------------------------------------------------------------------------------------------------

    if ( crypto_kx_client_session_keys( config->relay_secret_key, NULL, config->relay_public_key, config->relay_private_key, config->relay_backend_public_key ) != 0 )
    {
        printf( "\nerror: failed to generate relay secret key\n\n" );
        return RELAY_ERROR;
    }

    // -----------------------------------------------------------------------------------------------------------------------------

#if !RELAY_DEBUG

    const char * relay_backend_url = getenv( "RELAY_BACKEND_URL" );
    if ( !relay_backend_url )
    {
        printf( "\nerror: RELAY_BACKEND_URL not set\n\n" );
        return RELAY_ERROR;
    }

    printf( "Relay backend url is %s\n", relay_backend_url );

    strncpy( config->relay_backend_url, relay_backend_url, sizeof(config->relay_backend_url) - 1 );

#endif // #if !RELAY_DEBUG

    // -----------------------------------------------------------------------------------------------------------------------------

    return RELAY_OK;
}
