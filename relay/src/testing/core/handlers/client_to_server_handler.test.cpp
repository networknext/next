#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/client_to_server_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::PacketType;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_client_to_server_handler)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.set_timestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = PacketHeader::SIZE_OF_SIGNED + 100;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::ClientToServer;
    header.sequence = 123123130131LL;
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->next_addr = addr;
  session->expire_timestamp = 10;
  session->private_key = private_key;
  session->client_to_server_sequence = 0;
  session->client_to_server_protection.reset();
  session->server_to_client_protection.reset();

  map.set(header.hash(), session);

  size_t index = 0;

  CHECK(header.write(packet, index, PacketDirection::ClientToServer, private_key));
  CHECK(index == PacketHeader::SIZE_OF_SIGNED);

  core::handlers::client_to_server_handler(packet, map, recorder, router_info, socket);
  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(prev_len == packet.length);

  CHECK(recorder.client_to_server_tx.num_packets == 1);
  CHECK(recorder.client_to_server_tx.num_bytes == packet.length).on_fail([&] {
    std::cout << "packet len = " << packet.length << std::endl;
    std::cout << "byte count = " << recorder.client_to_server_tx.num_bytes << std::endl;
  });

  core::handlers::client_to_server_handler(packet, map, recorder, router_info, socket);
  // check already received
  CHECK(!socket.recv(packet));
}
