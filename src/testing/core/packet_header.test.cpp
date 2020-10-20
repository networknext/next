#include "includes.h"
#include "testing/test.hpp"

#include "core/packet_header.hpp"

#define CRYPTO_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::PacketType;
using crypto::GenericKey;

TEST(core_PacketHeaderV4_client_to_server)
{
  const GenericKey private_key = random_private_key();

  Packet packet;

  PacketHeader header;
  {
    header.type = PacketType::ClientToServer;
    header.sequence = 123123130131LL;
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  }

  size_t index = 0;

  CHECK(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  CHECK(index == PacketHeader::SIZE_OF_SIGNED);

  PacketHeader other;

  index = 0;
  CHECK(other.read(packet, index, PacketDirection::ClientToServer));

  CHECK(other.type == PacketType::ClientToServer);
  CHECK(other.sequence == header.sequence);
  CHECK(other.session_id == header.session_id);
  CHECK(other.session_version == header.session_version);

  index = 0;
  CHECK(header.verify(packet, index, PacketDirection::ClientToServer, private_key));
}

TEST(core_PacketHeaderV4_server_to_client)
{
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::random_bytes(private_key, private_key.size());
    return private_key;
  }();

  Packet packet;

  PacketHeader header;
  {
    header.type = PacketType::ServerToClient;
    header.sequence = 123123130131LL | (1ULL << 63);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  size_t index = 0;

  CHECK(header.write(packet, index, PacketDirection::ServerToClient, private_key));
  CHECK(index == PacketHeader::SIZE_OF_SIGNED);

  PacketHeader other;

  index = 0;
  CHECK(other.read(packet, index, PacketDirection::ServerToClient));

  CHECK(other.type == PacketType::ServerToClient);
  CHECK(other.sequence == header.sequence);
  CHECK(other.session_id == header.session_id);
  CHECK(other.session_version == header.session_version);

  index = 0;
  CHECK(header.verify(packet, index, PacketDirection::ServerToClient, private_key)).on_fail([&] {
    std::cout << header.sequence << std::endl;
  });
}
