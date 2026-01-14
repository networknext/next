/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_CONTINUE_TOKEN_H
#define NEXT_CONTINUE_TOKEN_H

#include "next.h"
#include "next_crypto.h"

struct next_continue_token_t
{
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
};

inline void next_read_continue_token( next_continue_token_t * token, const uint8_t * buffer )
{
    next_assert( token );
    next_assert( buffer );

    const uint8_t * p = buffer;

    token->expire_timestamp = next_read_uint64( &p );
    token->session_id = next_read_uint64( &p );
    token->session_version = next_read_uint8( &p );

    next_assert( p - buffer == NEXT_CONTINUE_TOKEN_BYTES );
}

inline int next_decrypt_continue_token( const uint8_t * key, const uint8_t * nonce, uint8_t * buffer, uint8_t * decrypted )
{
    next_assert( key );
    next_assert( buffer );
    unsigned long long decrypted_len;
    if ( next_crypto_aead_xchacha20poly1305_ietf_decrypt( decrypted, &decrypted_len, NULL, buffer, NEXT_CONTINUE_TOKEN_BYTES + 16, NULL, 0, nonce, key ) != 0 )
    {
        return NEXT_ERROR;
    }
    return NEXT_OK;
}

inline int next_read_encrypted_continue_token( uint8_t ** buffer, next_continue_token_t * token, const uint8_t * key )
{
    next_assert( buffer );
    next_assert( token );
    next_assert( key );

    const uint8_t * nonce = *buffer;

    *buffer += 24;

    uint8_t decrypted[NEXT_CONTINUE_TOKEN_BYTES];

    if ( next_decrypt_continue_token( key, nonce, *buffer, decrypted ) != NEXT_OK )
    {
        return NEXT_ERROR;
    }

    next_read_continue_token( token, decrypted );

    *buffer += NEXT_CONTINUE_TOKEN_BYTES + 16;

    return NEXT_OK;
}

#endif // #ifndef NEXT_CONTINUE_TOKEN_H
