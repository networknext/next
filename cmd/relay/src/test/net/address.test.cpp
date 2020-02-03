#include "address.test.hpp"

#include <iostream>
#include <cstdlib>
#include <array>

#include "test/macros.hpp"

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

        a.IPv4[0] = 1;
        check(a != b);
        a.IPv4[0] = 1;

        *a.IPv4.end() = 1;
        check(a != b);
        *a.IPv4.end() = 0;

        a.IPv6[0] = 1;
        check(a != b);
        a.IPv6[0] = 0;

        *a.IPv6.end() = 1;
        check(a != b);
        *a.IPv6.end() = 0;

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
        addr.parse("127.0.0.1:51034");
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
        addr.parse("[::1]:51034");
        check(addr.Type = RELAY_ADDRESS_IPV6);
        check(addr.Port == 51034);
        check(addr.IPv6[0] == 1);
        check(addr.IPv6[1] == 0);
        check(addr.IPv6[2] == 0);
        check(addr.IPv6[3] == 0);
        check(addr.IPv6[4] == 0);
        check(addr.IPv6[5] == 0);
        check(addr.IPv6[6] == 0);
        check(addr.IPv6[7] == 0);

        addr.reset();
        addr.parse("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:20000");
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
        addr.parse("2001:0db8:85a3:0000:0000:8a2e:0370:7334");
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
        addr.parse("127.0.:182a");
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

    // Type == none
    void TestAddressReadAndWrite1()
    {
        net::Address a, read_a;
        std::array<uint8_t, 1024> buffer;

        buffer.fill(0);
        size_t index = 0;
        encoding::WriteAddress(buffer, index, a);
        check(index == RELAY_ADDRESS_BYTES);
        index = 0;
        encoding::ReadAddress(buffer, index, read_a);
        check(index == RELAY_ADDRESS_BYTES);
        check(a == read_a);
    }

    // Type == ipv4
    void TestAddressReadAndWrite2()
    {
        net::Address b, read_b;
        std::array<uint8_t, 1024> buffer;

        buffer.fill(0);
        b.parse("127.0.0.1:50000");
        size_t index = 0;
        encoding::WriteAddress(buffer, index, b);
        check(index == RELAY_ADDRESS_BYTES);
        index = 0;
        encoding::ReadAddress(buffer, index, read_b);
        check(index == RELAY_ADDRESS_BYTES);
        check(b == read_b);
    }

    // Type == ipv6
    void TestAddressReadAndWrite3()
    {
        net::Address c, read_c;
        std::array<uint8_t, 1024> buffer;

        buffer.fill(0);
        c.parse("[::1]:50000");
        size_t index = 0;
        encoding::WriteAddress(buffer, index, c);
        check(index == RELAY_ADDRESS_BYTES);
        index = 0;
        encoding::ReadAddress(buffer, index, read_c);
        check(index == RELAY_ADDRESS_BYTES);
        check(c == read_c);
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

        TestAddressReadAndWrite1();
        TestAddressReadAndWrite2();
        TestAddressReadAndWrite3();
    }

    void test_address_read_and_write()
    {
        legacy::relay_address_t a, b, c;

        memset(&a, 0, sizeof(a));

        legacy::relay_address_parse(&b, "127.0.0.1:50000");

        legacy::relay_address_parse(&c, "[::1]:50000");

        uint8_t buffer[1024];

        uint8_t* p = buffer;

        encoding::write_address(&p, &a);
        encoding::write_address(&p, &b);
        encoding::write_address(&p, &c);

        struct legacy::relay_address_t read_a, read_b, read_c;

        const uint8_t* q = buffer;

        encoding::read_address(&q, &read_a);
        encoding::read_address(&q, &read_b);
        encoding::read_address(&q, &read_c);

        check(legacy::relay_address_equal(&a, &read_a));
        check(legacy::relay_address_equal(&b, &read_b));
        check(legacy::relay_address_equal(&c, &read_c));
    }
}  // namespace testing