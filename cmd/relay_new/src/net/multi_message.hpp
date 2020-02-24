#ifndef NET_MULTI_MESSAGE_HPP
#define NET_MULTI_MESSAGE_HPP

#include "address.hpp"

namespace net
{
  struct MultiMessage
  {
    Address Addr;
    std::vector<uint8_t> Msg;
  };
}  // namespace net
#endif