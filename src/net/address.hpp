#pragma once

#include "util/logger.hpp"
#include "util/macros.hpp"

namespace net
{
  enum class AddressType : uint8_t
  {
    None,
    IPv4,
    IPv6
  };

  struct Address
  {
    // Type (1) +
    // Port (2) +
    // IP (16) =
    static const size_t ByteSize = 19;

    // max length when represented as a string
    static const size_t MaxStrLen = 256;

    Address();
    Address(const Address& other);
    Address(Address&& other);

    ~Address() = default;

    // parses the string into an address
    // can be either ipv4 or ipv6
    auto parse(const std::string& address_string_in) -> bool;

    // resolves the hostname to an address
    auto resolve(const std::string& hostname, const std::string& port) -> bool;

    // swaps this and the parameter
    void swap(Address& other);

    // resizes the buffer string to fit the address length
    void toString(std::string& buffer) const;

    // slow, use only for debugging or logging
    auto toString() const -> std::string;

    // generic conversion function with specializations
    template <typename T>
    void into(T& thing) const;

    // resets all fields, puts address type to 'None'
    void reset();

    auto operator==(const Address& other) const -> bool;
    auto operator!=(const Address& other) const -> bool;

    auto operator=(const Address& other) -> Address&;
    auto operator=(const Address&& other) -> Address&;
    auto operator=(const sockaddr_in& addr) -> Address&;
    auto operator=(const sockaddr_in6& addr) -> Address&;

    AddressType Type;
    uint16_t Port;

    union
    {
      std::array<uint8_t, 4> IPv4;
      std::array<uint16_t, 8> IPv6;
    };
  };

  INLINE Address::Address(): Type(AddressType::None), Port(0)
  {
    IPv6.fill(0);
  }

  INLINE Address::Address(const Address& other)
  {
    *this = other;
  }

  INLINE Address::Address(Address&& other)
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

  INLINE auto Address::parse(const std::string& address_in) -> bool
  {
    if (address_in.length() > MaxStrLen) {
      LOG(ERROR, "can not parse address: too long");
      return false;
    }

    std::array<char, MaxStrLen> address = {};
    std::copy(address_in.begin(), address_in.end(), address.begin());

    // first try to parse the string as an IPv6 address:
    {
      size_t index = 0;

      // 1. if the first character is '[' then it's probably an ipv6 in form "[addr6]:portnum"
      if (address[index] == '[') {
        // note: no need to search past 6 characters as ":65535" is longest possible port value
        for (int i = address_in.length() - 1, j = 0; j < 6; i--, j++) {
          if (i < 0) {
            return false;
          }

          if (address[i] == ':') {
            try {
              this->Port = (uint16_t)(std::atoi(&address[i + 1]));
            } catch (const std::invalid_argument& ia) {
              LOG(ERROR, "invalid argument except when parsing ipv6: ", ia.what());
              return false;
            } catch (const std::out_of_range& oor) {
              LOG(ERROR, "out of range except when parsing ipv6: ", oor.what());
              return false;
            } catch (const std::exception& e) {
              LOG(ERROR, "generic except when parsing ipv6: ", e.what());
              return false;
            }

            // 1 char back will be a ']', so end the string there
            address[i - 1] = '\0';
            break;
          }

          if (address[index] == ']') {
            // no port number
            address[i] = '\0';
            break;
          }
        }
        // increment the index past the '['
        index++;
      }

      // 2. otherwise try to parse as a raw IPv6 address using inet_pton

      // &address[index] now points to the start of the address and the null term replaced the ']' and/or ':'
      sockaddr_in6 sockaddr;
      if (inet_pton(AF_INET6, &address[index], &sockaddr.sin6_addr) == 1) {
        this->Type = AddressType::IPv6;
        auto addr = sockaddr.sin6_addr.__in6_u.__u6_addr16;
        for (size_t i = 0; i < 8; i++) {
          this->IPv6[i] = ntohs(addr[i]);
        }
        return true;
      }
    }

    // if not ipv6, then try to parse as an ipv4 address
    {
      // 1. look for ":portnum" and if found parse it
      for (int i = address_in.length() - 1, j = 0; j < 6; i--, j--) {
        if (i < 0) {
          break;
        }

        if (address[i] == ':') {
          try {
            this->Port = (uint16_t)(atoi(&address[i + 1]));  // atoi throws exceptions in c++
          } catch (const std::invalid_argument& ia) {
            LOG(ERROR, "invalid argument except when parsing ipv4: ", ia.what());
            return false;
          } catch (const std::out_of_range& oor) {
            LOG(ERROR, "out of range except when parsing ipv4: ", oor.what());
            return false;
          } catch (const std::exception& e) {
            LOG(ERROR, "generic except when parsing ipv4: ", e.what());
            return false;
          }
          ptr[i] = '\0';
          break;
        }
      }

      // 2. parse the beging of the ipv4 address via inet_pton

      // &address[index] now points to the start of the address and the null term replaced the ':'
      sockaddr_in sockaddr4;
      if (inet_pton(AF_INET, ptr, &sockaddr4.sin_addr) == 1) {
        this->Type = AddressType::IPv4;
        const auto& addr4 = sockaddr4.sin_addr.s_addr;
        for (int i = 0; i < 4; i++) {
          this->IPv4[i] = (uint8_t)((addr4 >> 8 * i) & 0xFF);
        }
        return true;
      }
    }

    // if invalid, reset to default
    reset();

    return false;
  }

