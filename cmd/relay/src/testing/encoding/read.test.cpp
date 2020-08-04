#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "encoding/read.hpp"

Test(ReadAddress_ipv4)
{
  net::Address addr;
  std::array<uint8_t, net::Address::ByteSize> bin;
  bin.fill(0);
  bin[0] = static_cast<uint8_t>(net::AddressType::IPv4);
  bin[1] = 127;
  bin[2] = 0;
  bin[3] = 0;
  bin[4] = 1;
  bin[5] = 0x5A;
  bin[6] = 0xC7;

  size_t index = 0;
  encoding::ReadAddress(bin, index, addr);
  check(index == net::Address::ByteSize);
}

Test(ReadAddress_ipv6)
{
  net::Address addr;
  std::array<uint8_t, net::Address::ByteSize> bin;
  bin[0] = static_cast<uint8_t>(net::AddressType::IPv6);
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
  encoding::ReadAddress(bin, index, addr);
  check(index == net::Address::ByteSize);
  check(addr.toString() == "[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:51034");
}

Test(ReadAddress_none)
{
  net::Address before, after;
  std::array<uint8_t, net::Address::ByteSize> bin;
  bin.fill(0);
  bin[0] = static_cast<uint8_t>(net::AddressType::None);
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
  encoding::ReadAddress(bin, index, after);
  check(index == net::Address::ByteSize);
  check(after.toString() == "NONE");
  check(before == after);
}
