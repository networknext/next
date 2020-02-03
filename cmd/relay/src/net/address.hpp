#ifndef RELAY_RELAY_ADDRESS
#define RELAY_RELAY_ADDRESS

#include <array>
#include <string>
#include <cinttypes>

#include "config.hpp"

namespace net
{
    class Address
    {
       public:
        Address();
        ~Address() = default;

        bool parse(const std::string& address_string_in);

        void toString(std::string& buffer);

        bool operator==(const Address& other);

        void reset();

        uint8_t Type;
        uint16_t Port;

        union
        {
            std::array<uint8_t, 4> IPv4;
            std::array<uint16_t, 8> IPv6;
        };
    };

    inline void Address::reset()
    {
        if (Type == RELAY_ADDRESS_IPV4) {
            IPv4.fill(0);
        } else if (Type == RELAY_ADDRESS_IPV6) {
            IPv6.fill(0);
        }

        Type = 0;
        Port = 0;
    }
}  // namespace net

namespace legacy
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
}  // namespace legacy
#endif