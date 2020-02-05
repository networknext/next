#include "write.test.hpp"

#include <cstddef>

#include "testing/macros.hpp"

#include "net/address.hpp"
#include "encoding/write.hpp"

namespace testing
{
    void Test_WriteAddress_ipv4()
    {
        net::Address addr;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> bin;

        bin.fill(0);
        addr.parse("127.0.0.1:51034");
        size_t index = 0;
        encoding::WriteAddress(bin, index, addr);
        check(index == RELAY_ADDRESS_BYTES);
        check(bin[0] == RELAY_ADDRESS_IPV4);
        check(bin[1] == 127);
        check(bin[2] == 0);
        check(bin[3] == 0);
        check(bin[4] == 1);
        check(bin[5] == 0x5A);
        check(bin[6] == 0xC7);
    }

    void Test_WriteAddress_ipv6()
    {
        net::Address addr;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> bin;

        bin.fill(0);
        addr.parse("[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:51034");
        size_t index = 0;
        encoding::WriteAddress(bin, index, addr);
        check(index == RELAY_ADDRESS_BYTES);
        check(bin[0] == RELAY_ADDRESS_IPV6);
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
        check(bin[19] == 0x00);
    }

    void Test_WriteAddress_none()
    {
        net::Address addr;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> bin;

        bin.fill(0);
        addr.parse("1udai898haidfihe");
        size_t index = 0;
        encoding::WriteAddress(bin, index, addr);
        for (auto& i : bin) {
            check(i == 0);
        }
    }

    void TestWrite()
    {
        Test_WriteAddress_ipv4();
        Test_WriteAddress_ipv6();
        Test_WriteAddress_none();
    }
}  // namespace testing
