#include "relay_continue_token.hpp"

#include <sodium.h>

#include <cassert>

#include "encoding/binary.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"

namespace relay
{
  void relay_write_continue_token(relay_continue_token_t* token, uint8_t* buffer, int buffer_length)
  {
    (void)buffer_length;

    assert(token);
    assert(buffer);
    assert(buffer_length >= RELAY_CONTINUE_TOKEN_BYTES);

    uint8_t* start = buffer;

    (void)start;

    encoding::write_uint64(&buffer, token->expire_timestamp);
    encoding::write_uint64(&buffer, token->session_id);
    encoding::write_uint8(&buffer, token->session_version);
    encoding::write_uint8(&buffer, token->session_flags);

    assert(buffer - start == RELAY_CONTINUE_TOKEN_BYTES);
  }

  void relay_read_continue_token(relay_continue_token_t* token, const uint8_t* buffer)
  {
    assert(token);
    assert(buffer);

    const uint8_t* start = buffer;

    (void)start;

    token->expire_timestamp = encoding::read_uint64(&buffer);
    token->session_id = encoding::read_uint64(&buffer);
    token->session_version = encoding::read_uint8(&buffer);
    token->session_flags = encoding::read_uint8(&buffer);

    assert(buffer - start == RELAY_CONTINUE_TOKEN_BYTES);
  }

  int relay_encrypt_continue_token(
   uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length)
  {
    assert(sender_private_key);
    assert(receiver_public_key);
    assert(buffer);
    assert(buffer_length >= (int)(RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES));

    (void)buffer_length;

    if (crypto_box_easy(buffer, buffer, RELAY_CONTINUE_TOKEN_BYTES, nonce, receiver_public_key, sender_private_key) != 0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  int relay_decrypt_continue_token(
   const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer)
  {
    assert(sender_public_key);
    assert(receiver_private_key);
    assert(buffer);

    if (crypto_box_open_easy(
         buffer, buffer, RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key) !=
        0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  int relay_write_encrypted_continue_token(
   uint8_t** buffer, relay_continue_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key)
  {
    assert(buffer);
    assert(token);
    assert(sender_private_key);
    assert(receiver_public_key);

    unsigned char nonce[crypto_box_NONCEBYTES];
    encoding::relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* start = *buffer;

    encoding::write_bytes(buffer, nonce, crypto_box_NONCEBYTES);

    relay_write_continue_token(token, *buffer, RELAY_CONTINUE_TOKEN_BYTES);

    if (relay_encrypt_continue_token(
         sender_private_key, receiver_public_key, nonce, *buffer, RELAY_CONTINUE_TOKEN_BYTES + crypto_box_NONCEBYTES) !=
        RELAY_OK)
      return RELAY_ERROR;

    *buffer += RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    (void)start;

    assert((*buffer - start) == RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);

    return RELAY_OK;
  }

  int relay_read_encrypted_continue_token(
   uint8_t** buffer, relay_continue_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key)
  {
    assert(buffer);
    assert(token);
    assert(sender_public_key);
    assert(receiver_private_key);

    const uint8_t* nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if (relay_decrypt_continue_token(sender_public_key, receiver_private_key, nonce, *buffer) != RELAY_OK) {
      return RELAY_ERROR;
    }

    relay_read_continue_token(token, *buffer);

    *buffer += RELAY_CONTINUE_TOKEN_BYTES + crypto_box_MACBYTES;

    return RELAY_OK;
  }
}  // namespace relay