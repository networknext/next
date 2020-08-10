#ifndef ENCODING_READ_HPP
#define ENCODING_READ_HPP

#include "net/address.hpp"

#include "util/logger.hpp"

#include "binary.hpp"

namespace encoding
{
  template <typename T>
  uint8_t ReadUint8(const T& buff, size_t& index);

  template <typename T>
  uint16_t ReadUint16(const T& buff, size_t& index);

  template <typename T>
  uint32_t ReadUint32(const T& buff, size_t& index);

  template <typename T>
  uint64_t ReadUint64(const T& buff, size_t& index);

  void ReadBytes(const uint8_t* buff, size_t buffLength, size_t& index, uint8_t* storage, size_t storageLength, size_t len);

  template <typename T, typename U>
  void ReadBytes(const T& buff, size_t& index, U& storage, size_t len);

  void ReadAddress(const uint8_t* buff, size_t buffLength, size_t& index, net::Address& addr);

  template <typename T>
  void ReadAddress(const T& buff, size_t& index, net::Address& addr);

  template <typename T>
  auto ReadString(const T& buff, size_t& index, std::string& str) -> std::string;

  template <typename T>
  [[gnu::always_inline]] inline uint8_t ReadUint8(const T& buff, size_t& index)
  {
    return buff[index++];
  }

  template <typename T>
  [[gnu::always_inline]] inline uint16_t ReadUint16(const T& buff, size_t& index)
  {
    GCC_NO_OPT_OUT;
    uint16_t retval;
    retval = (buff)[index++];
    retval |= (static_cast<uint64_t>(buff[index++]) << 8);
    return retval;
  }

  template <typename T>
  [[gnu::always_inline]] inline uint32_t ReadUint32(const T& buff, size_t& index)
  {
    uint32_t retval;
    retval = buff[index++];
    retval |= (static_cast<uint64_t>(buff[index++]) << 8);
    retval |= (static_cast<uint64_t>(buff[index++]) << 16);
    retval |= (static_cast<uint64_t>(buff[index++]) << 24);
    return retval;
  }

  template <typename T>
  [[gnu::always_inline]] inline uint64_t ReadUint64(const T& buff, size_t& index)
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

  [[gnu::always_inline]] inline void ReadBytes(
   const uint8_t* buff, size_t buffLength, size_t& index, uint8_t* storage, size_t storageLength, size_t len)
  {
    (void)buffLength;
    (void)storageLength;
    assert(len <= storageLength);
    assert(index + len <= buffLength);
    std::copy(buff + index, buff + index + len, storage);
    index += len;
  }

  template <typename T, typename U>
  [[gnu::always_inline]] inline void ReadBytes(const T& buff, size_t& index, U& storage, size_t len)
  {
    assert(len <= storage.size());
    assert(index + len <= buff.size());
    std::copy(buff.begin() + index, buff.begin() + index + len, storage.begin());
    index += len;
  }

  [[gnu::always_inline]] inline void ReadAddress(const uint8_t* buff, size_t buffLength, size_t& index, net::Address& addr)
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

namespace legacy
{
  uint8_t read_uint8(const uint8_t** p);

  uint16_t read_uint16(const uint8_t** p);

  uint32_t read_uint32(const uint8_t** p);

  uint64_t read_uint64(const uint8_t** p);

  float read_float32(const uint8_t** p);

  double read_float64(const uint8_t** p);

  void read_bytes(const uint8_t** p, uint8_t* byte_array, int num_bytes);

  void read_string(const uint8_t** p, char* string_data, uint32_t max_length);

  void read_address(const uint8_t** buffer, legacy::relay_address_t* address);

}  // namespace legacy
#endif
