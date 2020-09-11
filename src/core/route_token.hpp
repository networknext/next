#pragma once

#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using crypto::GenericKey;

namespace core
{
  class RouteTokenV4: public TokenV4
  {
   public:
    RouteTokenV4() = default;
    virtual ~RouteTokenV4() override = default;
    // Token(17)
    // KbpsUp (4) +
    // KbpsDown (4) +
    // NextAddr (net::Address::size) +
    // PrivateKey (crypto_box_SECRETKEYBYTES) =
    static const size_t SIZE_OF = TokenV4::SIZE_OF + 4 + 4 + net::Address::SIZE_OF + crypto_box_SECRETKEYBYTES;
    static const size_t SIZE_OF_SIGNED = crypto_box_NONCEBYTES + RouteTokenV4::SIZE_OF + crypto_box_MACBYTES;
    static const size_t ENCRYPTION_LENGTH = RouteTokenV4::SIZE_OF + crypto_box_MACBYTES;

    uint32_t kbps_up;
    uint32_t kbps_down;
    net::Address next_addr;
    GenericKey private_key;

    auto write_encrypted(
     Packet& packet,
     size_t& index,
     const crypto::GenericKey& sender_private_key,
     const crypto::GenericKey& receiver_public_key) -> bool;

    auto read_encrypted(
     Packet& packet,
     size_t& index,
     const crypto::GenericKey& sender_public_key,
     const crypto::GenericKey& receiver_private_key) -> bool;

   private:
    auto write(Packet& packet, size_t& index) -> bool;

    auto read(const Packet& packet, size_t& index) -> bool;

    auto encrypt(
     Packet& packet,
     const size_t& index,
     const crypto::GenericKey& sender_private_key,
     const crypto::GenericKey& receiver_public_key,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool;

    auto decrypt(
     Packet& packet,
     const size_t& index,
     const crypto::GenericKey& sender_public_key,
     const crypto::GenericKey& receiver_private_key,
     const size_t nonce_index) -> bool;
  };

  INLINE auto RouteTokenV4::write_encrypted(
   Packet& packet,
   size_t& index,
   const crypto::GenericKey& sender_private_key,
   const crypto::GenericKey& receiver_public_key) -> bool
  {
    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    crypto::random_bytes(nonce, nonce.size());  // fill nonce

    if (!encoding::write_bytes(packet.buffer, index, nonce, nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t after_nonce = index;

    write(packet, index);  // write the token data to the buffer

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packet, after_nonce, sender_private_key, receiver_public_key, nonce)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // index at this point will be past nonce & token, so add the mac bytes to it

    assert(index - start == RouteTokenV4::SIZE_OF_SIGNED);  // TODO move to test

    return true;
  }

  INLINE auto RouteTokenV4::read_encrypted(
   Packet& packet,
   size_t& index,
   const crypto::GenericKey& sender_public_key,
   const crypto::GenericKey& receiver_private_key) -> bool
  {
    const auto nonce_index = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packet, index, sender_public_key, receiver_private_key, nonce_index)) {
      LOG(ERROR, "could not decrypt route token");
      return false;
    }

    if (!read(packet, index)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // adjust the offset past the decrypted data

    return true;
  }

  INLINE auto RouteTokenV4::write(Packet& packet, size_t& index) -> bool
  {
    if (index + RouteTokenV4::SIZE_OF > packet.buffer.size()) {
      return false;
    }

    const auto start = index;
    (void)start;

    if (!TokenV4::write(packet, index)) {
      return false;
    }

    if (!encoding::write_uint32(packet.buffer, index, this->kbps_up)) {
      return false;
    }

    if (!encoding::write_uint32(packet.buffer, index, this->kbps_down)) {
      return false;
    }

    if (!encoding::write_address(packet.buffer, index, this->next_addr)) {
      return false;
    }

    if (!encoding::write_bytes(packet.buffer, index, this->private_key, this->private_key.size())) {
      return false;
    }

    assert(index - start == RouteTokenV4::SIZE_OF);  // TODO move this to a test

    return true;
  }

