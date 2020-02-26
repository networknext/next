#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#ifndef OS_LINUX_SOCKET
#define OS_LINUX_SOCKET

#include "net/address.hpp"
#include "net/message.hpp"
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

    bool send(const net::Address& to, const uint8_t* data, size_t size) const;
    bool multisend(const std::vector<net::Message>& multiMessages, size_t count, int& messagesSent) const;

    // for compat only
    bool send(const legacy::relay_address_t& to, const uint8_t* data, size_t size) const;

    size_t recv(net::Address& from, uint8_t* data, size_t maxSize) const;

    // for compat only
    size_t recv(legacy::relay_address_t& from, uint8_t* data, size_t maxSize) const;

    void close();

    bool isOpen() const;

    void setBusy(bool val);

    const net::Address& getAddress() const;

   private:
    int mSockFD = 0;
    const SocketType mType;
    net::Address mAddress;

    std::atomic<bool> mOpen;
    std::atomic<bool> mIsBusy;
    std::mutex mWaitLock;
    std::condition_variable mWaitVar;

    bool setBufferSizes(size_t sendBufferSize, size_t recvBufferSize);
    bool setLingerTime(int lingerTime);
    bool setPortReuse(bool reuse);

    bool bindIPv4(const net::Address& addr);
    bool bindIPv6(const net::Address& addr);

    bool getPortIPv4(net::Address& addr);
    bool getPortIPv6(net::Address& addr);

    bool setSocketType(float timeout);
  };

  [[gnu::always_inline]] inline const net::Address& Socket::getAddress() const
  {
    return mAddress;
  }

  [[gnu::always_inline]] inline bool Socket::isOpen() const
  {
    return mOpen;
  }

  [[gnu::always_inline]] inline void Socket::close()
  {
    std::unique_lock<std::mutex> lk(mWaitLock);
    mWaitVar.wait(lk, [this]() -> bool {
      return !mIsBusy;
    });

    shutdown(mSockFD, SHUT_RDWR);
    mOpen = false;
  }

  [[gnu::always_inline]] inline void Socket::setBusy(bool val)
  {
    mIsBusy = val;
    if (!mIsBusy) {
      mWaitVar.notify_one();
    }
  }

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
