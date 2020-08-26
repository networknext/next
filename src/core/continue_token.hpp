#pragma once

#include "crypto/keychain.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using core::GenericPacket;

namespace core
{
  class ContinueToken: public Token
  {
   public:
    ContinueToken(const RouterInfo& routerInfo);
    virtual ~ContinueToken() override = default;

    static const size_t ByteSize = Token::ByteSize;
    static const size_t EncryptedByteSize = crypto_box_NONCEBYTES + ContinueToken::ByteSize + crypto_box_MACBYTES;
    static const size_t EncryptionLength = ContinueToken::ByteSize + crypto_box_MACBYTES;

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
     size_t packetLength,
     const size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey,
     const size_t nonceIndex) -> bool;
  };

  INLINE ContinueToken::ContinueToken(const RouterInfo& routerInfo): Token(routerInfo) {}

  INLINE auto ContinueToken::write_encrypted(
   GenericPacket<>& packet,
   size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey) -> bool
  {
    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    if (!crypto::CreateNonceBytes(nonce)) {
      return false;
    }

    uint8_t* packetData = &packet.Buffer[index];
    size_t packetLength = packet.Len - index;

    // write nonce to the buffer
    if (!encoding::WriteBytes(packetData, packetLength, index, nonce.data(), nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t afterNonce = index;

    // write the token data to the buffer
    if (!this->write(packetData, packetLength, index)) {
      return false;
    }

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packetData, packetLength, afterNonce, senderPrivateKey, receiverPublicKey, nonce)) {
      return false;
    }

    // index at this point will be past nonce & token, so add the mac bytes to it
    index += crypto_box_MACBYTES;

    assert(index - start == ContinueToken::EncryptedByteSize);  // TODO move this to a test

    return true;
  }

  INLINE auto ContinueToken::read_encrypted(
   const GenericPacket<>& packet,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey) -> bool
  {
    const uint8_t* const packetData = &packet.Buffer[index];
    size_t packetLength = packet.Len - index;

    const auto nonceIndex = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packetData, packetLength, index, senderPublicKey, receiverPrivateKey, nonceIndex)) {
      LOG(ERROR, "failed to decrypt continue token");
      return false;
    }

    read(packetData, packetLength, index);

    index += crypto_box_MACBYTES;  // adjust the index past the decrypted data

    return true;
  }

  INLINE auto ContinueToken::write(uint8_t* packetData, size_t packetLength, size_t& index) -> bool
  {
    assert(index + ContinueToken::ByteSize < packetLength);

    const size_t start = index;
    (void)start;

    if (!Token::write(packetData, packetLength, index)) {
      return false;
    }

    assert(index - start == ContinueToken::ByteSize);  // TODO implement a friend test that can assert this instead
    return true;
  }

  INLINE auto ContinueToken::read(const uint8_t* const packetData, size_t packetLength, size_t& index) -> bool
  {
    const size_t start = index;
    (void)start;

    if (!Token::read(packetData, packetLength, index)) {
      return false;
    }

    assert(index - start == ContinueToken::ByteSize);
    return true;
  }

  INLINE bool ContinueToken::encrypt(
   uint8_t* packetData,
   size_t packetLength,
   const size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce)
  {
    (void)packetLength;
    assert(packetLength >= ContinueToken::EncryptionLength);

    if (
     crypto_box_easy(
      &packetData[index],
      &packetData[index],
      ContinueToken::ByteSize,
      nonce.data(),
      receiverPublicKey.data(),
      senderPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }

  INLINE bool ContinueToken::decrypt(
   const uint8_t* const packetData,
   size_t packetLength,
   const size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey,
   const size_t nonceIndex)
  {
    (void)packetLength;
    assert(packetLength >= ContinueToken::EncryptionLength);

    if (
     crypto_box_open_easy(
      const_cast<uint8_t*>(&packetData[index]),
      &packetData[index],
      ContinueToken::EncryptionLength,
      &packetData[nonceIndex],
      senderPublicKey.data(),
      receiverPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }
}  // namespace core
