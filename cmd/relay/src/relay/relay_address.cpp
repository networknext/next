#include "relay_address.hpp"

#include <algorithm>
#include <sstream>

#include <cassert>
#include <cstdlib>
#include <cstring>
#include <cstdio>

#include "config.hpp"
#include "util.hpp"

#include "util/logger.hpp"

#include "relay_platform.hpp"

#include "net/net.hpp"

namespace relay
{
    RelayAddress::RelayAddress() : mType(0), mPort(0)
    {}

    bool RelayAddress::parse(const std::string& address)
    {
        __builtin_prefetch(address.c_str());
        // first try to parse the string as an IPv6 address:

        std::array<char, RELAY_MAX_ADDRESS_STRING_LENGTH + RELAY_ADDRESS_BUFFER_SAFETY * 2> buff;
        std::copy(address.begin(), address.end(), buff.begin());  // this is supposed to take advantage of MMX registers
        auto ptr = &buff[0];

        // 1. if the first character is '[' then it's probably an ipv6 in form "[addr6]:portnum"

        if (ptr[0] == '[') {
            // note: no need to search past 6 characters as ":65535" is longest possible port value
            int index = address.length() - 1;
            for (int i = 0; i < 6; i++, index--) {
                if (index < 0) {
                    return false;
                }

                if (ptr[index] == ':') {
                    try {
                        this->mPort = (uint16_t)(atoi(&ptr[index + 1]));  // atoi throws exceptions in c++
                    } catch (const std::invalid_argument& ia) {
                        LogDebug("Invalid argument except when parsing ipv6: ", ia.what());
                        return false;
                    } catch (const std::out_of_range& oor) {
                        LogDebug("Out of range except when parsing ipv6: ", oor.what());
                        return false;
                    } catch (const std::exception& e) {
                        LogDebug("Generic except when parsing ipv6: ", e.what());
                        return false;
                    }
                    ptr[index - 1] = '\0';
                    break;
                }

                if (ptr[index] == ']') {
                    // no port number
                    ptr[index] = '\0';
                    break;
                }
            }
            ptr++;
        }

        // 2. otherwise try to parse as a raw IPv6 address using inet_pton

        uint16_t addr6[8];
        if (relay_platform_inet_pton6(ptr, addr6) == RELAY_OK) {
            this->mType = RELAY_ADDRESS_IPV6;
            for (int i = 0; i < 8; ++i) {
                this->mIPv6[i] = relay_platform_ntohs(addr6[i]);
            }
            return true;
        }

        // otherwise it's probably an IPv4 address:

        // 1. look for ":portnum", if found save the portnum and strip it out
        int index = address.length() - 1;
        for (int i = 0; i < 6; i--, index--) {
            if (index < 0) {
                break;
            }

            if (ptr[index] == ':') {
                try {
                    this->mPort = (uint16_t)(atoi(&ptr[index + 1]));  // atoi throws exceptions in c++
                } catch (const std::invalid_argument& ia) {
                    LogDebug("Invalid argument except when parsing ipv4: ", ia.what());
                    return false;
                } catch (const std::out_of_range& oor) {
                    LogDebug("Out of range except when parsing ipv4: ", oor.what());
                    return false;
                } catch (const std::exception& e) {
                    LogDebug("Generic except when parsing ipv4: ", e.what());
                    return false;
                }
                ptr[index] = '\0';
                break;
            }
        }

        // 2. parse remaining ipv4 address via inet_pton

        uint32_t addr4;
        if (relay_platform_inet_pton4(ptr, &addr4) == RELAY_OK) {
            this->mType = RELAY_ADDRESS_IPV4;
            this->mIPv4[3] = (uint8_t)((addr4 & 0xFF000000) >> 24);
            this->mIPv4[2] = (uint8_t)((addr4 & 0x00FF0000) >> 16);
            this->mIPv4[1] = (uint8_t)((addr4 & 0x0000FF00) >> 8);
            this->mIPv4[0] = (uint8_t)((addr4 & 0x000000FF));
            return true;
        }

        return false;
    }

    void RelayAddress::toString(std::string& buffer)
    {
        std::stringstream ss;

        if (mType == RELAY_ADDRESS_IPV6) {
            // TODO check if c++17 is ok, can replace this with "if constexpr" for less preprocessor littered code
#if defined(WINVER) && WINVER <= 0x0502
            // ipv6 not supported
            buf[0] = '\0';
#else
            uint16_t ipv6_network_order[8];
            for (int i = 0; i < 8; ++i) {
                ipv6_network_order[i] = net::relay_htons(mIPv6[i]);
            }

            std::array<char, RELAY_MAX_ADDRESS_STRING_LENGTH> address_string;
            relay_platform_inet_ntop6(ipv6_network_order, address_string.data(), address_string.size() * sizeof(char));
            ss << '[' << address_string.data() << ']';
            if (mPort != 0) {
                ss << ':' << mPort;
            }
#endif
        } else if (mType == RELAY_ADDRESS_IPV4) {
            ss << (unsigned int)mIPv4[0] << '.' << (unsigned int)mIPv4[1] << '.' << (unsigned int)mIPv4[2] << '.'
               << (unsigned int)mIPv4[3];
            if (mPort != 0) {
                ss << ':' << mPort;
            }
        } else {
            ss << "NONE";
        }

        buffer = std::move(ss.str());
    }

