#ifndef NET_MESSAGE_HPP
#define NET_MESSAGE_HPP

#include "address.hpp"
#include "core/packet.hpp"

namespace net
{
  struct Message
  {
    Message() = default;

    Message(const Address& addr, std::vector<uint8_t>& msg);

    template <size_t DataBuffSize>
    Message(const Address& addr, std::array<uint8_t, DataBuffSize>& msg, size_t len);

    Message(const Address& addr, uint8_t* msg, size_t len);

    ~Message() = default;

    Address Addr;
    core::GenericPacket Data;
    size_t Len = 0;

    void swap(Message& other);
  };

  inline Message::Message(const Address& addr, std::vector<uint8_t>& msg)
  {
    Addr = addr;
    std::copy(msg.begin(), msg.end(), Data.begin());
  }

  template <size_t DataBuffSize>
  inline Message::Message(const Address& addr, std::array<uint8_t, DataBuffSize>& msg, size_t len)
  {
    Addr = addr;
    std::copy(msg.begin(), msg.begin() + len, Data.begin());
  }

  inline Message::Message(const Address& addr, uint8_t* msg, size_t len)
  {
    Addr = addr;
    std::copy(msg, msg + len, Data.begin());
  }

  [[gnu::always_inline]] inline void Message::swap(Message& other)
  {
    this->Addr.swap(other.Addr);
    this->Data.swap(other.Data);
  }

}  // namespace net
#endif