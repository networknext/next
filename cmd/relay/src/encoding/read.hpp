#ifndef ENCODING_READ_HPP
#define ENCODING_READ_HPP

#include <cinttypes>
#include <cstring>
#include <cassert>

#include "config.hpp"

#include "relay/relay_address.hpp"

namespace encoding
{
    inline uint8_t relay_read_uint8(const uint8_t** p)
    {
        uint8_t value = **p;
        ++(*p);
        return value;
    }

    inline uint16_t relay_read_uint16(const uint8_t** p)
    {
        uint16_t value;
        value = (*p)[0];
        value |= (((uint16_t)((*p)[1])) << 8);
        *p += 2;
        return value;
    }

    inline uint32_t relay_read_uint32(const uint8_t** p)
    {
        uint32_t value;
        value = (*p)[0];
        value |= (((uint32_t)((*p)[1])) << 8);
        value |= (((uint32_t)((*p)[2])) << 16);
        value |= (((uint32_t)((*p)[3])) << 24);
        *p += 4;
        return value;
    }

    inline uint64_t relay_read_uint64(const uint8_t** p)
    {
        uint64_t value;
        value = (*p)[0];
        value |= (((uint64_t)((*p)[1])) << 8);
        value |= (((uint64_t)((*p)[2])) << 16);
        value |= (((uint64_t)((*p)[3])) << 24);
        value |= (((uint64_t)((*p)[4])) << 32);
        value |= (((uint64_t)((*p)[5])) << 40);
        value |= (((uint64_t)((*p)[6])) << 48);
        value |= (((uint64_t)((*p)[7])) << 56);
        *p += 8;
        return value;
    }

    inline float relay_read_float32(const uint8_t** p)
    {
        uint32_t value_int = relay_read_uint32(p);
        float value_float = 0.0f;
        uint8_t* pointer_int = (uint8_t*)(&value_int);
        uint8_t* pointer_float = (uint8_t*)(&value_float);
        memcpy(pointer_float, pointer_int, sizeof(value_int));
        return value_float;
    }

    inline double relay_read_float64(const uint8_t** p)
    {
        uint64_t value_int = relay_read_uint64(p);
        double value_float = 0.0;
        uint8_t* pointer_int = (uint8_t*)(&value_int);
        uint8_t* pointer_float = (uint8_t*)(&value_float);
        memcpy(pointer_float, pointer_int, sizeof(value_int));
        return value_float;
    }

    inline void relay_read_bytes(const uint8_t** p, uint8_t* byte_array, int num_bytes)
    {
        for (int i = 0; i < num_bytes; ++i) {
            byte_array[i] = relay_read_uint8(p);
        }
    }

    inline void relay_read_string(const uint8_t** p, char* string_data, uint32_t max_length)
    {
        uint32_t length = relay_read_uint32(p);
        if (length > max_length) {
            length = 0;
            return;
        }
        uint32_t i = 0;
        for (; i < length; ++i) {
            string_data[i] = relay_read_uint8(p);
        }
        string_data[i] = 0;
    }

    inline void relay_read_address(const uint8_t** buffer, relay::relay_address_t* address)
    {
        const uint8_t* start = *buffer;

        address->type = relay_read_uint8(buffer);

        if (address->type == RELAY_ADDRESS_IPV4) {
            for (int j = 0; j < 4; ++j) {
                address->data.ipv4[j] = relay_read_uint8(buffer);
            }
            address->port = relay_read_uint16(buffer);
            for (int i = 0; i < 12; ++i) {
                uint8_t dummy = relay_read_uint8(buffer);
                (void)dummy;
            }
        } else if (address->type == RELAY_ADDRESS_IPV6) {
            for (int j = 0; j < 8; ++j) {
                address->data.ipv6[j] = relay_read_uint16(buffer);
            }
            address->port = relay_read_uint16(buffer);
        } else {
            for (int i = 0; i < RELAY_ADDRESS_BYTES - 1; ++i) {
                uint8_t dummy = relay_read_uint8(buffer);
                (void)dummy;
            }
        }

        (void)start;

        assert(*buffer - start == RELAY_ADDRESS_BYTES);
    }
}  // namespace encoding
#endif