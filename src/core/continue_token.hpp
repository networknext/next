#pragma once

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using core::GenericPacket;

namespace testing
{
  class _test_core_ContinueToken_write_;
  class _test_core_ContinueToken_read_;
}  // namespace testing

namespace core
{
  class ContinueToken: public Token
  {
    friend testing::_test_core_ContinueToken_write_;
    friend testing::_test_core_ContinueToken_read_;

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
     GenericPacket<>& packet,
     size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey) -> bool;

    auto operator==(const ContinueToken& other) const -> bool;

   private:
    auto write(GenericPacket<>& packet, size_t& index) -> bool;

    auto read(const GenericPacket<>& packet, size_t& index) -> bool;

    auto encrypt(
     GenericPacket<>& packet,
     const size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce) -> bool;

    auto decrypt(
     GenericPacket<>& packet,
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
    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    if (!crypto::CreateNonceBytes(nonce)) {
      return false;
    }

    // write nonce to the buffer
    if (!encoding::WriteBytes(packet.Buffer, index, nonce, nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t afterNonce = index;

    // write the token data to the buffer
    if (!this->write(packet, index)) {
      return false;
    }

    // encrypt at the start of the packet, function knows where to end
    if (!this->encrypt(packet, afterNonce, senderPrivateKey, receiverPublicKey, nonce)) {
      return false;
    }

    // index at this point will be past nonce & token, so add the mac bytes to it
    index += crypto_box_MACBYTES;

    return true;
  }

  INLINE auto ContinueToken::read_encrypted(
   GenericPacket<>& packet,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey) -> bool
  {
    const auto nonceIndex = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packet, index, senderPublicKey, receiverPrivateKey, nonceIndex)) {
      LOG(ERROR, "failed to decrypt continue token");
      return false;
    }

    if (!read(packet, index)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // adjust the index past the encryption data

    return true;
  }

  INLINE auto ContinueToken::operator==(const ContinueToken& other) const -> bool
  {
    return this->ExpireTimestamp == other.ExpireTimestamp && this->SessionID == other.SessionID &&
           this->SessionVersion == other.SessionVersion && this->SessionFlags == other.SessionFlags;
  }

  INLINE auto ContinueToken::write(GenericPacket<>& packet, size_t& index) -> bool
  {
    return Token::write(packet, index);
  }

  INLINE auto ContinueToken::read(const GenericPacket<>& packet, size_t& index) -> bool
  {
    return Token::read(packet, index);
  }

  INLINE bool ContinueToken::encrypt(
   GenericPacket<>& packet,
   const size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce)
  {
    if (index + ContinueToken::EncryptionLength > packet.Buffer.size()) {
      return false;
    }

    if (
     crypto_box_easy(
      &packet.Buffer[index],
      &packet.Buffer[index],
      ContinueToken::ByteSize,
      nonce.data(),
      receiverPublicKey.data(),
      senderPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }

  INLINE bool ContinueToken::decrypt(
   GenericPacket<>& packet,
   const size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey,
   const size_t nonceIndex)
  {
    if (index + ContinueToken::EncryptionLength > packet.Buffer.size()) {
      return false;
    }

    if (
     crypto_box_open_easy(
      &packet.Buffer[index],
      &packet.Buffer[index],
      ContinueToken::EncryptionLength,
      &packet.Buffer[nonceIndex],
      senderPublicKey.data(),
      receiverPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }
}  // namespace core
