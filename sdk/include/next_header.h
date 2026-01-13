/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_HEADER_H
#define NEXT_HEADER_H

#include "next.h"
#include "next_crypto.h"
#include "next_read_write.h"

#pragma pack(push, 1)
struct header_data
{
    uint8_t session_private_key[NEXT_SESSION_PRIVATE_KEY_BYTES];
    uint8_t packet_type;
    uint64_t packet_sequence;
    uint64_t session_id;
    uint8_t session_version;
};
#pragma pack(pop)

inline int next_write_header( uint8_t packet_type, uint64_t packet_sequence, uint64_t session_id, uint8_t session_version, const uint8_t * private_key, uint8_t * header )
{
    next_assert( private_key );
    next_assert( header );

    uint8_t * p = header;

    next_write_uint64( &p, packet_sequence );
    next_write_uint64( &p, session_id );
    next_write_uint8( &p, session_version );

    struct header_data data;
    memcpy( data.session_private_key, private_key, NEXT_SESSION_PRIVATE_KEY_BYTES );
    data.packet_type = packet_type;
    data.packet_sequence = packet_sequence;
    data.session_id = session_id;
    data.session_version = session_version;

    next_crypto_hash_sha256( p, (const unsigned char*) &data, sizeof(struct header_data) );

    return NEXT_OK;
}

inline void next_peek_header( uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, const uint8_t * header, int header_length )
{
    (void) header_length;

    next_assert( header );
    next_assert( header_length >= NEXT_HEADER_BYTES );

    uint64_t packet_sequence = next_read_uint64( &header );

    *sequence = packet_sequence;
    *session_id = next_read_uint64( &header );
    *session_version = next_read_uint8( &header );
}

inline int next_read_header( int packet_type, uint64_t * sequence, uint64_t * session_id, uint8_t * session_version, const uint8_t * private_key, uint8_t * header, int header_length )
{
    next_assert( private_key );
    next_assert( header );
    next_assert( header_length >= NEXT_HEADER_BYTES );

    (void) header_length;

    struct header_data data;

    memcpy( data.session_private_key, private_key, NEXT_SESSION_PRIVATE_KEY_BYTES );

    data.packet_type = packet_type;

    data.packet_sequence  = header[0];
    data.packet_sequence |= ( ( (uint64_t)( header[1] ) ) << 8  );
    data.packet_sequence |= ( ( (uint64_t)( header[2] ) ) << 16 );
    data.packet_sequence |= ( ( (uint64_t)( header[3] ) ) << 24 );
    data.packet_sequence |= ( ( (uint64_t)( header[4] ) ) << 32 );
    data.packet_sequence |= ( ( (uint64_t)( header[5] ) ) << 40 );
    data.packet_sequence |= ( ( (uint64_t)( header[6] ) ) << 48 );
    data.packet_sequence |= ( ( (uint64_t)( header[7] ) ) << 56 );

    data.session_id  = header[8];
    data.session_id |= ( ( (uint64_t)( header[8+1] ) ) << 8  );
    data.session_id |= ( ( (uint64_t)( header[8+2] ) ) << 16 );
    data.session_id |= ( ( (uint64_t)( header[8+3] ) ) << 24 );
    data.session_id |= ( ( (uint64_t)( header[8+4] ) ) << 32 );
    data.session_id |= ( ( (uint64_t)( header[8+5] ) ) << 40 );
    data.session_id |= ( ( (uint64_t)( header[8+6] ) ) << 48 );
    data.session_id |= ( ( (uint64_t)( header[8+7] ) ) << 56 );

    data.session_version = header[8+8];

    uint8_t hash[32];
    next_crypto_hash_sha256( hash, (const unsigned char*) &data, sizeof(struct header_data) );

    if ( memcmp( hash, header + 8 + 8 + 1, 8 ) != 0 )
    {
        return NEXT_ERROR;
    }

    *sequence = data.packet_sequence;
    *session_id = data.session_id;
    *session_version = data.session_version;

    return NEXT_OK;
}

#endif // #ifndef NEXT_HEADER_H
