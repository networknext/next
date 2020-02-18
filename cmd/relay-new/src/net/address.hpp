#ifndef RELAY_RELAY_ADDRESS
#define RELAY_RELAY_ADDRESS

#include "config.hpp"

namespace net
{
  enum class AddressType : uint8_t
  {
    None,
    IPv4,
    IPv6
  };

  class Address
  {
   public:
    Address();
    Address(const Address& other);
    ~Address() = default;

    bool parse(const std::string& address_string_in);

    void toString(std::string& buffer) const;
    auto toString() -> std::string;  // slow, use only for debugging

    void reset();

    auto operator==(const Address& other) const -> bool;
    auto operator!=(const Address& other) const -> bool;

    auto operator=(const Address& other) -> Address&;

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

  inline auto Address::toString() -> std::string
  {
    std::string buff;
    toString(buff);
    return buff;
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
}  // namespace legacy
#endif
