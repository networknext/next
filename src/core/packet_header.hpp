#pragma once

#include "core/session.hpp"
#include "crypto/keychain.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "packet.hpp"
#include "packet_types.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

using crypto::GenericKey;

namespace core
{
  enum class PacketDirection
  {
    ClientToServer,
    ServerToClient,
  };

  struct PacketHeaderV4: public SessionHasher
  {
    // type (1) +
    // SessionHasher (9) +
    // sequence (8)
    static const size_t SIZE_OF = 18;
    static const size_t SIZE_OF_SIGNED = SIZE_OF + crypto_aead_chacha20poly1305_IETF_ABYTES;

    PacketType type;

    uint64_t sequence;

    auto read(Packet& packet, size_t& index, PacketDirection direction) -> bool;
    auto write(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& public_key) -> bool;
    auto verify(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& public_key) -> bool;

    auto clean_sequence() -> uint64_t;
  };

  INLINE auto PacketHeaderV4::read(Packet& packet, size_t& index, PacketDirection direction) -> bool
  {
    if (index + PacketHeaderV4::SIZE_OF_SIGNED > packet.buffer.size()) {
      LOG(ERROR, "header read, buffer is too small");
      return false;
    }

    uint8_t type;
    if (!encoding::read_uint8(packet.buffer, index, type)) {
      LOG(ERROR, "header read, unable to read packet type");
      return false;
    }
    this->type = static_cast<PacketType>(type);

    if (!encoding::read_uint64(packet.buffer, index, this->sequence)) {
      LOG(ERROR, "header read, unable to read packet sequence");
      return false;
    }

    if (direction == PacketDirection::ServerToClient) {
      // high bit must be set
      if ((this->sequence & (1ULL << 63)) == 0) {
        LOG(ERROR, "header read, high bit unset");
        return false;
      }
    } else {
      // high bit must be clear
      if ((this->sequence & (1ULL << 63)) != 0) {
        LOG(ERROR, "header read, high bit set");
        return false;
      }
    }

    if (
     this->type == PacketType::SessionPing4 || this->type == PacketType::SessionPong4 ||
     this->type == PacketType::RouteResponse4 || this->type == PacketType::ContinueResponse4) {
      // second highest bit must be set
      if ((this->sequence & (1ULL << 62)) == 0) {
        LOG(ERROR, "header read, second high bit unset");
        return false;
      }
    } else {
      // second highest bit must be clear
      if ((this->sequence & (1ULL << 62)) != 0) {
        LOG(ERROR, "header read, second high bit set");
        return false;
      }
    }

    if (!encoding::read_uint64(packet.buffer, index, this->session_id)) {
      LOG(ERROR, "header read, could not read session id");
      return false;
    }

    if (!encoding::read_uint8(packet.buffer, index, this->session_version)) {
      LOG(ERROR, "header read, could not read session version");
      return false;
    }

    return true;
  }

