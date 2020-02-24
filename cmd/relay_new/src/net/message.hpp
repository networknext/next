#ifndef NET_MESSAGE_HPP
#define NET_MESSAGE_HPP

#include "address.hpp"

namespace net
{
  struct Message
  {
    Address Addr;
    std::vector<uint8_t> Msg;
  };
}  // namespace net
#endif