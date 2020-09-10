#pragma once

#include "core/packet.hpp"
#include "net/address.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

using core::Packet;
using core::PacketBuffer;
using net::Address;
using net::AddressType;

namespace os
{
  enum class SocketType : uint8_t
  {
    NonBlocking,
    Blocking
  };

  INLINE auto operator==(SocketType st, int i) -> bool
  {
    return static_cast<int>(st) == i;
  }

  INLINE auto operator==(int i, SocketType st) -> bool
  {
    return static_cast<int>(st) == i;
  }

  struct SocketConfig
  {
    SocketType socket_type;
    size_t send_buffer_size;
    size_t recv_buffer_size;
    std::optional<float> recv_timeout;
    bool reuse_port;
  };

  class Socket
  {
   public:
    Socket();
    ~Socket();

    auto create(Address& bind_addr, SocketConfig config) -> bool;

    // uses sendto()
    auto send(const Packet& packet) const -> bool;

    // uses sendto()
    auto send(const Address& to, const uint8_t* data, size_t size) const -> bool;

    // uses sendmmsg()
    template <size_t BuffSize>
    auto multisend(PacketBuffer<BuffSize>& packet_buff) const -> bool;

    // uses sendmmsg()
    template <size_t BuffSize>
    auto multisend(std::array<mmsghdr, BuffSize>& packet_buff, int count) const -> bool;

    // uses recvfrom()
    auto recv(Address& from, uint8_t* data, size_t max_size) const -> int;

    // uses recvfrom()
    auto recv(Packet& packet) const -> bool;

    // uses recvmmsg()
    template <size_t BuffSize>
    auto multirecv(PacketBuffer<BuffSize>& packet_buff) const -> bool;

    // close the socket
    void close();

    // returns if the socket is closed
    auto closed() const -> bool;

   private:
    int socket_fd = 0;
    SocketType type;
    std::atomic<bool> is_closed;

    auto set_buffer_sizes(size_t send_buffer_size, size_t recv_buffer_size) -> bool;
    auto set_port_reuse(bool reuse) -> bool;

    auto bind_ipv4(const Address& addr) -> bool;
    auto bind_ipv6(const Address& addr) -> bool;

    auto get_port_ipv4(Address& addr) -> bool;
    auto get_port_ipv6(Address& addr) -> bool;

    auto set_socket_type(SocketType type, std::optional<float> timeout) -> bool;
  };

  using SocketPtr = std::shared_ptr<Socket>;

  INLINE Socket::Socket(): is_closed(false) {}

  INLINE Socket::~Socket()
  {
    if (this->socket_fd) {
      close();
    }
  }

  INLINE auto Socket::create(Address& bind_addr, SocketConfig config) -> bool
  {
    if (bind_addr.Type == AddressType::None) {
      return false;
    }

    // create socket
    {
      this->socket_fd = socket((bind_addr.Type == AddressType::IPv4) ? PF_INET : PF_INET6, SOCK_DGRAM, IPPROTO_UDP);

      if (this->socket_fd < 0) {
        LOG(ERROR, "failed to create socket");
        perror("OS msg:");
        return false;
      }
    }

    // force IPv6 only if necessary
    {
      if (bind_addr.Type == AddressType::IPv6) {
        int enable = 1;
        if (setsockopt(this->socket_fd, IPPROTO_IPV6, IPV6_V6ONLY, &enable, sizeof(enable)) != 0) {
          LOG(ERROR, "failed to set socket ipv6 only");
          perror("OS msg:");
          close();
          return false;
        }
      }
    }

    if (!this->set_buffer_sizes(config.send_buffer_size, config.recv_buffer_size)) {
      return false;
    }

    if (!set_port_reuse(config.reuse_port)) {
      return false;
    }

    // bind to port
    {
      if (bind_addr.Type == AddressType::IPv6) {
        if (!bind_ipv6(bind_addr)) {
          return false;
        }
      } else {
        if (!bind_ipv4(bind_addr)) {
          return false;
        }
      }
    }

    // if bound to port 0, find the actual port we got
    // port 0 is a "wildcard" so using it will bind to any available port
    {
      if (bind_addr.Port == 0) {
        if (bind_addr.Type == AddressType::IPv6) {
          if (!get_port_ipv6(bind_addr)) {
            return false;
          }
        } else {
          if (!get_port_ipv4(bind_addr)) {
            return false;
          }
        }
      }
    }

    if (!set_socket_type(config.socket_type, config.recv_timeout)) {
      return false;
    }

    this->type = config.socket_type;

    return true;
  }

