#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/relay_ping_handler.hpp"

#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::packets::RELAY_PING_PACKET_SIZE;
using core::packets::Type;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_relay_ping_handler)
{
  Packet packet;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.Addr = addr;
  packet.Len = RELAY_PING_PACKET_SIZE;

  core::handlers::relay_ping_handler(packet, recorder, socket, true);

  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(prev_len == packet.Len);

  check(recorder.InboundPingTx.PacketCount == 1);
  check(recorder.InboundPingTx.ByteCount == RELAY_PING_PACKET_SIZE);
  check(static_cast<Type>(packet.Buffer[crypto::PacketHashLength]) == Type::RelayPong);
}
