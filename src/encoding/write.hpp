#pragma once

#include "net/address.hpp"
#include "util/macros.hpp"

namespace encoding
{
  auto write_uint8(const uint8_t* buff, size_t buff_length, size_t& index, uint8_t value) -> bool;

  template <typename T>
  auto write_uint8(T& buff, size_t& index, uint8_t value) -> bool;

  auto write_uint16(const uint8_t* buff, size_t buff_length, size_t& index, uint16_t value) -> bool;

  template <typename T>
  auto write_uint16(T& buff, size_t& index, uint16_t value) -> bool;

  auto write_uint32(const uint8_t* buff, size_t buff_length, size_t& index, uint32_t value) -> bool;

  template <typename T>
  auto write_uint32(T& buff, size_t& index, uint32_t value) -> bool;

  auto write_uint64(const uint8_t* buff, size_t buff_length, size_t& index, uint64_t value) -> bool;

  template <typename T>
  auto write_uint64(T& buff, size_t& index, uint64_t value) -> bool;

  template <typename T>
  auto write_double(T& buff, size_t& index, double value) -> bool;

  auto write_bytes(uint8_t* buff, size_t buff_length, size_t& index, const uint8_t* const data, size_t data_length) -> bool;

  template <typename T, typename U>
  auto write_bytes(T& buff, size_t& index, const U& data, size_t len) -> bool;

  auto write_address(uint8_t* buff, size_t buff_length, size_t& index, const net::Address& addr) -> bool;

  template <typename T>
  auto write_address(T& buff, size_t& index, const net::Address& addr) -> bool;

  template <typename T>
  auto write_string(T& buff, size_t& index, const std::string& str) -> bool;

  INLINE auto write_uint8(uint8_t* buff, size_t buff_length, size_t& index, uint8_t value) -> bool
  {
    if (index + 1 > buff_length) {
      LOG(DEBUG, "index out of range: goal = ", index + 1, ", buff size = ", buff_length);
      return false;
    }

    buff[index++] = value;

    return true;
  }

  template <typename T>
  INLINE auto write_uint8(T& buff, size_t& index, uint8_t value) -> bool
  {
    if (index + 1 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 1, ", buff size = ", buff.size());
      return false;
    }

    buff[index++] = value;

