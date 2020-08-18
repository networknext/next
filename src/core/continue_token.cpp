#include "includes.h"
#include "continue_token.hpp"

#include "crypto/bytes.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "util/dump.hpp"

namespace core
{
  bool ContinueToken::writeEncrypted(
   uint8_t* packetData,
   size_t packetLength,
   size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey)
  {
    const size_t start = index;
    (void)start;

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    crypto::RandomBytes(nonce, nonce.size());  // fill nonce

    // write nonce to the buffer
    if (!encoding::WriteBytes(packetData, packetLength, index, nonce.data(), nonce.size())) {
      LOG("could not write nonce");
      return false;
    }

    const size_t afterNonce = index;

    write(packetData, packetLength, index);  // write the token data to the buffer

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
      LOG("failed to decrypt continue token");
      return false;
    }

    read(packetData, packetLength, index);

    index += crypto_box_MACBYTES;  // adjust the index past the decrypted data

    return true;
  }

  void ContinueToken::write(uint8_t* packetData, size_t packetLength, size_t& index)
  {
    assert(index + ContinueToken::ByteSize < packetLength);

    const size_t start = index;
    (void)start;

    Token::write(packetData, packetLength, index);

    assert(index - start == ContinueToken::ByteSize);
  }

  void ContinueToken::read(uint8_t* packetData, size_t packetLength, size_t& index)
  {
    const size_t start = index;
    (void)start;

    Token::read(packetData, packetLength, index);

    assert(index - start == ContinueToken::ByteSize);
  }

  bool ContinueToken::encrypt(
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

  bool ContinueToken::decrypt(
   uint8_t* packetData,
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

namespace legacy
{
  void relay_write_continue_token(relay_continue_token_t* token, uint8_t* buffer, int buffer_length)
  {
    (void)buffer_length;

    assert(token);
    assert(buffer);
    assert((size_t)buffer_length >= core::ContinueToken::ByteSize);

    uint8_t* start = buffer;

    (void)start;

    write_uint64(&buffer, token->expire_timestamp);
    write_uint64(&buffer, token->session_id);
    write_uint8(&buffer, token->session_version);
    write_uint8(&buffer, token->session_flags);

    assert(buffer - start == core::ContinueToken::ByteSize);
  }

  void relay_read_continue_token(relay_continue_token_t* token, const uint8_t* buffer)
  {
    assert(token);
    assert(buffer);

    const uint8_t* start = buffer;

    (void)start;

    token->expire_timestamp = read_uint64(&buffer);
    token->session_id = read_uint64(&buffer);
    token->session_version = read_uint8(&buffer);
    token->session_flags = read_uint8(&buffer);

    assert(buffer - start == core::ContinueToken::ByteSize);
  }

  int relay_encrypt_continue_token(
   uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length)
  {
    assert(sender_private_key);
    assert(receiver_public_key);
    assert(buffer);
    assert(buffer_length >= (int)(core::ContinueToken::ByteSize + crypto_box_MACBYTES));

    (void)buffer_length;

    if (crypto_box_easy(buffer, buffer, core::ContinueToken::ByteSize, nonce, receiver_public_key, sender_private_key) != 0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  int relay_decrypt_continue_token(
   const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer)
  {
    assert(sender_public_key);
    assert(receiver_private_key);
    assert(buffer);

    if (
     crypto_box_open_easy(
      buffer, buffer, core::ContinueToken::ByteSize + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key) !=
     0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  int relay_write_encrypted_continue_token(
   uint8_t** buffer, relay_continue_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key)
  {
    assert(buffer);
    assert(token);
    assert(sender_private_key);
    assert(receiver_public_key);

    unsigned char nonce[crypto_box_NONCEBYTES];
    legacy::relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* start = *buffer;

    write_bytes(buffer, nonce, crypto_box_NONCEBYTES);

    relay_write_continue_token(token, *buffer, core::ContinueToken::ByteSize);

    if (
     relay_encrypt_continue_token(
      sender_private_key, receiver_public_key, nonce, *buffer, core::ContinueToken::ByteSize + crypto_box_NONCEBYTES) !=
     RELAY_OK)
      return RELAY_ERROR;

    *buffer += core::ContinueToken::ByteSize + crypto_box_MACBYTES;

    (void)start;

    assert((*buffer - start) == core::ContinueToken::EncryptedByteSize);

    return RELAY_OK;
  }

  int relay_read_encrypted_continue_token(
   uint8_t** buffer, relay_continue_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key)
  {
    assert(buffer);
    assert(token);
    assert(sender_public_key);
    assert(receiver_private_key);

    const uint8_t* nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if (relay_decrypt_continue_token(sender_public_key, receiver_private_key, nonce, *buffer) != RELAY_OK) {
      return RELAY_ERROR;
    }

    relay_read_continue_token(token, *buffer);

    *buffer += core::ContinueToken::ByteSize + crypto_box_MACBYTES;

    return RELAY_OK;
  }
}  // namespace legacy
