#include "includes.h"
#include "testing/test.hpp"

#include "core/packet_types.hpp"
#include "core/handlers/relay_ping_handler.hpp"

#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::RELAY_PING_PACKET_SIZE;
using core::PacketType;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_relay_ping_handler)
{
  Packet packet;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.addr = addr;
  packet.length = RELAY_PING_PACKET_SIZE;

  core::handlers::relay_ping_handler(packet, recorder, socket, true);

  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(prev_len == packet.length);

  check(recorder.inbound_ping_tx.num_packets == 1);
  check(recorder.inbound_ping_tx.num_bytes == RELAY_PING_PACKET_SIZE);
  check(static_cast<PacketType>(packet.buffer[0]) == PacketType::RelayPong);
}
