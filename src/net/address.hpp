#ifndef RELAY_RELAY_ADDRESS
#define RELAY_RELAY_ADDRESS

#include "relay/relay_platform.hpp"

#include "util/logger.hpp"

namespace net
{
  const uint8_t IPv4UDPHeaderSize = 28;
  const uint8_t IPv6UDPHeaderSize = 48;

  enum class AddressType : uint8_t
  {
    None,
    IPv4,
    IPv6
  };

  class Address
  {
   public:
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

    bool parse(const std::string& address_string_in);
    bool resolve(const std::string& hostname, const std::string& port);

    void swap(Address& other);

    void toString(std::string& buffer) const;
    auto toString() const -> std::string;  // slow, use only for debugging or logging

    // generic conversion function with specializations, looks better this way
    template <typename T>
    void to(T& thing) const;

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

  [[gnu::always_inline]] inline Address::Address(const Address& other)
  {
    *this = other;
  }

  [[gnu::always_inline]] inline void Address::swap(Address& other)
  {
    std::swap(this->Type, other.Type);
    std::swap(this->Port, other.Port);

    if (this->Type == AddressType::IPv4) {
      this->IPv4.swap(other.IPv4);
    } else if (this->Type == AddressType::IPv6) {
      this->IPv6.swap(other.IPv6);
    }
  }

  inline void Address::reset()
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

  inline auto Address::toString() const -> std::string
  {
    std::string buff;
    toString(buff);
    return buff;
  }

  // TODO cache this, likely these won't change
  template <>
  inline void Address::to(sockaddr_in& sin) const
  {
    sin = {};
    sin.sin_family = AF_INET;
    sin.sin_addr.s_addr =
     (((uint32_t)IPv4[0])) | (((uint32_t)IPv4[1]) << 8) | (((uint32_t)IPv4[2]) << 16) | (((uint32_t)IPv4[3]) << 24);

    sin.sin_port = htons(Port);
  }

  template <>
  inline void Address::to(sockaddr_in6& sin) const
  {
    sin = {};
    sin.sin6_family = AF_INET6;

    for (int i = 0; i < 8; i++) {
      reinterpret_cast<uint16_t*>(&sin.sin6_addr)[i] = htons(IPv6[i]);
    }

    sin.sin6_port = htons(Port);
  }

  template <>
  inline void Address::to(mmsghdr& hdr) const
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

  inline auto Address::operator==(const Address& other) const -> bool
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

  inline auto Address::operator!=(const Address& other) const -> bool
  {
    return !(*this == other);
  }

  [[gnu::always_inline]] inline auto Address::operator=(const Address& other) -> Address&
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

  [[gnu::always_inline]] inline auto Address::operator=(const Address&& other) -> Address&
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

  inline auto operator==(const AddressType at, uint8_t t) -> bool
  {
    return static_cast<uint8_t>(at) == t;
  }

  inline auto operator==(const uint8_t t, const AddressType at) -> bool
  {
    return static_cast<uint8_t>(at) == t;
  }

  inline auto operator!=(const AddressType at, uint8_t t) -> bool
  {
    return static_cast<uint8_t>(at) != t;
  }

  inline auto operator!=(const uint8_t t, const AddressType at) -> bool
  {
    return static_cast<uint8_t>(at) != t;
  }

  inline auto Address::operator=(const sockaddr_in& addr) -> Address&
  {
    this->Type = net::AddressType::IPv4;
    this->IPv4[0] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x000000FF));
    this->IPv4[1] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x0000FF00) >> 8);
    this->IPv4[2] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0x00FF0000) >> 16);
    this->IPv4[3] = static_cast<uint8_t>((addr.sin_addr.s_addr & 0xFF000000) >> 24);
    this->Port = relay::relay_platform_ntohs(addr.sin_port);
    return *this;
  }

  inline auto Address::operator=(const sockaddr_in6& addr) -> Address&
  {
    this->Type = net::AddressType::IPv6;
    for (int i = 0; i < 8; i++) {
      this->IPv6[i] = relay::relay_platform_ntohs(reinterpret_cast<const uint16_t*>(&addr.sin6_addr)[i]);
    }
    this->Port = relay::relay_platform_ntohs(addr.sin6_port);
    return *this;
  }

  inline std::ostream& operator<<(std::ostream& os, const Address& addr)
  {
    std::string str;
    addr.toString(str);
    return os << str;
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

  inline std::ostream& operator<<(std::ostream& os, const relay_address_t& addr)
  {
    char buff[128];
    memset(buff, 0, sizeof(buff));
    relay_address_to_string(&addr, buff);
    return os << buff;
  }
}  // namespace legacy
#endif
