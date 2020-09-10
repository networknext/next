#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "encoding/write.hpp"

using net::Address;

Test(WriteAddress_ipv4)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;

  bin.fill(0);
  addr.parse("127.0.0.1:51034");
  size_t index = 0;
  encoding::write_address(bin, index, addr);
  check(index == Address::SIZE_OF);
  check(bin[0] == net::AddressType::IPv4);
  check(bin[1] == 127);
  check(bin[2] == 0);
  check(bin[3] == 0);
  check(bin[4] == 1);
  check(bin[5] == 0x5A);
  check(bin[6] == 0xC7);
}

Test(WriteAddress_ipv6)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;

  bin.fill(0);
  addr.parse("[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:51034");
  size_t index = 0;
  encoding::write_address(bin, index, addr);
  check(index == Address::SIZE_OF);
  check(bin[0] == net::AddressType::IPv6);
  check(bin[1] == 0x1F);
  check(bin[2] == 0x3B);
  check(bin[3] == 0x33);
  check(bin[4] == 0x3C);
  check(bin[5] == 0x28);
  check(bin[6] == 0x99);
  check(bin[7] == 0xFF);
  check(bin[8] == 0xFF);
  check(bin[9] == 0xFF);
  check(bin[10] == 0xFF);
  check(bin[11] == 0xFF);
  check(bin[12] == 0xFF);
  check(bin[13] == 0xFF);
  check(bin[14] == 0xFF);
  check(bin[15] == 0xFF);
  check(bin[16] == 0xFF);
  check(bin[17] == 0x5A);
  check(bin[18] == 0xC7);
}

Test(WriteAddress_none)
{
  Address addr;
  std::array<uint8_t, Address::SIZE_OF> bin;

  bin.fill(0);
  addr.parse("1udai898haidfihe");
  size_t index = 0;
  encoding::write_address(bin, index, addr);
  for (auto& i : bin) {
    check(i == 0);
  }
}

Test(WriteBytes_array)
{
  std::array<uint8_t, 32> buff;
  std::array<uint8_t, 8> data;
  size_t index = 8;

  for (uint8_t i = 0; i < data.size(); i++) {
    data[i] = i;
  }

  encoding::write_bytes(buff, index, data, 4);

  check(index == 12);

  for (size_t i = 8, j = 0; i < 4; i++) {
    check(buff[i] == data[j]);
  }
}
