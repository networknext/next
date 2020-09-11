#pragma once

#include "net/address.hpp"
#include "util/logger.hpp"

using net::Address;
using net::AddressType;

namespace encoding
{
  auto read_uint8(const char* const buff, size_t bufflen, size_t& index, uint8_t& value) -> bool;

  template <typename T>
  auto read_uint8(const T& buff, size_t& index, uint8_t& value) -> bool;

  auto read_uint16(const uint8_t* const buff, size_t bufflen, size_t& index, uint16_t& value) -> bool;

  template <typename T>
  auto read_uint16(const T& buff, size_t& index, uint16_t& value) -> bool;

  auto read_uint32(const uint8_t* const buff, size_t bufflen, size_t& index, uint32_t& value) -> bool;

  template <typename T>
  auto read_uint32(const T& buff, size_t& index, uint32_t& value) -> bool;

  auto read_uint64(const uint8_t* const buff, size_t bufflen, size_t& index, uint64_t& value) -> bool;

  template <typename T>
  auto read_uint64(const T& buff, size_t& index, uint64_t& value) -> bool;

  template <typename T>
  auto read_double(const T& buff, size_t& index, double& value) -> bool;

  auto read_bytes(
   const uint8_t* const buff, size_t buff_length, size_t& index, uint8_t* storage, size_t storage_length, size_t len) -> bool;

  template <typename T, typename U>
  auto read_bytes(const T& buff, size_t& index, U& storage, size_t len) -> bool;

  auto read_address(const uint8_t* const buff, size_t buff_length, size_t& index, net::Address& addr) -> bool;

  template <typename T>
  auto read_address(const T& buff, size_t& index, net::Address& addr) -> bool;

  template <typename T>
  auto read_string(const T& buff, size_t& index, std::string& str) -> bool;

  INLINE auto read_uint8(const uint8_t* const buff, size_t bufflen, size_t& index, uint8_t& value) -> bool
  {
    if (index + 1 > bufflen) {
      return false;
    }
    value = buff[index++];
    return true;
  }

  template <typename T>
  INLINE auto read_uint8(const T& buff, size_t& index, uint8_t& value) -> bool
  {
    if (index + 1 > buff.size()) {
      return false;
    }
    value = buff[index++];
    return true;
  }

  INLINE auto read_uint16(const uint8_t* const buff, size_t bufflen, size_t& index, uint16_t& value) -> bool
  {
    if (index + 2 > bufflen) {
      return false;
    }
    value = (buff)[index++];
    value |= (static_cast<uint64_t>(buff[index++]) << 8);
    return true;
  }

  template <typename T>
  INLINE auto read_uint16(const T& buff, size_t& index, uint16_t& value) -> bool
  {
    if (index + 2 > buff.size()) {
      return false;
    }
    value = (buff)[index++];
    value |= (static_cast<uint64_t>(buff[index++]) << 8);
    return true;
  }

  INLINE auto read_uint32(const uint8_t* const buff, size_t bufflen, size_t& index, uint32_t& value) -> bool
  {
    if (index + 4 > bufflen) {
      return false;
    }
    value = buff[index++];
    value |= (static_cast<uint32_t>(buff[index++]) << 8);
    value |= (static_cast<uint32_t>(buff[index++]) << 16);
    value |= (static_cast<uint32_t>(buff[index++]) << 24);
    return true;
  }

  template <typename T>
  INLINE auto read_uint32(const T& buff, size_t& index, uint32_t& value) -> bool
  {
    if (index + 4 > buff.size()) {
      return false;
    }
    value = buff[index++];
    value |= (static_cast<uint32_t>(buff[index++]) << 8);
    value |= (static_cast<uint32_t>(buff[index++]) << 16);
    value |= (static_cast<uint32_t>(buff[index++]) << 24);
    return true;
  }

  INLINE auto read_uint64(const uint8_t* const buff, size_t bufflen, size_t& index, uint64_t& value) -> bool
  {
    if (index + 8 > bufflen) {
      return false;
    }
    value = buff[index++];
    value |= (static_cast<uint64_t>(buff[index++]) << 8);
    value |= (static_cast<uint64_t>(buff[index++]) << 16);
    value |= (static_cast<uint64_t>(buff[index++]) << 24);
    value |= (static_cast<uint64_t>(buff[index++]) << 32);
    value |= (static_cast<uint64_t>(buff[index++]) << 40);
    value |= (static_cast<uint64_t>(buff[index++]) << 48);
    value |= (static_cast<uint64_t>(buff[index++]) << 56);
    return true;
  }

