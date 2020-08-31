#pragma once

#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"
#include "util/macros.hpp"

namespace core
{
  class RouteToken: public Token
  {
   public:
    RouteToken() = default;
    virtual ~RouteToken() override = default;
    // KbpsUp (4) +
    // KbpsDown (4) +
    // NextAddr (net::Address::size) +
    // PrivateKey (crypto_box_SECRETKEYBYTES) =
    static const size_t ByteSize = Token::ByteSize + 4 + 4 + net::Address::ByteSize + crypto_box_SECRETKEYBYTES;
    static const size_t EncryptedByteSize = crypto_box_NONCEBYTES + RouteToken::ByteSize + crypto_box_MACBYTES;
    static const size_t EncryptionLength = RouteToken::ByteSize + crypto_box_MACBYTES;

    uint32_t KbpsUp;
    uint32_t KbpsDown;
    net::Address NextAddr;
    std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;

    auto write_encrypted(
     Packet& packet,
     size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey) -> bool;

    auto read_encrypted(
     Packet& packet,
     size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey) -> bool;

   private:
    auto write(Packet& packet, size_t& index) -> bool;

    auto read(const Packet& packet, size_t& index) -> bool;

    auto encrypt(
     Packet& packet,
     const size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool;

    auto decrypt(
     Packet& packet,
     const size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey,
     const size_t nonceIndex) -> bool;
  };

  INLINE auto RouteToken::write_encrypted(
   Packet& packet,
   size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey) -> bool
  {
    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    crypto::RandomBytes(nonce, nonce.size());  // fill nonce

    if (!encoding::WriteBytes(packet.Buffer, index, nonce, nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t afterNonce = index;

    write(packet, index);  // write the token data to the buffer

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packet, afterNonce, senderPrivateKey, receiverPublicKey, nonce)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // index at this point will be past nonce & token, so add the mac bytes to it

    assert(index - start == RouteToken::EncryptedByteSize);  // TODO move to test

    return true;
  }

  INLINE auto RouteToken::read_encrypted(
   Packet& packet,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey) -> bool
  {
    const auto nonceIndex = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packet, index, senderPublicKey, receiverPrivateKey, nonceIndex)) {
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
    if (index + RouteToken::ByteSize > packet.Buffer.size()) {
      return false;
    }

    const auto start = index;
    (void)start;

    if (!Token::write(packet, index)) {
      return false;
    }

    if (!encoding::WriteUint32(packet.Buffer, index, this->KbpsUp)) {
      return false;
    }

    if (!encoding::WriteUint32(packet.Buffer, index, this->KbpsDown)) {
      return false;
    }

    if (!encoding::WriteAddress(packet.Buffer, index, this->NextAddr)) {
      return false;
    }

    if (!encoding::WriteBytes(packet.Buffer, index, this->PrivateKey, this->PrivateKey.size())) {
      return false;
    }

    assert(index - start == RouteToken::ByteSize);  // TODO move this to a test

    return true;
  }

  INLINE auto RouteToken::read(const Packet& packet, size_t& index) -> bool
  {
    const size_t start = index;

    (void)start;

    if (!Token::read(packet, index)) {
      return false;
    }

    if (!encoding::ReadUint32(packet.Buffer, index, this->KbpsUp)) {
      return false;
    }

    if (!encoding::ReadUint32(packet.Buffer, index, this->KbpsDown)) {
      return false;
    }

    if (!encoding::ReadAddress(packet.Buffer, index, this->NextAddr)) {
      return false;
    }

    if (!encoding::ReadBytes(packet.Buffer, index, this->PrivateKey, this->PrivateKey.size())) {
      return false;
    }

    assert(index - start == RouteToken::ByteSize);
    return true;
  }

  INLINE auto RouteToken::encrypt(
   Packet& packet,
   const size_t& encryption_start,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool
  {
    if (encryption_start + RouteToken::EncryptionLength > packet.Buffer.size()) {
      return false;
    }

    if (
     crypto_box_easy(
      &packet.Buffer[encryption_start],
      &packet.Buffer[encryption_start],
      RouteToken::ByteSize,
      nonce.data(),
      receiverPublicKey.data(),
      senderPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }

  INLINE auto RouteToken::decrypt(
   Packet& packet,
   const size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey,
   const size_t nonceIndex) -> bool
  {
    if (index + RouteToken::EncryptionLength > packet.Buffer.size()) {
      return false;
    }

    if (
     crypto_box_open_easy(
      &packet.Buffer[index],
      &packet.Buffer[index],
      RouteToken::EncryptionLength,
      &packet.Buffer[nonceIndex],
      senderPublicKey.data(),
      receiverPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }
}  // namespace core
