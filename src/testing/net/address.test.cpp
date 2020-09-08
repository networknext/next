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

  addr.reset();
  check(addr.parse("[::]"));
  check(addr.Type == net::AddressType::IPv6);
  check(addr.Port == 0);
  check(addr.IPv6[0] == 0);
  check(addr.IPv6[1] == 0);
  check(addr.IPv6[2] == 0);
  check(addr.IPv6[3] == 0);
  check(addr.IPv6[4] == 0);
  check(addr.IPv6[5] == 0);
  check(addr.IPv6[6] == 0);
  check(addr.IPv6[7] == 0);
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
  addr.reset();
  check(addr.parse(base) == true);
  check(addr.to_string(output));
  check(output == expected);

  base = "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000";
  expected = "[2001:db8:85a3::8a2e:370:7334]:20000";
  addr.reset();
  check(addr.parse(base) == true);
  check(addr.to_string(output));
  check(output == expected);

  base = "2001:0db8:85a3:0000:0000:8a2e:0370:7334";
  expected = "2001:db8:85a3::8a2e:370:7334";
  addr.reset();
  check(addr.parse(base) == true);
  check(addr.to_string(output));
  check(output == expected).onFail([&] {
    std::cout << "output = '" << output << '\'' << std::endl;
    std::cout << "expected = '" << expected << '\'' << std::endl;
  });

  base = "127.0djaid?.0.sad1:5as1034";
  expected = "NONE";
  addr.reset();
  check(addr.parse(base) == false);
  check(addr.to_string(output));
  check(output == expected);
}

Test(Address_parse_additional)
{
  net::Address addr;

  // all should fail
  {
    addr.reset();
    check(addr.parse("") == false);
    addr.reset();
    check(addr.parse("[") == false);
    addr.reset();
    check(addr.parse("[]") == false);
    addr.reset();
    check(addr.parse("[]:") == false);
    addr.reset();
    check(addr.parse(":") == false);
    addr.reset();
    check(addr.parse("1") == false);
    addr.reset();
    check(addr.parse("12") == false);
    addr.reset();
    check(addr.parse("123") == false);
    addr.reset();
    check(addr.parse("1234") == false);
    addr.reset();
    check(addr.parse("1234.0.12313.0000") == false);
    addr.reset();
    check(addr.parse("1234.0.12313.0000.0.0.0.0.0") == false);
    addr.reset();
    check(addr.parse("1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131:1312313:123131") == false);
    addr.reset();
    check(addr.parse(".") == false);
    addr.reset();
    check(addr.parse("..") == false);
    addr.reset();
    check(addr.parse("...") == false);
    addr.reset();
    check(addr.parse("....") == false);
    addr.reset();
    check(addr.parse(".....") == false);
    addr.reset();
    check(addr.parse("0.0.0.0") == true);
  }

  // reset should pass

  {
    addr.reset();
    check(addr.parse("107.77.207.77") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 0);
    check(addr.IPv4[0] == 107);
    check(addr.IPv4[1] == 77);
    check(addr.IPv4[2] == 207);
    check(addr.IPv4[3] == 77);
  }

  {
    addr.reset();
    check(addr.parse("127.0.0.1") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 0);
    check(addr.IPv4[0] == 127);
    check(addr.IPv4[1] == 0);
    check(addr.IPv4[2] == 0);
    check(addr.IPv4[3] == 1);
  }

  {
    addr.reset();
    check(addr.parse("107.77.207.77:40000") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 40000);
    check(addr.IPv4[0] == 107);
    check(addr.IPv4[1] == 77);
    check(addr.IPv4[2] == 207);
    check(addr.IPv4[3] == 77);
  }

  {
    addr.reset();
    check(addr.parse("127.0.0.1:40000") == true);
    check(addr.Type == net::AddressType::IPv4);
    check(addr.Port == 40000);
    check(addr.IPv4[0] == 127);
    check(addr.IPv4[1] == 0);
    check(addr.IPv4[2] == 0);
    check(addr.IPv4[3] == 1);
    addr.reset();
  }

  {
    addr.reset();
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
    addr.reset();
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
    addr.reset();
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
    addr.reset();
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
    addr.reset();
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
    addr.reset();
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
