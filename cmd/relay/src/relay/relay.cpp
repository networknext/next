#include "includes.h"
#include "relay.hpp"

#include "core/packets/types.hpp"
#include "core/relay_stats.hpp"
#include "encoding/binary.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "net/http.hpp"
#include "util/logger.hpp"

namespace relay
{
  int relay_initialize()
  {
    if (relay::relay_platform_init() != RELAY_OK) {
      Log("failed to initialize platform");
      return RELAY_ERROR;
    }

    if (sodium_init() == -1) {
      Log("failed to initialize sodium");
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  void relay_term()
  {
    relay::relay_platform_term();
  }

  int relay_write_header(
   int direction,
   core::packets::Type type,
   uint64_t sequence,
   uint64_t session_id,
   uint8_t session_version,
   const uint8_t* private_key,
   uint8_t* buffer,
   int buffer_length)
  {
    assert(private_key);
    assert(buffer);
    assert(RELAY_HEADER_BYTES <= buffer_length);

    (void)buffer_length;

    uint8_t* start = buffer;

    (void)start;

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
      // high bit must be set
      assert(sequence & (1ULL << 63));
    } else {
      // high bit must be clear
      assert((sequence & (1ULL << 63)) == 0);
    }

    if (
     type == core::packets::Type::SessionPing || type == core::packets::Type::SessionPong ||
     type == core::packets::Type::RouteResponse || type == core::packets::Type::ContinueResponse) {
      // second highest bit must be set
      assert(sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((sequence & (1ULL << 62)) == 0);
    }

    legacy::write_uint8(&buffer, static_cast<uint8_t>(type));

    legacy::write_uint64(&buffer, sequence);

    uint8_t* additional = buffer;
    const int additional_length = 8 + 2;

    legacy::write_uint64(&buffer, session_id);
    legacy::write_uint8(&buffer, session_version);
    legacy::write_uint8(&buffer, 0);  // todo: remove this once we fully switch to new relay

    uint8_t nonce[12];
    {
      uint8_t* p = nonce;
      legacy::write_uint32(&p, 0);
      legacy::write_uint64(&p, sequence);
    }

    unsigned long long encrypted_length = 0;

    int result = crypto_aead_chacha20poly1305_ietf_encrypt(
     buffer, &encrypted_length, buffer, 0, additional, (unsigned long long)additional_length, NULL, nonce, private_key);

    if (result != 0)
      return RELAY_ERROR;

    buffer += encrypted_length;

    assert(int(buffer - start) == RELAY_HEADER_BYTES);

    return RELAY_OK;
  }

  int relay_peek_header(
   int direction,
   core::packets::Type* type,
   uint64_t* sequence,
   uint64_t* session_id,
   uint8_t* session_version,
   const uint8_t* buffer,
   int buffer_length)
  {
    core::packets::Type packet_type;
    uint64_t packet_sequence;

    assert(buffer);

    if (buffer_length < RELAY_HEADER_BYTES)
      return RELAY_ERROR;

    packet_type = static_cast<core::packets::Type>(legacy::read_uint8(&buffer));

    packet_sequence = legacy::read_uint64(&buffer);

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
      // high bit must be set
      if (!(packet_sequence & (1ULL << 63)))
        return RELAY_ERROR;
    } else {
      // high bit must be clear
      if (packet_sequence & (1ULL << 63))
        return RELAY_ERROR;
    }

    *type = packet_type;

    if (
     *type == core::packets::Type::SessionPing || *type == core::packets::Type::SessionPong ||
     *type == core::packets::Type::RouteResponse || *type == core::packets::Type::ContinueResponse) {
      // second highest bit must be set
      assert(packet_sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((packet_sequence & (1ULL << 62)) == 0);
    }

    *sequence = packet_sequence;
    *session_id = legacy::read_uint64(&buffer);
    *session_version = legacy::read_uint8(&buffer);

    return RELAY_OK;
  }

  int relay_verify_header(int direction, const uint8_t* private_key, uint8_t* buffer, int buffer_length)
  {
    assert(private_key);
    assert(buffer);

    if (buffer_length < RELAY_HEADER_BYTES) {
      return RELAY_ERROR;
    }

    const uint8_t* p = buffer;

    core::packets::Type packet_type = static_cast<core::packets::Type>(legacy::read_uint8(&p));

    uint64_t packet_sequence = legacy::read_uint64(&p);

    if (direction == RELAY_DIRECTION_SERVER_TO_CLIENT) {
      // high bit must be set
      if (!(packet_sequence & (1ULL << 63))) {
        return RELAY_ERROR;
      }
    } else {
      // high bit must be clear
      if (packet_sequence & (1ULL << 63)) {
        return RELAY_ERROR;
      }
    }

    if (
     packet_type == core::packets::Type::SessionPing || packet_type == core::packets::Type::SessionPong ||
     packet_type == core::packets::Type::RouteResponse || packet_type == core::packets::Type::ContinueResponse) {
      // second highest bit must be set
      assert(packet_sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((packet_sequence & (1ULL << 62)) == 0);
    }

    const uint8_t* additional = p;

    const int additional_length = 8 + 2;

    uint64_t packet_session_id = legacy::read_uint64(&p);
    uint8_t packet_session_version = legacy::read_uint8(&p);
    uint8_t packet_session_flags = legacy::read_uint8(&p);  // todo: remove once we fully switch over to new relay

    (void)packet_session_id;
    (void)packet_session_version;
    (void)packet_session_flags;

    uint8_t nonce[12];
    {
      uint8_t* q = nonce;
      legacy::write_uint32(&q, 0);
      legacy::write_uint64(&q, packet_sequence);
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt(
     buffer + 19,
     &decrypted_length,
     NULL,
     buffer + 19,
     (unsigned long long)crypto_aead_chacha20poly1305_IETF_ABYTES,
     additional,
     (unsigned long long)additional_length,
     nonce,
     private_key);

    if (result != 0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  uint64_t relay_clean_sequence(uint64_t sequence)
  {
    uint64_t mask = ~((1ULL << 63) | (1ULL << 62));
    return sequence & mask;
  }
}  // namespace relay
