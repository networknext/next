#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/server_to_client_handler.hpp"

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

TEST(core_handlers_server_to_client_handler_sdk4)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  router_info.set_timestamp(0);

  packet.length = PacketHeader::SIZE_OF_SIGNED + 100;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::ServerToClient4;
    header.sequence = 123123130131LL | (1ULL << 63);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->client_to_server_sequence = 0;
  session->expire_timestamp = 10;
  session->next_addr = addr;
  session->prev_addr = addr;
  session->private_key = private_key;
  session->session_id = header.session_id;
  session->session_version = header.session_version;

  map.set(header.hash(), session);

  size_t index = 0;
  CHECK(header.write(packet, index, PacketDirection::ServerToClient, private_key));

  core::handlers::server_to_client_handler_sdk4(packet, map, recorder, router_info, socket);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(prev_len == packet.length);

  CHECK(recorder.server_to_client_tx.num_packets == 1);
  CHECK(recorder.server_to_client_tx.num_bytes == packet.length);

  core::handlers::server_to_client_handler_sdk4(packet, map, recorder, router_info, socket);
  CHECK(!socket.recv(packet));
}
