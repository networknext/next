#include "includes.h"
#include "continue_token.hpp"

#include "crypto/bytes.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "util/dump.hpp"

namespace core
{
  bool ContinueToken::writeEncrypted(
   GenericPacket<> packet,
   size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey)
  {
    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    crypto::RandomBytes(nonce, nonce.size());  // fill nonce

    const uint8_t* packetData = &packet.Buffer[index];
    size_t packetLength = packet.Len - index;

    // write nonce to the buffer
    if (!encoding::WriteBytes(packetData, packetLength, index, nonce.data(), nonce.size())) {
      LOG(ERROR, "could not write nonce");
      return false;
    }

    const size_t afterNonce = index;

    this->write(packetData, packetLength, index);  // write the token data to the buffer

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packetData, packetLength, afterNonce, senderPrivateKey, receiverPublicKey, nonce)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // index at this point will be past nonce & token, so add the mac bytes to it

    assert(index - start == ContinueToken::EncryptedByteSize);

    return true;
  }

  bool ContinueToken::readEncrypted(
   uint8_t* packetData,
   size_t packetLength,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey)
  {
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

  void ContinueToken::write(const uint8_t* packetData, size_t packetLength, size_t& index)
  {
    assert(index + ContinueToken::ByteSize < packetLength);

    const size_t start = index;
    (void)start;

    Token::write(packetData, packetLength, index);

    assert(index - start == ContinueToken::ByteSize); // TODO implement a friend test that can assert this instead
  }

  void ContinueToken::read(const uint8_t* packetData, size_t packetLength, size_t& index)
  {
    const size_t start = index;
    (void)start;

    Token::read(packetData, packetLength, index);

    assert(index - start == ContinueToken::ByteSize);
  }

  bool ContinueToken::encrypt(
   const uint8_t* packetData,
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

  bool ContinueToken::decrypt(
   const uint8_t* packetData,
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
      &packetData[index],
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
