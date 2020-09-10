#include "includes.h"
#include "testing/test.hpp"

#include "crypto/bytes.hpp"
#include "os/socket.hpp"

using core::Packet;
using net::Address;
using os::Socket;
using os::SocketConfig;
using os::SocketType;

Test(os_socket_nonblocking_ipv4)
{
  Address bind_address, local_address;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::NonBlocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  check(socket.create(bind_address, config));
  local_address.port = bind_address.port;

  // regular buffer
  {
    net::Address from;
    std::array<uint8_t, 256> out, in;

    crypto::RandomBytes(out, out.size());

    check(socket.send(local_address, out.data(), out.size()));

    size_t packets_received = 0;
    while (socket.recv(from, in.data(), in.size()) == out.size()) {
      check(from == local_address);
      packets_received++;
    }

    check(packets_received == 1);
    check(in == out);
  }

  // generic packet
  {
    Packet out, in;

    out.addr = local_address;
    out.length = 256;
    // randomize past 256
    crypto::RandomBytes(out.buffer, out.buffer.size());

    check(socket.send(out));

    size_t packets_received = 0;
    while (socket.recv(in)) {
      check(in.addr == local_address);
      check(in.length == 256);
      packets_received++;
    }

    check(packets_received == 1);
    // only check the range
    std::equal(out.buffer.begin(), out.buffer.begin() + out.length, in.buffer.begin(), in.buffer.begin() + in.length);
  }
}

Test(os_socket_blocking_ipv4_with_timeout)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.recv_timeout = 0.01f;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  check(socket.create(bind_address, config));

  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size()) == packet.size()) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(os_socket_blocking_with_no_timeout_ipv4)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  check(bind_address.parse("0.0.0.0"));
  check(local_address.parse("127.0.0.1"));
  check(socket.create(bind_address, config));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  check(socket.recv(from, packet.data(), packet.size()) == packet.size());
  check(from == local_address);
}

Test(os_socket_nonblocking_ipv6)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::NonBlocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  check(socket.create(bind_address, config));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size()) == packet.size()) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(os_socket_blocking_ipv6)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.recv_timeout = 0.01f;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  check(socket.create(bind_address, config));

  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size()) == packet.size()) {
    check(from == local_address);
    packets_received++;
  }
  check(packets_received == 1);
}

Test(os_socket_blocking_with_no_timeout_ipv6)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  check(bind_address.parse("[::]"));
  check(local_address.parse("[::1]"));
  check(socket.create(bind_address, config));

  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  check(socket.send(local_address, packet.data(), packet.size()));
  check(socket.recv(from, packet.data(), packet.size()) == packet.size());
  check(from == local_address);
}
