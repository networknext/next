#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"
#include "encoding/read.hpp"

using net::Address;
using net::AddressType;

Test(ReadAddress_ipv4)
{
  Address addr;
  std::array<uint8_t, Address::ByteSize> bin;
  bin.fill(0);
  bin[0] = static_cast<uint8_t>(AddressType::IPv4);
  bin[1] = 127;
  bin[2] = 0;
  bin[3] = 0;
  bin[4] = 1;
  bin[5] = 0x5A;
  bin[6] = 0xC7;

  size_t index = 0;
  check(encoding::read_address(bin, index, addr));
  check(index == Address::ByteSize);
}

Test(ReadAddress_ipv6)
{
  Address addr;
  std::array<uint8_t, Address::ByteSize> bin;
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
  check(encoding::read_address(bin, index, addr));

  check(addr.Type == AddressType::IPv6);
  check(addr.IPv6[0] == 0x3b1f);
  check(addr.IPv6[1] == 0x3c33);
  check(addr.IPv6[2] == 0x9928);
  check(addr.IPv6[3] == 0xffff);
  check(addr.IPv6[4] == 0xffff);
  check(addr.IPv6[5] == 0xffff);
  check(addr.IPv6[6] == 0xffff);
  check(addr.IPv6[7] == 0xffff);
  check(addr.Port == 51034);

  check(index == Address::ByteSize);
  check(addr.to_string() == "[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:51034").onFail([&] {
    std::cout << addr.to_string() << std::endl;
  });
}

Test(ReadAddress_none)
{
  Address before, after;
  std::array<uint8_t, Address::ByteSize> bin;
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
  check(encoding::read_address(bin, index, after));
  check(index == Address::ByteSize);
  check(after.to_string() == "NONE").onFail([&] {
    std::cout << "\n'" << after.to_string() << '\'' << std::endl;
  });
  check(before == after);
}
