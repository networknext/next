#ifndef CORE_TOKEN_HPP
#define CORE_TOKEN_HPP

#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "expireable.hpp"
#include "packet.hpp"

namespace core
{
  class Token: public Expireable
  {
   public:
    Token(const util::Clock& relayClock);
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
    void write(GenericPacket<>& packet, size_t& index);
    void read(GenericPacket<>& packet, size_t& index);
  };

  inline Token::Token(const util::Clock& relayClock): Expireable(relayClock) {}

  [[gnu::always_inline]] inline uint64_t Token::key()
  {
    return SessionID ^ SessionVersion;
  }

  [[gnu::always_inline]] inline void Token::write(GenericPacket<>& packet, size_t& index)
  {
    assert(index + Token::ByteSize < packet.Buffer.size());

    if (!encoding::WriteUint64(packet.Buffer, index, ExpireTimestamp)) {
      LogDebug("could not write expire timestamp");
    }

    if (!encoding::WriteUint64(packet.Buffer, index, SessionID)) {
      LogDebug("could not write session id");
    }

    if (!encoding::WriteUint8(packet.Buffer, index, SessionVersion)) {
      LogDebug("could not write session version");
    }

    if (!encoding::WriteUint8(packet.Buffer, index, SessionFlags)) {
      LogDebug("could not write session flags");
    }
  }

  [[gnu::always_inline]] inline void Token::read(GenericPacket<>& packet, size_t& index)
  {
    assert(index + Token::ByteSize < packet.Buffer.size());
    ExpireTimestamp = encoding::ReadUint64(packet.Buffer, index);
    SessionID = encoding::ReadUint64(packet.Buffer, index);
    SessionVersion = encoding::ReadUint8(packet.Buffer, index);
    SessionFlags = encoding::ReadUint8(packet.Buffer, index);
  }
}  // namespace core
#endif
