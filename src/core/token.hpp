#ifndef CORE_TOKEN_HPP
#define CORE_TOKEN_HPP

#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "expireable.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "util/macros.hpp"

using core::Packet;

namespace core
{
  class Token: public Expireable
  {
   public:
    Token() = default;
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
    auto write(Packet& packet, size_t& index) -> bool;
    auto read(const Packet& packet, size_t& index) -> bool;
  };

  INLINE uint64_t Token::hash()
  {
    return SessionID ^ SessionVersion;
  }

  INLINE auto Token::write(Packet& packet, size_t& index) -> bool
  {
    if (index + Token::ByteSize > packet.Buffer.size()) {
      return false;
    }

    if (!encoding::WriteUint64(packet.Buffer, index, ExpireTimestamp)) {
      return false;
    }

    if (!encoding::WriteUint64(packet.Buffer, index, SessionID)) {
      return false;
    }

    if (!encoding::WriteUint8(packet.Buffer, index, SessionVersion)) {
      return false;
    }

    if (!encoding::WriteUint8(packet.Buffer, index, SessionFlags)) {
      return false;
    }

    return true;
  }

  INLINE auto Token::read(const Packet& packet, size_t& index) -> bool
  {
    if (index + Token::ByteSize > packet.Buffer.size()) {
      return false;
    }

    if (!encoding::ReadUint64(packet.Buffer, index, this->ExpireTimestamp)) {
      return false;
    }

    if (!encoding::ReadUint64(packet.Buffer, index, SessionID)) {
      return false;
    }

    if (!encoding::ReadUint8(packet.Buffer, index, SessionVersion)) {
      return false;
    }

    if (!encoding::ReadUint8(packet.Buffer, index, SessionFlags)) {
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
