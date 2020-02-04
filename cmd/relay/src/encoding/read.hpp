#ifndef ENCODING_READ_HPP
#define ENCODING_READ_HPP

#include <cinttypes>
#include <cstddef>
#include <array>
#include <cassert>

#include "net/address.hpp"

#include "config.hpp"

#include "util/logger.hpp"

#include "binary.hpp"

namespace encoding
{
    // Prototypes

    template <size_t BuffSize>
    uint8_t ReadUint8(std::array<uint8_t, BuffSize>& buff, size_t& index);

    uint8_t read_uint8(const uint8_t** p);

    template <size_t BuffSize>
    uint16_t ReadUint16(std::array<uint8_t, BuffSize>& buff, size_t& index);

    uint16_t read_uint16(const uint8_t** p);

    uint32_t read_uint32(const uint8_t** p);

    uint64_t read_uint64(const uint8_t** p);

    float read_float32(const uint8_t** p);

    double read_float64(const uint8_t** p);

    void read_bytes(const uint8_t** p, uint8_t* byte_array, int num_bytes);

    void read_string(const uint8_t** p, char* string_data, uint32_t max_length);

    template <size_t BuffSize>
    void ReadAddress(std::array<uint8_t, BuffSize>& buff, size_t& index, net::Address& addr);

    void read_address(const uint8_t** buffer, legacy::relay_address_t* address);

    template <size_t BuffSize>
    uint8_t ReadUint8(std::array<uint8_t, BuffSize>& buff, size_t& index)
    {
        return buff[index++];
    }

    template <size_t BuffSize>
    uint16_t ReadUint16(std::array<uint8_t, BuffSize>& buff, size_t& index)
    {
        GCC_NO_OPT_OUT;
        auto retval = *reinterpret_cast<uint16_t*>(&buff[index]);
        index += 2;
        return retval;
    }

    template <size_t BuffSize>
    void ReadAddress(std::array<uint8_t, BuffSize>& buff, size_t& index, net::Address& addr)
    {
        GCC_NO_OPT_OUT;
#ifndef NDEBUG
        auto start = index;
#endif
        addr.Type = ReadUint8(buff, index);  // read the type

        if (addr.Type == RELAY_ADDRESS_IPV4) {
            std::copy(buff.begin() + index, buff.begin() + index + 4, addr.IPv4.begin());  // copy the address
            index += 4;                                                                    // increment the index
            addr.Port = ReadUint16(buff, index);                                           // read the port
            index += 12;  // increment the index past the reserved area
        } else if (addr.Type == RELAY_ADDRESS_IPV6) {
            for (int i = 0; i < 8; i++) {
                addr.IPv6[i] = ReadUint16(buff, index);
            }
            addr.Port = ReadUint16(buff, index);  // read the port
        } else {
            index += RELAY_ADDRESS_BYTES - 1;  // if no type, increment the index past the address area
        }

        assert(index - start == RELAY_ADDRESS_BYTES);
    }
}  // namespace encoding
#endif