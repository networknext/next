#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

Test(Address_operator_eq_and_neq)
{
  net::Address a, b;

  check(a == b);

  a.Type = net::AddressType::IPv4;
  check(a != b);
  a.Type = net::AddressType::None;

  a.Port = 1;
  check(a != b);
  a.Port = 0;

  a.Type = net::AddressType::IPv4;

  a.IPv4[0] = 1;
  check(a != b);
  a.IPv4[0] = 0;

  *a.IPv4.end() = 1;
  check(a != b);
  *a.IPv4.end() = 0;

  a.Type = net::AddressType::IPv6;

  a.IPv6[0] = 1;
  check(a != b);
  a.IPv6[0] = 0;

  *a.IPv6.end() = 1;
  check(a != b);
  *a.IPv6.end() = 0;

  a.Type = net::AddressType::None;

  check(a == b);
}

Test(Address_reset)
{
  net::Address addr, base;
  addr.Type = static_cast<net::AddressType>(77);
  addr.Port = 123;
  addr.IPv4[0] = 18;
  addr.IPv4[1] = 82;
  addr.IPv4[2] = 53;
  addr.IPv4[3] = 67;

  addr.reset();
  check(addr == base);
}

Test(Address_parse_ipv4)
{
  net::Address addr;
  check(addr.parse("127.0.0.1:51034") == true);
  check(addr.Type == net::AddressType::IPv4);
  check(addr.Port == 51034);
  check(addr.IPv4[0] == 127);
  check(addr.IPv4[1] == 0);
  check(addr.IPv4[2] == 0);
  check(addr.IPv4[3] == 1);
}

Test(Address_parse_ipv6_with_braces)
{
  net::Address addr;
  check(addr.parse("[::1]:51034") == true);
  check(addr.Type == net::AddressType::IPv6);
  check(addr.Port == 51034);
  check(addr.IPv6[0] == 0);
  check(addr.IPv6[1] == 0);
  check(addr.IPv6[2] == 0);
  check(addr.IPv6[3] == 0);
  check(addr.IPv6[4] == 0);
  check(addr.IPv6[5] == 0);
  check(addr.IPv6[6] == 0);
  check(addr.IPv6[7] == 1);

  addr.reset();
  check(addr.parse("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000") == true);
  check(addr.Type == net::AddressType::IPv6);
  check(addr.Port == 20000);
  check(addr.IPv6[0] == 0x2001);
  check(addr.IPv6[1] == 0x0db8);
  check(addr.IPv6[2] == 0x85a3);
  check(addr.IPv6[3] == 0x0000);
  check(addr.IPv6[4] == 0x0000);
  check(addr.IPv6[5] == 0x8a2e);
  check(addr.IPv6[6] == 0x0370);
  check(addr.IPv6[7] == 0x7334);
}

Test(Address_parse_ipv6_without_braces)
{
  net::Address addr;
  check(addr.parse("2001:0db8:85a3:0000:0000:8a2e:0370:7334") == true);
  check(addr.Type == net::AddressType::IPv6);
  check(addr.Port == 0);
  check(addr.IPv6[0] == 0x2001);
  check(addr.IPv6[1] == 0x0db8);
  check(addr.IPv6[2] == 0x85a3);
  check(addr.IPv6[3] == 0x0000);
  check(addr.IPv6[4] == 0x0000);
  check(addr.IPv6[5] == 0x8a2e);
  check(addr.IPv6[6] == 0x0370);
  check(addr.IPv6[7] == 0x7334);
}

Test(Address_parse_invalid_ips)
{
  net::Address addr;
  check(addr.parse("127.0.:182a") == false);
  check(addr.Type == net::AddressType::None);
  check(addr.Port == 0);

  check(addr.IPv4[0] == 0);
  check(addr.IPv4[1] == 0);
  check(addr.IPv4[2] == 0);
  check(addr.IPv4[3] == 0);

  check(addr.IPv6[0] == 0);
  check(addr.IPv6[1] == 0);
  check(addr.IPv6[2] == 0);
  check(addr.IPv6[3] == 0);
  check(addr.IPv6[4] == 0);
  check(addr.IPv6[5] == 0);
  check(addr.IPv6[6] == 0);
  check(addr.IPv6[7] == 0);
}

