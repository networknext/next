#ifndef CORE_TOKEN_HPP
#define CORE_TOKEN_HPP

#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "expireable.hpp"
#include "packet.hpp"
#include "router_info.hpp"

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

    uint64_t key();

   protected:
    void write(uint8_t* packetData, size_t packetLength, size_t& index);
    void read(uint8_t* packetData, size_t packetLength, size_t& index);
  };

  inline Token::Token(const RouterInfo& routerInfo): Expireable(routerInfo) {}

  [[gnu::always_inline]] inline uint64_t Token::key()
  {
    return SessionID ^ SessionVersion;
  }

  [[gnu::always_inline]] inline void Token::write(uint8_t* packetData, size_t packetLength, size_t& index)
  {
    assert(index + Token::ByteSize < packetLength);

    if (!encoding::WriteUint64(packetData, packetLength, index, ExpireTimestamp)) {
      LogDebug("could not write expire timestamp");
      assert(false);
    }

    if (!encoding::WriteUint64(packetData, packetLength, index, SessionID)) {
      LogDebug("could not write session id");
      assert(false);
    }

    if (!encoding::WriteUint8(packetData, packetLength, index, SessionVersion)) {
      LogDebug("could not write session version");
      assert(false);
    }

    if (!encoding::WriteUint8(packetData, packetLength, index, SessionFlags)) {
      LogDebug("could not write session flags");
      assert(false);
    }
  }

  [[gnu::always_inline]] inline void Token::read(uint8_t* packetData, size_t packetLength, size_t& index)
  {
    (void)packetLength;
    assert(index + Token::ByteSize < packetLength);
    ExpireTimestamp = encoding::ReadUint64(packetData, index);
    SessionID = encoding::ReadUint64(packetData, index);
    SessionVersion = encoding::ReadUint8(packetData, index);
    SessionFlags = encoding::ReadUint8(packetData, index);
  }

  inline std::ostream& operator<<(std::ostream& os, const Token& token)
  {
    return os << std::hex << token.SessionID << '.' << std::dec << static_cast<unsigned int>(token.SessionVersion);
  }
}  // namespace core
#endif
