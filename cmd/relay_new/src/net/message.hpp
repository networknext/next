#ifndef NET_MESSAGE_HPP
#define NET_MESSAGE_HPP

#include "address.hpp"
#include "core/packet.hpp"

namespace net
{
  struct Message
  {
    Message() = default;

    template <size_t DataBuffSize>
    Message(const Address& addr, std::array<uint8_t, DataBuffSize>& msg, size_t index, size_t len);

    ~Message() = default;

    Address Addr;
    std::vector<uint8_t> Data = std::vector<uint8_t>(0);
    size_t Len = 0;

    void swap(Message& other);
  };

  template <size_t DataBuffSize>
  inline Message::Message(const Address& addr, std::array<uint8_t, DataBuffSize>& msg, size_t index, size_t len)
  {
    Addr = addr;
    Data.resize(msg.size());
    std::copy(msg.begin() + index, msg.begin() + index + len, Data.begin());
  }

  [[gnu::always_inline]] inline void Message::swap(Message& other)
  {
    this->Addr.swap(other.Addr);
    this->Data.swap(other.Data);
  }

}  // namespace net
#endif