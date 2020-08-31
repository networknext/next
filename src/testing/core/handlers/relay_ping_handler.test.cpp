#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/relay_ping_handler.hpp"

#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::packets::Type;
using net::Address;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_relay_ping_handler_unsigned)
{
  Packet packet;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.Len = crypto::PacketHashLength + core::packets::RELAY_PING_PACKET_SIZE;
}
