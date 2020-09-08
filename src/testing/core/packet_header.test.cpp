#include "includes.h"
#include "testing/test.hpp"

#include "core/packet_header.hpp"
#include "crypto/bytes.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::PacketType;
using crypto::GenericKey;

Test(core_packets_Header_client_to_server)
{
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
    return private_key;
  }();

  Packet packet;

  PacketHeader header;
  {
    header.type = PacketType::ClientToServer;
    header.sequence = 123123130131LL;
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  }

  size_t index = 0;

  check(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  check(index == PacketHeader::SIZE_OF_ENCRYPTED);

  PacketHeader other;

  index = 0;
  check(other.read(packet, index, PacketDirection::ClientToServer));

  check(other.type == PacketType::ClientToServer);
  check(other.sequence == header.sequence);
  check(other.session_id == header.session_id);
  check(other.session_version == header.session_version);

  index = 0;
  check(header.verify(packet, index, PacketDirection::ClientToServer, private_key));
}

Test(core_packets_Header_server_to_client)
{
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
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

  check(header.write(packet, index, PacketDirection::ServerToClient, private_key));

  PacketHeader other;

  index = 0;
  check(other.read(packet, index, PacketDirection::ServerToClient));

  check(other.type == PacketType::ServerToClient);
  check(other.sequence == header.sequence);
  check(other.session_id == header.session_id);
  check(other.session_version == header.session_version);

  index = 0;
  check(header.verify(packet, index, PacketDirection::ServerToClient, private_key)).onFail([&] {
    std::cout << header.sequence << std::endl;
  });
}