    std::string RelayAddress::toString()
    {
        std::string buffer;
        this->toString(buffer);
        return std::move(buffer);
    }

    bool RelayAddress::operator==(const RelayAddress& other)
    {
        if (this->mType != other.mType || this->mPort != other.mPort) {
            return false;
        }

        switch (this->mType) {
            case RELAY_ADDRESS_IPV4:
                return this->mIPv4 == other.mIPv4;
            case RELAY_ADDRESS_IPV6:
                return this->mIPv6 == other.mIPv6;
            default:
                return false;
        }
    }

    // TODO make this inline or remove
    bool RelayAddress::equal(const RelayAddress& other)
    {
        return (*this) == other;
    }

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
                    try {
                        address->port = (uint16_t)(atoi(&address_string[index + 1]));  // atoi throws exceptions in c++
                    } catch (const std::invalid_argument& ia) {
                        LogDebug("Invalid argument except when parsing ipv6: ", ia.what());
                        return RELAY_ERROR;
                    } catch (const std::out_of_range& oor) {
                        LogDebug("Out of range except when parsing ipv6: ", oor.what());
                        return RELAY_ERROR;
                    } catch (const std::exception& e) {
                        LogDebug("Generic except when parsing ipv6: ", e.what());
                        return RELAY_ERROR;
                    }
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
                try {
                    address->port = (uint16_t)(atoi(&address_string[index + 1]));  // for same reason as above
                } catch (const std::invalid_argument& ia) {
                    LogDebug("Invalid argument except when parsing ipv4: ", ia.what());
                    return RELAY_ERROR;
                } catch (const std::out_of_range& oor) {
                    LogDebug("Out of range except when parsing ipv4: ", oor.what());
                    return RELAY_ERROR;
                } catch (const std::exception& e) {
                    LogDebug("Generic except when parsing ipv4: ", e.what());
                    return RELAY_ERROR;
                }
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

    const char* relay_address_to_string(const relay::relay_address_t* address, char* buffer)
    {
        assert(buffer);

        if (address->type == RELAY_ADDRESS_IPV6) {
#if defined(WINVER) && WINVER <= 0x0502
            // ipv6 not supported
            buffer[0] = '\0';
            return buffer;
#else
            uint16_t ipv6_network_order[8];
            for (int i = 0; i < 8; ++i)
                ipv6_network_order[i] = net::relay_htons(address->data.ipv6[i]);
            char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
            relay_platform_inet_ntop6(ipv6_network_order, address_string, sizeof(address_string));
            if (address->port == 0) {
                strncpy(buffer, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH);
                return buffer;
            } else {
                if (snprintf(buffer, RELAY_MAX_ADDRESS_STRING_LENGTH, "[%s]:%hu", address_string, address->port) < 0) {
                    relay_printf("address string truncated: [%s]:%hu", address_string, address->port);
                }
                return buffer;
            }
#endif
        } else if (address->type == RELAY_ADDRESS_IPV4) {
            if (address->port != 0) {
                snprintf(buffer,
                    RELAY_MAX_ADDRESS_STRING_LENGTH,
                    "%d.%d.%d.%d:%d",
                    address->data.ipv4[0],
                    address->data.ipv4[1],
                    address->data.ipv4[2],
                    address->data.ipv4[3],
                    address->port);
            } else {
                snprintf(buffer,
                    RELAY_MAX_ADDRESS_STRING_LENGTH,
                    "%d.%d.%d.%d",
                    address->data.ipv4[0],
                    address->data.ipv4[1],
                    address->data.ipv4[2],
                    address->data.ipv4[3]);
            }
            return buffer;
        } else {
            snprintf(buffer, RELAY_MAX_ADDRESS_STRING_LENGTH, "%s", "NONE");
            return buffer;
        }
    }

    int relay_address_equal(const relay::relay_address_t* a, const relay::relay_address_t* b)
    {
        assert(a);
        assert(b);

        if (a->type != b->type)
            return 0;

        if (a->type == RELAY_ADDRESS_IPV4) {
            if (a->port != b->port)
                return 0;

            for (int i = 0; i < 4; ++i) {
                if (a->data.ipv4[i] != b->data.ipv4[i])
                    return 0;
            }
        } else if (a->type == RELAY_ADDRESS_IPV6) {
            if (a->port != b->port)
                return 0;

            for (int i = 0; i < 8; ++i) {
                if (a->data.ipv6[i] != b->data.ipv6[i])
                    return 0;
            }
        }

        return 1;
    }
}  // namespace relay

namespace legacy
{}