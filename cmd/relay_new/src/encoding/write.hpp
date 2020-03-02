#ifndef ENCODING_WRITE_HPP
#define ENCODING_WRITE_HPP

#include "net/net.hpp"

#include "net/address.hpp"

namespace encoding
{
  // TODO move these to legacy namespace
  void write_uint8(uint8_t** p, uint8_t value);

  void write_uint16(uint8_t** p, uint16_t value);

  void write_uint32(uint8_t** p, uint32_t value);

  void write_uint64(uint8_t** p, uint64_t value);

  void write_float32(uint8_t** p, float value);

  void write_float64(uint8_t** p, double value);

  void write_bytes(uint8_t** p, const uint8_t* byte_array, int num_bytes);

  void write_string(uint8_t** p, const char* string_data, uint32_t max_length);

  void write_address(uint8_t** buffer, const legacy::relay_address_t* address);

  template <size_t StorageBufferSize>
  void WriteUint8(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint8_t value);

  template <size_t StorageBufferSize>
  void WriteUint16(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint16_t value);

  template <size_t StorageBufferSize>
  void WriteUint32(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint32_t value);

  template <size_t StorageBufferSize>
  void WriteUint64(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint64_t value);

  template <size_t StorageBufferSize, size_t DataBufferSize>
  void WriteBytes(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, const std::array<uint8_t, DataBufferSize>& data, size_t len);

  template <size_t StorageBufferSize>
  void WriteAddress(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, const net::Address& addr);

  template <size_t StorageBufferSize>
  [[gnu::always_inline]] inline void WriteUint8(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint8_t value)
  {
    buff[index++] = value;
  }

  template <size_t StorageBufferSize>
  [[gnu::always_inline]] inline void WriteUint16(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint16_t value)
  {
    buff[index++] = value & 0xFF;
    buff[index++] = value >> 8;
  }

  template <size_t StorageBufferSize>
  [[gnu::always_inline]] inline void WriteUint32(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint32_t value)
  {
    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = value >> 24;
  }

  // TODO consider #pragma GCC unroll n, cleaner code same perf

  template <size_t StorageBufferSize>
  [[gnu::always_inline]] inline void WriteUint64(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, uint64_t value)
  {
    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = (value >> 24) & 0xFF;
    buff[index++] = (value >> 32) & 0xFF;
    buff[index++] = (value >> 40) & 0xFF;
    buff[index++] = (value >> 48) & 0xFF;
    buff[index++] = value >> 56;
  }

  template <size_t StorageBufferSize, size_t DataBufferSize>
  [[gnu::always_inline]] inline void WriteBytes(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, const std::array<uint8_t, DataBufferSize>& data, size_t len)
  {
    assert(index + len < buff.size());
    std::copy(data.begin(), data.begin() + len, buff.begin() + index);
    index += len;
  }

  template <size_t StorageBufferSize>
  [[gnu::always_inline]] inline void WriteAddress(std::array<uint8_t, StorageBufferSize>& buff, size_t& index, const net::Address& addr)
  {
    GCC_NO_OPT_OUT;
#ifndef NDEBUG
    auto start = index;
#endif

    if (addr.Type == net::AddressType::IPv4) {
      WriteUint8(buff, index, static_cast<uint8_t>(net::AddressType::IPv4));  // write the type

      std::copy(addr.IPv4.begin(), addr.IPv4.end(), buff.begin() + index);  // copy the address
      index += addr.IPv4.size() * sizeof(uint8_t);                          // increment the index

      WriteUint16(buff, index, addr.Port);  // write the port

      index += 12;  // increment the index past the address section
    } else if (addr.Type == net::AddressType::IPv6) {
      WriteUint8(buff, index, static_cast<uint8_t>(net::AddressType::IPv6));  // write the type

      for (const auto& ip : addr.IPv6) {
        WriteUint16(buff, index, ip);
      }

      WriteUint16(buff, index, addr.Port);
    } else {
      std::fill(buff.begin() + index, buff.begin() + index + net::Address::ByteSize, 0);
      index += net::Address::ByteSize;
    }

    assert(index - start == net::Address::ByteSize);
  }
}  // namespace encoding
#endif
