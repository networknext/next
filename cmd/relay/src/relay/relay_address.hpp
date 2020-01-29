#ifndef RELAY_RELAY_ADDRESS
#define RELAY_RELAY_ADDRESS

#include <cinttypes>

namespace relay
{
    struct relay_address_t
    {
        union
        {
            uint8_t ipv4[4];
            uint16_t ipv6[8];
        } data;
        uint16_t port;
        uint8_t type;
    };

    int relay_address_parse(relay_address_t* address, const char* address_string_in);

    const char* relay_address_to_string(const relay_address_t* address, char* buffer);

    int relay_address_equal(const relay_address_t* a, const relay_address_t* b);

    const char* relay_address_to_string(const relay_address_t* address, char* buffer);

    int relay_address_equal(const relay_address_t* a, const relay_address_t* b);
}  // namespace relay
#endif