  INLINE auto RouteTokenV4::read(const Packet& packet, size_t& index) -> bool
  {
    const size_t start = index;

    (void)start;

    if (!TokenV4::read(packet, index)) {
      return false;
    }

    if (!encoding::read_uint32(packet.buffer, index, this->kbps_up)) {
      return false;
    }

    if (!encoding::read_uint32(packet.buffer, index, this->kbps_down)) {
      return false;
    }

    if (!encoding::read_address(packet.buffer, index, this->next_addr)) {
      return false;
    }

    if (!encoding::read_bytes(packet.buffer, index, this->private_key, this->private_key.size())) {
      return false;
    }

    assert(index - start == RouteTokenV4::SIZE_OF);
    return true;
  }

  INLINE auto RouteTokenV4::encrypt(
   Packet& packet,
   const size_t& encryption_start,
   const crypto::GenericKey& sender_private_key,
   const crypto::GenericKey& receiver_public_key,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool
  {
    if (encryption_start + RouteTokenV4::ENCRYPTION_LENGTH > packet.buffer.size()) {
      return false;
    }

    if (
     crypto_box_easy(
      &packet.buffer[encryption_start],
      &packet.buffer[encryption_start],
      RouteTokenV4::SIZE_OF,
      nonce.data(),
      receiver_public_key.data(),
      sender_private_key.data()) != 0) {
      return false;
    }

    return true;
  }

  INLINE auto RouteTokenV4::decrypt(
   Packet& packet,
   const size_t& index,
   const crypto::GenericKey& sender_public_key,
   const crypto::GenericKey& receiver_private_key,
   const size_t nonce_index) -> bool
  {
    if (index + RouteTokenV4::ENCRYPTION_LENGTH > packet.buffer.size()) {
      return false;
    }

    if (
     crypto_box_open_easy(
      &packet.buffer[index],
      &packet.buffer[index],
      RouteTokenV4::ENCRYPTION_LENGTH,
      &packet.buffer[nonce_index],
      sender_public_key.data(),
      receiver_private_key.data()) != 0) {
      return false;
    }

    return true;
  }

  class RouteToken: public Token
  {
   public:
    RouteToken() = default;
    virtual ~RouteToken() override = default;
    // Token(17)
    // KbpsUp (4) +
    // KbpsDown (4) +
    // NextAddr (net::Address::size) +
    // PrivateKey (crypto_box_SECRETKEYBYTES) =
    static const size_t SIZE_OF = Token::SIZE_OF + 4 + 4 + net::Address::SIZE_OF + crypto_box_SECRETKEYBYTES;
    static const size_t SIZE_OF_ENCRYPTED = crypto_box_NONCEBYTES + RouteToken::SIZE_OF + crypto_box_MACBYTES;
    static const size_t ENCRYPTION_LENGTH = RouteToken::SIZE_OF + crypto_box_MACBYTES;

    uint32_t kbps_up;
    uint32_t kbps_down;
    net::Address next_addr;
    GenericKey private_key;

    auto write_encrypted(
     Packet& packet,
     size_t& index,
     const crypto::GenericKey& sender_private_key,
     const crypto::GenericKey& receiver_public_key) -> bool;

    auto read_encrypted(
     Packet& packet,
     size_t& index,
     const crypto::GenericKey& sender_public_key,
     const crypto::GenericKey& receiver_private_key) -> bool;

   private:
    auto write(Packet& packet, size_t& index) -> bool;

    auto read(const Packet& packet, size_t& index) -> bool;

    auto encrypt(
     Packet& packet,
     const size_t& index,
     const crypto::GenericKey& sender_private_key,
     const crypto::GenericKey& receiver_public_key,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool;

    auto decrypt(
     Packet& packet,
     const size_t& index,
     const crypto::GenericKey& sender_public_key,
     const crypto::GenericKey& receiver_private_key,
     const size_t nonce_index) -> bool;
  };

  INLINE auto RouteToken::write_encrypted(
   Packet& packet,
   size_t& index,
   const crypto::GenericKey& sender_private_key,
   const crypto::GenericKey& receiver_public_key) -> bool
  {
    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    crypto::random_bytes(nonce, nonce.size());  // fill nonce

    if (!encoding::write_bytes(packet.buffer, index, nonce, nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t after_nonce = index;

    write(packet, index);  // write the token data to the buffer

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packet, after_nonce, sender_private_key, receiver_public_key, nonce)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // index at this point will be past nonce & token, so add the mac bytes to it

    assert(index - start == RouteToken::SIZE_OF_ENCRYPTED);  // TODO move to test

    return true;
  }

  INLINE auto RouteToken::read_encrypted(
   Packet& packet,
   size_t& index,
   const crypto::GenericKey& sender_public_key,
   const crypto::GenericKey& receiver_private_key) -> bool
  {
    const auto nonce_index = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packet, index, sender_public_key, receiver_private_key, nonce_index)) {
      LOG(ERROR, "could not decrypt route token");
      return false;
    }

    if (!read(packet, index)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // adjust the offset past the decrypted data

    return true;
  }

  INLINE auto RouteToken::write(Packet& packet, size_t& index) -> bool
  {
    if (index + RouteToken::SIZE_OF > packet.buffer.size()) {
      return false;
    }

    const auto start = index;
    (void)start;

    if (!Token::write(packet, index)) {
      return false;
    }

    if (!encoding::write_uint32(packet.buffer, index, this->kbps_up)) {
      return false;
    }

    if (!encoding::write_uint32(packet.buffer, index, this->kbps_down)) {
      return false;
    }

    if (!encoding::write_address(packet.buffer, index, this->next_addr)) {
      return false;
    }

    if (!encoding::write_bytes(packet.buffer, index, this->private_key, this->private_key.size())) {
      return false;
    }

    assert(index - start == RouteToken::SIZE_OF);  // TODO move this to a test

    return true;
  }

  INLINE auto RouteToken::read(const Packet& packet, size_t& index) -> bool
  {
    const size_t start = index;

    (void)start;

    if (!Token::read(packet, index)) {
      return false;
    }

    if (!encoding::read_uint32(packet.buffer, index, this->kbps_up)) {
      return false;
    }

    if (!encoding::read_uint32(packet.buffer, index, this->kbps_down)) {
      return false;
    }

    if (!encoding::read_address(packet.buffer, index, this->next_addr)) {
      return false;
    }

    if (!encoding::read_bytes(packet.buffer, index, this->private_key, this->private_key.size())) {
      return false;
    }

    assert(index - start == RouteToken::SIZE_OF);
    return true;
  }

  INLINE auto RouteToken::encrypt(
   Packet& packet,
   const size_t& encryption_start,
   const crypto::GenericKey& sender_private_key,
   const crypto::GenericKey& receiver_public_key,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool
  {
    if (encryption_start + RouteToken::ENCRYPTION_LENGTH > packet.buffer.size()) {
      return false;
    }

    if (
     crypto_box_easy(
      &packet.buffer[encryption_start],
      &packet.buffer[encryption_start],
      RouteToken::SIZE_OF,
      nonce.data(),
      receiver_public_key.data(),
      sender_private_key.data()) != 0) {
      return false;
    }

    return true;
  }

  INLINE auto RouteToken::decrypt(
   Packet& packet,
   const size_t& index,
   const crypto::GenericKey& sender_public_key,
   const crypto::GenericKey& receiver_private_key,
   const size_t nonce_index) -> bool
  {
    if (index + RouteToken::ENCRYPTION_LENGTH > packet.buffer.size()) {
      return false;
    }

    if (
     crypto_box_open_easy(
      &packet.buffer[index],
      &packet.buffer[index],
      RouteToken::ENCRYPTION_LENGTH,
      &packet.buffer[nonce_index],
      sender_public_key.data(),
      receiver_private_key.data()) != 0) {
      return false;
    }

    return true;
  }
}  // namespace core
