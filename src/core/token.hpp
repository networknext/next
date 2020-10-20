#pragma once

#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "expireable.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "util/macros.hpp"
#include "core/session.hpp"

using core::Packet;

namespace testing
{
  class _test_core_Token_write_;
  class _test_core_Token_read_;
}  // namespace testing

namespace core
{
  class Token: public Expireable, public SessionHasher
  {
    friend testing::_test_core_Token_write_;
    friend testing::_test_core_Token_read_;

   public:
    Token() = default;
    virtual ~Token() override = default;

    static const size_t SIZE_OF = Expireable::SIZE_OF + SessionHasher::SIZE_OF;

    auto operator==(const Token& other) -> bool;

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

    return true;
  }

  INLINE auto Token::operator==(const Token& other) -> bool
  {
    return this->session_id == other.session_id && this->session_version == other.session_version &&
           this->expire_timestamp == other.expire_timestamp;
  }

  INLINE auto operator<<(std::ostream& os, const Token& token) -> std::ostream&
  {
    return os << std::hex << token.session_id << '.' << std::dec << static_cast<unsigned int>(token.session_version);
  }
}  // namespace core
