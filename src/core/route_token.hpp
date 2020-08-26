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
    RouteToken(const RouterInfo& routerInfo);
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
     GenericPacket<>& packet,
     size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey) -> bool;

    auto read_encrypted(
     const GenericPacket<>& packet,
     size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey) -> bool;

   private:
    auto write(uint8_t* packetData, size_t packetLength, size_t& index) -> bool;

    auto read(const uint8_t* const packetData, size_t packetLength, size_t& index) -> bool;

    auto encrypt(
     uint8_t* packetData,
     size_t packetLength,
     const size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool;

    auto decrypt(
     const uint8_t* const packetData,
     const size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey,
     const size_t nonceIndex) -> bool;
  };

  INLINE RouteToken::RouteToken(const RouterInfo& routerInfo): Token(routerInfo) {}

  INLINE auto RouteToken::write_encrypted(
   GenericPacket<>& packet,
   size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey) -> bool
  {
    uint8_t* packetData = &packet.Buffer[index];
    size_t packetLength = packet.Len - index;

    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    crypto::RandomBytes(nonce, nonce.size());  // fill nonce

    if (!encoding::WriteBytes(packetData, packet.Len, index, nonce.data(), nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t afterNonce = index;

    write(packetData, packetLength, index);  // write the token data to the buffer

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packetData, packetLength, afterNonce, senderPrivateKey, receiverPublicKey, nonce)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // index at this point will be past nonce & token, so add the mac bytes to it

    assert(index - start == RouteToken::EncryptedByteSize);

    return true;
  }

  INLINE auto RouteToken::read_encrypted(
   const GenericPacket<>& packet,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey) -> bool
  {
    const uint8_t* const packetData = &packet.Buffer[index];
    size_t packetLength = packet.Len - index;

    const auto nonceIndex = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packetData, index, senderPublicKey, receiverPrivateKey, nonceIndex)) {
      LOG(ERROR, "could not decrypt route token");
      return false;
    }

    read(packetData, packetLength, index);

    index += crypto_box_MACBYTES;  // adjust the offset past the decrypted data

    return true;
  }

  INLINE auto RouteToken::write(uint8_t* packetData, size_t packetLength, size_t& index) -> bool
  {
    assert(packetLength >= RouteToken::ByteSize);

    const auto start = index;
    (void)start;

    Token::write(packetData, packetLength, index);
    if (!encoding::WriteUint32(packetData, packetLength, index, this->KbpsUp)) {
      return false;
    }

    if (!encoding::WriteUint32(packetData, packetLength, index, this->KbpsDown)) {
      return false;
    }

    if (!encoding::WriteAddress(packetData, packetLength, index, this->NextAddr)) {
      return false;
    }

    if (!encoding::WriteBytes(packetData, packetLength, index, this->PrivateKey.data(), this->PrivateKey.size())) {
      return false;
    }

    assert(index - start == RouteToken::ByteSize); // TODO move this to a test

    return true;
  }

  INLINE auto RouteToken::read(const uint8_t* const packetData, size_t packetLength, size_t& index) -> bool
  {
    const size_t start = index;

    (void)start;

    Token::read(packetData, packetLength, index);
    if (!encoding::ReadUint32(packetData, packetLength, index, this->KbpsUp)) {
      return false;
    }
    if (!encoding::ReadUint32(packetData, packetLength, index, this->KbpsDown)) {
      return false;
    }
    if (!encoding::ReadAddress(packetData, packetLength, index, this->NextAddr)) {
      return false;
    }
    if (!encoding::ReadBytes(
         packetData, packetLength, index, this->PrivateKey.data(), this->PrivateKey.size(), this->PrivateKey.size())) {
      return false;
    }

    assert(index - start == RouteToken::ByteSize);
    return true;
  }

  INLINE auto RouteToken::encrypt(
   uint8_t* packetData,
   size_t packetLength,
   const size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool
  {
    (void)packetLength;
    assert(packetLength >= RouteToken::EncryptionLength);

    if (
     crypto_box_easy(
      &packetData[index],
      &packetData[index],
      RouteToken::ByteSize,
      nonce.data(),
      receiverPublicKey.data(),
      senderPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }

  INLINE auto RouteToken::decrypt(
   const uint8_t* const packetData,
   const size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey,
   const size_t nonceIndex) -> bool
  {
    if (
     crypto_box_open_easy(
      const_cast<uint8_t*>(&packetData[index]),
      &packetData[index],
      RouteToken::EncryptionLength,
      &packetData[nonceIndex],
      senderPublicKey.data(),
      receiverPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }
}  // namespace core
