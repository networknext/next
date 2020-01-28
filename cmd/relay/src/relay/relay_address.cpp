#include "relay_address.hpp"

#include <cassert>
#include <cstdlib>
#include <cstring>

#include "config.hpp"

namespace relay
{
    int relay_address_parse(relay_address_t* address, const char* address_string_in)
    {
        assert(address);
        assert(address_string_in);

        if (!address)
            return RELAY_ERROR;

        if (!address_string_in)
            return RELAY_ERROR;

        memset(address, 0, sizeof(relay::relay_address_t));

        // first try to parse the string as an IPv6 address:
        // 1. if the first character is '[' then it's probably an ipv6 in form "[addr6]:portnum"
        // 2. otherwise try to parse as a raw IPv6 address using inet_pton

        char buffer[RELAY_MAX_ADDRESS_STRING_LENGTH + RELAY_ADDRESS_BUFFER_SAFETY * 2];

        char* address_string = buffer + RELAY_ADDRESS_BUFFER_SAFETY;
        strncpy(address_string, address_string_in, RELAY_MAX_ADDRESS_STRING_LENGTH - 1);
        address_string[RELAY_MAX_ADDRESS_STRING_LENGTH - 1] = '\0';

        int address_string_length = (int)strlen(address_string);

        if (address_string[0] == '[') {
            const int base_index = address_string_length - 1;

            // note: no need to search past 6 characters as ":65535" is longest possible port value
            for (int i = 0; i < 6; ++i) {
                const int index = base_index - i;
                if (index < 0) {
                    return RELAY_ERROR;
                }
                if (address_string[index] == ':') {
                    address->port = (uint16_t)(atoi(&address_string[index + 1]));
                    address_string[index - 1] = '\0';
                    break;
                } else if (address_string[index] == ']') {
                    // no port number
                    address->port = 0;
                    address_string[index] = '\0';
                    break;
                }
            }
            address_string += 1;
        }
        uint16_t addr6[8];
        if (relay_platform_inet_pton6(address_string, addr6) == RELAY_OK) {
            address->type = RELAY_ADDRESS_IPV6;
            for (int i = 0; i < 8; ++i) {
                address->data.ipv6[i] = relay_platform_ntohs(addr6[i]);
            }
            return RELAY_OK;
        }

        // otherwise it's probably an IPv4 address:
        // 1. look for ":portnum", if found save the portnum and strip it out
        // 2. parse remaining ipv4 address via inet_pton

        address_string_length = (int)strlen(address_string);
        const int base_index = address_string_length - 1;
        for (int i = 0; i < 6; ++i) {
            const int index = base_index - i;
            if (index < 0)
                break;
            if (address_string[index] == ':') {
                address->port = (uint16_t)(atoi(&address_string[index + 1]));
                address_string[index] = '\0';
            }
        }

        uint32_t addr4;
        if (relay_platform_inet_pton4(address_string, &addr4) == RELAY_OK) {
            address->type = RELAY_ADDRESS_IPV4;
            address->data.ipv4[3] = (uint8_t)((addr4 & 0xFF000000) >> 24);
            address->data.ipv4[2] = (uint8_t)((addr4 & 0x00FF0000) >> 16);
            address->data.ipv4[1] = (uint8_t)((addr4 & 0x0000FF00) >> 8);
            address->data.ipv4[0] = (uint8_t)((addr4 & 0x000000FF));
            return RELAY_OK;
        }

        return RELAY_ERROR;
    }

}  // namespace relay