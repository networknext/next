#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#ifndef OS_LINUX_SOCKET
#define OS_LINUX_SOCKET

#include "core/packet.hpp"

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
     net::Address& addr, size_t sendBuffSize, size_t recvBuffSize, float timeout, bool reuse);

    bool send(const net::Address& to, const uint8_t* data, size_t size) const;

    template <typename T>
    bool send(const core::Packet<T>& packet) const;

    template <size_t BuffSize>
    bool multisend(std::array<mmsghdr, BuffSize>& packetBuff, int count) const;

    template <size_t BuffSize, size_t PacketSize>
    bool multisend(core::GenericPacketBuffer<BuffSize, PacketSize>& packetBuff) const;

    auto recv(net::Address& from, uint8_t* data, size_t maxSize) const -> int;

    template <typename T>
    auto recv(core::Packet<T>& packet) const -> bool;

    template <size_t BuffSize, size_t PacketSize>
    bool multirecv(core::GenericPacketBuffer<BuffSize, PacketSize>& packetBuff) const;

    void close();
    auto closed() const -> bool;

    const net::Address& getAddress() const;

   private:
    int mSockFD = 0;
    const SocketType mType;
    std::atomic<bool> mClosed;

    bool setBufferSizes(size_t sendBufferSize, size_t recvBufferSize);
    bool setLingerTime(int lingerTime);
    bool setPortReuse(bool reuse);

    bool bindIPv4(const net::Address& addr);
    bool bindIPv6(const net::Address& addr);

    bool getPortIPv4(net::Address& addr);
    bool getPortIPv6(net::Address& addr);

    bool setSocketType(float timeout);
  };

  [[gnu::always_inline]] inline void Socket::close()
  {
    mClosed = true;
    shutdown(mSockFD, SHUT_RDWR);
  }

  [[gnu::always_inline]] inline auto Socket::closed() const -> bool
  {
    return mClosed;
  }

  template <typename T>
  bool Socket::send(const core::Packet<T>& packet) const
  {
    return send(packet.Addr, packet.Buffer.data(), packet.Len);
  }

  template <size_t BuffSize>
  bool Socket::multisend(std::array<mmsghdr, BuffSize>& headers, int count) const
  {
    static_assert(BuffSize <= 1024);  // max sendmmsg will allow

    assert(count > 0);
    assert(count <= 1024);

    auto toSend = count;
    count = sendmmsg(mSockFD, headers.data(), toSend, 0);

    if (count < 0) {
      LogError("sendmmsg() failed");
      return false;
    }

    return toSend == count;
  }

  template <size_t BuffSize, size_t PacketSize>
  inline bool Socket::multisend(core::GenericPacketBuffer<BuffSize, PacketSize>& packetBuff) const
  {
    return multisend(packetBuff.Headers, packetBuff.Count);
  }

  template <typename T>
  inline auto Socket::recv(core::Packet<T>& packet) const -> bool
  {
    auto len = this->recv(packet.Addr, packet.Buffer.data(), packet.Buffer.size());
    packet.Len = static_cast<size_t>(len);
    return len > 0;
  }

  template <size_t BuffSize, size_t PacketSize>
  bool Socket::multirecv(core::GenericPacketBuffer<BuffSize, PacketSize>& packetBuff) const
  {
    packetBuff.Count = recvmmsg(
     mSockFD,
     packetBuff.Headers.data(),
     BuffSize,
     MSG_WAITFORONE,
     nullptr);  // DON'T EVER USE TIMEOUT, linux man pages state it is broken

    if (packetBuff.Count < 0) {
      LogError("recvmmsg failed");
      return false;
    }

    return true;
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
