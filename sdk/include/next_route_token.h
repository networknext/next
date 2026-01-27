/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_ROUTE_TOKEN_H
#define NEXT_ROUTE_TOKEN_H

#include "next.h"
#include "next_util.h"
#include "next_crypto.h"
#include "next_read_write.h"

struct next_route_token_t
{
    uint8_t private_key[NEXT_SESSION_PRIVATE_KEY_BYTES];
    uint64_t expire_timestamp;
    uint64_t session_id;
    int kbps_up;
    int kbps_down;
    uint32_t next_address;              // big endian
    uint32_t prev_address;              // big endian
    uint16_t next_port;                 // big endian
    uint16_t prev_port;                 // big endian
    uint8_t session_version;
    uint8_t next_internal;
    uint8_t prev_internal;
};

inline void next_read_route_token( next_route_token_t * token, const uint8_t * buffer )
{
    next_assert( token );
    next_assert( buffer );

    const uint8_t * p = buffer;

    next_read_bytes( &p, token->private_key, NEXT_SESSION_PRIVATE_KEY_BYTES );
    token->expire_timestamp = next_read_uint64( &p );
    token->session_id = next_read_uint64( &p );
    token->kbps_up = next_read_uint32( &p );
    token->kbps_down = next_read_uint32( &p );
    token->next_address = next_read_uint32( &p );
    token->prev_address = next_read_uint32( &p );
    token->next_port = next_read_uint16( &p );
    token->prev_port = next_read_uint16( &p );
    token->session_version = next_read_uint8( &p );
    token->next_internal = next_read_uint8( &p );
    token->prev_internal = next_read_uint8( &p );

    // IMPORTANT: ports are stored in big endian order because it works better with xdp relay
#if NEXT_LITTLE_ENDIAN
    token->next_port = next::bswap( token->next_port );
    token->prev_port = next::bswap( token->prev_port );
#endif // #if NEXT_LITTLE_ENDIAN

    next_assert( p - buffer == NEXT_ROUTE_TOKEN_BYTES );
}

inline int next_decrypt_route_token( const uint8_t * key, const uint8_t * nonce, uint8_t * buffer, uint8_t * decrypted )
{
    next_assert( key );
    next_assert( buffer );
    unsigned long long decrypted_len;
    if ( next_crypto_aead_xchacha20poly1305_ietf_decrypt( decrypted, &decrypted_len, NULL, buffer, NEXT_ROUTE_TOKEN_BYTES + 16, NULL, 0, nonce, key ) != 0 )
    {
        return NEXT_ERROR;
    }
    return NEXT_OK;
}

inline int next_read_encrypted_route_token( uint8_t ** buffer, next_route_token_t * token, const uint8_t * key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( key );

    const uint8_t * nonce = *buffer;

    *buffer += 24;

    uint8_t decrypted[NEXT_ROUTE_TOKEN_BYTES];

    if ( next_decrypt_route_token( key, nonce, *buffer, decrypted ) != NEXT_OK )
    {
        return NEXT_ERROR;
    }

    next_read_route_token( token, decrypted );

    *buffer += NEXT_ROUTE_TOKEN_BYTES + 16;

    return NEXT_OK;
}

#endif // #ifndef NEXT_ROUTE_TOKEN_H