Test(Address_toString)
{
  net::Address addr;
  std::string base, output, expected;

  base = "127.0.0.1:51034";
  expected = base;
  check(addr.parse(base) == true);
  addr.toString(output);
  check(output == expected);

  base = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";
  expected = "[2001:db8:85a3::8a2e:370:7334]:20000";
  check(addr.parse(base) == true);
  addr.toString(output);
  check(output == expected);

  base = "2001:0db8:85a3:0000:0000:8a2e:0370:7334";
  expected = "2001:db8:85a3::8a2e:370:7334";
  check(addr.parse(base) == true);
  addr.toString(output);
  check(output == expected);

  base = "127.0djaid?.0.sad1:5as1034";
  expected = "NONE";
  check(addr.parse(base) == false);
  addr.toString(output);
  check(output == expected);
}

Test(Address_parse_additional)
{
  net::Address addr;

  // all should fail
  {
    check(addr.parse("") == false);
    check(addr.parse("[") == false);
    check(addr.parse("[]") == false);
    check(addr.parse("[]:") == false);
    check(addr.parse(":") == false);
    check(addr.parse("1") == false);
    check(addr.parse("12") == false);
    check(addr.parse("123") == false);
    check(addr.parse("1234") == false);
    check(addr.parse("1234.0.12313.0000") == false);
    check(addr.parse("1234.0.12313.0000.0.0.0.0.0") == false);
    check(addr.parse("1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131") == false);
    check(addr.parse(".") == false);
    check(addr.parse("..") == false);
    check(addr.parse("...") == false);
    check(addr.parse("....") == false);
    check(addr.parse(".....") == false);
    check(addr.parse("0.0.0.0") == true);
    check(addr.parse("107.77.207.77") == true);
  }

  // reset should pass

  {
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 0);
    check(addr.IPv4[0] == 107);
    check(addr.IPv4[1] == 77);
    check(addr.IPv4[2] == 207);
    check(addr.IPv4[3] == 77);
  }

  {
    check(addr.parse("127.0.0.1") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 0);
    check(addr.IPv4[0] == 127);
    check(addr.IPv4[1] == 0);
    check(addr.IPv4[2] == 0);
    check(addr.IPv4[3] == 1);
  }

  {
    check(addr.parse("107.77.207.77:40000") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 40000);
    check(addr.IPv4[0] == 107);
    check(addr.IPv4[1] == 77);
    check(addr.IPv4[2] == 207);
    check(addr.IPv4[3] == 77);
  }

  {
    check(addr.parse("127.0.0.1:40000") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 40000);
    check(addr.IPv4[0] == 127);
    check(addr.IPv4[1] == 0);
    check(addr.IPv4[2] == 0);
    check(addr.IPv4[3] == 1);
  }

  {
    check(addr.parse("fe80::202:b3ff:fe1e:8329") == true);
    check(addr.Type == net::AddressType::IPv6);
    check(addr.Port == 0);
    check(addr.IPv6[0] == 0xfe80);
    check(addr.IPv6[1] == 0x0000);
    check(addr.IPv6[2] == 0x0000);
    check(addr.IPv6[3] == 0x0000);
    check(addr.IPv6[4] == 0x0202);
    check(addr.IPv6[5] == 0xb3ff);
    check(addr.IPv6[6] == 0xfe1e);
    check(addr.IPv6[7] == 0x8329);
  }

  {
    check(addr.parse("::") == true);
    check(addr.Type == net::AddressType::IPv6);
    check(addr.Port == 0);
    check(addr.IPv6[0] == 0x0000);
    check(addr.IPv6[1] == 0x0000);
    check(addr.IPv6[2] == 0x0000);
    check(addr.IPv6[3] == 0x0000);
    check(addr.IPv6[4] == 0x0000);
    check(addr.IPv6[5] == 0x0000);
    check(addr.IPv6[6] == 0x0000);
    check(addr.IPv6[7] == 0x0000);
  }

  {
    check(addr.parse("::1") == true);
    check(addr.Type == net::AddressType::IPv6);
    check(addr.Port == 0);
    check(addr.IPv6[0] == 0x0000);
    check(addr.IPv6[1] == 0x0000);
    check(addr.IPv6[2] == 0x0000);
    check(addr.IPv6[3] == 0x0000);
    check(addr.IPv6[4] == 0x0000);
    check(addr.IPv6[5] == 0x0000);
    check(addr.IPv6[6] == 0x0000);
    check(addr.IPv6[7] == 0x0001);
  }

  {
    check(addr.parse("[fe80::202:b3ff:fe1e:8329]:40000") == true);
    check(addr.Type == net::AddressType::IPv6);
    check(addr.Port == 40000);
    check(addr.IPv6[0] == 0xfe80);
    check(addr.IPv6[1] == 0x0000);
    check(addr.IPv6[2] == 0x0000);
    check(addr.IPv6[3] == 0x0000);
    check(addr.IPv6[4] == 0x0202);
    check(addr.IPv6[5] == 0xb3ff);
    check(addr.IPv6[6] == 0xfe1e);
    check(addr.IPv6[7] == 0x8329);
  }

  {
    check(addr.parse("[::]:40000") == true);
    check(addr.Type == net::AddressType::IPv6);
    check(addr.Port == 40000);
    check(addr.IPv6[0] == 0x0000);
    check(addr.IPv6[1] == 0x0000);
    check(addr.IPv6[2] == 0x0000);
    check(addr.IPv6[3] == 0x0000);
    check(addr.IPv6[4] == 0x0000);
    check(addr.IPv6[5] == 0x0000);
    check(addr.IPv6[6] == 0x0000);
    check(addr.IPv6[7] == 0x0000);
  }

  {
    check(addr.parse("[::1]:40000") == true);
    check(addr.Type == net::AddressType::IPv6);
    check(addr.Port == 40000);
    check(addr.IPv6[0] == 0x0000);
    check(addr.IPv6[1] == 0x0000);
    check(addr.IPv6[2] == 0x0000);
    check(addr.IPv6[3] == 0x0000);
    check(addr.IPv6[4] == 0x0000);
    check(addr.IPv6[5] == 0x0000);
    check(addr.IPv6[6] == 0x0000);
    check(addr.IPv6[7] == 0x0001);
  }
}

