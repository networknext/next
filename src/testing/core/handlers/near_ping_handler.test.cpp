#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/near_ping_handler.hpp"

#define OS_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using os::Socket;
using os::SocketConfig;
using util::ThroughputRecorder;

TEST(core_handlers_near_ping_handler_sdk4)
{
  Packet packet;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  SocketConfig config = default_socket_config();

  CHECK(addr.parse("127.0.0.1"));
  CHECK(socket.create(addr, config));

  packet.length = 1 + 8 + 8 + 8 + 8;
  packet.addr = addr;

  core::handlers::near_ping_handler_sdk4(packet, recorder, socket);
  size_t prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(packet.length == prev_len - 16);

  // no check for already received
  packet.length = 1 + 8 + 8 + 8 + 8;

  core::handlers::near_ping_handler_sdk4(packet, recorder, socket);
  prev_len = packet.length;
  CHECK(socket.recv(packet));
  CHECK(packet.length == prev_len - 16);
}
