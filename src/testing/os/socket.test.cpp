#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "os/socket.hpp"

Test(socket_nonblocking_ipv4)
{
  net::Address bind_address, local_address, from;
  os::Socket socket;

  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.Port = bind_address.Port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_ipv4)
{
  net::Address bind_address, local_address, from;
  os::Socket socket;

  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.01, false));
  local_address.Port = bind_address.Port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_with_no_timeout_ipv4)
{
  net::Address bind_address, local_address, from;
  os::Socket socket;

  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.Port = bind_address.Port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  check(socket.recv(from, packet.data(), packet.size()) == packet.size());
  check(from == local_address);
}

Test(socket_nonblocking_ipv6)
{
  net::Address bind_address, local_address, from;
  os::Socket socket;

  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  check(socket.create(os::SocketType::NonBlocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.Port = bind_address.Port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_ipv6)
{
  net::Address bind_address, local_address, from;
  os::Socket socket;

  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.01, false));
  local_address.Port = bind_address.Port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_with_no_timeout_ipv6)
{
  net::Address bind_address, local_address, from;
  os::Socket socket;

  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.Port = bind_address.Port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  check(socket.recv(from, packet.data(), packet.size()) == packet.size());
  check(from == local_address);
}
