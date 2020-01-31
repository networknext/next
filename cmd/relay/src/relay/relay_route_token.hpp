#ifndef RELAY_RELAY_ROUTE_TOKEN
#define RELAY_RELAY_ROUTE_TOKEN

#include <cinttypes>

#include <sodium.h>

#include "net/address.hpp"

namespace relay
{
    struct relay_route_token_t
    {
        uint64_t expire_timestamp;
        uint64_t session_id;
        uint8_t session_version;
        uint8_t session_flags;
        int kbps_up;
        int kbps_down;
        legacy::relay_address_t next_address;
        uint8_t private_key[crypto_box_SECRETKEYBYTES];
    };

    void relay_write_route_token(relay_route_token_t* token, uint8_t* buffer, int buffer_length);

    void relay_read_route_token(relay_route_token_t* token, const uint8_t* buffer);

    int relay_encrypt_route_token(
        uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length);

    int relay_decrypt_route_token(
        const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer);

    int relay_write_encrypted_route_token(
        uint8_t** buffer, relay_route_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key);

    int relay_read_encrypted_route_token(
        uint8_t** buffer, relay_route_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key);
}  // namespace relay
#endif