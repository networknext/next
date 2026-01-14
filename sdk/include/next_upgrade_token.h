/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_UPGRADE_TOKEN_H
#define NEXT_UPGRADE_TOKEN_H

struct NextUpgradeToken
{
    uint64_t session_id;
    uint64_t expire_timestamp;
    next_address_t client_address;
    next_address_t server_address;

    NextUpgradeToken()
    {
        session_id = 0;
        expire_timestamp = 0;
        memset( &client_address, 0, sizeof(next_address_t) );
        memset( &server_address, 0, sizeof(next_address_t) );
    }

    int Write( uint8_t * buffer, const uint8_t * private_key )
    {
        next_assert( buffer );
        next_assert( private_key );

        memset( buffer, 0, NEXT_UPGRADE_TOKEN_BYTES );

        uint8_t * nonce = buffer;
        next_crypto_random_bytes( nonce, NEXT_CRYPTO_SECRETBOX_NONCEBYTES );
        buffer += NEXT_CRYPTO_SECRETBOX_NONCEBYTES;

        uint8_t * p = buffer;

        next_write_uint64( &p, session_id );
        next_write_uint64( &p, expire_timestamp );
        next_write_address( &p, &client_address );
        next_write_address( &p, &server_address );

        int bytes_written = p - buffer;

        next_crypto_secretbox_easy( buffer, buffer, NEXT_UPGRADE_TOKEN_BYTES - NEXT_CRYPTO_SECRETBOX_NONCEBYTES - NEXT_CRYPTO_SECRETBOX_MACBYTES, nonce, private_key );

        next_assert( NEXT_CRYPTO_SECRETBOX_NONCEBYTES + bytes_written + NEXT_CRYPTO_SECRETBOX_MACBYTES <= NEXT_UPGRADE_TOKEN_BYTES );

        return NEXT_CRYPTO_SECRETBOX_NONCEBYTES + bytes_written + NEXT_CRYPTO_SECRETBOX_MACBYTES;
    }

    bool Read( const uint8_t * buffer, const uint8_t * private_key )
    {
        next_assert( buffer );
        next_assert( private_key );

        const uint8_t * nonce = buffer;

        uint8_t decrypted[NEXT_UPGRADE_TOKEN_BYTES];
        memcpy( decrypted, buffer + NEXT_CRYPTO_SECRETBOX_NONCEBYTES, NEXT_UPGRADE_TOKEN_BYTES - NEXT_CRYPTO_SECRETBOX_NONCEBYTES );

        if ( next_crypto_secretbox_open_easy( decrypted, decrypted, NEXT_UPGRADE_TOKEN_BYTES - NEXT_CRYPTO_SECRETBOX_NONCEBYTES, nonce, private_key ) != 0 )
            return false;

        const uint8_t * p = decrypted;

        session_id = next_read_uint64( &p );
        expire_timestamp = next_read_uint64( &p );
        next_read_address( &p, &client_address );
        next_read_address( &p, &server_address );

        return true;
    }
};

#endif // #ifndef NEXT_UPGRADE_TOKEN_H