  INLINE auto Socket::send(const Packet& packet) const -> bool
  {
    return send(packet.addr, packet.buffer.data(), packet.length);
  }

  INLINE auto Socket::send(const Address& to, const uint8_t* data, size_t size) const -> bool
  {
    if (to.Type != AddressType::IPv4 && to.Type != AddressType::IPv6) {
      return false;
    }

    if (data == nullptr) {
      return false;
    }

    if (size == 0) {
      return false;
    }

    if (this->closed()) {
      return false;
    }

    if (to.Type == AddressType::IPv6) {
      sockaddr_in6 socket_address;
      to.into(socket_address);

      auto res = sendto(this->socket_fd, data, size, 0, reinterpret_cast<sockaddr*>(&socket_address), sizeof(sockaddr_in6));
      if (res < 0) {
        LOG(ERROR, "sendto (", to, ") failed");
        return false;
      }
    } else if (to.Type == AddressType::IPv4) {
      sockaddr_in socket_address;
      to.into(socket_address);

      auto res = sendto(this->socket_fd, data, size, 0, reinterpret_cast<sockaddr*>(&socket_address), sizeof(sockaddr_in6));
      if (res < 0) {
        LOG(ERROR, "sendto (", to, ") failed");
        return false;
      }
    } else {
      LOG(ERROR, "invalid address type, could not send packet");
      return false;
    }

    return true;
  }

  template <size_t BuffSize>
  INLINE auto Socket::multisend(std::array<mmsghdr, BuffSize>& headers, int num_packets_to_send) const -> bool
  {
    static_assert(BuffSize <= 1024);  // max sendmmsg will allow

    if (num_packets_to_send < 0) {
      LOG(ERROR, "number of packets to send is less than 0: ", num_packets_to_send);
      return false;
    }

    if (num_packets_to_send > 1024) {
      LOG(ERROR, "number of packets to send is greater than 1024: ", num_packets_to_send);
      return false;
    }

    int actual_sent = sendmmsg(this->socket_fd, headers.data(), num_packets_to_send, 0);

    if (actual_sent < 0) {
      LOG(ERROR, "sendmmsg() failed: ");
      perror("OS msg:");
      return false;
    }

    return num_packets_to_send == num_packets_to_send;
  }

  template <size_t BuffSize>
  INLINE auto Socket::multisend(PacketBuffer<BuffSize>& packet_buff) const -> bool
  {
    return multisend(packet_buff.Headers, packet_buff.Count);
  }

  INLINE auto Socket::recv(Packet& packet) const -> bool
  {
    auto len = this->recv(packet.addr, packet.buffer.data(), packet.buffer.size());
    packet.length = static_cast<size_t>(len);
    return len > 0;
  }

  INLINE auto Socket::recv(Address& from, uint8_t* data, size_t max_size) const -> int
  {
    assert(data != nullptr);
    assert(max_size > 0);

    if (this->closed()) {
      return 0;
    }

    sockaddr_storage sockaddr_from = {};

    socklen_t len = sizeof(sockaddr_from);
    auto res = recvfrom(
     this->socket_fd,
     data,
     max_size,
     (this->type == SocketType::NonBlocking) ? MSG_DONTWAIT : 0,
     reinterpret_cast<sockaddr*>(&sockaddr_from),
     &len);

    if (res > 0) {
      if (sockaddr_from.ss_family == AF_INET6) {
        from = reinterpret_cast<sockaddr_in6&>(sockaddr_from);
      } else if (sockaddr_from.ss_family == AF_INET) {
        from = reinterpret_cast<sockaddr_in&>(sockaddr_from);
      } else {
        LOG(ERROR, "received packet with invalid ss family: ", sockaddr_from.ss_family);
        return 0;
      }
    } else {
      // if not a timeout, log the error
      if (errno != EAGAIN && errno != EINTR) {
        LOG(ERROR, "recvfrom failed");
      }
    }

    return res;
  }

  template <size_t BuffSize>
  INLINE auto Socket::multirecv(PacketBuffer<BuffSize>& packet_buff) const -> bool
  {
    packet_buff.count = recvmmsg(
     this->socket_fd,
     packet_buff.headers.data(),
     BuffSize,
     MSG_WAITFORONE,
     nullptr);  // DON'T EVER USE TIMEOUT, linux man pages state it is broken

    if (packet_buff.count < 0) {
      LOG(ERROR, "recvmmsg failed");
      return false;
    }

    return true;
  }

