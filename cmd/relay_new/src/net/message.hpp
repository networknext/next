#ifndef NET_MESSAGE_HPP
#define NET_MESSAGE_HPP

#include "address.hpp"

namespace net
{
  struct Message
  {
    Address Addr;
    std::vector<uint8_t> Msg;

    void swap(Message& other);
  };

  [[gnu::always_inline]] inline void Message::swap(Message& other) {
    this->Addr.swap(other.Addr);
    this->Msg.swap(other.Msg);
  }

}  // namespace net
#endif