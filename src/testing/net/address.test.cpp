#include "includes.h"
#include "testing/test.hpp"

#include "net/address.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

TEST(net_Address_operator_eq_and_neq)
{
  net::Address a, b;

  CHECK(a == b);

  a.type = net::AddressType::IPv4;
  CHECK(a != b);
  a.type = net::AddressType::None;

  a.port = 1;
  CHECK(a != b);
  a.port = 0;

  a.type = net::AddressType::IPv4;

  a.ipv4[0] = 1;
  CHECK(a != b);
  a.ipv4[0] = 0;

  *a.ipv4.end() = 1;
  CHECK(a != b);
  *a.ipv4.end() = 0;

  a.type = net::AddressType::IPv6;

  a.ipv6[0] = 1;
  CHECK(a != b);
  a.ipv6[0] = 0;

  *a.ipv6.end() = 1;
  CHECK(a != b);
  *a.ipv6.end() = 0;

  a.type = net::AddressType::None;

  CHECK(a == b);
}

TEST(net_Address_reset)
{
  net::Address addr, base;
  addr.type = static_cast<net::AddressType>(77);
  addr.port = 123;
  addr.ipv4[0] = 18;
  addr.ipv4[1] = 82;
  addr.ipv4[2] = 53;
  addr.ipv4[3] = 67;

  addr.reset();
  CHECK(addr == base);
}

TEST(net_Address_parse_ipv4)
{
  net::Address addr;
  CHECK(addr.parse("127.0.0.1:51034") == true);
  CHECK(addr.type == net::AddressType::IPv4);
  CHECK(addr.port == 51034);
  CHECK(addr.ipv4[0] == 127);
  CHECK(addr.ipv4[1] == 0);
  CHECK(addr.ipv4[2] == 0);
  CHECK(addr.ipv4[3] == 1);
}

TEST(net_Address_parse_ipv6_with_braces)
{
  net::Address addr;
  CHECK(addr.parse("[::1]:51034") == true);
  CHECK(addr.type == net::AddressType::IPv6);
  CHECK(addr.port == 51034);
  CHECK(addr.ipv6[0] == 0);
  CHECK(addr.ipv6[1] == 0);
  CHECK(addr.ipv6[2] == 0);
  CHECK(addr.ipv6[3] == 0);
  CHECK(addr.ipv6[4] == 0);
  CHECK(addr.ipv6[5] == 0);
  CHECK(addr.ipv6[6] == 0);
  CHECK(addr.ipv6[7] == 1);

  addr.reset();
  CHECK(addr.parse("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000") == true);
  CHECK(addr.type == net::AddressType::IPv6);
  CHECK(addr.port == 20000);
  CHECK(addr.ipv6[0] == 0x2001);
  CHECK(addr.ipv6[1] == 0x0db8);
  CHECK(addr.ipv6[2] == 0x85a3);
  CHECK(addr.ipv6[3] == 0x0000);
  CHECK(addr.ipv6[4] == 0x0000);
  CHECK(addr.ipv6[5] == 0x8a2e);
  CHECK(addr.ipv6[6] == 0x0370);
  CHECK(addr.ipv6[7] == 0x7334);

  addr.reset();
  CHECK(addr.parse("[::]"));
  CHECK(addr.type == net::AddressType::IPv6);
  CHECK(addr.port == 0);
  CHECK(addr.ipv6[0] == 0);
  CHECK(addr.ipv6[1] == 0);
  CHECK(addr.ipv6[2] == 0);
  CHECK(addr.ipv6[3] == 0);
  CHECK(addr.ipv6[4] == 0);
  CHECK(addr.ipv6[5] == 0);
  CHECK(addr.ipv6[6] == 0);
  CHECK(addr.ipv6[7] == 0);
}

TEST(net_Address_parse_ipv6_without_braces)
{
  net::Address addr;
  CHECK(addr.parse("2001:0db8:85a3:0000:0000:8a2e:0370:7334") == true);
  CHECK(addr.type == net::AddressType::IPv6);
  CHECK(addr.port == 0);
  CHECK(addr.ipv6[0] == 0x2001);
  CHECK(addr.ipv6[1] == 0x0db8);
  CHECK(addr.ipv6[2] == 0x85a3);
  CHECK(addr.ipv6[3] == 0x0000);
  CHECK(addr.ipv6[4] == 0x0000);
  CHECK(addr.ipv6[5] == 0x8a2e);
  CHECK(addr.ipv6[6] == 0x0370);
  CHECK(addr.ipv6[7] == 0x7334);
}

