#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "encoding/write.hpp"

using net::Address;

TEST(WriteAddress_ipv4)
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

TEST(WriteAddress_ipv6)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;

  bin.fill(0);
  addr.parse("[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:51034");
  size_t index = 0;
  encoding::write_address(bin, index, addr);
  CHECK(index == Address::SIZE_OF);
  CHECK(bin[0] == net::AddressType::IPv6);
  CHECK(bin[1] == 0x1F);
  CHECK(bin[2] == 0x3B);
  CHECK(bin[3] == 0x33);
  CHECK(bin[4] == 0x3C);
  CHECK(bin[5] == 0x28);
  CHECK(bin[6] == 0x99);
  CHECK(bin[7] == 0xFF);
  CHECK(bin[8] == 0xFF);
  CHECK(bin[9] == 0xFF);
  CHECK(bin[10] == 0xFF);
  CHECK(bin[11] == 0xFF);
  CHECK(bin[12] == 0xFF);
  CHECK(bin[13] == 0xFF);
  CHECK(bin[14] == 0xFF);
  CHECK(bin[15] == 0xFF);
  CHECK(bin[16] == 0xFF);
  CHECK(bin[17] == 0x5A);
  CHECK(bin[18] == 0xC7);
}

TEST(WriteAddress_none)
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

TEST(WriteBytes_array)
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