  INLINE void Socket::close()
  {
    this->is_closed = true;
    shutdown(this->socket_fd, SHUT_RDWR);
    this->socket_fd = -1;
  }

  INLINE auto Socket::closed() const -> bool
  {
    return this->is_closed;
  }

  INLINE auto Socket::set_buffer_sizes(size_t send_buff_size, size_t recv_buff_size) -> bool
  {
    if (setsockopt(this->socket_fd, SOL_SOCKET, SO_SNDBUF, &send_buff_size, sizeof(send_buff_size)) != 0) {
      LOG(ERROR, "failed to set socket send buffer size");
      this->close();
      return false;
    }

    if (setsockopt(this->socket_fd, SOL_SOCKET, SO_RCVBUF, &recv_buff_size, sizeof(recv_buff_size)) != 0) {
      LOG(ERROR, "failed to set socket receive buffer size");
      this->close();
      return false;
    }

    return true;
  }

  // good read - https://stackoverflow.com/questions/14388706/how-do-so-reuseaddr-and-so-reuseport-differ
  INLINE auto Socket::set_port_reuse(bool reuse) -> bool
  {
    if (reuse) {
      int enable = 1;
      if (setsockopt(this->socket_fd, SOL_SOCKET, SO_REUSEPORT, &enable, sizeof(enable)) < 0) {
        LOG(ERROR, "could not set port reuse");
        perror("OS msg:");
        close();
        return false;
      }
    }

    return true;
  }

  INLINE auto Socket::bind_ipv4(const net::Address& addr) -> bool
  {
    sockaddr_in socket_address = {};
    socket_address.sin_family = AF_INET;
    socket_address.sin_addr.s_addr = (((uint32_t)addr.IPv4[0])) | (((uint32_t)addr.IPv4[1]) << 8) |
                                     (((uint32_t)addr.IPv4[2]) << 16) | (((uint32_t)addr.IPv4[3]) << 24);
    socket_address.sin_port = htons(addr.Port);

    if (bind(this->socket_fd, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
      LOG(ERROR, "failed to bind to address ", addr, " (ipv4)");
      perror("OS msg:");
      close();
      return false;
    }

    return true;
  }

  INLINE auto Socket::bind_ipv6(const net::Address& addr) -> bool
  {
    sockaddr_in6 socket_address = {};

    socket_address.sin6_family = AF_INET6;
    for (int i = 0; i < 8; i++) {
      reinterpret_cast<uint16_t*>(&socket_address.sin6_addr)[i] = htons(addr.IPv6[i]);
    }

    socket_address.sin6_port = htons(addr.Port);

    if (bind(this->socket_fd, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
      LOG(ERROR, "failed to bind socket (ipv6)");
      perror("OS msg:");
      close();
      return false;
    }

    return true;
  }

  INLINE auto Socket::get_port_ipv4(net::Address& addr) -> bool
  {
    sockaddr_in sin;
    socklen_t len = sizeof(len);
    if (getsockname(this->socket_fd, reinterpret_cast<sockaddr*>(&sin), &len) < 0) {
      LOG(ERROR, "failed to get socket port (ipv4)");
      perror("OS msg:");
      close();
      return false;
    }
    addr.Port = ntohs(sin.sin_port);
    return true;
  }

  INLINE auto Socket::get_port_ipv6(net::Address& addr) -> bool
  {
    sockaddr_in6 sin;
    socklen_t len = sizeof(sin);
    if (getsockname(this->socket_fd, reinterpret_cast<sockaddr*>(&sin), &len) < 0) {
      LOG(ERROR, "failed to get socket port (ipv6)");
      perror("OS msg:");
      close();
      return false;
    }
    addr.Port = ntohs(sin.sin6_port);
    return true;
  }

  INLINE auto Socket::set_socket_type(SocketType type, std::optional<float> timeout) -> bool
  {
    // set non-blocking io or receive timeout, or if neither then just blocking with no timeout
    if (type == SocketType::NonBlocking) {
      if (fcntl(this->socket_fd, F_SETFL, O_NONBLOCK, 1) < 0) {
        LOG(ERROR, "failed to set socket to non blocking");
        perror("OS msg:");
        close();
        return false;
      }
    } else if (type == SocketType::Blocking && timeout.has_value()) {
      timeval tv;
      tv.tv_sec = 0;
      tv.tv_usec = (int)(*timeout * 1000000.0f);
      if (setsockopt(this->socket_fd, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv)) < 0) {
        LOG(ERROR, "failed to set socket receive timeout");
        perror("OS msg:");
        close();
        return false;
      }
    }

    return true;
  }
}  // namespace os
