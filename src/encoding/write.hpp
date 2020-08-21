#pragma once

#include "net/address.hpp"
#include "util/macros.hpp"

namespace encoding
{
  auto WriteUint8(const uint8_t* buff, size_t buffLength, size_t& index, uint8_t value) -> bool;

  template <typename T>
  auto WriteUint8(T& buff, size_t& index, uint8_t value) -> bool;

  auto WriteUint16(const uint8_t* buff, size_t buffLength, size_t& index, uint16_t value) -> bool;

  template <typename T>
  auto WriteUint16(T& buff, size_t& index, uint16_t value) -> bool;

  auto WriteUint32(const uint8_t* buff, size_t buffLength, size_t& index, uint32_t value) -> bool;

  template <typename T>
  auto WriteUint32(T& buff, size_t& index, uint32_t value) -> bool;

  auto WriteUint64(const uint8_t* buff, size_t buffLength, size_t& index, uint64_t value) -> bool;

  template <typename T>
  auto WriteUint64(T& buff, size_t& index, uint64_t value) -> bool;

  template <typename T>
  auto WriteDouble(T& buff, size_t& index, double value) -> bool;

  auto WriteBytes(const uint8_t* buff, size_t buffLength, size_t& index, const uint8_t* const data, size_t dataLength) -> bool;

  template <typename T, typename U>
  auto WriteBytes(T& buff, size_t& index, const U& data, size_t len) -> bool;

  auto WriteAddress(uint8_t* buff, size_t buffLength, size_t& index, const net::Address& addr) -> bool;

  template <typename T>
  auto WriteAddress(T& buff, size_t& index, const net::Address& addr) -> bool;

  template <typename T>
  auto WriteString(T& buff, size_t& index, const std::string& str) -> bool;

  INLINE auto WriteUint8(uint8_t* buff, size_t buffLength, size_t& index, uint8_t value) -> bool
  {
    if (index + 1 > buffLength) {
      LOG(DEBUG, "index out of range: goal = ", index + 1, ", buff size = ", buffLength);
      return false;
    }

    buff[index++] = value;

    return true;
  }

  template <typename T>
  INLINE auto WriteUint8(T& buff, size_t& index, uint8_t value) -> bool
  {
    if (index + 1 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 1, ", buff size = ", buff.size());
      return false;
    }

    buff[index++] = value;

