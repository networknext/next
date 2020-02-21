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

    bool create(
     net::Address& addr, size_t sendBuffSize, size_t recvBuffSize, float timeout, bool reuse, int lingerTimeInSeconds);

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

    bool setBufferSizes(size_t sendBufferSize, size_t recvBufferSize);
    bool setLingerTime(int lingerTime);
    bool setPortReuse(bool reuse);

    bool bindIPv4(const net::Address& addr);
    bool bindIPv6(const net::Address& addr);

    bool getPortIPv4(net::Address& addr);
    bool getPortIPv6(net::Address& addr);

    bool setSocketType(float timeout);
  };

  // helpers to reduce static cast's

  inline bool operator==(SocketType st, int i)
  {
    return static_cast<int>(st) == i;
  }

  inline bool operator==(int i, SocketType st)
  {
    return static_cast<int>(st) == i;
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
