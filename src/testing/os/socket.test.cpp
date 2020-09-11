#include "includes.h"
#include "testing/test.hpp"

#include "crypto/bytes.hpp"
#include "os/socket.hpp"

using core::Packet;
using net::Address;
using os::Socket;
using os::SocketConfig;
using os::SocketType;

TEST(os_Socket_nonblocking_ipv4)
{
  Address bind_address, local_address;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::NonBlocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  CHECK(bind_address.parse("0.0.0.0"));
  CHECK(local_address.parse("127.0.0.1"));
  CHECK(socket.create(bind_address, config));
  local_address.port = bind_address.port;

  // regular buffer
  {
    net::Address from;
    std::array<uint8_t, 256> out, in;

    crypto::RandomBytes(out, out.size());

    CHECK(socket.send(local_address, out.data(), out.size()));

    size_t packets_received = 0;
    while (socket.recv(from, in.data(), in.size()) == out.size()) {
      CHECK(from == local_address);
      packets_received++;
    }

    CHECK(packets_received == 1);
    CHECK(in == out);
  }

  // generic packet
  {
    Packet out, in;

    out.addr = local_address;
    out.length = 256;
    // randomize past 256
    crypto::RandomBytes(out.buffer, out.buffer.size());

    CHECK(socket.send(out));

    size_t packets_received = 0;
    while (socket.recv(in)) {
      CHECK(in.addr == local_address);
      CHECK(in.length == 256);
      packets_received++;
    }

    CHECK(packets_received == 1);
    // only check the range
    std::equal(out.buffer.begin(), out.buffer.begin() + out.length, in.buffer.begin(), in.buffer.begin() + in.length);
  }
}

TEST(os_Socket_blocking_ipv4_with_timeout)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.recv_timeout = 0.01f;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  CHECK(bind_address.parse("0.0.0.0"));
  CHECK(local_address.parse("127.0.0.1"));
  CHECK(socket.create(bind_address, config));

  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  CHECK(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size()) == packet.size()) {
    CHECK(from == local_address);
    packets_received++;
  }
  CHECK(packets_received == 1);
}

TEST(os_Socket_blocking_with_no_timeout_ipv4)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  CHECK(bind_address.parse("0.0.0.0"));
  CHECK(local_address.parse("127.0.0.1"));
  CHECK(socket.create(bind_address, config));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  CHECK(socket.send(local_address, packet.data(), packet.size()));
  CHECK(socket.recv(from, packet.data(), packet.size()) == packet.size());
  CHECK(from == local_address);
}

TEST(os_Socket_nonblocking_ipv6)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::NonBlocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  CHECK(bind_address.parse("[::]"));
  CHECK(local_address.parse("[::1]"));
  CHECK(socket.create(bind_address, config));
  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  CHECK(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size()) == packet.size()) {
    CHECK(from == local_address);
    packets_received++;
  }
  CHECK(packets_received == 1);
}

TEST(os_Socket_blocking_ipv6)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.recv_timeout = 0.01f;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  CHECK(bind_address.parse("[::]"));
  CHECK(local_address.parse("[::1]"));
  CHECK(socket.create(bind_address, config));

  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  CHECK(socket.send(local_address, packet.data(), packet.size()));
  size_t packets_received = 0;
  while (socket.recv(from, packet.data(), packet.size()) == packet.size()) {
    CHECK(from == local_address);
    packets_received++;
  }
  CHECK(packets_received == 1);
}

TEST(os_Socket_blocking_with_no_timeout_ipv6)
{
  Address bind_address, local_address, from;
  Socket socket;
  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.send_buffer_size = 1000000;
  config.recv_buffer_size = 1000000;
  config.reuse_port = false;

  CHECK(bind_address.parse("[::]"));
  CHECK(local_address.parse("[::1]"));
  CHECK(socket.create(bind_address, config));

  local_address.port = bind_address.port;
  std::array<uint8_t, 256> packet = {};
  CHECK(socket.send(local_address, packet.data(), packet.size()));
  CHECK(socket.recv(from, packet.data(), packet.size()) == packet.size());
  CHECK(from == local_address);
}
