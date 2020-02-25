#ifndef CORE_TOKEN_HPP
#define CORE_TOKEN_HPP

#include "packet.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

namespace core
{
  struct Token
  {
    // timestamp (8) +
    // session id (8) +
    // session version (1) +
    // session flags (1) =
    static const size_t ByteSize = 18;

    uint64_t ExpireTimestamp;
    uint64_t SessionID;
    uint8_t SessionVersion;
    uint8_t SessionFlags;

    uint64_t key();

   protected:
    void write(GenericPacket& buffer, size_t& index);
    void read(GenericPacket& buffer, size_t& index);
  };

  [[gnu::always_inline]] inline uint64_t Token::key()
  {
    return SessionID ^ SessionVersion;
  }

  [[gnu::always_inline]] inline void Token::write(GenericPacket& packet, size_t& index)
  {
    assert(index + Token::ByteSize < packet.size());
    encoding::WriteUint64(packet, index, ExpireTimestamp);
    encoding::WriteUint64(packet, index, SessionID);
    encoding::WriteUint8(packet, index, SessionVersion);
    encoding::WriteUint8(packet, index, SessionFlags);
  }

  [[gnu::always_inline]] inline void Token::read(GenericPacket& packet, size_t& index)
  {
    assert(index + Token::ByteSize < packet.size());
    ExpireTimestamp = encoding::ReadUint64(packet, index);
    SessionID = encoding::ReadUint64(packet, index);
    SessionVersion = encoding::ReadUint8(packet, index);
    SessionFlags = encoding::ReadUint8(packet, index);
  }
}  // namespace core
#endif