Test(legacy_test_address)
{
  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "[") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "[]") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "[]:") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, ":") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "1") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "12") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "123") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "1234") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "1234.0.12313.0000") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "1234.0.12313.0000.0.0.0.0.0") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address,
           "1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, ".") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "..") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "...") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, "....") == RELAY_ERROR);
    check(legacy::relay_address_parse(&address, ".....") == RELAY_ERROR);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "107.77.207.77") == RELAY_OK);
    check(address.type == net::AddressType::IPv4);
    check(address.port == 0);
    check(address.data.ipv4[0] == 107);
    check(address.data.ipv4[1] == 77);
    check(address.data.ipv4[2] == 207);
    check(address.data.ipv4[3] == 77);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "127.0.0.1") == RELAY_OK);
    check(address.type == net::AddressType::IPv4);
    check(address.port == 0);
    check(address.data.ipv4[0] == 127);
    check(address.data.ipv4[1] == 0);
    check(address.data.ipv4[2] == 0);
    check(address.data.ipv4[3] == 1);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "107.77.207.77:40000") == RELAY_OK);
    check(address.type == net::AddressType::IPv4);
    check(address.port == 40000);
    check(address.data.ipv4[0] == 107);
    check(address.data.ipv4[1] == 77);
    check(address.data.ipv4[2] == 207);
    check(address.data.ipv4[3] == 77);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "127.0.0.1:40000") == RELAY_OK);
    check(address.type == net::AddressType::IPv4);
    check(address.port == 40000);
    check(address.data.ipv4[0] == 127);
    check(address.data.ipv4[1] == 0);
    check(address.data.ipv4[2] == 0);
    check(address.data.ipv4[3] == 1);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "fe80::202:b3ff:fe1e:8329") == RELAY_OK);
    check(address.type == net::AddressType::IPv6);
    check(address.port == 0);
    check(address.data.ipv6[0] == 0xfe80);
    check(address.data.ipv6[1] == 0x0000);
    check(address.data.ipv6[2] == 0x0000);
    check(address.data.ipv6[3] == 0x0000);
    check(address.data.ipv6[4] == 0x0202);
    check(address.data.ipv6[5] == 0xb3ff);
    check(address.data.ipv6[6] == 0xfe1e);
    check(address.data.ipv6[7] == 0x8329);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "::") == RELAY_OK);
    check(address.type == net::AddressType::IPv6);
    check(address.port == 0);
    check(address.data.ipv6[0] == 0x0000);
    check(address.data.ipv6[1] == 0x0000);
    check(address.data.ipv6[2] == 0x0000);
    check(address.data.ipv6[3] == 0x0000);
    check(address.data.ipv6[4] == 0x0000);
    check(address.data.ipv6[5] == 0x0000);
    check(address.data.ipv6[6] == 0x0000);
    check(address.data.ipv6[7] == 0x0000);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "::1") == RELAY_OK);
    check(address.type == net::AddressType::IPv6);
    check(address.port == 0);
    check(address.data.ipv6[0] == 0x0000);
    check(address.data.ipv6[1] == 0x0000);
    check(address.data.ipv6[2] == 0x0000);
    check(address.data.ipv6[3] == 0x0000);
    check(address.data.ipv6[4] == 0x0000);
    check(address.data.ipv6[5] == 0x0000);
    check(address.data.ipv6[6] == 0x0000);
    check(address.data.ipv6[7] == 0x0001);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "[fe80::202:b3ff:fe1e:8329]:40000") == RELAY_OK);
    check(address.type == net::AddressType::IPv6);
    check(address.port == 40000);
    check(address.data.ipv6[0] == 0xfe80);
    check(address.data.ipv6[1] == 0x0000);
    check(address.data.ipv6[2] == 0x0000);
    check(address.data.ipv6[3] == 0x0000);
    check(address.data.ipv6[4] == 0x0202);
    check(address.data.ipv6[5] == 0xb3ff);
    check(address.data.ipv6[6] == 0xfe1e);
    check(address.data.ipv6[7] == 0x8329);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "[::]:40000") == RELAY_OK);
    check(address.type == net::AddressType::IPv6);
    check(address.port == 40000);
    check(address.data.ipv6[0] == 0x0000);
    check(address.data.ipv6[1] == 0x0000);
    check(address.data.ipv6[2] == 0x0000);
    check(address.data.ipv6[3] == 0x0000);
    check(address.data.ipv6[4] == 0x0000);
    check(address.data.ipv6[5] == 0x0000);
    check(address.data.ipv6[6] == 0x0000);
    check(address.data.ipv6[7] == 0x0000);
  }

  {
    legacy::relay_address_t address;
    check(legacy::relay_address_parse(&address, "[::1]:40000") == RELAY_OK);
    check(address.type == net::AddressType::IPv6);
    check(address.port == 40000);
    check(address.data.ipv6[0] == 0x0000);
    check(address.data.ipv6[1] == 0x0000);
    check(address.data.ipv6[2] == 0x0000);
    check(address.data.ipv6[3] == 0x0000);
    check(address.data.ipv6[4] == 0x0000);
    check(address.data.ipv6[5] == 0x0000);
    check(address.data.ipv6[6] == 0x0000);
    check(address.data.ipv6[7] == 0x0001);
  }
}
