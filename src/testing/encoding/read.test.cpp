#include "includes.h"
#include "testing/test.hpp"
#include "testing/helpers.hpp"

#include "encoding/read.hpp"

using net::Address;
using net::AddressType;

TEST(encoding_read_uint8)
{
  std::array<uint8_t, 16> buff{};

  for (size_t i = 0; i < buff.size(); i++) {
    buff[i] = 0xFF - 0x11 * i;
  }

  uint8_t val = 0;
  size_t index = 1;
  CHECK(encoding::read_uint8(buff, index, val));
  CHECK(val == 0xEE);
}

TEST(encoding_read_uint16)
{
  std::array<uint8_t, 16> buff{};

  for (size_t i = 0; i < buff.size(); i++) {
    buff[i] = 0xFF - 0x11 * i;
  }

  uint16_t val = 0;
  size_t index = 1;

  CHECK(encoding::read_uint16(buff, index, val));
  CHECK(val == 0xDDEE);
}

TEST(encoding_read_uint32)
{
  std::array<uint8_t, 16> buff{};

  for (size_t i = 0; i < buff.size(); i++) {
    buff[i] = 0xFF - 0x11 * i;
  }

  uint32_t val = 0;
  size_t index = 1;

  CHECK(encoding::read_uint32(buff, index, val));
  CHECK(val == 0xBBCCDDEE);
}

TEST(encoding_read_uint64)
{
  std::array<uint8_t, 16> buff{};

  for (size_t i = 0; i < buff.size(); i++) {
    buff[i] = 0xFF - 0x11 * i;
  }

  uint64_t val = 0;
  size_t index = 1;

  CHECK(encoding::read_uint64(buff, index, val));
  CHECK(val == 0x778899AABBCCDDEE);
}

TEST(encoding_read_double)
{
  union
  {
    std::array<uint8_t, 16> buff;
    double num;
  } values;

  double num = 123.456;
  values.num = num;

  double val;
  size_t index = 0;
  CHECK(encoding::read_double(values.buff, index, val));
  CHECK(val == num);
}

TEST(encoding_read_bytes)
{
  std::array<uint8_t, 16> buff{};
  std::array<uint8_t, 8> actual{}, expected{};

  for (size_t i = 0; i < buff.size(); i++) {
    buff[i] = 0xFF - 0x11 * i;
  }

  for (size_t i = 0; i < expected.size(); i++) {
    expected[i] = 0xCC - 0x11 * i;
  }

  size_t index = 3;
  CHECK(encoding::read_bytes(buff, index, actual, actual.size()));
  CHECK(actual == expected);
}

TEST(encoding_read_address_ipv4)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;
  bin.fill(0);
  bin[0] = static_cast<uint8_t>(AddressType::IPv4);
  bin[1] = 127;
  bin[2] = 0;
  bin[3] = 0;
  bin[4] = 1;
  bin[5] = 0x5A;
  bin[6] = 0xC7;

  size_t index = 0;
  CHECK(encoding::read_address(bin, index, addr));
  CHECK(index == Address::SIZE_OF);
}

TEST(encoding_read_address_ipv6)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;
  bin[0] = static_cast<uint8_t>(AddressType::IPv6);
  bin[1] = 0x1F;
  bin[2] = 0x3B;
  bin[3] = 0x33;
  bin[4] = 0x3C;
  bin[5] = 0x28;
  bin[6] = 0x99;
  bin[7] = 0xFF;
  bin[8] = 0xFF;
  bin[9] = 0xFF;
  bin[10] = 0xFF;
  bin[11] = 0xFF;
  bin[12] = 0xFF;
  bin[13] = 0xFF;
  bin[14] = 0xFF;
  bin[15] = 0xFF;
  bin[16] = 0xFF;
  bin[17] = 0x5A;
  bin[18] = 0xC7;

  size_t index = 0;
  CHECK(encoding::read_address(bin, index, addr));

  CHECK(addr.type == AddressType::IPv6);
  CHECK(addr.ipv6[0] == 0x3b1f);
  CHECK(addr.ipv6[1] == 0x3c33);
  CHECK(addr.ipv6[2] == 0x9928);
  CHECK(addr.ipv6[3] == 0xffff);
  CHECK(addr.ipv6[4] == 0xffff);
  CHECK(addr.ipv6[5] == 0xffff);
  CHECK(addr.ipv6[6] == 0xffff);
  CHECK(addr.ipv6[7] == 0xffff);
  CHECK(addr.port == 51034);

  CHECK(index == Address::SIZE_OF);
  CHECK(addr.to_string() == "[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:51034").on_fail([&] {
    std::cout << addr.to_string() << std::endl;
  });
}

TEST(encoding_read_address_none)
{
  Address before, after;
  std::array<uint8_t, Address::SIZE_OF> bin;
  bin.fill(0);
  bin[0] = static_cast<uint8_t>(AddressType::None);
  bin[1] = 0x1F;
  bin[2] = 0x3B;
  bin[3] = 0x33;
  bin[4] = 0x3C;
  bin[5] = 0x28;
  bin[6] = 0x99;
  bin[7] = 0xFF;
  bin[8] = 0xFF;
  bin[9] = 0xFF;
  bin[10] = 0xFF;
  bin[11] = 0xFF;
  bin[12] = 0xFF;
  bin[13] = 0xFF;
  bin[14] = 0xFF;
  bin[15] = 0xFF;
  bin[16] = 0xFF;
  bin[17] = 0xFF;
  bin[18] = 0xFF;

  size_t index = 0;
  CHECK(encoding::read_address(bin, index, after));
  CHECK(index == Address::SIZE_OF);
  CHECK(after.to_string() == "NONE").on_fail([&] {
    std::cout << "\n'" << after.to_string() << '\'' << std::endl;
  });
  CHECK(before == after);
}

TEST(encoding_read_string)
{
  std::array<char, 32> buff = {0x08, 0x00, 0x00, 0x00, 'a', ' ', 's', 't', 'r', 'i', 'n', 'g'};

  std::string val;
  size_t index = 0;
  CHECK(encoding::read_string(buff, index, val));
  CHECK(val == "a string");
}
