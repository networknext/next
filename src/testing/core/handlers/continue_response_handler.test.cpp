#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/continue_response_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::PacketDirection;
using core::PacketHeader;
using core::PacketHeaderV4;
using core::PacketType;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using crypto::GenericKey;
using crypto::PACKET_HASH_LENGTH;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_continue_response_handler_sdk4)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  Socket socket;

  const GenericKey private_key = random_private_key();

  RouterInfo router_info;
  router_info.set_timestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = PacketHeaderV4::SIZE_OF_SIGNED;

  PacketHeaderV4 header;
  {
    header.type = PacketType::ContinueResponse4;
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

  core::handlers::continue_response_handler_sdk4(packet, map, recorder, router_info, socket);
  size_t prev_len = packet.length;
  CHECK(socket.recv(packet)).on_fail([&] {
    std::cout << "unable to receive packet\n";
  });
  CHECK(prev_len == packet.length);

  CHECK(recorder.continue_response_tx.num_packets == 1);
  CHECK(recorder.continue_response_tx.num_bytes == packet.length).on_fail([&] {
    std::cout << "packet len = " << packet.length << '\n';
    std::cout << "byte count = " << recorder.continue_response_rx.num_bytes << '\n';
  });

  core::handlers::continue_response_handler_sdk4(packet, map, recorder, router_info, socket);
  CHECK(!socket.recv(packet));
}