  template <typename T>
  INLINE auto read_uint64(const T& buff, size_t& index, uint64_t& value) -> bool
  {
    if (index + 8 > buff.size()) {
      return false;
    }
    value = buff[index++];
    value |= (static_cast<uint64_t>(buff[index++]) << 8);
    value |= (static_cast<uint64_t>(buff[index++]) << 16);
    value |= (static_cast<uint64_t>(buff[index++]) << 24);
    value |= (static_cast<uint64_t>(buff[index++]) << 32);
    value |= (static_cast<uint64_t>(buff[index++]) << 40);
    value |= (static_cast<uint64_t>(buff[index++]) << 48);
    value |= (static_cast<uint64_t>(buff[index++]) << 56);
    return true;
  }

  template <typename T>
  INLINE auto read_double(const T& buff, size_t& index, double& value) -> bool
  {
    union
    {
      uint64_t fake;
      double actual;
    } values = {};
    bool retval = encoding::read_uint64(buff, index, values.fake);
    value = values.actual;
    return retval;
  }

  INLINE auto read_bytes(
   const uint8_t* const buff, size_t buff_length, size_t& index, uint8_t* storage, size_t storage_length, size_t read_len)
   -> bool
  {
    if (index + read_len > buff_length) {
      return false;
    }
    if (read_len > storage_length) {
      return false;
    }
    std::copy(buff + index, buff + index + read_len, storage);
    index += read_len;
    return true;
  }

  template <typename T, typename U>
  INLINE auto read_bytes(const T& buff, size_t& index, U& storage, size_t read_len) -> bool
  {
    if (index + read_len > buff.size()) {
      return false;
    }
    if (read_len > storage.size()) {
      return false;
    }
    std::copy(buff.begin() + index, buff.begin() + index + read_len, storage.begin());
    index += read_len;
    return true;
  }

  INLINE auto read_address(const uint8_t* const buff, size_t buff_length, size_t& index, net::Address& addr) -> bool
  {
    if (buff_length < Address::SIZE_OF) {
      return false;
    }

    uint8_t type;
    if (!read_uint8(buff, buff_length, index, type)) {
      return false;
    }
    addr.type = static_cast<AddressType>(type);

    switch (addr.type) {
      case AddressType::IPv4: {
        // read address parts
        std::copy(buff + index, buff + index + 4, addr.ipv4.data());
        index += 4;
        // read the port
        if (!read_uint16(buff, buff_length, index, addr.port)) {
          return false;
        }
        index += 12;  // increment the index past the reserved area
      } break;
      case AddressType::IPv6: {
        // read address parts
        for (int i = 0; i < 8; i++) {
          if (!read_uint16(buff, buff_length, index, addr.ipv6[i])) {
            return false;
          }
        }
        // read the port
        if (!read_uint16(buff, buff_length, index, addr.port)) {
          return false;
        }
      } break;
      default: {
        // if no type, increment the index past the address area
        index += Address::SIZE_OF - 1;
        addr.reset();
      } break;
    }

    return true;
  }

  template <typename T>
  INLINE auto read_address(const T& buff, size_t& index, net::Address& addr) -> bool
  {
    uint8_t type;
    if (!read_uint8(buff, index, type)) {
      return false;
    }
    addr.type = static_cast<AddressType>(type);

    if (addr.type == net::AddressType::IPv4) {
      // copy the address parts
      std::copy(buff.begin() + index, buff.begin() + index + 4, addr.ipv4.begin());
      index += 4;
      // read the port
      if (!read_uint16(buff, index, addr.port)) {
        return false;
      }
      index += 12;  // increment the index past the reserved area
    } else if (addr.type == net::AddressType::IPv6) {
      for (int i = 0; i < 8; i++) {
        if (!read_uint16(buff, index, addr.ipv6[i])) {
          return false;
        }
      }
      if (!read_uint16(buff, index, addr.port)) {
        return false;
      }
    } else {
      addr.reset();
      index += net::Address::SIZE_OF - 1;  // if no type, increment the index past the address area
    }

    return true;
  }

  template <typename T>
  INLINE auto read_string(const T& buff, size_t& index, std::string& value) -> bool
  {
    uint32_t len;
    if (!read_uint32(buff, index, len)) {
      return false;
    }
    value = std::move(std::string(buff.begin() + index, buff.begin() + index + len));
    index += len;
    return true;
  }
}  // namespace encoding