  INLINE auto PacketHeaderV4::write(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& private_key)
   -> bool
  {
    if (index + PacketHeaderV4::SIZE_OF_SIGNED > packet.buffer.size()) {
      LOG(ERROR, "could not write header, buffer is too small");
      return false;
    }

    if (direction == PacketDirection::ServerToClient) {
      // high bit must be set
      if ((this->sequence & (1ULL << 63)) == 0) {
        return false;
      }
    } else {
      // high bit must be clear
      if ((this->sequence & (1ULL << 63)) != 0) {
        return false;
      }
    }

    if (
     this->type == PacketType::SessionPing4 || this->type == PacketType::SessionPong4 ||
     this->type == PacketType::RouteResponse4 || this->type == PacketType::ContinueResponse4) {
      // second highest bit must be set
      if ((sequence & (1ULL << 62)) == 0) {
        return false;
      }
    } else {
      // second highest bit must be clear
      if ((sequence & (1ULL << 62)) != 0) {
        return false;
      }
    }

    if (!encoding::write_uint8(packet.buffer, index, static_cast<uint8_t>(this->type))) {
      return false;
    }

    if (!encoding::write_uint64(packet.buffer, index, this->sequence)) {
      return false;
    }

    uint8_t* additional = &packet.buffer[index];
    const int additional_length = 8 + 1;

    if (!encoding::write_uint64(packet.buffer, index, this->session_id)) {
      return false;
    }

    if (!encoding::write_uint8(packet.buffer, index, this->session_version)) {
      return false;
    }

    std::array<uint8_t, 12> nonce;
    {
      size_t index = 0;
      if (!encoding::write_uint32(nonce, index, 0)) {
        return false;
      }

      if (!encoding::write_uint64(nonce, index, this->sequence)) {
        return false;
      }
    }

    unsigned long long encrypted_length = 0;

    int result = crypto_aead_chacha20poly1305_ietf_encrypt(
     &packet.buffer[index],
     &encrypted_length,
     &packet.buffer[index],
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

  // TODO consider removing the direction & encoding reads. verify() is called after read() in all cases so index will be set to
  // the right spot and read() performs those same checks
  INLINE auto PacketHeaderV4::verify(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& private_key)
   -> bool
  {
    if (index + PacketHeaderV4::SIZE_OF_SIGNED > packet.buffer.size()) {
      LOG(ERROR, "could not verify header, buffer is too small");
      return false;
    }

    size_t begin = index;

    uint8_t type;
    if (!encoding::read_uint8(packet.buffer, index, type)) {
      return false;
    }
    PacketType packet_type = static_cast<PacketType>(type);

    uint64_t packet_sequence;
    if (!encoding::read_uint64(packet.buffer, index, packet_sequence)) {
      LOG(ERROR, "could not verify header, could not read sequence");
      return false;
    }

    if (direction == PacketDirection::ServerToClient) {
      // high bit must be set
      if ((packet_sequence & (1ULL << 63)) == 0) {
        LOG(ERROR, "could not verify header, server to client sequence check failed");
        return false;
      }
    } else {
      // high bit must be clear
      if ((packet_sequence & (1ULL << 63)) != 0) {
        LOG(ERROR, "could not verify header, client to server sequence check failed");
        return false;
      }
    }

    // TODO change this to if checks and put in a test
    if (
     packet_type == PacketType::SessionPing4 || packet_type == PacketType::SessionPong4 ||
     packet_type == PacketType::RouteResponse4 || packet_type == PacketType::ContinueResponse4) {
      // second highest bit must be set
      assert(packet_sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((packet_sequence & (1ULL << 62)) == 0);
    }

    size_t additional_index = index;
    const int additional_length = 8 + 1;

    index += 12;

    std::array<uint8_t, 12> nonce;
    {
      size_t index = 0;
      encoding::write_uint32(nonce, index, 0);
      encoding::write_uint64(nonce, index, packet_sequence);
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt(
     &packet.buffer[begin + 18],
     &decrypted_length,
     nullptr,
     &packet.buffer[begin + 18],
     (unsigned long long)crypto_aead_chacha20poly1305_IETF_ABYTES,
     &packet.buffer[additional_index],
     (unsigned long long)additional_length,
     nonce.data(),
     private_key.data());

    if (result != 0) {
      LOG(ERROR, "could not verify header, crypto aead check failed");
      return false;
    }

    return true;
  }

  INLINE auto PacketHeaderV4::clean_sequence() -> uint64_t
  {
    static const uint64_t mask = ~((1ULL << 63) | (1ULL << 62));
    return this->sequence & mask;
  }

  INLINE std::ostream& operator<<(std::ostream& stream, const PacketHeaderV4& header)
  {
    stream << std::hex << header.session_id << '.' << std::dec << static_cast<uint32_t>(header.session_version);
    return stream;
  }

  struct PacketHeader: public SessionHasher
  {
    // type (1) +
    // SessionHasher (9) +
    // sequence (8) +
    // ??? (1)
    static const size_t SIZE_OF = 19;
    static const size_t SIZE_OF_SIGNED = SIZE_OF + crypto_aead_chacha20poly1305_IETF_ABYTES;

    PacketType type;

    uint64_t sequence;

    auto read(Packet& packet, size_t& index, PacketDirection direction) -> bool;
    auto write(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& public_key) -> bool;
    auto verify(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& public_key) -> bool;

    auto clean_sequence() -> uint64_t;
  };

  INLINE auto PacketHeader::read(Packet& packet, size_t& index, PacketDirection direction) -> bool
  {
    if (index + PacketHeader::SIZE_OF_SIGNED > packet.buffer.size()) {
      LOG(ERROR, "header read, buffer is too small");
      return false;
    }

    uint8_t type;
    if (!encoding::read_uint8(packet.buffer, index, type)) {
      LOG(ERROR, "header read, unable to read packet type");
      return false;
    }
    PacketType packet_type = static_cast<PacketType>(type);

    uint64_t packet_sequence;
    if (!encoding::read_uint64(packet.buffer, index, packet_sequence)) {
      LOG(ERROR, "header read, unable to read packet sequence");
      return false;
    }

    if (direction == PacketDirection::ServerToClient) {
      // high bit must be set
      if ((packet_sequence & (1ULL << 63)) == 0) {
        LOG(ERROR, "header read, high bit unset");
        return false;
      }
    } else {
      // high bit must be clear
      if ((packet_sequence & (1ULL << 63)) != 0) {
        LOG(ERROR, "header read, high bit set");
        return false;
      }
    }

    this->type = packet_type;

    if (
     this->type == PacketType::SessionPing || this->type == PacketType::SessionPong ||
     this->type == PacketType::RouteResponse || this->type == PacketType::ContinueResponse) {
      // second highest bit must be set
      if ((packet_sequence & (1ULL << 62)) == 0) {
        LOG(ERROR, "header read, second high bit unset");
        return false;
      }
    } else {
      // second highest bit must be clear
      if ((packet_sequence & (1ULL << 62)) != 0) {
        LOG(ERROR, "header read, second high bit set");
        return false;
      }
    }

    this->sequence = packet_sequence;

    if (!encoding::read_uint64(packet.buffer, index, this->session_id)) {
      LOG(ERROR, "header read, could not read session id");
      return false;
    }

    if (!encoding::read_uint8(packet.buffer, index, this->session_version)) {
      LOG(ERROR, "header read, could not read session version");
      return false;
    }

    return true;
  }

  INLINE auto PacketHeader::write(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& private_key)
   -> bool
  {
    if (index + PacketHeader::SIZE_OF_SIGNED > packet.buffer.size()) {
      LOG(ERROR, "could not write header, buffer is too small");
      return false;
    }

    if (direction == PacketDirection::ServerToClient) {
      // high bit must be set
      if ((this->sequence & (1ULL << 63)) == 0) {
        return false;
      }
    } else {
      // high bit must be clear
      if ((this->sequence & (1ULL << 63)) != 0) {
        return false;
      }
    }

    if (
     this->type == PacketType::SessionPing || this->type == PacketType::SessionPong ||
     this->type == PacketType::RouteResponse || this->type == PacketType::ContinueResponse) {
      // second highest bit must be set
      if ((sequence & (1ULL << 62)) == 0) {
        return false;
      }
    } else {
      // second highest bit must be clear
      if ((sequence & (1ULL << 62)) != 0) {
        return false;
      }
    }

    if (!encoding::write_uint8(packet.buffer, index, static_cast<uint8_t>(this->type))) {
      return false;
    }

    if (!encoding::write_uint64(packet.buffer, index, this->sequence)) {
      return false;
    }

    uint8_t* additional = &packet.buffer[index];
    const int additional_length = 8 + 1 + 1;

    if (!encoding::write_uint64(packet.buffer, index, this->session_id)) {
      return false;
    }

    if (!encoding::write_uint8(packet.buffer, index, this->session_version)) {
      return false;
    }

    // todo: remove this once we fully switch to new relay
    // todo: ^ still applicable?
    if (!encoding::write_uint8(packet.buffer, index, 0)) {
      return false;
    }

    std::array<uint8_t, 12> nonce;
    {
      size_t index = 0;
      if (!encoding::write_uint32(nonce, index, 0)) {
        return false;
      }

      if (!encoding::write_uint64(nonce, index, this->sequence)) {
        return false;
      }
    }

    unsigned long long encrypted_length = 0;

    int result = crypto_aead_chacha20poly1305_ietf_encrypt(
     &packet.buffer[index],
     &encrypted_length,
     &packet.buffer[index],
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

  INLINE auto PacketHeader::verify(Packet& packet, size_t& index, PacketDirection direction, const GenericKey& private_key)
   -> bool
  {
    if (index + PacketHeader::SIZE_OF_SIGNED > packet.buffer.size()) {
      LOG(ERROR, "could not verify header, buffer is too small");
      return false;
    }

    size_t begin = index;

    uint8_t type;
    if (!encoding::read_uint8(packet.buffer, index, type)) {
      return false;
    }
    PacketType packet_type = static_cast<PacketType>(type);

    uint64_t packet_sequence;
    if (!encoding::read_uint64(packet.buffer, index, packet_sequence)) {
      LOG(ERROR, "could not verify header, could not read sequence");
      return false;
    }

    if (direction == PacketDirection::ServerToClient) {
      // high bit must be set
      if ((packet_sequence & (1ULL << 63)) == 0) {
        LOG(ERROR, "could not verify header, server to client sequence check failed");
        return false;
      }
    } else {
      // high bit must be clear
      if ((packet_sequence & (1ULL << 63)) != 0) {
        LOG(ERROR, "could not verify header, client to server sequence check failed");
        return false;
      }
    }

    // TODO change this to if checks
    if (
     packet_type == PacketType::SessionPing || packet_type == PacketType::SessionPong ||
     packet_type == PacketType::RouteResponse || packet_type == PacketType::ContinueResponse) {
      // second highest bit must be set
      assert(packet_sequence & (1ULL << 62));
    } else {
      // second highest bit must be clear
      assert((packet_sequence & (1ULL << 62)) == 0);
    }

    size_t additional_index = index;
    const int additional_length = 8 + 1 + 1;

    index += 12;

    std::array<uint8_t, 12> nonce;
    {
      size_t index = 0;
      encoding::write_uint32(nonce, index, 0);
      encoding::write_uint64(nonce, index, packet_sequence);
    }

    unsigned long long decrypted_length;

    int result = crypto_aead_chacha20poly1305_ietf_decrypt(
     &packet.buffer[begin + 19],
     &decrypted_length,
     nullptr,
     &packet.buffer[begin + 19],
     (unsigned long long)crypto_aead_chacha20poly1305_IETF_ABYTES,
     &packet.buffer[additional_index],
     (unsigned long long)additional_length,
     nonce.data(),
     private_key.data());

    if (result != 0) {
      LOG(ERROR, "could not verify header, crypto aead check failed");
      return false;
    }

    return true;
  }

  INLINE auto PacketHeader::clean_sequence() -> uint64_t
  {
    static const uint64_t mask = ~((1ULL << 63) | (1ULL << 62));
    return this->sequence & mask;
  }

  INLINE std::ostream& operator<<(std::ostream& stream, const PacketHeader& header)
  {
    stream << std::hex << header.session_id << '.' << std::dec << static_cast<uint32_t>(header.session_version);
    return stream;
  }
}  // namespace core