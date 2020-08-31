#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/near_ping_handler.hpp"

using core::Packet;
using os::Socket;
using util::ThroughputRecorder;

Test(core_handlers_near_ping_handler_unsigned)
{
  Packet packet;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  check(addr.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, addr, 64 * 1024, 64 * 1024, 0.0, false));

  packet.Len = 1 + 8 + 8 + 8 + 8;
  packet.Addr = addr;

  core::handlers::near_ping_handler(packet, recorder, socket, false);
  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(packet.Len == prev_len - 16);

  // no check for already received
  packet.Len = 1 + 8 + 8 + 8 + 8;

  core::handlers::near_ping_handler(packet, recorder, socket, false);
  prev_len = packet.Len;
  check(socket.recv(packet));
  check(packet.Len == prev_len - 16);
}

Test(core_handlers_near_ping_handler_signed)
{
  Packet packet;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  check(addr.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, addr, 64 * 1024, 64 * 1024, 0.0, false));

  packet.Len = crypto::PacketHashLength + 1 + 8 + 8 + 8 + 8;
  packet.Addr = addr;

  core::handlers::near_ping_handler(packet, recorder, socket, true);
  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(packet.Len == prev_len - 16);
  check(crypto::IsNetworkNextPacket(packet.Buffer, packet.Len));

  // no check for already received
  packet.Len = crypto::PacketHashLength + 1 + 8 + 8 + 8 + 8;

  core::handlers::near_ping_handler(packet, recorder, socket, true);
  prev_len = packet.Len;
  check(socket.recv(packet));
  check(packet.Len == prev_len - 16);
  check(crypto::IsNetworkNextPacket(packet.Buffer, packet.Len));
}