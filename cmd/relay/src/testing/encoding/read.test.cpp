#include "read.test.hpp"

#include <array>

#include "testing/macros.hpp"
#include "net/address.hpp"
#include "encoding/read.hpp"

namespace testing
{
    void Test_ReadAddress_ipv4()
    {
        net::Address addr;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> bin;
        bin.fill(0);
        bin[0] = RELAY_ADDRESS_IPV4;
        bin[1] = 127;
        bin[2] = 0;
        bin[3] = 0;
        bin[4] = 1;
        bin[5] = 0xFF;
        bin[6] = 0xFF;

        size_t index = 0;
        encoding::ReadAddress(bin, index, addr);
        check(index == RELAY_ADDRESS_BYTES);
        check(addr.toString() == "127.0.0.1:65535");
    }

    void Test_ReadAddress_ipv6()
    {
        net::Address addr;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> bin;
        bin.fill(0);
        bin[0] = RELAY_ADDRESS_IPV6;
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
        bin[19] = 0xFF;

        size_t index = 0;
        encoding::ReadAddress(bin, index, addr);
        check(index == RELAY_ADDRESS_BYTES);
        check(addr.toString() == "[3b1f:3c33:9928:ffff:ffff:ffff:ffff:ffff]:65535");
    }

    void Test_ReadAddress_none()
    {
        net::Address before, after;
        std::array<uint8_t, RELAY_ADDRESS_BYTES> bin;
        bin.fill(0);
        bin[0] = RELAY_ADDRESS_NONE;
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
        bin[19] = 0xFF;

        size_t index = 0;
        encoding::ReadAddress(bin, index, after);
        check(index == RELAY_ADDRESS_BYTES);
        check(after.toString() == "NONE");
        check(before == after);
    }

    void TestRead()
    {
        Test_ReadAddress_ipv4();
        Test_ReadAddress_ipv6();
        Test_ReadAddress_none();
    }
}  // namespace testing