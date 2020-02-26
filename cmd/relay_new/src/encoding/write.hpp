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

  /* StorageBuffer & DataBuffer can be anything that has operator[], begin(), and end(), so array, vector, etc...
   *
   * technically even maps work here though it's not recomended because what would be the point
   */

  template <class StorageBuffer>
  void WriteUint8(StorageBuffer& buff, size_t& index, uint8_t value);

  template <class StorageBuffer>
  void WriteUint16(StorageBuffer& buff, size_t& index, uint16_t value);

  template <class StorageBuffer>
  void WriteUint32(StorageBuffer& buff, size_t& index, uint32_t value);

  template <class StorageBuffer>
  void WriteUint64(StorageBuffer& buff, size_t& index, uint64_t value);

  template <class StorageBuffer, class DataBuffer>
  void WriteBytes(StorageBuffer& buff, size_t& index, const DataBuffer& data, size_t len);

  template <class StorageBuffer>
  void WriteAddress(StorageBuffer& buff, size_t& index, const net::Address& addr);

  template <class StorageBuffer>
  [[gnu::always_inline]] inline void WriteUint8(StorageBuffer& buff, size_t& index, uint8_t value)
  {
    buff[index++] = value;
  }

  template <class StorageBuffer>
  [[gnu::always_inline]] inline void WriteUint16(StorageBuffer& buff, size_t& index, uint16_t value)
  {
    buff[index++] = value & 0xFF;
    buff[index++] = value >> 8;
  }

  template <class StorageBuffer>
  void WriteUint32(StorageBuffer& buff, size_t& index, uint32_t value)
  {
    buff[index++] = value & 0xFF;
    buff[index++] = (value >> 8) & 0xFF;
    buff[index++] = (value >> 16) & 0xFF;
    buff[index++] = value >> 24;
  }

  // TODO consider #pragma GCC unroll n, cleaner code same perf

  template <class StorageBuffer>
  [[gnu::always_inline]] inline void WriteUint64(StorageBuffer& buff, size_t& index, uint64_t value)
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

  template <class StorageBuffer, class DataBuffer>
  [[gnu::always_inline]] inline void WriteBytes(StorageBuffer& buff, size_t& index, const DataBuffer& data, size_t len)
  {
    assert(index + len < buff.size());
    std::copy(data.begin(), data.begin() + len, buff.begin() + index);
    index += len;
  }

  template <class StorageBuffer>
  [[gnu::always_inline]] inline void WriteAddress(StorageBuffer& buff, size_t& index, const net::Address& addr)
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

      /* hack to write the data faster, only use if we're getting desperate for performance */
      // std::copy(addr.IPv6.begin(), addr.IPv6.end(), reinterpret_cast<uint16_t*>(buff.data() + index));
      // index += addr.IPv6.size() * sizeof(uint16_t);

      WriteUint16(buff, index, addr.Port);
    } else {
      std::fill(buff.begin() + index, buff.begin() + index + net::Address::ByteSize, 0);
      index += net::Address::ByteSize;
    }

    assert(index - start == net::Address::ByteSize);
  }
}  // namespace encoding
#endif
