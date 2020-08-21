#pragma once

#include "core/packet.hpp"
#include "crypto/keychain.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "types.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

using core::GenericPacketContainer;
using core::packets::Type;
using crypto::GenericKey;

namespace core
{
  namespace packets
  {
    enum class Direction
    {
      ClientToServer,
      ServerToClient,
    };

    struct Header
    {
      static const size_t ByteSize = 35;

      Direction direction;
      Type type;

      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;

      auto read(GenericPacketContainer<>& buffer, size_t& index) -> bool;
      auto write(GenericPacketContainer<>& buffer, size_t& index, const GenericKey& public_key) -> bool;
      auto verify(GenericPacketContainer<>& buffer, size_t& index, const GenericKey& public_key) -> bool;

      auto Header::clean_sequence() -> uint64_t;
    };

    INLINE auto Header::read(GenericPacketContainer<>& buffer, size_t& index) -> bool
    {
      if (index + ByteSize > buffer.size()) {
        LOG(ERROR, "could not read header, buffer is too small");
        return false;
      }

      Type packet_type = static_cast<Type>(encoding::ReadUint8(buffer, index));

      uint64_t packet_sequence = encoding::ReadUint64(buffer, index);

      if (direction == Direction::ServerToClient) {
        // high bit must be set
        if (!(packet_sequence & (1ULL << 63)))
          return false;
      } else {
        // high bit must be clear
        if (packet_sequence & (1ULL << 63))
          return false;
      }

      this->type = packet_type;

      if (
       this->type == Type::SessionPing || this->type == Type::SessionPong || this->type == Type::RouteResponse ||
       this->type == Type::ContinueResponse) {
        // second highest bit must be set
        assert(packet_sequence & (1ULL << 62));
      } else {
        // second highest bit must be clear
        assert((packet_sequence & (1ULL << 62)) == 0);
      }

      this->sequence = packet_sequence;
      this->session_id = encoding::ReadUint64(buffer, index);
      this->session_version = encoding::ReadUint8(buffer, index);

      return true;
    }

    INLINE auto Header::write(GenericPacketContainer<>& buffer, size_t& index, const GenericKey& private_key) -> bool
    {
      if (index + ByteSize > buffer.size()) {
        LOG(ERROR, "could not write header, buffer is too small");
        return false;
      }

      if (this->direction == Direction::ServerToClient) {
        // high bit must be set
        assert(this->sequence & (1ULL << 63));
      } else {
        // high bit must be clear
        assert((this->sequence & (1ULL << 63)) == 0);
      }

      if (
       this->type == Type::SessionPing || this->type == Type::SessionPong || this->type == Type::RouteResponse ||
       this->type == Type::ContinueResponse) {
        // second highest bit must be set
        assert(sequence & (1ULL << 62));
      } else {
        // second highest bit must be clear
        assert((sequence & (1ULL << 62)) == 0);
      }

      if (!encoding::WriteUint8(buffer, index, static_cast<uint8_t>(this->type))) {
        return false;
      }

      if (!encoding::WriteUint64(buffer, index, this->sequence)) {
        return false;
      }

      uint8_t* additional = &buffer[index];
      const int additional_length = 8 + 1 + 1;

      if (!encoding::WriteUint64(buffer, index, this->session_id)) {
        return false;
      }

      if (!encoding::WriteUint8(buffer, index, this->session_version)) {
        return false;
      }

      // todo: remove this once we fully switch to new relay
      // todo: still applicable?
      if (!encoding::WriteUint8(buffer, index, 0)) {
        return false;
      }

      std::array<uint8_t, 12> nonce;
      {
        size_t index = 0;
        if (!encoding::WriteUint32(nonce, index, 0)) {
          return false;
        }

        if (!encoding::WriteUint64(nonce, index, this->sequence)) {
          return false;
        }
      }

      unsigned long long encrypted_length = 0;

      int result = crypto_aead_chacha20poly1305_ietf_encrypt(
       &buffer[index],
       &encrypted_length,
       &buffer[index],
       0,
       additional,
       (unsigned long long)additional_length,
       NULL,
       nonce.data(),
       private_key.data());

      if (result != 0) {
        return false;
      }

      index += encrypted_length;

      return true;
    }

    INLINE auto Header::verify(GenericPacketContainer<>& buffer, size_t& index, const GenericKey& private_key) -> bool
    {
      if (index + ByteSize > buffer.size()) {
        LOG(ERROR, "could not verify header, buffer is too small");
        return false;
      }

      Type packet_type = static_cast<Type>(encoding::ReadUint8(buffer, index));

      uint64_t packet_sequence = encoding::ReadUint64(buffer, index);

      if (this->direction == Direction::ServerToClient) {
        // high bit must be set
        if ((packet_sequence & (1ULL << 63)) == 0) {
          return false;
        }
      } else {
        // high bit must be clear
        if (packet_sequence & (1ULL << 63) != 0) {
          return false;
        }
      }

      if (
       packet_type == Type::SessionPing || packet_type == Type::SessionPong || packet_type == Type::RouteResponse ||
       packet_type == Type::ContinueResponse) {
        // second highest bit must be set
        assert(packet_sequence & (1ULL << 62));
      } else {
        // second highest bit must be clear
        assert((packet_sequence & (1ULL << 62)) == 0);
      }

      const uint8_t* additional = &buffer[index];
      const int additional_length = 8 + 1 + 1;

      index += 12;

      std::array<uint8_t, 12> nonce;
      {
        size_t index = 0;
        encoding::WriteUint32(nonce, index, 0);
        encoding::WriteUint64(nonce, index, packet_sequence);
      }

      unsigned long long decrypted_length;

      int result = crypto_aead_chacha20poly1305_ietf_decrypt(
       &buffer[index + 19],
       &decrypted_length,
       nullptr,
       &buffer[index + 19],
       (unsigned long long)crypto_aead_chacha20poly1305_IETF_ABYTES,
       additional,
       (unsigned long long)additional_length,
       nonce.data(),
       private_key.data());

      if (result != 0) {
        return false;
      }

      return true;
    }

    INLINE auto Header::clean_sequence() -> uint64_t
    {
      static const uint64_t mask = ~((1ULL << 63) | (1ULL << 62));
      return this->sequence & mask;
    }

    INLINE std::ostream& operator<<(std::ostream& stream, const Header& header)
    {
      stream << std::hex << header.session_id << '.' << std::dec << static_cast<uint32_t>(header.session_version);
      return stream;
    }
  }  // namespace packets
}  // namespace core