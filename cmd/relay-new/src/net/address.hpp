#ifndef RELAY_RELAY_ADDRESS
#define RELAY_RELAY_ADDRESS

#include <array>
#include <string>
#include <cinttypes>

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
    ~Address() = default;

    bool parse(const std::string& address_string_in);

    void toString(std::string& buffer) const;
    std::string toString();  // slow, use only for debugging

    bool operator==(const Address& other) const;
    bool operator!=(const Address& other) const;

    void reset();

    AddressType Type;
    uint16_t Port;

    union
    {
      std::array<uint8_t, 4> IPv4;
      std::array<uint16_t, 8> IPv6;
    };
  };

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

  inline std::string Address::toString()
  {
    std::string buff;
    toString(buff);
    return buff;
  }

  inline bool Address::operator!=(const Address& other) const
  {
    return !(*this == other);
  }

  /* Helpers to reduce the amount of times static_cast has to be written */

  inline bool operator==(const AddressType at, uint8_t t)
  {
    return static_cast<uint8_t>(at) == t;
  }

  inline bool operator==(const uint8_t t, const AddressType at)
  {
    return static_cast<uint8_t>(at) == t;
  }

  inline bool operator!=(const AddressType at, uint8_t t)
  {
    return static_cast<uint8_t>(at) != t;
  }

  inline bool operator!=(const uint8_t t, const AddressType at)
  {
    return static_cast<uint8_t>(at) != t;
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