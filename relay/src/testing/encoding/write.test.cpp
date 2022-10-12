#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "encoding/write.hpp"

using net::Address;

TEST(encoding_write_uint8)
{
  std::array<uint8_t, 16> buff{};

  uint8_t val = 0xFF;
  size_t index = 1;

  CHECK(encoding::write_uint8(buff, index, val));
  CHECK(buff[1] == 0xFF);
}

TEST(encoding_write_uint16)
{
  std::array<uint8_t, 16> buff{};

  uint16_t val = 0xEEFF;
  size_t index = 1;

  CHECK(encoding::write_uint16(buff, index, val));
  CHECK(buff[1] == 0xFF);
  CHECK(buff[2] == 0xEE);
}

TEST(encoding_write_uint32)
{
  std::array<uint8_t, 16> buff{};

  uint32_t val = 0xCCDDEEFF;
  size_t index = 1;

  CHECK(encoding::write_uint32(buff, index, val));
  CHECK(buff[1] == 0xFF);
  CHECK(buff[2] == 0xEE);
  CHECK(buff[3] == 0xDD);
  CHECK(buff[4] == 0xCC);
}

TEST(encoding_write_uint64)
{
  std::array<uint8_t, 16> buff{};

  uint64_t val = 0x8899AABBCCDDEEFF;
  size_t index = 1;

  CHECK(encoding::write_uint64(buff, index, val));
  CHECK(buff[1] == 0xFF);
  CHECK(buff[2] == 0xEE);
  CHECK(buff[3] == 0xDD);
  CHECK(buff[4] == 0xCC);
  CHECK(buff[5] == 0xBB);
  CHECK(buff[6] == 0xAA);
  CHECK(buff[7] == 0x99);
  CHECK(buff[8] == 0x88);
}

TEST(encoding_write_double)
{
  union
  {
    std::array<uint8_t, 16> buff;
    double num;
  } values;

  double num = 123.456;
  size_t index = 0;

  CHECK(encoding::write_double(values.buff, index, num));
  CHECK(values.num == num);
}

TEST(encoding_write_bytes)
{
  std::array<uint8_t, 32> buff;
  std::array<uint8_t, 8> data;
  size_t index = 8;

  for (uint8_t i = 0; i < data.size(); i++) {
    data[i] = i;
  }

  encoding::write_bytes(buff, index, data, 4);

  CHECK(index == 12);

  for (size_t i = 8, j = 0; i < 4; i++) {
    CHECK(buff[i] == data[j]);
  }
}

TEST(encoding_write_address_ipv4)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;

  bin.fill(0);
  addr.parse("127.0.0.1:51034");
  size_t index = 0;
  encoding::write_address(bin, index, addr);
  CHECK(index == Address::SIZE_OF);
  CHECK(bin[0] == net::AddressType::IPv4);
  CHECK(bin[1] == 127);
  CHECK(bin[2] == 0);
  CHECK(bin[3] == 0);
  CHECK(bin[4] == 1);
  CHECK(bin[5] == 0x5A);
  CHECK(bin[6] == 0xC7);
}

TEST(encoding_write_address_none)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;

  bin.fill(0);
  addr.parse("1udai898haidfihe");
  size_t index = 0;
  encoding::write_address(bin, index, addr);
  for (auto& i : bin) {
    CHECK(i == 0);
  }
}

TEST(encoding_write_string)
{
  std::array<char, 32> buff{};  // = {0x08, 0x00, 0x00, 0x00, 'a', ' ', 's', 't', 'r', 'i', 'n', 'g'};

  std::string val = "a string";
  size_t index = 0;
  CHECK(encoding::write_string(buff, index, val));
  CHECK(buff[0] == 0x08);
  CHECK(buff[1] == 0x00);
  CHECK(buff[2] == 0x00);
  CHECK(buff[3] == 0x00);
  CHECK(buff[4] == 'a');
  CHECK(buff[5] == ' ');
  CHECK(buff[6] == 's');
  CHECK(buff[7] == 't');
  CHECK(buff[8] == 'r');
  CHECK(buff[9] == 'i');
  CHECK(buff[10] == 'n');
  CHECK(buff[11] == 'g');
}
