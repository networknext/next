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

  class Socket
  {
   public:
    Socket();
    ~Socket();

    auto create(SocketType type, Address& addr, size_t sendBuffSize, size_t recvBuffSize, float timeout, bool reuse) -> bool;

    // uses sendto()
    auto send(const Packet& packet) const -> bool;

    // uses sendto()
    auto send(const Address& to, const uint8_t* data, size_t size) const -> bool;

    // uses sendmmsg()
    template <size_t BuffSize>
    auto multisend(PacketBuffer<BuffSize>& packetBuff) const -> bool;

    // uses sendmmsg()
    template <size_t BuffSize>
    auto multisend(std::array<mmsghdr, BuffSize>& packetBuff, int count) const -> bool;

    // uses recvfrom()
    auto recv(Address& from, uint8_t* data, size_t maxSize) const -> int;

    // uses recvfrom()
    auto recv(Packet& packet) const -> bool;

    // uses recvmmsg()
    template <size_t BuffSize>
    auto multirecv(PacketBuffer<BuffSize>& packetBuff) const -> bool;

    // close the socket
    void close();

    // returns if the socket is closed
    auto closed() const -> bool;

   private:
    int mSockFD = 0;
    SocketType mType;
    std::atomic<bool> mClosed;

    auto set_buffer_sizes(size_t sendBufferSize, size_t recvBufferSize) -> bool;
    auto set_port_reuse(bool reuse) -> bool;

    auto bind_ipv4(const Address& addr) -> bool;
    auto bind_ipv6(const Address& addr) -> bool;

    auto get_port_ipv4(Address& addr) -> bool;
    auto get_port_ipv6(Address& addr) -> bool;

    auto set_socket_type(float timeout) -> bool;
  };

  using SocketPtr = std::shared_ptr<Socket>;

  INLINE Socket::Socket(): mClosed(false) {}

  INLINE Socket::~Socket()
  {
    if (mSockFD) {
      close();
    }
  }

  INLINE auto Socket::create(
   SocketType type, Address& addr, size_t sendBuffSize, size_t recvBuffSize, float timeout, bool reuse) -> bool
  {
    assert(addr.Type != AddressType::None);

    // create socket
    {
      mSockFD = socket((addr.Type == AddressType::IPv4) ? PF_INET : PF_INET6, SOCK_DGRAM, IPPROTO_UDP);

      if (mSockFD < 0) {
        LOG(ERROR, "failed to create socket");
        perror("OS msg:");
        return false;
      }
    }

    // force IPv6 only if necessary
    {
      if (addr.Type == AddressType::IPv6) {
        int enable = 1;
        if (setsockopt(mSockFD, IPPROTO_IPV6, IPV6_V6ONLY, &enable, sizeof(enable)) != 0) {
          LOG(ERROR, "failed to set socket ipv6 only");
          close();
          return false;
        }
      }
    }

    if (!this->set_buffer_sizes(sendBuffSize, recvBuffSize)) {
      return false;
    }

    if (!set_port_reuse(reuse)) {
      return false;
    }

    // bind to port
    {
      if (addr.Type == AddressType::IPv6) {
        if (!bind_ipv6(addr)) {
          return false;
        }
      } else {
        if (!bind_ipv4(addr)) {
          return false;
        }
      }
    }

    // if bound to port 0, find the actual port we got
    // port 0 is a "wildcard" so using it will bind to any available port
    {
      if (addr.Port == 0) {
        if (addr.Type == AddressType::IPv6) {
          if (!get_port_ipv6(addr)) {
            return false;
          }
        } else {
          if (!get_port_ipv4(addr)) {
            return false;
          }
        }
      }
    }

    mType = type;

    if (!set_socket_type(timeout)) {
      return false;
    }

    return true;
  }

  INLINE auto Socket::send(const Packet& packet) const -> bool
  {
    return send(packet.Addr, packet.Buffer.data(), packet.Len);
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

    if (mClosed) {
      return false;
    }

    if (to.Type == AddressType::IPv6) {
      sockaddr_in6 socket_address;
      to.into(socket_address);

      auto res = sendto(mSockFD, data, size, 0, reinterpret_cast<sockaddr*>(&socket_address), sizeof(sockaddr_in6));
      if (res < 0) {
        LOG(ERROR, "sendto (", to, ") failed");
        return false;
      }
    } else if (to.Type == AddressType::IPv4) {
      sockaddr_in socket_address;
      to.into(socket_address);

      auto res = sendto(mSockFD, data, size, 0, reinterpret_cast<sockaddr*>(&socket_address), sizeof(sockaddr_in6));
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
  INLINE auto Socket::multisend(std::array<mmsghdr, BuffSize>& headers, int count) const -> bool
  {
    static_assert(BuffSize <= 1024);  // max sendmmsg will allow

    assert(count > 0);
    assert(count <= 1024);

    auto toSend = count;
    count = sendmmsg(mSockFD, headers.data(), toSend, 0);

    if (count < 0) {
      LOG(ERROR, "sendmmsg() failed");
      return false;
    }

    return toSend == count;
  }

  template <size_t BuffSize>
  INLINE auto Socket::multisend(PacketBuffer<BuffSize>& packetBuff) const -> bool
  {
    return multisend(packetBuff.Headers, packetBuff.Count);
  }

  INLINE auto Socket::recv(Packet& packet) const -> bool
  {
    auto len = this->recv(packet.Addr, packet.Buffer.data(), packet.Buffer.size());
    packet.Len = static_cast<size_t>(len);
    return len > 0;
  }

  INLINE auto Socket::recv(Address& from, uint8_t* data, size_t maxSize) const -> int
  {
    assert(data != nullptr);
    assert(maxSize > 0);

    if (mClosed) {
      return 0;
    }

    sockaddr_storage sockaddr_from = {};

    socklen_t len = sizeof(sockaddr_from);
    auto res = recvfrom(
     mSockFD,
     data,
     maxSize,
     (mType == SocketType::NonBlocking) ? MSG_DONTWAIT : 0,
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
  INLINE auto Socket::multirecv(PacketBuffer<BuffSize>& packetBuff) const -> bool
  {
    packetBuff.Count = recvmmsg(
     mSockFD,
     packetBuff.Headers.data(),
     BuffSize,
     MSG_WAITFORONE,
     nullptr);  // DON'T EVER USE TIMEOUT, linux man pages state it is broken

    if (packetBuff.Count < 0) {
      LOG(ERROR, "recvmmsg failed");
      return false;
    }

    return true;
  }

  INLINE void Socket::close()
  {
    mClosed = true;
    shutdown(mSockFD, SHUT_RDWR);
    mSockFD = -1;
  }

  INLINE auto Socket::closed() const -> bool
  {
    return mClosed;
  }

  INLINE auto Socket::set_buffer_sizes(size_t sendBuffSize, size_t recvBuffSize) -> bool
  {
    if (setsockopt(mSockFD, SOL_SOCKET, SO_SNDBUF, &sendBuffSize, sizeof(sendBuffSize)) != 0) {
      LOG(ERROR, "failed to set socket send buffer size");
      this->close();
      return false;
    }

    if (setsockopt(mSockFD, SOL_SOCKET, SO_RCVBUF, &recvBuffSize, sizeof(recvBuffSize)) != 0) {
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
      if (setsockopt(mSockFD, SOL_SOCKET, SO_REUSEPORT, &enable, sizeof(enable)) < 0) {
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

    if (bind(mSockFD, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
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

    if (bind(mSockFD, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
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
    if (getsockname(mSockFD, reinterpret_cast<sockaddr*>(&sin), &len) < 0) {
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
    if (getsockname(mSockFD, reinterpret_cast<sockaddr*>(&sin), &len) < 0) {
      LOG(ERROR, "failed to get socket port (ipv6)");
      perror("OS msg:");
      close();
      return false;
    }
    addr.Port = ntohs(sin.sin6_port);
    return true;
  }

  INLINE auto Socket::set_socket_type(float timeout) -> bool
  {
    // set non-blocking io or receive timeout, or if neither then just blocking with no timeout
    if (mType == SocketType::NonBlocking) {
      if (fcntl(mSockFD, F_SETFL, O_NONBLOCK, 1) < 0) {
        LOG(ERROR, "failed to set socket to non blocking");
        perror("OS msg:");
        close();
        return false;
      }
    } else if (mType == SocketType::Blocking && timeout > 0.0f) {
      timeval tv;
      tv.tv_sec = 0;
      tv.tv_usec = (int)(timeout * 1000000.0f);
      if (setsockopt(mSockFD, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv)) < 0) {
        LOG(ERROR, "failed to set socket receive timeout");
        perror("OS msg:");
        close();
        return false;
      }
    }

    return true;
  }
}  // namespace os
