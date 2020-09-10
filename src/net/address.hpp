#pragma once

#include "util/logger.hpp"
#include "util/macros.hpp"

namespace net
{
  // Ethernet Frame | IP | Proto
  static const uint8_t IPv4UDPHeaderSize = 18 + 20 + 8;
  static const uint8_t IPv6UDPHeaderSize = 18 + 40 + 8;

  enum class AddressType : uint8_t
  {
    None,
    IPv4,
    IPv6
  };

  struct Address
  {
    // type (1) +
    // port (2) +
    // IP (16) =
    static const size_t SIZE_OF = 19;

    // max length when represented as a string
    static const size_t MAX_LEN = 256;

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
    auto to_string(std::string& buffer) const -> bool;

    // slow, use only for debugging or logging
    auto to_string() const -> std::string;

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

    AddressType type;
    uint16_t port;

    union
    {
      std::array<uint8_t, 4> ipv4;
      std::array<uint16_t, 8> ipv6;
    };
  };

  INLINE Address::Address(): type(AddressType::None), port(0)
  {
    this->ipv6.fill(0);
  }

  INLINE Address::Address(const Address& other)
  {
    *this = other;
  }

  INLINE Address::Address(Address&& other)
  {
    this->type = other.type;
    this->port = other.port;

    switch (other.type) {
      case AddressType::IPv4: {
        this->ipv4 = std::move(other.ipv4);
      } break;
      case AddressType::IPv6: {
        this->ipv6 = std::move(other.ipv6);
      } break;
      case AddressType::None:
        break;
    }
  }

  INLINE auto Address::parse(const std::string& address_in) -> bool
  {
    if (address_in.length() > MAX_LEN) {
      LOG(ERROR, "can not parse address: too long");
      return false;
    }

    std::array<char, Address::MAX_LEN> address = {};
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
            char* end = nullptr;
            this->port = (uint16_t)(std::strtol(&address[i + 1], &end, 10));
            if (end == nullptr) {
              return false;
            }
            // 1 char back will be a ']', so end the string there
            address[i - 1] = '\0';
            break;
          }

          if (address[i] == ']') {
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
        this->type = AddressType::IPv6;
        auto addr = sockaddr.sin6_addr.__in6_u.__u6_addr16;
        for (size_t i = 0; i < 8; i++) {
          this->ipv6[i] = ntohs(addr[i]);
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
          char* end = nullptr;
          this->port = (uint16_t)(std::strtol(&address[i + 1], &end, 10));
          if (end == nullptr) {
            return false;
          }
          address[i] = '\0';
          break;
        }
      }

      // 2. parse the beging of the ipv4 address via inet_pton