    return true;
  }

  INLINE auto WriteUint16(uint8_t* buff, size_t buffLength, size_t& index, uint16_t value) -> bool
  {
    if (index + 2 > buffLength) {
      LOG(DEBUG, "index out of range: goal = ", index + 2, ", buff size = ", buffLength);
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = value >> 8;

    return true;
  }

  template <typename T>
  INLINE auto WriteUint16(T& buff, size_t& index, uint16_t value) -> bool
  {
    if (index + 2 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 2, ", buff size = ", buff.size());
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = value >> 8;

    return true;
  }

  INLINE auto WriteUint32(uint8_t* buff, size_t buffLength, size_t& index, uint32_t value) -> bool
  {
    if (index + 4 > buffLength) {
      LOG(DEBUG, "index out of range: goal = ", index + 4, ", buff size = ", buffLength);
      return false;
    }

    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = value >> 24;

    return true;
  }

  template <typename T>
  INLINE auto WriteUint32(T& buff, size_t& index, uint32_t value) -> bool
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

  INLINE auto WriteUint64(uint8_t* buff, size_t buffLength, size_t& index, uint64_t value) -> bool
  {
    if (index + 8 > buffLength) {
      LOG(DEBUG, "index out of range: goal = ", index + 8, ", buff size = ", buffLength);
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
  INLINE auto WriteUint64(T& buff, size_t& index, uint64_t value) -> bool
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
  INLINE auto WriteDouble(T& buff, size_t& index, double value) -> bool
  {
    if (index + 8 > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + 8, ", buff size = ", buff.size());
      return false;
    }
    return encoding::WriteBytes(buff.data(), buff.size(), index, reinterpret_cast<uint8_t*>(&value), sizeof(double));
  }

  INLINE auto WriteBytes(const uint8_t* buff, size_t buffLength, size_t& index, const uint8_t* const data, size_t dataLength)
   -> bool
  {
    if (index + dataLength > buffLength) {
      LOG(DEBUG, "index out of range: goal = ", index + buffLength, ", buff size = ", buffLength);
      return false;
    }

    std::copy(data, data + dataLength, buff + index);
    index += dataLength;

    return true;
  }

  template <typename T, typename U>
  INLINE auto WriteBytes(T& buff, size_t& index, const U& data, size_t len) -> bool
  {
    if (index + len > buff.size()) {
      LOG(DEBUG, "index out of range: goal = ", index + len, ", buff size = ", buff.size());
      return false;
    }

    std::copy(data.begin(), data.begin() + len, buff.begin() + index);
    index += len;

    return true;
  }

  INLINE auto WriteAddress(uint8_t* buff, size_t buffLength, size_t& index, const net::Address& addr)
   -> bool
  {
#ifndef NDEBUG
    auto start = index;
#endif

    if (index + net::Address::ByteSize > buffLength) {
      LOG(DEBUG, "buffer too small for address");
      LOG(DEBUG, "index end = ", index + net::Address::ByteSize, ", buffer size = ", buffLength);
      return false;
    }

    if (addr.Type == net::AddressType::IPv4) {
      // write the type
      if (!WriteUint8(buff, buffLength, index, static_cast<uint8_t>(net::AddressType::IPv4))) {
        LOG(DEBUG, "buffer too small for address type");
        LOG(DEBUG, "index end = ", index + 1, ", buffer size = ", buffLength);
        return false;
      }

      std::copy(addr.IPv4.begin(), addr.IPv4.end(), buff + index);  // copy the address
      index += addr.IPv4.size() * sizeof(uint8_t);                  // increment the index

      // write the port
      if (!WriteUint16(buff, buffLength, index, addr.Port)) {
        LOG(DEBUG, "buffer too small for address port");
        LOG(DEBUG, "index end = ", index + 2, ", buffer size = ", buffLength);
        return false;
      }

      index += 12;  // increment the index past the address section
    } else if (addr.Type == net::AddressType::IPv6) {
      // write the type
      if (!WriteUint8(buff, buffLength, index, static_cast<uint8_t>(net::AddressType::IPv6))) {
        LOG(DEBUG, "buffer too small for address type");
        LOG(DEBUG, "index end = ", index + 1, ", buffer size = ", buffLength);
        return false;
      }

      for (const auto& ip : addr.IPv6) {
        if (!WriteUint16(buff, buffLength, index, ip)) {
          LOG(DEBUG, "buffer too small for address part");
          LOG(DEBUG, "index end = ", index + 2, ", buffer size = ", buffLength);
          return false;
        }
      }

      if (!WriteUint16(buff, buffLength, index, addr.Port)) {
        LOG(DEBUG, "buffer too small for address port");
        LOG(DEBUG, "index end = ", index + 2, ", buffer size = ", buffLength);
        return false;
      }
    } else {
      std::fill(buff + index, buff + index + net::Address::ByteSize, 0);
      index += net::Address::ByteSize;
    }

    assert(index - start == net::Address::ByteSize);

    return true;
  }

  template <typename T>
  INLINE auto WriteAddress(T& buff, size_t& index, const net::Address& addr) -> bool
  {
    GCC_NO_OPT_OUT;
#ifndef NDEBUG
    auto start = index;
#endif

    if (index + net::Address::ByteSize > buff.size()) {
      LOG("buffer too small for address");
      LOG("index end = ", index + net::Address::ByteSize, ", buffer size = ", buff.size());
      return false;
    }

    if (addr.Type == net::AddressType::IPv4) {
      // write the type
      if (!WriteUint8(buff, index, static_cast<uint8_t>(net::AddressType::IPv4))) {
        LOG("buffer too small for address type");
        LOG("index end = ", index + 1, ", buffer size = ", buff.size());
        return false;
      }

      std::copy(addr.IPv4.begin(), addr.IPv4.end(), buff.begin() + index);  // copy the address
      index += addr.IPv4.size() * sizeof(uint8_t);                          // increment the index

      // write the port
      if (!WriteUint16(buff, index, addr.Port)) {
        LOG("buffer too small for address port");
        LOG("index end = ", index + 2, ", buffer size = ", buff.size());
        return false;
      }

      index += 12;  // increment the index past the address section
    } else if (addr.Type == net::AddressType::IPv6) {
      // write the type
      if (!WriteUint8(buff, index, static_cast<uint8_t>(net::AddressType::IPv6))) {
        LOG("buffer too small for address type");
        LOG("index end = ", index + 1, ", buffer size = ", buff.size());
        return false;
      }

      for (const auto& ip : addr.IPv6) {
        if (!WriteUint16(buff, index, ip)) {
          LOG("buffer too small for address part");
          LOG("index end = ", index + 2, ", buffer size = ", buff.size());
          return false;
        }
      }

      if (!WriteUint16(buff, index, addr.Port)) {
        LOG("buffer too small for address port");
        LOG("index end = ", index + 2, ", buffer size = ", buff.size());
        return false;
      }
    } else {
      std::fill(buff.begin() + index, buff.begin() + index + net::Address::ByteSize, 0);
      index += net::Address::ByteSize;
    }

    assert(index - start == net::Address::ByteSize);

    return true;
  }

  template <typename T>
  INLINE auto WriteString(T& buff, size_t& index, const std::string& str) -> bool
  {
    if (index + 4 + str.length() > buff.size()) {
      LOG("buffer too small for string");
      return false;
    }

    // sanity check
    if (!encoding::WriteUint32(buff, index, str.length())) {
      LOG("could not write string length");
      return false;
    }

    for (const auto c : str) {
      buff[index++] = c;
    }

    return true;
  }
}  // namespace encoding
