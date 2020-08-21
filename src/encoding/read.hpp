#pragma once

#include "binary.hpp"
#include "net/address.hpp"
#include "util/logger.hpp"

namespace encoding
{
  template <typename T>
  auto ReadUint8(const T& buff, size_t& index, uint8_t& value) -> bool;

  template <typename T>
  auto ReadUint16(const T& buff, size_t& index, uint16_t& value) -> bool;

  template <typename T>
  auto ReadUint32(const T& buff, size_t& index, uint32_t& value) -> bool;

  template <typename T>
  auto ReadUint64(const T& buff, size_t& index, uint64_t& value) -> bool;

  template <typename T>
  auto ReadDouble(const T& buff, size_t& index, double& value) -> bool;

  auto ReadBytes(
   const uint8_t* const buff, size_t buffLength, size_t& index, const uint8_t* storage, size_t storageLength, size_t len)
   -> bool;

  template <typename T, typename U>
  auto ReadBytes(const T& buff, size_t& index, U& storage, size_t len) -> bool;

  auto ReadAddress(const uint8_t* buff, size_t buffLength, size_t& index, net::Address& addr) -> bool;

  template <typename T>
  auto ReadAddress(const T& buff, size_t& index, net::Address& addr) -> bool;

  template <typename T>
  auto ReadString(const T& buff, size_t& index, std::string& str) -> std::string;

  template <typename T>
  INLINE auto ReadUint8(const T& buff, size_t& index, uint8_t& value) -> bool
  {
    if (index + 1 > buff.size()) {
      return false;
    }
    value = buff[index++];
    return true;
  }

  template <typename T>
  INLINE auto ReadUint16(const T& buff, size_t& index, uint16_t& value) -> bool
  {
    if (index + 2 > buff.size()) {
      return false;
    }
    value = (buff)[index++];
    value |= (static_cast<uint64_t>(buff[index++]) << 8);
    return true;
  }

  template <typename T>
  INLINE auto ReadUint32(const T& buff, size_t& index, uint32_t& value) -> bool
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

  template <typename T>
  INLINE auto ReadUint64(const T& buff, size_t& index, uint64_t& value) -> bool
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
  INLINE auto ReadDouble(const T& buff, size_t& index, double& value) -> bool
  {
    return encoding::ReadBytes(
     buff.data(), buff.size(), index, reinterpret_cast<uint8_t*>(&value), sizeof(double), sizeof(double));
  }

  INLINE auto ReadBytes(
   const uint8_t* const buff, size_t buff_length, size_t& index, uint8_t* storage, size_t storage_length, size_t read_len) -> bool
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
  INLINE auto ReadBytes(const T& buff, size_t& index, U& storage, size_t read_len) -> bool
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

  INLINE auto ReadAddress(const uint8_t* buff, size_t buffLength, size_t& index, net::Address& addr) -> bool
  {
    (void)buffLength;
#ifndef NDEBUG
    auto start = index;
#endif
    assert(buffLength >= net::Address::ByteSize);
    addr.Type = static_cast<net::AddressType>(ReadUint8(buff, index));  // read the type

    if (addr.Type == net::AddressType::IPv4) {
      std::copy(buff + index, buff + index + 4, addr.IPv4.begin());  // copy the address
      index += 4;                                                    // increment the index
      addr.Port = ReadUint16(buff, index);                           // read the port
      index += 12;                                                   // increment the index past the reserved area
    } else if (addr.Type == net::AddressType::IPv6) {
      for (int i = 0; i < 8; i++) {
        addr.IPv6[i] = ReadUint16(buff, index);
      }
      addr.Port = ReadUint16(buff, index);  // read the port
    } else {
      addr.reset();
      index += net::Address::ByteSize - 1;  // if no type, increment the index past the address area
    }

    assert(index - start == net::Address::ByteSize);
  }

  template <typename T>
  [[gnu::always_inline]] inline void ReadAddress(const T& buff, size_t& index, net::Address& addr)
  {
    GCC_NO_OPT_OUT;
#ifndef NDEBUG
    auto start = index;
#endif
    addr.Type = static_cast<net::AddressType>(ReadUint8(buff, index));  // read the type

    if (addr.Type == net::AddressType::IPv4) {
      std::copy(buff.begin() + index, buff.begin() + index + 4, addr.IPv4.begin());  // copy the address
      index += 4;                                                                    // increment the index
      addr.Port = ReadUint16(buff, index);                                           // read the port
      index += 12;  // increment the index past the reserved area
    } else if (addr.Type == net::AddressType::IPv6) {
      for (int i = 0; i < 8; i++) {
        addr.IPv6[i] = ReadUint16(buff, index);
      }
      addr.Port = ReadUint16(buff, index);  // read the port
    } else {
      addr.reset();
      index += net::Address::ByteSize - 1;  // if no type, increment the index past the address area
    }

    assert(index - start == net::Address::ByteSize);
  }

  template <typename T>
  auto ReadString(const T& buff, size_t& index) -> std::string
  {
    size_t len = ReadUint32(buff, index);
    std::string str(buff.begin() + index, buff.begin() + index + len);
    index += len;
    return str;
  }
}  // namespace encoding