TEST(net_Address_parse_invalid_ips)
{
  net::Address addr;
  CHECK(addr.parse("127.0.:182a") == false);
  CHECK(addr.type == net::AddressType::None);
  CHECK(addr.port == 0);

  CHECK(addr.ipv4[0] == 0);
  CHECK(addr.ipv4[1] == 0);
  CHECK(addr.ipv4[2] == 0);
  CHECK(addr.ipv4[3] == 0);

  CHECK(addr.ipv6[0] == 0);
  CHECK(addr.ipv6[1] == 0);
  CHECK(addr.ipv6[2] == 0);
  CHECK(addr.ipv6[3] == 0);
  CHECK(addr.ipv6[4] == 0);
  CHECK(addr.ipv6[5] == 0);
  CHECK(addr.ipv6[6] == 0);
  CHECK(addr.ipv6[7] == 0);
}

TEST(net_Address_toString)
{
  net::Address addr;
  std::string base, output, expected;

  base = "127.0.0.1:51034";
  expected = base;
  addr.reset();
  CHECK(addr.parse(base) == true);
  CHECK(addr.to_string(output));
  CHECK(output == expected);

  base = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";
  expected = "[2001:db8:85a3::8a2e:370:7334]:20000";
  addr.reset();
  CHECK(addr.parse(base) == true);
  CHECK(addr.to_string(output));
  CHECK(output == expected);

  base = "2001:0db8:85a3:0000:0000:8a2e:0370:7334";
  expected = "2001:db8:85a3::8a2e:370:7334";
  addr.reset();
  CHECK(addr.parse(base) == true);
  CHECK(addr.to_string(output));
  CHECK(output == expected).on_fail([&] {
    std::cout << "output = '" << output << '\'' << std::endl;
    std::cout << "expected = '" << expected << '\'' << std::endl;
  });

  base = "127.0djaid?.0.sad1:5as1034";
  expected = "NONE";
  addr.reset();
  CHECK(addr.parse(base) == false);
  CHECK(addr.to_string(output));
  CHECK(output == expected);
}

