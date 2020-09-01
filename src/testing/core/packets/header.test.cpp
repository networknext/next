#include "includes.h"
#include "testing/test.hpp"

#include "core/packet_header.hpp"
#include "crypto/bytes.hpp"

using core::Packet;
using core::packets::Direction;
using core::packets::Header;
using core::packets::Type;
using crypto::GenericKey;

Test(core_packets_Header_client_to_server)
{
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
    return private_key;
  }();

  Packet packet;

  Header header = {
   .type = Type::ClientToServer,
   .sequence = 123123130131LL,
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  size_t index = 0;

  check(header.write(packet, index, Direction::ClientToServer, private_key));
  check(index == Header::ByteSize);

  Header other;

  index = 0;
  check(other.read(packet, index, Direction::ClientToServer));

  check(other.type == Type::ClientToServer);
  check(other.sequence == header.sequence);
  check(other.session_id == header.session_id);
  check(other.session_version == header.session_version);

  index = 0;
  check(header.verify(packet, index, Direction::ClientToServer, private_key));
}

Test(core_packets_Header_server_to_client)
{
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
    return private_key;
  }();

  Packet packet;

  Header header = {
   .type = Type::ServerToClient,
   .sequence = 123123130131LL | (1ULL << 63),
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  size_t index = 0;

  check(header.write(packet, index, Direction::ServerToClient, private_key));

  Header other;

  index = 0;
  check(other.read(packet, index, Direction::ServerToClient));

  check(other.type == Type::ServerToClient);
  check(other.sequence == header.sequence);
  check(other.session_id == header.session_id);
  check(other.session_version == header.session_version);

  index = 0;
  check(header.verify(packet, index, Direction::ServerToClient, private_key)).onFail([&] {
    std::cout << header.sequence << std::endl;
  });
}
