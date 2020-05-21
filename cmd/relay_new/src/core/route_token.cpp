#include "includes.h"
#include "route_token.hpp"

#include "crypto/bytes.hpp"

#include "encoding/binary.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"

namespace core
{
  bool RouteToken::writeEncrypted(
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

    if (!encoding::WriteBytes(packetData, packetLength, index, nonce.data(), nonce.size())) {
      Log("could not write nonce");
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

  bool RouteToken::readEncrypted(
   uint8_t* packetData,
   size_t packetLength,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey)
  {
    const auto nonceIndex = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packetData, index, senderPublicKey, receiverPrivateKey, nonceIndex)) {
      Log("could not decrypt route token");
      return false;
    }

    read(packetData, packetLength, index);

    index += crypto_box_MACBYTES;  // adjust the offset past the decrypted data

    return true;
  }

  void RouteToken::write(uint8_t* packetData, size_t packetLength, size_t& index)
  {
    assert(packetLength >= RouteToken::ByteSize);

    const auto start = index;

    (void)start;

    Token::write(packetData, packetLength, index);
    if (!encoding::WriteUint32(packetData, packetLength, index, KbpsUp)) {
      LogDebug("could not write kbps up");
      assert(false);
    }

    if (!encoding::WriteUint32(packetData, packetLength, index, KbpsDown)) {
      LogDebug("could not write kbps down");
      assert(false);
    }

    if (!encoding::WriteAddress(packetData, packetLength, index, NextAddr)) {
      LogDebug("could not write next addr");
      assert(false);
    }

    if (!encoding::WriteBytes(packetData, packetLength, index, PrivateKey.data(), crypto_box_SECRETKEYBYTES)) {
      LogDebug("could not write prev addr");
      assert(false);
    }

    assert(index - start == RouteToken::ByteSize);
  }

  void RouteToken::read(uint8_t* packetData, size_t packetLength, size_t& index)
  {
    const size_t start = index;

    (void)start;

    Token::read(packetData, packetLength, index);
    KbpsUp = encoding::ReadUint32(packetData, index);
    KbpsDown = encoding::ReadUint32(packetData, index);
    encoding::ReadAddress(packetData, packetLength, index, NextAddr);
    encoding::ReadBytes(packetData, packetLength, index, PrivateKey.data(), PrivateKey.size(), crypto_box_SECRETKEYBYTES);

    assert(index - start == RouteToken::ByteSize);
  }

  bool RouteToken::encrypt(
   uint8_t* packetData,
   size_t packetLength,
   const size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce)
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

  bool RouteToken::decrypt(
   uint8_t* packetData,
   const size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey,
   const size_t nonceIndex)
  {
    if (
     crypto_box_open_easy(
      &packetData[index],
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

namespace legacy
{
  void relay_write_route_token(relay_route_token_t* token, uint8_t* buffer, int buffer_length)
  {
    (void)buffer_length;

    assert(token);
    assert(buffer);
    assert((size_t)buffer_length >= core::RouteToken::ByteSize);

    uint8_t* start = buffer;

    (void)start;

    write_uint64(&buffer, token->expire_timestamp);
    write_uint64(&buffer, token->session_id);
    write_uint8(&buffer, token->session_version);
    write_uint8(&buffer, token->session_flags);
    write_uint32(&buffer, token->kbps_up);
    write_uint32(&buffer, token->kbps_down);
    write_address(&buffer, &token->next_address);
    write_bytes(&buffer, token->private_key, crypto_box_SECRETKEYBYTES);

    assert(buffer - start == core::RouteToken::ByteSize);
  }

  void relay_read_route_token(relay_route_token_t* token, const uint8_t* buffer)
  {
    assert(token);
    assert(buffer);

    const uint8_t* start = buffer;

    (void)start;

    token->expire_timestamp = read_uint64(&buffer);
    token->session_id = read_uint64(&buffer);
    token->session_version = read_uint8(&buffer);
    token->session_flags = read_uint8(&buffer);
    token->kbps_up = read_uint32(&buffer);
    token->kbps_down = read_uint32(&buffer);
    read_address(&buffer, &token->next_address);
    read_bytes(&buffer, token->private_key, crypto_box_SECRETKEYBYTES);
    assert(buffer - start == core::RouteToken::ByteSize);
  }

  int relay_encrypt_route_token(
   uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length)
  {
    assert(sender_private_key);
    assert(receiver_public_key);
    assert(buffer);
    assert(buffer_length >= (int)(core::RouteToken::ByteSize + crypto_box_MACBYTES));

    (void)buffer_length;

    if (crypto_box_easy(buffer, buffer, core::RouteToken::ByteSize, nonce, receiver_public_key, sender_private_key) != 0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  int relay_decrypt_route_token(
   const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer)
  {
    assert(sender_public_key);
    assert(receiver_private_key);
    assert(buffer);

    if (
     crypto_box_open_easy(
      buffer, buffer, core::RouteToken::ByteSize + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key) != 0) {
      return RELAY_ERROR;
    }

    return RELAY_OK;
  }

  int relay_write_encrypted_route_token(
   uint8_t** buffer, relay_route_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key)
  {
    assert(buffer);
    assert(token);
    assert(sender_private_key);
    assert(receiver_public_key);

    unsigned char nonce[crypto_box_NONCEBYTES];
    relay_random_bytes(nonce, crypto_box_NONCEBYTES);

    uint8_t* start = *buffer;

    (void)start;

    write_bytes(buffer, nonce, crypto_box_NONCEBYTES);

    relay_write_route_token(token, *buffer, core::RouteToken::ByteSize);

    if (
     relay_encrypt_route_token(
      sender_private_key, receiver_public_key, nonce, *buffer, core::RouteToken::ByteSize + crypto_box_NONCEBYTES) != RELAY_OK)
      return RELAY_ERROR;

    *buffer += core::RouteToken::ByteSize + crypto_box_MACBYTES;

    assert((*buffer - start) == core::RouteToken::EncryptedByteSize);

    return RELAY_OK;
  }

  int relay_read_encrypted_route_token(
   uint8_t** buffer, relay_route_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key)
  {
    assert(buffer);
    assert(token);
    assert(sender_public_key);
    assert(receiver_private_key);

    const uint8_t* nonce = *buffer;

    *buffer += crypto_box_NONCEBYTES;

    if (relay_decrypt_route_token(sender_public_key, receiver_private_key, nonce, *buffer) != RELAY_OK) {
      return RELAY_ERROR;
    }

    relay_read_route_token(token, *buffer);

    *buffer += core::RouteToken::ByteSize + crypto_box_MACBYTES;

    return RELAY_OK;
  }
}  // namespace legacy