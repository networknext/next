#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/route_response_handler.hpp"

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
using crypto::GenericKey;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_route_response_handler)
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

  packet.length = PacketHeader::SIZE_OF_SIGNED;
  packet.addr = addr;

  PacketHeader header;
  {
    header.type = PacketType::RouteResponse;
    header.sequence = 123123130131LL | (1ULL << 63) | (1ULL << 62);
    header.session_id = 0x12313131;
    header.session_version = 0x12;
  };

  auto session = std::make_shared<Session>();
  session->session_id = header.session_id;
  session->session_version = header.session_version;
  session->private_key = private_key;
  session->prev_addr = addr;
  session->expire_timestamp = 10;
  session->server_to_client_sequence = 0;

  map.set(header.hash(), session);

  size_t index = 0;
  CHECK(header.write(packet, index, PacketDirection::ServerToClient, private_key));

  core::handlers::route_response_handler(packet, map, recorder, router_info, socket);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(prev_len == packet.length);

  CHECK(recorder.route_response_tx.num_packets == 1);
  CHECK(recorder.route_response_tx.num_bytes == PacketHeader::SIZE_OF_SIGNED);

  core::handlers::route_response_handler(packet, map, recorder, router_info, socket);
  CHECK(!socket.recv(packet));
}
