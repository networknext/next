#ifndef RELAY_RELAY_HPP
#define RELAY_RELAY_HPP

#include "net/address.hpp"
#include "core/relay_manager.hpp"
#include "relay_platform.hpp"

#include "core/replay_protection.hpp"
#include "core/packets/types.hpp"

namespace relay
{
  int relay_initialize();

  void relay_term();

  int relay_write_header(int direction,
   core::packets::Type type,
   uint64_t sequence,
   uint64_t session_id,
   uint8_t session_version,
   const uint8_t* private_key,
   uint8_t* buffer,
   int buffer_length);

  int relay_peek_header(int direction,
   core::packets::Type* type,
   uint64_t* sequence,
   uint64_t* session_id,
   uint8_t* session_version,
   const uint8_t* buffer,
   int buffer_length);

  int relay_verify_header(int direction, const uint8_t* private_key, uint8_t* buffer, int buffer_length);

  uint64_t relay_clean_sequence(uint64_t sequence);
}  // namespace relay
#endif
