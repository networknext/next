#include "includes.h"
#include "write.hpp"

namespace legacy
{
  void write_uint8(uint8_t** p, uint8_t value)
  {
    **p = value;
    ++(*p);
  }

  void write_uint16(uint8_t** p, uint16_t value)
  {
    (*p)[0] = value & 0xFF;
    (*p)[1] = value >> 8;
    *p += 2;
  }

  void write_uint32(uint8_t** p, uint32_t value)
  {
    (*p)[0] = value & 0xFF;
    (*p)[1] = (value >> 8) & 0xFF;
    (*p)[2] = (value >> 16) & 0xFF;
    (*p)[3] = value >> 24;
    *p += 4;
  }

  void write_uint64(uint8_t** p, uint64_t value)
  {
    (*p)[0] = value & 0xFF;
    (*p)[1] = (value >> 8) & 0xFF;
    (*p)[2] = (value >> 16) & 0xFF;
    (*p)[3] = (value >> 24) & 0xFF;
    (*p)[4] = (value >> 32) & 0xFF;
    (*p)[5] = (value >> 40) & 0xFF;
    (*p)[6] = (value >> 48) & 0xFF;
    (*p)[7] = value >> 56;
    *p += 8;
  }

  void write_float32(uint8_t** p, float value)
  {
    uint32_t value_int = 0;
    char* p_value = (char*)(&value);
    char* p_value_int = (char*)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint32_t));
    write_uint32(p, value_int);
  }

  void write_float64(uint8_t** p, double value)
  {
    uint64_t value_int = 0;
    char* p_value = (char*)(&value);
    char* p_value_int = (char*)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint64_t));
    write_uint64(p, value_int);
  }

  void write_bytes(uint8_t** p, const uint8_t* byte_array, int num_bytes)
  {
    for (int i = 0; i < num_bytes; ++i) {
      write_uint8(p, byte_array[i]);
    }
  }

  void write_string(uint8_t** p, const char* string_data, uint32_t max_length)
  {
    uint32_t length = strlen(string_data);
    assert(length <= max_length);
    if (length > max_length)
      length = max_length;
    write_uint32(p, length);
    for (uint32_t i = 0; i < length; ++i) {
      write_uint8(p, string_data[i]);
    }
  }

  void write_address(uint8_t** buffer, const legacy::relay_address_t* address)
  {
    assert(buffer);
    assert(*buffer);
    assert(address);

#ifndef NDEBUG
    uint8_t* start = *buffer;
#endif

    if (address->type == net::AddressType::IPv4) {
      write_uint8(buffer, static_cast<uint8_t>(net::AddressType::IPv4));
      for (int i = 0; i < 4; ++i) {
        write_uint8(buffer, address->data.ipv4[i]);
      }
      write_uint16(buffer, address->port);
      for (int i = 0; i < 12; ++i) {
        write_uint8(buffer, 0);
      }
    } else if (address->type == net::AddressType::IPv6) {
      write_uint8(buffer, static_cast<uint8_t>(net::AddressType::IPv6));
      for (int i = 0; i < 8; ++i) {
        write_uint16(buffer, address->data.ipv6[i]);
      }
      write_uint16(buffer, address->port);
    } else {
      for (size_t i = 0; i < net::Address::ByteSize; ++i) {
        write_uint8(buffer, 0);
      }
    }

    assert(*buffer - start == net::Address::ByteSize);
  }
}  // namespace legacy