TEST(net_Address_parse_additional)
{
  net::Address addr;

  // all should fail
  {
    addr.reset();
    CHECK(addr.parse("") == false);
    addr.reset();
    CHECK(addr.parse("[") == false);
    addr.reset();
    CHECK(addr.parse("[]") == false);
    addr.reset();
    CHECK(addr.parse("[]:") == false);
    addr.reset();
    CHECK(addr.parse(":") == false);
    addr.reset();
    CHECK(addr.parse("1") == false);
    addr.reset();
    CHECK(addr.parse("12") == false);
    addr.reset();
    CHECK(addr.parse("123") == false);
    addr.reset();
    CHECK(addr.parse("1234") == false);
    addr.reset();
    CHECK(addr.parse("1234.0.12313.0000") == false);
    addr.reset();
    CHECK(addr.parse("1234.0.12313.0000.0.0.0.0.0") == false);
    addr.reset();
    CHECK(addr.parse("1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131") == false);
    addr.reset();
    CHECK(addr.parse(".") == false);
    addr.reset();
    CHECK(addr.parse("..") == false);
    addr.reset();
    CHECK(addr.parse("...") == false);
    addr.reset();
    CHECK(addr.parse("....") == false);
    addr.reset();
    CHECK(addr.parse(".....") == false);
    addr.reset();
    CHECK(addr.parse("0.0.0.0") == true);
  }

  // reset should pass

  {
    addr.reset();
    CHECK(addr.parse("107.77.207.77") == true);
    CHECK(addr.type == net::AddressType::IPv4);
    CHECK(addr.port == 0);
    CHECK(addr.ipv4[0] == 107);
    CHECK(addr.ipv4[1] == 77);
    CHECK(addr.ipv4[2] == 207);
    CHECK(addr.ipv4[3] == 77);
  }

  {
    addr.reset();
    CHECK(addr.parse("127.0.0.1") == true);
    CHECK(addr.type == net::AddressType::IPv4);
    CHECK(addr.port == 0);
    CHECK(addr.ipv4[0] == 127);
    CHECK(addr.ipv4[1] == 0);
    CHECK(addr.ipv4[2] == 0);
    CHECK(addr.ipv4[3] == 1);
  }

  {
    addr.reset();
    CHECK(addr.parse("107.77.207.77:40000") == true);
    CHECK(addr.type == net::AddressType::IPv4);
    CHECK(addr.port == 40000);
    CHECK(addr.ipv4[0] == 107);
    CHECK(addr.ipv4[1] == 77);
    CHECK(addr.ipv4[2] == 207);
    CHECK(addr.ipv4[3] == 77);
  }

  {
    addr.reset();
    CHECK(addr.parse("127.0.0.1:40000") == true);
    CHECK(addr.type == net::AddressType::IPv4);
    CHECK(addr.port == 40000);
    CHECK(addr.ipv4[0] == 127);
    CHECK(addr.ipv4[1] == 0);
    CHECK(addr.ipv4[2] == 0);
    CHECK(addr.ipv4[3] == 1);
    addr.reset();
  }

  {
    addr.reset();
    CHECK(addr.parse("fe80::202:b3ff:fe1e:8329") == true);
    CHECK(addr.type == net::AddressType::IPv6);
    CHECK(addr.port == 0);
    CHECK(addr.ipv6[0] == 0xfe80);
    CHECK(addr.ipv6[1] == 0x0000);
    CHECK(addr.ipv6[2] == 0x0000);
    CHECK(addr.ipv6[3] == 0x0000);
    CHECK(addr.ipv6[4] == 0x0202);
    CHECK(addr.ipv6[5] == 0xb3ff);
    CHECK(addr.ipv6[6] == 0xfe1e);
    CHECK(addr.ipv6[7] == 0x8329);
  }

  {
    addr.reset();
    CHECK(addr.parse("::") == true);
    CHECK(addr.type == net::AddressType::IPv6);
    CHECK(addr.port == 0);
    CHECK(addr.ipv6[0] == 0x0000);
    CHECK(addr.ipv6[1] == 0x0000);
    CHECK(addr.ipv6[2] == 0x0000);
    CHECK(addr.ipv6[3] == 0x0000);
    CHECK(addr.ipv6[4] == 0x0000);
    CHECK(addr.ipv6[5] == 0x0000);
    CHECK(addr.ipv6[6] == 0x0000);
    CHECK(addr.ipv6[7] == 0x0000);
  }

  {
    addr.reset();
    CHECK(addr.parse("::1") == true);
    CHECK(addr.type == net::AddressType::IPv6);
    CHECK(addr.port == 0);
    CHECK(addr.ipv6[0] == 0x0000);
    CHECK(addr.ipv6[1] == 0x0000);
    CHECK(addr.ipv6[2] == 0x0000);
    CHECK(addr.ipv6[3] == 0x0000);
    CHECK(addr.ipv6[4] == 0x0000);
    CHECK(addr.ipv6[5] == 0x0000);
    CHECK(addr.ipv6[6] == 0x0000);
    CHECK(addr.ipv6[7] == 0x0001);
  }

  {
    addr.reset();
    CHECK(addr.parse("[fe80::202:b3ff:fe1e:8329]:40000") == true);
    CHECK(addr.type == net::AddressType::IPv6);
    CHECK(addr.port == 40000);
    CHECK(addr.ipv6[0] == 0xfe80);
    CHECK(addr.ipv6[1] == 0x0000);
    CHECK(addr.ipv6[2] == 0x0000);
    CHECK(addr.ipv6[3] == 0x0000);
    CHECK(addr.ipv6[4] == 0x0202);
    CHECK(addr.ipv6[5] == 0xb3ff);
    CHECK(addr.ipv6[6] == 0xfe1e);
    CHECK(addr.ipv6[7] == 0x8329);
  }

  {
    addr.reset();
    CHECK(addr.parse("[::]:40000") == true);
    CHECK(addr.type == net::AddressType::IPv6);
    CHECK(addr.port == 40000);
    CHECK(addr.ipv6[0] == 0x0000);
    CHECK(addr.ipv6[1] == 0x0000);
    CHECK(addr.ipv6[2] == 0x0000);
    CHECK(addr.ipv6[3] == 0x0000);
    CHECK(addr.ipv6[4] == 0x0000);
    CHECK(addr.ipv6[5] == 0x0000);
    CHECK(addr.ipv6[6] == 0x0000);
    CHECK(addr.ipv6[7] == 0x0000);
  }

  {
    addr.reset();
    CHECK(addr.parse("[::1]:40000") == true);
    CHECK(addr.type == net::AddressType::IPv6);
    CHECK(addr.port == 40000);
    CHECK(addr.ipv6[0] == 0x0000);
    CHECK(addr.ipv6[1] == 0x0000);
    CHECK(addr.ipv6[2] == 0x0000);
    CHECK(addr.ipv6[3] == 0x0000);
    CHECK(addr.ipv6[4] == 0x0000);
    CHECK(addr.ipv6[5] == 0x0000);
    CHECK(addr.ipv6[6] == 0x0000);
    CHECK(addr.ipv6[7] == 0x0001);
  }
}