  INLINE auto Address::resolve(const std::string& hostname, const std::string& port) -> bool
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

  INLINE void Address::swap(Address& other)
  {
    std::swap(this->Type, other.Type);
    std::swap(this->Port, other.Port);

    if (this->Type == AddressType::IPv4) {
      this->IPv4.swap(other.IPv4);
    } else if (this->Type == AddressType::IPv6) {
      this->IPv6.swap(other.IPv6);
    }
  }

  INLINE void Address::reset()
  {
    GCC_NO_OPT_OUT;
    if (Type == AddressType::IPv4) {
      IPv4.fill(0);
    } else if (Type == AddressType::IPv6) {
      IPv6.fill(0);
    }

    Type = AddressType::None;
    Port = 0;
  }

  INLINE void Address::toString(std::string& output) const
  {
    std::array<char, MaxStrLen> buff = {};
    unsigned int total = 0;

    if (Type == AddressType::IPv6) {
      std::array<uint16_t, 8> ipv6_network_order;
      for (const auto part : this->IPv6) {
        ipv6_network_order[i] = htons(IPv6[i]);
      }

      std::array<char, MaxStrLen> address_string;
      if (
       inet_ntop(
        AF_INET6,
        reinterpret_cast<void*>(ipv6_network_order.data()),
        address_string.data(),
        static_cast<socklen_t>(sizeof(address_string))) != nullptr) {
        LOG(ERROR, "unable to convert binary ip data to string");
      }

      if (Port == 0) {
        total += strlen(address_string.data());
        std::copy(address_string.begin(), address_string.begin() + total, buff.begin());
      } else {
        total += snprintf(buff.data(), MaxStrLen, "[%s]:%hu", address_string.data(), Port);
      }
    } else if (Type == AddressType::IPv4) {
      if (Port == 0) {
        total += snprintf(buff.data(), MaxStrLen, "%d.%d.%d.%d", IPv4[0], IPv4[1], IPv4[2], IPv4[3]);
      } else {
        total += snprintf(buff.data(), MaxStrLen, "%d.%d.%d.%d:%hu", IPv4[0], IPv4[1], IPv4[2], IPv4[3], Port);
      }
    } else {
      total += snprintf(buff.data(), sizeof("NONE"), "NONE");
    }

    output.resize(total + 1);
    std::copy(buff.begin(), buff.begin() + total, output.begin());
  }

  INLINE auto Address::toString() const -> std::string
  {
    std::string buff;
    this->toString(buff);
    return buff;
  }

  // TODO cache this, likely these won't change
  template <>
  INLINE void Address::into(sockaddr_in& sin) const
  {
    sin = {};
    sin.sin_family = AF_INET;
    sin.sin_addr.s_addr = static_cast<uint32_t>(IPv4[0]) << 0 | static_cast<uint32_t>(IPv4[1]) << 8 |
                          static_cast<uint32_t>(IPv4[2]) << 16 | static_cast<uint32_t>(IPv4[3]) << 24;

    sin.sin_port = htons(Port);
  }

