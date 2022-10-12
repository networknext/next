#pragma once

#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using crypto::GenericKey;
using crypto::Nonce;
using net::Address;

namespace testing
{
  class _test_core_RouteToken_write_;
}  // namespace testing

namespace core
{
  class RouteToken: public Token
  {
    friend testing::_test_core_RouteToken_write_;

   public:
    RouteToken() = default;
    virtual ~RouteToken() override = default;
    // Token(17)
    // kbps_up (4) +
    // kbps_down (4) +
    // next_addr (Address::size) +
    // private_key (crypto_box_SECRETKEYBYTES) =
    static const size_t SIZE_OF = Token::SIZE_OF + 4 + 4 + Address::SIZE_OF + crypto_box_SECRETKEYBYTES;
    static const size_t SIZE_OF_SIGNED = crypto_box_NONCEBYTES + RouteToken::SIZE_OF + crypto_box_MACBYTES;
    static const size_t ENCRYPTION_LENGTH = RouteToken::SIZE_OF + crypto_box_MACBYTES;

    uint32_t kbps_up;
    uint32_t kbps_down;
    Address next_addr;
    GenericKey private_key;

    auto write_encrypted(
     Packet& packet, size_t& index, const GenericKey& sender_private_key, const GenericKey& receiver_public_key) -> bool;

    auto read_encrypted(
     Packet& packet, size_t& index, const GenericKey& sender_public_key, const GenericKey& receiver_private_key) -> bool;

   private:
    auto write(Packet& packet, size_t& index) -> bool;

    auto read(const Packet& packet, size_t& index) -> bool;

    auto encrypt(
     Packet& packet,
     const size_t& index,
     const GenericKey& sender_private_key,
     const GenericKey& receiver_public_key,
     const Nonce& nonce) -> bool;

    auto decrypt(
     Packet& packet,
     const size_t& index,
     const GenericKey& sender_public_key,
     const GenericKey& receiver_private_key,
     const size_t nonce_index) -> bool;
  };

  INLINE auto RouteToken::write_encrypted(
   Packet& packet, size_t& index, const GenericKey& sender_private_key, const GenericKey& receiver_public_key) -> bool
  {
    Nonce nonce;
    crypto::make_nonce(nonce);

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

    return true;
  }

  INLINE auto RouteToken::read_encrypted(
   Packet& packet, size_t& index, const GenericKey& sender_public_key, const GenericKey& receiver_private_key) -> bool
  {
    const auto nonce_index = index;  // nonce is first in the packet's data
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
