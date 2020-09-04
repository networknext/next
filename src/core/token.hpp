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
    static const size_t SIZE_OF = 18;

    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;

    uint64_t hash();

   protected:
    auto write(Packet& packet, size_t& index) -> bool;
    auto read(const Packet& packet, size_t& index) -> bool;
  };

  INLINE uint64_t Token::hash()
  {
    return session_id ^ session_version;
  }

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

  inline std::ostream& operator<<(std::ostream& os, const Token& token)
  {
    return os << std::hex << token.session_id << '.' << std::dec << static_cast<unsigned int>(token.session_version);
  }
}  // namespace core
#endif
