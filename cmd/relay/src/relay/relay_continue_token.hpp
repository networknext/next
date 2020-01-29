#ifndef RELAY_RELAY_CONTINUE_TOKEN_HPP
#define RELAY_RELAY_CONTINUE_TOKEN_HPP

#include <cinttypes>

namespace relay
{
    struct relay_continue_token_t
    {
        uint64_t expire_timestamp;
        uint64_t session_id;
        uint8_t session_version;
        uint8_t session_flags;
    };

    void relay_write_continue_token(relay_continue_token_t* token, uint8_t* buffer, int buffer_length);

    void relay_read_continue_token(relay_continue_token_t* token, const uint8_t* buffer);

    int relay_encrypt_continue_token(
        uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length);

    int relay_decrypt_continue_token(
        const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer);

    int relay_write_encrypted_continue_token(
        uint8_t** buffer, relay_continue_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key);

    int relay_read_encrypted_continue_token(
        uint8_t** buffer, relay_continue_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key);
}  // namespace relay
#endif