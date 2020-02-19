#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#ifndef OS_LINUX_SOCKET
#define OS_LINUX_SOCKET

#include "net/address.hpp"
#include "net/net.hpp"
#include "util/logger.hpp"

#include "relay/relay_platform.hpp"

namespace os
{
  enum class SocketType : uint8_t
  {
    NonBlocking,
    Blocking
  };

  class Socket
  {
   public:
    Socket(SocketType type);
    ~Socket();

    bool create(net::Address& addr, size_t sendBuffSize, size_t recvBuffSize, float timeout, bool reuse);

    // for compat only
    bool create(legacy::relay_address_t& addr, size_t sendBuffSize, size_t recvBuffSize, float timeout, bool reuse);

    bool send(const net::Address& to, const void* data, size_t size);

    // for compat only
    bool send(const legacy::relay_address_t& to, const void* data, size_t size);

    size_t recv(net::Address& from, void* data, size_t maxSize);

    // for compat only
    size_t recv(legacy::relay_address_t& from, void* data, size_t maxSize);

    void close();

   private:
    int mSockFD = 0;
    const SocketType mType;
    std::vector<uint8_t> mSendBuffer;
    std::vector<uint8_t> mReceiveBuffer;

    bool bindIPv4(const net::Address& addr);
    bool bindIPv6(const net::Address& addr);

    bool getPortIPv4(net::Address& addr);
    bool getPortIPv6(net::Address& addr);
  };

  inline bool operator==(SocketType st, int i)
  {
    return static_cast<int>(st) == i;
  }

  inline bool operator==(int i, SocketType st)
  {
    return static_cast<int>(st) == i;
  }

  inline bool Socket::bindIPv4(const net::Address& addr)
  {
    sockaddr_in socket_address;
    bzero(&socket_address, sizeof(socket_address));
    socket_address.sin_family = AF_INET;
    socket_address.sin_addr.s_addr = (((uint32_t)addr.IPv4[0])) | (((uint32_t)addr.IPv4[1]) << 8) |
                                     (((uint32_t)addr.IPv4[2]) << 16) | (((uint32_t)addr.IPv4[3]) << 24);
    socket_address.sin_port = net::relay_htons(addr.Port);

    if (bind(mSockFD, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
      LogError("failed to bind socket (ipv4)");
      close();
      return false;
    }

    return true;
  }

  inline bool Socket::bindIPv6(const net::Address& addr)
  {
    sockaddr_in6 socket_address;
    bzero(&socket_address, sizeof(socket_address));

    socket_address.sin6_family = AF_INET6;
    for (int i = 0; i < 8; i++) {
      reinterpret_cast<uint16_t*>(&socket_address.sin6_addr)[i] = net::relay_htons(addr.IPv6[i]);
    }

    socket_address.sin6_port = net::relay_htons(addr.Port);

    if (bind(mSockFD, reinterpret_cast<sockaddr*>(&socket_address), sizeof(socket_address)) < 0) {
      LogError("failed to bind socket (ipv6)");
      close();
      return false;
    }

    return true;
  }

  inline bool Socket::getPortIPv4(net::Address& addr)
  {
    sockaddr_in sin;
    socklen_t len = sizeof(len);
    if (getsockname(mSockFD, reinterpret_cast<sockaddr*>(&sin), &len) < 0) {
      LogError("failed to get socket port (ipv4)");
      close();
      return false;
    }
    addr.Port = relay::relay_platform_ntohs(sin.sin_port);
    return true;
  }

  inline bool Socket::getPortIPv6(net::Address& addr)
  {
    sockaddr_in6 sin;
    socklen_t len = sizeof(sin);
    if (getsockname(mSockFD, reinterpret_cast<sockaddr*>(&sin), &len) < 0) {
      LogError("failed to get socket port (ipv6)");
      close();
      return false;
    }
    addr.Port = relay::relay_platform_ntohs(sin.sin6_port);
    return true;
  }

  using SocketPtr = std::shared_ptr<Socket>;
}  // namespace os

namespace legacy
{
  typedef int relay_platform_socket_handle_t;

  struct relay_platform_socket_t
  {
    int type;
    relay_platform_socket_handle_t handle;
  };

  relay_platform_socket_t* relay_platform_socket_create(
   legacy::relay_address_t* address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size);

  void relay_platform_socket_destroy(relay_platform_socket_t* socket);

  void relay_platform_socket_send_packet(
   relay_platform_socket_t* socket, legacy::relay_address_t* to, const void* packet_data, int packet_bytes);

  int relay_platform_socket_receive_packet(
   relay_platform_socket_t* socket, legacy::relay_address_t* from, void* packet_data, int max_packet_size);
}  // namespace legacy
#endif

#endif