      // &address[index] now points to the start of the address and the null term replaced the ':'
      sockaddr_in sockaddr4;
      if (inet_pton(AF_INET, address.data(), &sockaddr4.sin_addr) == 1) {
        this->type = AddressType::IPv4;
        const auto& addr4 = sockaddr4.sin_addr.s_addr;
        for (int i = 0; i < 4; i++) {
          this->ipv4[i] = (uint8_t)((addr4 >> 8 * i) & 0xFF);
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
    std::swap(this->type, other.type);
    std::swap(this->port, other.port);

    if (this->type == AddressType::IPv4) {
      this->ipv4.swap(other.ipv4);
    } else if (this->type == AddressType::IPv6) {
      this->ipv6.swap(other.ipv6);
    }
  }

  INLINE void Address::reset()
  {
    GCC_NO_OPT_OUT;
    if (this->type == AddressType::IPv4) {
      this->ipv4.fill(0);
    } else if (this->type == AddressType::IPv6) {
      this->ipv6.fill(0);
    }

    this->type = AddressType::None;
    this->port = 0;
  }

  INLINE auto Address::to_string(std::string& output) const -> bool
  {
    std::array<char, Address::MAX_LEN> buff = {};
    unsigned int total = 0;

    if (this->type == AddressType::IPv6) {
      std::array<uint16_t, 8> ipv6_network_order;
      for (size_t i = 0; i < 8; i++) {
        ipv6_network_order[i] = htons(this->ipv6[i]);
      }

      std::array<char, Address::MAX_LEN> addr_buff = {};
      if (
       inet_ntop(
        AF_INET6,
        reinterpret_cast<void*>(ipv6_network_order.data()),
        addr_buff.data(),
        static_cast<socklen_t>(sizeof(addr_buff))) == nullptr) {
        LOG(ERROR, "unable to convert binary ip data to string");
        return false;
      }

      if (this->port == 0) {
        total += strlen(addr_buff.data());
        std::copy(addr_buff.begin(), addr_buff.begin() + total, buff.begin());
      } else {
        total += snprintf(buff.data(), MAX_LEN, "[%s]:%hu", addr_buff.data(), this->port);
      }
    } else if (this->type == AddressType::IPv4) {
      if (this->port == 0) {
        total += snprintf(buff.data(), MAX_LEN, "%d.%d.%d.%d", this->ipv4[0], this->ipv4[1], this->ipv4[2], this->ipv4[3]);
      } else {
        total += snprintf(
         buff.data(), MAX_LEN, "%d.%d.%d.%d:%hu", this->ipv4[0], this->ipv4[1], this->ipv4[2], this->ipv4[3], this->port);
      }
    } else {
      total += snprintf(buff.data(), sizeof("NONE"), "NONE");
    }

    // output.resize(total);
    output.assign(buff.begin(), buff.begin() + total);
    // std::copy(buff.begin(), buff.begin() + total, output.begin());

    return true;
  }

  INLINE auto Address::to_string() const -> std::string
  {
    std::string buff;
    this->to_string(buff);
    return buff;
  }

  // TODO cache this, likely these won't change
  template <>
  INLINE void Address::into(sockaddr_in& sin) const
  {
    sin = {};
    sin.sin_family = AF_INET;
    sin.sin_addr.s_addr = static_cast<uint32_t>(this->ipv4[0]) << 0 | static_cast<uint32_t>(this->ipv4[1]) << 8 |
                          static_cast<uint32_t>(this->ipv4[2]) << 16 | static_cast<uint32_t>(this->ipv4[3]) << 24;

    sin.sin_port = htons(this->port);
  }

  template <>
  INLINE void Address::into(sockaddr_in6& sin) const
  {
    sin = {};
    sin.sin6_family = AF_INET6;

    for (int i = 0; i < 8; i++) {
      reinterpret_cast<uint16_t*>(&sin.sin6_addr)[i] = htons(this->ipv6[i]);
    }

    sin.sin6_port = htons(this->port);
  }

  template <>
  INLINE void Address::into(mmsghdr& hdr) const
  {
    assert(hdr.msg_hdr.msg_name != nullptr);

    switch (this->type) {
      case AddressType::IPv4: {
        this->into(*reinterpret_cast<sockaddr_in*>(hdr.msg_hdr.msg_name));
        hdr.msg_hdr.msg_namelen = sizeof(sockaddr_in);
      } break;
      case AddressType::IPv6: {
        this->into(*reinterpret_cast<sockaddr_in6*>(hdr.msg_hdr.msg_name));
        hdr.msg_hdr.msg_namelen = sizeof(sockaddr_in6);
      } break;
      case AddressType::None: {
        // TODO log something?
      } break;
    }
  }

  INLINE auto Address::operator==(const Address& other) const -> bool
  {
    if (this->type != other.type || this->port != other.port) {
      return false;
    }

    switch (this->type) {
      case AddressType::IPv4:
        for (unsigned int i = 0; i < this->ipv4.size(); i++) {
          if (this->ipv4[i] != other.ipv4[i]) {
            return false;
          }
        }
        return true;
      case AddressType::IPv6:
        for (unsigned int i = 0; i < this->ipv6.size(); i++) {
          if (this->ipv6[i] != other.ipv6[i]) {
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
    this->type = other.type;
    this->port = other.port;

    if (this->type == AddressType::IPv4) {
      std::copy(other.ipv4.begin(), other.ipv4.end(), this->ipv4.begin());
    } else if (this->type == AddressType::IPv6) {
      std::copy(other.ipv6.begin(), other.ipv6.end(), this->ipv6.begin());
    }

    return *this;
  }

  INLINE auto Address::operator=(const Address&& other) -> Address&
  {
    this->type = other.type;
    this->port = other.port;

    switch (other.type) {
      case AddressType::IPv4: {
        this->ipv4 = other.ipv4;
      } break;
      case AddressType::IPv6: {
        this->ipv6 = other.ipv6;
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
    this->type = net::AddressType::IPv4;
    this->ipv4[0] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x000000FF));
    this->ipv4[1] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x0000FF00) >> 8);
    this->ipv4[2] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x00FF0000) >> 16);
    this->ipv4[3] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0xFF000000) >> 24);
    this->port = ntohs(addr.sin_port);
    return *this;
  }

  INLINE auto Address::operator=(const sockaddr_in6& addr) -> Address&
  {
    this->type = net::AddressType::IPv6;
    for (int i = 0; i < 8; i++) {
      this->ipv6[i] = ntohs(addr.sin6_addr.s6_addr16[i]);
    }
    this->port = ntohs(addr.sin6_port);
    return *this;
  }

  INLINE std::ostream& operator<<(std::ostream& os, const Address& addr)
  {
    std::string str;
    addr.to_string(str);
    return os << str;
  }
}  // namespace net