    return true;
  }

  INLINE auto write_uint16(uint8_t* buff, size_t buff_length, size_t& index, uint16_t value) -> bool
  {
    if (index + 2 > buff_length) {
      LOG(DEBUG, "index out of range: goal = ", index + 2, ", buff size = ", buff_length);
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = value >> 8;

    return true;
  }

  template <typename T>
  INLINE auto write_uint16(T& buff, size_t& index, uint16_t value) -> bool
  {
    if (index + 2 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 2, ", buff size = ", buff.size());
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = value >> 8;

    return true;
  }

  INLINE auto write_uint32(uint8_t* buff, size_t buff_length, size_t& index, uint32_t value) -> bool
  {
    if (index + 4 > buff_length) {
      LOG(DEBUG, "index out of range: goal = ", index + 4, ", buff size = ", buff_length);
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = value >> 24;

    return true;
  }

  template <typename T>
  INLINE auto write_uint32(T& buff, size_t& index, uint32_t value) -> bool
  {
    if (index + 4 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 4, ", buff size = ", buff.size());
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = value >> 24;

    return true;
  }

  INLINE auto write_uint64(uint8_t* buff, size_t buff_length, size_t& index, uint64_t value) -> bool
  {
    if (index + 8 > buff_length) {
      LOG(DEBUG, "index out of range: goal = ", index + 8, ", buff size = ", buff_length);
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = (value >> 24) & 0xFF;
    buff[index++] = (value >> 32) & 0xFF;
    buff[index++] = (value >> 40) & 0xFF;
    buff[index++] = (value >> 48) & 0xFF;
    buff[index++] = value >> 56;

    return true;
  }

  template <typename T>
  INLINE auto write_uint64(T& buff, size_t& index, uint64_t value) -> bool
  {
    if (index + 8 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 8, ", buff size = ", buff.size());
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = (value >> 24) & 0xFF;
    buff[index++] = (value >> 32) & 0xFF;
    buff[index++] = (value >> 40) & 0xFF;
    buff[index++] = (value >> 48) & 0xFF;
    buff[index++] = value >> 56;

    return true;
  }

  template <typename T>
  INLINE auto write_double(T& buff, size_t& index, double value) -> bool
  {
    union
    {
      uint64_t fake;
      double actual;
    } values = {};
    values.actual = value;
    return encoding::write_uint64(buff, index, values.fake);
  }

  INLINE auto write_bytes(uint8_t* buff, size_t buff_length, size_t& index, const uint8_t* const data, size_t data_length)
   -> bool
  {
    if (index + data_length > buff_length) {
      LOG(DEBUG, "index out of range: goal = ", index + buff_length, ", buff size = ", buff_length);
      return false;
    }

    std::copy(data, data + data_length, buff + index);
    index += data_length;

    return true;
  }

  template <typename T, typename U>
  INLINE auto write_bytes(T& buff, size_t& index, const U& data, size_t len) -> bool
  {
    if (index + len > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + len, ", buff size = ", buff.size());
      return false;
    }

    std::copy(data.begin(), data.begin() + len, buff.begin() + index);
    index += len;

    return true;
  }

  INLINE auto write_address(uint8_t* buff, size_t buff_length, size_t& index, const net::Address& addr) -> bool
  {
    if (index + net::Address::SIZE_OF > buff_length) {
      LOG(DEBUG, "buffer too small for address");
      LOG(DEBUG, "index end = ", index + net::Address::SIZE_OF, ", buffer size = ", buff_length);
      return false;
    }

    if (addr.type == net::AddressType::IPv4) {
      // write the type
      if (!write_uint8(buff, buff_length, index, static_cast<uint8_t>(net::AddressType::IPv4))) {
        LOG(DEBUG, "buffer too small for address type");
        LOG(DEBUG, "index end = ", index + 1, ", buffer size = ", buff_length);
        return false;
      }

      std::copy(addr.ipv4.begin(), addr.ipv4.end(), buff + index);  // copy the address
      index += addr.ipv4.size() * sizeof(uint8_t);                  // increment the index

      // write the port
      if (!write_uint16(buff, buff_length, index, addr.port)) {
        LOG(DEBUG, "buffer too small for address port");
        LOG(DEBUG, "index end = ", index + 2, ", buffer size = ", buff_length);
        return false;
      }

      index += 12;  // increment the index past the address section
    } else if (addr.type == net::AddressType::IPv6) {
      // write the type
      if (!write_uint8(buff, buff_length, index, static_cast<uint8_t>(net::AddressType::IPv6))) {
        LOG(DEBUG, "buffer too small for address type");
        LOG(DEBUG, "index end = ", index + 1, ", buffer size = ", buff_length);
        return false;
      }

      for (const auto& ip : addr.ipv6) {
        if (!write_uint16(buff, buff_length, index, ip)) {
          LOG(DEBUG, "buffer too small for address part");
          LOG(DEBUG, "index end = ", index + 2, ", buffer size = ", buff_length);
          return false;
        }
      }

      if (!write_uint16(buff, buff_length, index, addr.port)) {
        LOG(DEBUG, "buffer too small for address port");
        LOG(DEBUG, "index end = ", index + 2, ", buffer size = ", buff_length);
        return false;
      }
    } else {
      std::fill(buff + index, buff + index + net::Address::SIZE_OF, 0);
      index += net::Address::SIZE_OF;
    }

    return true;
  }

  template <typename T>
  INLINE auto write_address(T& buff, size_t& index, const net::Address& addr) -> bool
  {
    if (index + net::Address::SIZE_OF > buff.size()) {
      LOG(TRACE, "buffer too small for address");
      LOG(TRACE, "index end = ", index + net::Address::SIZE_OF, ", buffer size = ", buff.size());
      return false;
    }

    if (addr.type == net::AddressType::IPv4) {
      // write the type
      if (!write_uint8(buff, index, static_cast<uint8_t>(net::AddressType::IPv4))) {
        LOG(TRACE, "buffer too small for address type");
        LOG(TRACE, "index end = ", index + 1, ", buffer size = ", buff.size());
        return false;
      }

      std::copy(addr.ipv4.begin(), addr.ipv4.end(), buff.begin() + index);  // copy the address
      index += addr.ipv4.size() * sizeof(uint8_t);                          // increment the index

      // write the port
      if (!write_uint16(buff, index, addr.port)) {
        LOG(TRACE, "buffer too small for address port");
        LOG(TRACE, "index end = ", index + 2, ", buffer size = ", buff.size());
        return false;
      }

      index += 12;  // increment the index past the address section
    } else if (addr.type == net::AddressType::IPv6) {
      // write the type
      if (!write_uint8(buff, index, static_cast<uint8_t>(net::AddressType::IPv6))) {
        LOG(TRACE, "buffer too small for address type");
        LOG(TRACE, "index end = ", index + 1, ", buffer size = ", buff.size());
        return false;
      }

      for (const auto& ip : addr.ipv6) {
        if (!write_uint16(buff, index, ip)) {
          LOG(TRACE, "buffer too small for address part");
          LOG(TRACE, "index end = ", index + 2, ", buffer size = ", buff.size());
          return false;
        }
      }

      if (!write_uint16(buff, index, addr.port)) {
        LOG(TRACE, "buffer too small for address port");
        LOG(TRACE, "index end = ", index + 2, ", buffer size = ", buff.size());
        return false;
      }
    } else {
      std::fill(buff.begin() + index, buff.begin() + index + net::Address::SIZE_OF, 0);
      index += net::Address::SIZE_OF;
    }

    return true;
  }

  template <typename T>
  INLINE auto write_string(T& buff, size_t& index, const std::string& str) -> bool
  {
    if (index + 4 + str.length() > buff.size()) {
      LOG(TRACE, "buffer too small for string");
      return false;
    }

    // sanity check
    if (!encoding::write_uint32(buff, index, str.length())) {
      LOG(TRACE, "could not write string length");
      return false;
    }

    for (const auto c : str) {
      buff[index++] = c;
    }

    return true;
  }
}  // namespace encoding
