#ifndef CORE_TOKEN_HPP
#define CORE_TOKEN_HPP

#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "expireable.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "util/macros.hpp"
#include "core/session.hpp"

using core::Packet;

namespace core
{
  class TokenV4: public Expireable, public SessionHasher
  {
   public:
    TokenV4() = default;
    virtual ~TokenV4() override = default;

    static const size_t SIZE_OF = Expireable::SIZE_OF + SessionHasher::SIZE_OF;

   protected:
    auto write(Packet& packet, size_t& index) -> bool;
    auto read(const Packet& packet, size_t& index) -> bool;
  };

  INLINE auto TokenV4::write(Packet& packet, size_t& index) -> bool
  {
    if (index + TokenV4::SIZE_OF > packet.buffer.size()) {
      return false;
    }

    if (!encoding::write_uint64(packet.buffer, index, expire_timestamp)) {
      return false;
    }

    if (!encoding::write_uint64(packet.buffer, index, session_id)) {
      return false;
    }

    if (!encoding::write_uint8(packet.buffer, index, session_version)) {
      return false;
    }

    return true;
  }

  INLINE auto TokenV4::read(const Packet& packet, size_t& index) -> bool
  {
    if (index + TokenV4::SIZE_OF > packet.buffer.size()) {
      return false;
    }

    if (!encoding::read_uint64(packet.buffer, index, this->expire_timestamp)) {
      return false;
    }

    if (!encoding::read_uint64(packet.buffer, index, session_id)) {
      return false;
    }

    if (!encoding::read_uint8(packet.buffer, index, session_version)) {
      return false;
    }

    return true;
  }

  INLINE std::ostream& operator<<(std::ostream& os, const TokenV4& token)
  {
    return os << std::hex << token.session_id << '.' << std::dec << static_cast<unsigned int>(token.session_version);
  }

  class Token: public Expireable, public SessionHasher
  {
   public:
    Token() = default;
    virtual ~Token() override = default;
    // Expireable (8) +
    // SessionHasher (9) +
    // session flags (1) =
    static const size_t SIZE_OF = Expireable::SIZE_OF + SessionHasher::SIZE_OF + 1;

    uint8_t session_flags;

   protected:
    auto write(Packet& packet, size_t& index) -> bool;
    auto read(const Packet& packet, size_t& index) -> bool;
  };

  INLINE auto Token::write(Packet& packet, size_t& index) -> bool
  {
    if (index + Token::SIZE_OF > packet.buffer.size()) {
      return false;
    }

    if (!encoding::write_uint64(packet.buffer, index, expire_timestamp)) {
      return false;
    }

    if (!encoding::write_uint64(packet.buffer, index, session_id)) {
      return false;
    }

    if (!encoding::write_uint8(packet.buffer, index, session_version)) {
      return false;
    }

    if (!encoding::write_uint8(packet.buffer, index, session_flags)) {
      return false;
    }

    return true;
  }

  INLINE auto Token::read(const Packet& packet, size_t& index) -> bool
  {
    if (index + Token::SIZE_OF > packet.buffer.size()) {
      return false;
    }

    if (!encoding::read_uint64(packet.buffer, index, this->expire_timestamp)) {
      return false;
    }

    if (!encoding::read_uint64(packet.buffer, index, session_id)) {
      return false;
    }

    if (!encoding::read_uint8(packet.buffer, index, session_version)) {
      return false;
    }

    if (!encoding::read_uint8(packet.buffer, index, session_flags)) {
      return false;
    }

    return true;
  }

  INLINE std::ostream& operator<<(std::ostream& os, const Token& token)
  {
    return os << std::hex << token.session_id << '.' << std::dec << static_cast<unsigned int>(token.session_version);
  }
}  // namespace core
#endif
