#include "includes.h"
#include "testing/test.hpp"

#include "core/packets/header.hpp"
#include "crypto/bytes.hpp"

using core::packets::Direction;
using core::packets::Header;
using core::packets::Type;

Test(core_packets_header_general)
{
  std::array<uint8_t, crypto_box_SECRETKEYBYTES> private_key;

  crypto::RandomBytes(private_key, private_key.size());

  std::array<uint8_t, RELAY_MTU> buffer;

  // client -> server
  {
    Header header = {
     .direction = Direction::ClientToServer,
     .sequence = 123123130131LL,
     .session_id = 0x12313131,
     .session_version = 0x12,
    };

    check(header.write(buffer, buffer.size(), Type::ClientToServer, private_key));

    Type read_type = {};
    uint64_t read_sequence = 0;
    uint64_t read_session_id = 0;
    uint8_t read_session_version = 0;

    Header other = {
      .direction = Direction::ClientToServer,
    };

    check(other.read(buffer, buffer.size(), ))

    check(
     relay::relay_peek_header(
      RELAY_DIRECTION_CLIENT_TO_SERVER,
      &read_type,
      &read_sequence,
      &read_session_id,
      &read_session_version,
      buffer,
      sizeof(buffer)) == RELAY_OK);

    check(read_type == static_cast<uint8_t>(Type::ClientToServer));
    check(read_sequence == sequence);
    check(read_session_id == session_id);
    check(read_session_version == session_version);

    check(relay::relay_verify_header(RELAY_DIRECTION_CLIENT_TO_SERVER, private_key, buffer, sizeof(buffer)) == RELAY_OK);
  }

  // server -> client
  {
    uint64_t sequence = 123123130131LL | (1ULL << 63);
    uint64_t session_id = 0x12313131;
    uint8_t session_version = 0x12;

    check(
     relay::relay_write_header(
      RELAY_DIRECTION_SERVER_TO_CLIENT,
      core::packets::Type::ServerToClient,
      sequence,
      session_id,
      session_version,
      private_key,
      buffer,
      sizeof(buffer)) == RELAY_OK);

    core::packets::Type read_type = {};
    uint64_t read_sequence = 0;
    uint64_t read_session_id = 0;
    uint8_t read_session_version = 0;

    check(
     relay::relay_peek_header(
      RELAY_DIRECTION_SERVER_TO_CLIENT,
      &read_type,
      &read_sequence,
      &read_session_id,
      &read_session_version,
      buffer,
      sizeof(buffer)) == RELAY_OK);

    check(read_type == static_cast<uint8_t>(core::packets::Type::ServerToClient));
    check(read_sequence == sequence);
    check(read_session_id == session_id);
    check(read_session_version == session_version);

    check(relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, private_key, buffer, sizeof(buffer)) == RELAY_OK);
  }
}
