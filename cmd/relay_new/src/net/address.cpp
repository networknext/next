#include "includes.h"
#include "address.hpp"

#include "util/logger.hpp"

#include "net.hpp"

namespace net
{
  Address::Address(): Type(AddressType::None), Port(0)
  {
    IPv6.fill(0);
  }

  Address::Address(Address&& other)
  {
    this->Type = other.Type;
    this->Port = other.Port;

    switch (other.Type) {
      case AddressType::IPv4: {
        this->IPv4 = std::move(other.IPv4);
      } break;
      case AddressType::IPv6: {
        this->IPv6 = std::move(other.IPv6);
      } break;
      case AddressType::None:
        break;
    }
  }

  auto Address::parse(const std::string& address) -> bool
  {
    reset();
#ifdef __GNUC__
    __builtin_prefetch(address.c_str());
#endif
    // first try to parse the string as an IPv6 address:

    std::array<char, RELAY_MAX_ADDRESS_STRING_LENGTH + RELAY_ADDRESS_BUFFER_SAFETY * 2> buff;
    buff.fill(0);
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
            this->Port = (uint16_t)(atoi(&ptr[index + 1]));  // atoi throws exceptions in c++
          } catch (const std::invalid_argument& ia) {
            Log("Invalid argument except when parsing ipv6: ", ia.what());
            std::cout << std::flush;
            return false;
          } catch (const std::out_of_range& oor) {
            Log("Out of range except when parsing ipv6: ", oor.what());
            std::cout << std::flush;
            return false;
          } catch (const std::exception& e) {
            Log("Generic except when parsing ipv6: ", e.what());
            std::cout << std::flush;
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

    std::array<uint16_t, 8> addr6;
    if (relay::relay_platform_inet_pton6(ptr, addr6.data()) == RELAY_OK) {
      this->Type = AddressType::IPv6;
      for (int i = 0; i < 8; ++i) {
        this->IPv6[i] = relay::relay_platform_ntohs(addr6[i]);
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
          this->Port = (uint16_t)(atoi(&ptr[index + 1]));  // atoi throws exceptions in c++
        } catch (const std::invalid_argument& ia) {
          Log("Invalid argument except when parsing ipv4: ", ia.what());
          std::cout << std::flush;
          return false;
        } catch (const std::out_of_range& oor) {
          Log("Out of range except when parsing ipv4: ", oor.what());
          std::cout << std::flush;
          return false;
        } catch (const std::exception& e) {
          Log("Generic except when parsing ipv4: ", e.what());
          std::cout << std::flush;
          return false;
        }
        ptr[index] = '\0';
        break;
      }
    }

    // 2. parse remaining ipv4 address via inet_pton

    uint32_t addr4;
    if (relay::relay_platform_inet_pton4(ptr, &addr4) == RELAY_OK) {
      this->Type = AddressType::IPv4;
      this->IPv4[3] = (uint8_t)((addr4 & 0xFF000000) >> 24);
      this->IPv4[2] = (uint8_t)((addr4 & 0x00FF0000) >> 16);
      this->IPv4[1] = (uint8_t)((addr4 & 0x0000FF00) >> 8);
      this->IPv4[0] = (uint8_t)((addr4 & 0x000000FF));
      return true;
    }

    reset();  // if invalid, reset to 0
    return false;
  }

  auto Address::resolve(const std::string& hostname, const std::string& port) -> bool
  {
    bool success = false;
    addrinfo hints = {};
    addrinfo* result = nullptr;

    if (getaddrinfo(hostname.c_str(), port.c_str(), &hints, &result) == 0 && result != nullptr) {
      if (result->ai_addr->sa_family == AF_INET6) {
        sockaddr_in6* addr_ipv6 = (sockaddr_in6*)(result->ai_addr);
        *this = *addr_ipv6;
        success = true;
      } else if (result->ai_addr->sa_family == AF_INET) {
        sockaddr_in* addr_ipv4 = (sockaddr_in*)(result->ai_addr);
        *this = *addr_ipv4;
        success = true;
      }

      freeaddrinfo(result);
    }

    return success;
  }

  // TODO consider making this a bool retval. Since some windows versions can't do ipv6 then that would be the only case it
  // returns false
  void Address::toString(std::string& output) const
  {
    std::array<char, RELAY_MAX_ADDRESS_STRING_LENGTH> buff;
    unsigned int total = 0;

    if (Type == AddressType::IPv6) {
#if defined(WINVER) && WINVER <= 0x0502
      // ipv6 not supported
      return;
#else
      std::array<uint16_t, 8> ipv6_network_order;
      for (int i = 0; i < 8; ++i) {
        ipv6_network_order[i] = net::relay_htons(IPv6[i]);
      }

      std::array<char, RELAY_MAX_ADDRESS_STRING_LENGTH> address_string;
      relay::relay_platform_inet_ntop6(ipv6_network_order.data(), address_string.data(), address_string.size() * sizeof(char));
      if (Port == 0) {
        std::copy(address_string.begin(), address_string.end(), buff.begin());
        total += strlen(address_string.data());
      } else {
        total += snprintf(&buff[total], RELAY_MAX_ADDRESS_STRING_LENGTH - total, "[%s]:%hu", address_string.data(), Port);
      }
#endif
    } else if (Type == AddressType::IPv4) {
      if (Port == 0) {
        total += snprintf(buff.data(), RELAY_MAX_ADDRESS_STRING_LENGTH, "%d.%d.%d.%d", IPv4[0], IPv4[1], IPv4[2], IPv4[3]);
      } else {
        total += snprintf(
         &buff[total], RELAY_MAX_ADDRESS_STRING_LENGTH - total, "%d.%d.%d.%d:%hu", IPv4[0], IPv4[1], IPv4[2], IPv4[3], Port);
      }
    } else {
      total += snprintf(buff.data(), sizeof("NONE"), "NONE");
    }

    // method 1 - 1st fastest

    // resize because std::copy doesn't do that for strings
    output.resize(total);

    // can't use end() because end() doesn't point to the end of the string
    std::copy(buff.begin(), buff.begin() + total, output.begin());

    // method 2 - slow
    // output = std::move(std::string(buff.begin(), buff.begin() + total));

    // method 3 - almost tied with method 1
    // output.assign(buff.begin(), buff.begin() + total);
  }
}  // namespace net

