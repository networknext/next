#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "os/socket.hpp"

Test(socket_nonblocking_ipv4)
{
  net::Address bind_address, local_address, from;
  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  os::Socket socket;
  check(socket.create(os::SocketType::NonBlocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(relay_address_equal(&from, &local_address) != 0);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_ipv4)
{
  net::Address bind_address, local_address, from;
  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  os::Socket socket;
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.01, false));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(relay_address_equal(&from, &local_address) != 0);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_with_no_timeout_ipv4)
{
  net::Address bind_address, local_address, from;
  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  os::Socket socket;
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  check(socket.recv(from, paket.data(), packet.size()) == packet.size());
  legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet));
  check(legacy::relay_address_equal(&from, &local_address) != 0);
  legacy::relay_platform_socket_destroy(socket);
}

Test(socket_nonblocking_ipv6)
{
  net::Address bind_address, local_address, from;
  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  os::Socket socket;
  check(socket.create(os::SocketType::NonBlocking, bind_address, 64 * 1024, 64 * 1024, 0.0, false));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(relay_address_equal(&from, &local_address) != 0);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_ipv6)
{
  net::Address bind_address, local_address, from;
  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  os::Socket socket;
  check(socket.create(os::SocketType::Blocking, bind_address, 64 * 1024, 64 * 1024, 0.01, false));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size())) {
    check(relay_address_equal(&from, &local_address) != 0);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(socket_blocking_with_no_timeout_ipv6)
{
  legacy::relay_address_t bind_address;
  legacy::relay_address_t local_address;
  legacy::relay_address_parse(&bind_address, "[::]");
  legacy::relay_address_parse(&local_address, "[::1]");
  legacy::relay_platform_socket_t* socket =
   legacy::relay_platform_socket_create(&bind_address, RELAY_PLATFORM_SOCKET_BLOCKING, -1.0f, 64 * 1024, 64 * 1024);
  local_address.port = bind_address.port;
  check(socket != 0);
  uint8_t packet[256];
  memset(packet, 0, sizeof(packet));
  legacy::relay_platform_socket_send_packet(socket, &local_address, packet, sizeof(packet));
  legacy::relay_address_t from;
  legacy::relay_platform_socket_receive_packet(socket, &from, packet, sizeof(packet));
  check(legacy::relay_address_equal(&from, &local_address) != 0);
  legacy::relay_platform_socket_destroy(socket);
}
