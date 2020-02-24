#ifndef NET_BUFFERED_SENDER_HPP
#define NET_BUFFERED_SENDER_HPP

#include "net/message.hpp"
#include "os/platform.hpp"

namespace net
{
  template <size_t MaxCapacity>
  class BufferedSender
  {
   public:
    BufferedSender(const os::Socket& socket);
    ~BufferedSender() = default;

   private:
    const os::Socket& mSocket;
    std::array<net::Message, MaxCapacity> mBuffer;
  };
}  // namespace net
#endif