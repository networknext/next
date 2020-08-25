#ifndef CORE_TOKEN_HPP
#define CORE_TOKEN_HPP

#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "expireable.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "util/macros.hpp"

namespace core
{
  class Token: public Expireable
  {
   public:
    Token(const RouterInfo& routerInfo);
    virtual ~Token() override = default;
    // Expireable (8) +
    // session id (8) +
    // session version (1) +
    // session flags (1) =
    static const size_t ByteSize = 18;

    uint64_t SessionID;
    uint8_t SessionVersion;
    uint8_t SessionFlags;

    uint64_t hash();

   protected:
    auto write(const uint8_t* packetData, size_t packetLength, size_t& index) -> bool;
    auto read(const uint8_t* packetData, size_t packetLength, size_t& index) -> bool;
  };

  INLINE Token::Token(const RouterInfo& routerInfo): Expireable(routerInfo) {}

  INLINE uint64_t Token::hash()
  {
    return SessionID ^ SessionVersion;
  }

  INLINE auto Token::write(const uint8_t* packetData, size_t packetLength, size_t& index) -> bool
  {
    if (packetLength < index + Token::ByteSize) {
      return false;
    }
    if (!encoding::WriteUint64(packetData, packetLength, index, ExpireTimestamp)) {
      return false;
    }
    if (!encoding::WriteUint64(packetData, packetLength, index, SessionID)) {
      return false;
    }
    if (!encoding::WriteUint8(packetData, packetLength, index, SessionVersion)) {
      return false;
    }
    if (!encoding::WriteUint8(packetData, packetLength, index, SessionFlags)) {
      return false;
    }
  }

  INLINE auto Token::read(const uint8_t* packetData, size_t packetLength, size_t& index) -> bool
  {
    if (packetLength < index + Token::ByteSize) {
      return false;
    }
    if (!encoding::ReadUint64(packetData, packetLength, index, this->ExpireTimestamp)) {
      return false;
    }
    if (!encoding::ReadUint64(packetData, packetLength, index, SessionID)) {
      return false;
    }
    if (!encoding::ReadUint8(packetData, packetLength, index, SessionVersion)) {
      return false;
    }
    if (!encoding::ReadUint8(packetData, packetLength, index, SessionFlags)) {
      return false;
    }

    return true;
  }

  inline std::ostream& operator<<(std::ostream& os, const Token& token)
  {
    return os << std::hex << token.SessionID << '.' << std::dec << static_cast<unsigned int>(token.SessionVersion);
  }
}  // namespace core
#endif