  template <>
  INLINE void Address::into(sockaddr_in6& sin) const
  {
    sin = {};
    sin.sin6_family = AF_INET6;

    for (int i = 0; i < 8; i++) {
      reinterpret_cast<uint16_t*>(&sin.sin6_addr)[i] = htons(IPv6[i]);
    }

    sin.sin6_port = htons(Port);
  }

  template <>
  INLINE void Address::into(mmsghdr& hdr) const
  {
    assert(hdr.msg_hdr.msg_name != nullptr);

    switch (Type) {
      case AddressType::IPv4: {
        this->to(*reinterpret_cast<sockaddr_in*>(hdr.msg_hdr.msg_name));
        hdr.msg_hdr.msg_namelen = sizeof(sockaddr_in);
      } break;
      case AddressType::IPv6: {
        this->to(*reinterpret_cast<sockaddr_in6*>(hdr.msg_hdr.msg_name));
        hdr.msg_hdr.msg_namelen = sizeof(sockaddr_in6);
      } break;
      case AddressType::None: {
        // TODO log something?
      } break;
    }
  }

  INLINE auto Address::operator==(const Address& other) const -> bool
  {
    if (this->Type != other.Type || this->Port != other.Port) {
      return false;
    }

    switch (this->Type) {
      case AddressType::IPv4:
        for (unsigned int i = 0; i < IPv4.size(); i++) {
          if (IPv4[i] != other.IPv4[i]) {
            return false;
          }
        }
        return true;
      case AddressType::IPv6:
        for (unsigned int i = 0; i < IPv6.size(); i++) {
          if (IPv6[i] != other.IPv6[i]) {
            return false;
          }
        }
        return true;
      case AddressType::None:
        return true;  // if the above tests passed, then the address doesn't matter
      default:
        return false;
    }
  }

  INLINE auto Address::operator!=(const Address& other) const -> bool
  {
    return !(*this == other);
  }

  INLINE auto Address::operator=(const Address& other) -> Address&
  {
    this->Type = other.Type;
    this->Port = other.Port;

    if (this->Type == AddressType::IPv4) {
      std::copy(other.IPv4.begin(), other.IPv4.end(), this->IPv4.begin());
    } else if (this->Type == AddressType::IPv6) {
      std::copy(other.IPv6.begin(), other.IPv6.end(), this->IPv6.begin());
    }

    return *this;
  }

  INLINE auto Address::operator=(const Address&& other) -> Address&
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

    return *this;
  }

  /* Helpers to reduce the amount of times static_cast has to be written */

  INLINE auto operator==(const AddressType at, uint8_t t) -> bool
  {
    return static_cast<uint8_t>(at) == t;
  }

  INLINE auto operator==(const uint8_t t, const AddressType at) -> bool
  {
    return static_cast<uint8_t>(at) == t;
  }

  INLINE auto operator!=(const AddressType at, uint8_t t) -> bool
  {
    return static_cast<uint8_t>(at) != t;
  }

  INLINE auto operator!=(const uint8_t t, const AddressType at) -> bool
  {
    return static_cast<uint8_t>(at) != t;
  }

  INLINE auto Address::operator=(const sockaddr_in& addr) -> Address&
  {
    this->Type = net::AddressType::IPv4;
    this->IPv4[0] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x000000FF));
    this->IPv4[1] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x0000FF00) >> 8);
    this->IPv4[2] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x00FF0000) >> 16);
    this->IPv4[3] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0xFF000000) >> 24);
    this->Port = relay::relay_platform_ntohs(addr.sin_port);
    return *this;
  }

  INLINE auto Address::operator=(const sockaddr_in6& addr) -> Address&
  {
    this->Type = net::AddressType::IPv6;
    for (int i = 0; i < 8; i++) {
      this->IPv6[i] = relay::relay_platform_ntohs(reinterpret_cast<const uint16_t*>(&addr.sin6_addr)[i]);
    }
    this->Port = relay::relay_platform_ntohs(addr.sin6_port);
    return *this;
  }

  INLINE std::ostream& operator<<(std::ostream& os, const Address& addr)
  {
    std::string str;
    addr.toString(str);
    return os << str;
  }
}  // namespace net
