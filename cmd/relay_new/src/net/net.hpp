#ifndef UTIL_NET_H
#define UTIL_NET_H

#include "encoding/binary.hpp"

namespace net
{
  /**
      Template to convert an integer value from local byte order to network byte order.
      IMPORTANT: Because most machines running relay are little endian, relay defines network byte order to be little endian.
      @param value The input value in local byte order. Supported integer types: uint64_t, uint32_t, uint16_t.
      @returns The input value converted to network byte order. If this processor is little endian the output is the same as
     the input. If the processor is big endian, the output is the input byte swapped.
   */

  template <typename T>
  T host_to_network(T value)
  {
#if RELAY_BIG_ENDIAN
    return bswap(value);
#else   // #if RELAY_BIG_ENDIAN
    return value;
#endif  // #if RELAY_BIG_ENDIAN
  }

  /**
      Template to convert an integer value from network byte order to local byte order.
      IMPORTANT: Because most machines running relay are little endian, relay defines network byte order to be little endian.
      @param value The input value in network byte order. Supported integer types: uint64_t, uint32_t, uint16_t.
      @returns The input value converted to local byte order. If this processor is little endian the output is the same as the
     input. If the processor is big endian, the output is the input byte swapped.
   */

  template <typename T>
  T network_to_host(T value)
  {
#if RELAY_BIG_ENDIAN
    return bswap(value);
#else   // #if RELAY_BIG_ENDIAN
    return value;
#endif  // #if RELAY_BIG_ENDIAN
  }

  // TODO same as the funcs in relay_platform_*.hpp/cpp, which were previously relay_platform_ntohs/htons()
  // figure out if these are needed
  inline uint16_t relay_ntohs(uint16_t in)
  {
    return (uint16_t)(((in << 8) & 0xFF00) | ((in >> 8) & 0x00FF));
  }

  inline uint16_t relay_htons(uint16_t in)
  {
    return (uint16_t)(((in << 8) & 0xFF00) | ((in >> 8) & 0x00FF));
  }
}  // namespace net
#endif
