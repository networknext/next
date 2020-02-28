#ifndef ENCODING_READ_HPP
#define ENCODING_READ_HPP

#include "net/address.hpp"

#include "util/logger.hpp"

#include "binary.hpp"

namespace encoding
{
  // Prototypes

  uint8_t read_uint8(const uint8_t** p);

  uint16_t read_uint16(const uint8_t** p);

  uint32_t read_uint32(const uint8_t** p);

  uint64_t read_uint64(const uint8_t** p);

  float read_float32(const uint8_t** p);

  double read_float64(const uint8_t** p);

  void read_bytes(const uint8_t** p, uint8_t* byte_array, int num_bytes);

  void read_string(const uint8_t** p, char* string_data, uint32_t max_length);

  void read_address(const uint8_t** buffer, legacy::relay_address_t* address);

  template <size_t BuffSize>
  uint8_t ReadUint8(const std::array<uint8_t, BuffSize>& buff, size_t& index);

  template <size_t BuffSize>
  uint16_t ReadUint16(const std::array<uint8_t, BuffSize>& buff, size_t& index);

  template <size_t BuffSize>
  uint32_t ReadUint32(const std::array<uint8_t, BuffSize>& buff, size_t& index);

  template <size_t BuffSize>
  uint64_t ReadUint64(const std::array<uint8_t, BuffSize>& buff, size_t& index);

  template <size_t BuffSize, size_t StorageBufferSize>
  void ReadBytes(
   const std::array<uint8_t, BuffSize>& buff, size_t& index, std::array<uint8_t, StorageBufferSize>& storage, size_t len);

  template <size_t BuffSize>
  void ReadAddress(const std::array<uint8_t, BuffSize>& buff, size_t& index, net::Address& addr);

  template <size_t BuffSize>
  [[gnu::always_inline]] inline uint8_t ReadUint8(const std::array<uint8_t, BuffSize>& buff, size_t& index)
  {
    return buff[index++];
  }

  template <size_t BuffSize>
  [[gnu::always_inline]] inline uint16_t ReadUint16(const std::array<uint8_t, BuffSize>& buff, size_t& index)
  {
    GCC_NO_OPT_OUT;
    uint16_t retval;
    retval = (buff)[index++];
    retval |= (static_cast<uint64_t>(buff[index++]) << 8);
    return retval;
  }

  template <size_t BuffSize>
  [[gnu::always_inline]] inline uint32_t ReadUint32(const std::array<uint8_t, BuffSize>& buff, size_t& index)
  {
    uint32_t retval;
    retval = buff[index++];
    retval |= (static_cast<uint64_t>(buff[index++]) << 8);
    retval |= (static_cast<uint64_t>(buff[index++]) << 16);
    retval |= (static_cast<uint64_t>(buff[index++]) << 24);
    return retval;
  }

  template <size_t BuffSize>
  [[gnu::always_inline]] inline uint64_t ReadUint64(const std::array<uint8_t, BuffSize>& buff, size_t& index)
  {
    uint64_t retval;
    retval = buff[index++];
    retval |= (static_cast<uint64_t>(buff[index++]) << 8);
    retval |= (static_cast<uint64_t>(buff[index++]) << 16);
    retval |= (static_cast<uint64_t>(buff[index++]) << 24);
    retval |= (static_cast<uint64_t>(buff[index++]) << 32);
    retval |= (static_cast<uint64_t>(buff[index++]) << 40);
    retval |= (static_cast<uint64_t>(buff[index++]) << 48);
    retval |= (static_cast<uint64_t>(buff[index++]) << 56);
    return retval;
  }

  template <size_t BuffSize, size_t StorageBufferSize>
  [[gnu::always_inline]] inline void ReadBytes(
   const std::array<uint8_t, BuffSize>& buff, size_t& index, std::array<uint8_t, StorageBufferSize>& storage, size_t len)
  {
    assert(len <= StorageBufferSize);
    assert(index + len <= BuffSize);
    std::copy(buff.begin() + index, buff.begin() + index + len, storage.begin());
    index += len;
  }

  template <size_t BuffSize>
  [[gnu::always_inline]] inline void ReadAddress(const std::array<uint8_t, BuffSize>& buff, size_t& index, net::Address& addr)
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
}  // namespace encoding
#endif