namespace legacy
{
  int relay_address_parse(relay_address_t* address, const char* address_string_in)
  {
    assert(address);
    assert(address_string_in);

    if (!address)
      return RELAY_ERROR;

    if (!address_string_in)
      return RELAY_ERROR;

    memset(address, 0, sizeof(relay_address_t));

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
            Log("Invalid argument except when parsing ipv6: ", ia.what());
            return RELAY_ERROR;
          } catch (const std::out_of_range& oor) {
            Log("Out of range except when parsing ipv6: ", oor.what());
            return RELAY_ERROR;
          } catch (const std::exception& e) {
            Log("Generic except when parsing ipv6: ", e.what());
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
    if (relay::relay_platform_inet_pton6(address_string, addr6) == RELAY_OK) {
      address->type = static_cast<uint8_t>(net::AddressType::IPv6);
      for (int i = 0; i < 8; ++i) {
        address->data.ipv6[i] = relay::relay_platform_ntohs(addr6[i]);
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
          Log("Invalid argument except when parsing ipv4: ", ia.what());
          return RELAY_ERROR;
        } catch (const std::out_of_range& oor) {
          Log("Out of range except when parsing ipv4: ", oor.what());
          return RELAY_ERROR;
        } catch (const std::exception& e) {
          Log("Generic except when parsing ipv4: ", e.what());
          return RELAY_ERROR;
        }
        address_string[index] = '\0';
      }
    }

    uint32_t addr4;
    if (relay::relay_platform_inet_pton4(address_string, &addr4) == RELAY_OK) {
      address->type = static_cast<uint8_t>(net::AddressType::IPv4);
      address->data.ipv4[3] = (uint8_t)((addr4 & 0xFF000000) >> 24);
      address->data.ipv4[2] = (uint8_t)((addr4 & 0x00FF0000) >> 16);
      address->data.ipv4[1] = (uint8_t)((addr4 & 0x0000FF00) >> 8);
      address->data.ipv4[0] = (uint8_t)((addr4 & 0x000000FF));
      return RELAY_OK;
    }

    return RELAY_ERROR;
  }

  const char* relay_address_to_string(const relay_address_t* address, char* buffer)
  {
    assert(buffer);

    if (address->type == net::AddressType::IPv6) {
#if defined(WINVER) && WINVER <= 0x0502
      // ipv6 not supported
      buffer[0] = '\0';
      return buffer;
#else
      uint16_t ipv6_network_order[8];
      for (int i = 0; i < 8; ++i)
        ipv6_network_order[i] = net::relay_htons(address->data.ipv6[i]);
      char address_string[RELAY_MAX_ADDRESS_STRING_LENGTH];
      relay::relay_platform_inet_ntop6(ipv6_network_order, address_string, sizeof(address_string));
      if (address->port == 0) {
        strncpy(buffer, address_string, RELAY_MAX_ADDRESS_STRING_LENGTH);
        return buffer;
      } else {
        if (snprintf(buffer, RELAY_MAX_ADDRESS_STRING_LENGTH, "[%s]:%hu", address_string, address->port) < 0) {
          Log("address string truncated: [", address_string, "]:", static_cast<uint32_t>(address->port));
        }
        return buffer;
      }
#endif
    } else if (address->type == net::AddressType::IPv4) {
      if (address->port != 0) {
        snprintf(
         buffer,
         RELAY_MAX_ADDRESS_STRING_LENGTH,
         "%d.%d.%d.%d:%d",
         address->data.ipv4[0],
         address->data.ipv4[1],
         address->data.ipv4[2],
         address->data.ipv4[3],
         address->port);
      } else {
        snprintf(
         buffer,
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

  int relay_address_equal(const relay_address_t* a, const relay_address_t* b)
  {
    assert(a);
    assert(b);

    if (a->type != b->type)
      return 0;

    if (a->type == net::AddressType::IPv4) {
      if (a->port != b->port)
        return 0;

      for (int i = 0; i < 4; ++i) {
        if (a->data.ipv4[i] != b->data.ipv4[i])
          return 0;
      }
    } else if (a->type == net::AddressType::IPv6) {
      if (a->port != b->port)
        return 0;

      for (int i = 0; i < 8; ++i) {
        if (a->data.ipv6[i] != b->data.ipv6[i])
          return 0;
      }
    }

    return 1;
  }
}  // namespace legacy
