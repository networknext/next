#include "address.test.hpp"

#include <iostream>
#include <cstdlib>
#include <array>

#include "testing/macros.hpp"

#include "net/address.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

namespace testing
{
    void Test_Address_operator_eq_and_neq()
    {
        net::Address a, b;

        check(a == b);

        a.Type = 1;
        check(a != b);
        a.Type = 0;

        a.Port = 1;
        check(a != b);
        a.Port = 0;

        a.Type = RELAY_ADDRESS_IPV4;

        a.IPv4[0] = 1;
        check(a != b);
        a.IPv4[0] = 0;

        *a.IPv4.end() = 1;
        check(a != b);
        *a.IPv4.end() = 0;

        a.Type = RELAY_ADDRESS_IPV6;

        a.IPv6[0] = 1;
        check(a != b);
        a.IPv6[0] = 0;

        *a.IPv6.end() = 1;
        check(a != b);
        *a.IPv6.end() = 0;

        a.Type = RELAY_ADDRESS_NONE;

        check(a == b);
    }

    void Test_Address_reset()
    {
        net::Address addr, base;
        addr.Type = 77;
        addr.Port = 123;
        addr.IPv4[0] = 18;
        addr.IPv4[1] = 82;
        addr.IPv4[2] = 53;
        addr.IPv4[3] = 67;

        addr.reset();
        check(addr == base);
    }

    void Test_Address_parse_ipv4()
    {
        net::Address addr;
        check(addr.parse("127.0.0.1:51034") == true);
        check(addr.Type == RELAY_ADDRESS_IPV4);
        check(addr.Port == 51034);
        check(addr.IPv4[0] == 127);
        check(addr.IPv4[1] == 0);
        check(addr.IPv4[2] == 0);
        check(addr.IPv4[3] == 1);
    }

    void Test_Address_parse_ipv6_with_braces()
    {
        net::Address addr;
        check(addr.parse("[::1]:51034") == true);
        check(addr.Type = RELAY_ADDRESS_IPV6);
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
        check(addr.Type = RELAY_ADDRESS_IPV6);
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

    void Test_Address_parse_ipv6_without_braces()
    {
        net::Address addr;
        check(addr.parse("2001:0db8:85a3:0000:0000:8a2e:0370:7334") == true);
        check(addr.Type = RELAY_ADDRESS_IPV6);
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

    void Test_Address_parse_invalid_ips()
    {
        net::Address addr;
        check(addr.parse("127.0.:182a") == false);
        check(addr.Type == RELAY_ADDRESS_NONE);
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

    void Test_Address_toString()
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

    void TestAddress()
    {
        Test_Address_operator_eq_and_neq();  // Always test first, eq ops are used in subsequent tests and you don't want to
                                             // spend time debugging the wrong thing.
        Test_Address_reset();                // And then this one.
        Test_Address_parse_ipv4();
        Test_Address_parse_ipv6_with_braces();
        Test_Address_parse_ipv6_without_braces();
        Test_Address_parse_invalid_ips();
        Test_Address_toString();
    }
}  // namespace testing