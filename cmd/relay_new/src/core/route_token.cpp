#include "includes.h"
#include "route_token.hpp"

#include "encoding/binary.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"

namespace core
{
  bool RouteToken::writeEncrypted(GenericPacket& packet,
   size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey)
  {
    const size_t start = index;  // keep track of the start of the token

    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    encoding::RandomBytes(nonce, nonce.size());  // fill nonce

    encoding::WriteBytes(packet, index, nonce, nonce.size());

    const size_t afterNonce = index;

    write(packet, index);  // write the token data to the buffer

    // encrypt at the start of the packet, function knows where to end
    if (!encrypt(packet, afterNonce, senderPrivateKey, receiverPublicKey, nonce)) {
      return false;
    }

    index += crypto_box_MACBYTES;  // index at this point will be past nonce & token, so add the mac bytes to it

    assert(index - start == RouteToken::EncryptedByteSize);

    return true;
  }

  bool RouteToken::readEncrypted(GenericPacket& packet,
   size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey)
  {
    const auto nonceIndex = index;   // nonce is first in the packet's data
    index += crypto_box_NONCEBYTES;  // followed by actual data

    if (!decrypt(packet, index, senderPublicKey, receiverPrivateKey, nonceIndex)) {
      return false;
    }

    read(packet, index);

    index += crypto_box_MACBYTES;  // adjust the offset past the decrypted data

    return true;
  }

  void RouteToken::write(GenericPacket& packet, size_t& index)
  {
    assert(packet.size() >= RouteToken::ByteSize);

    const auto start = index;

    (void)start;

    Token::write(packet, index);
    encoding::WriteUint32(packet, index, KbpsUp);
    encoding::WriteUint32(packet, index, KbpsDown);
    encoding::WriteAddress(packet, index, NextAddr);
    encoding::WriteBytes(packet, index, PrivateKey, crypto_box_SECRETKEYBYTES);

    assert(index - start == RouteToken::ByteSize);
  }

  void RouteToken::read(GenericPacket& packet, size_t& index)
  {
    const size_t start = index;

    (void)start;

    Token::read(packet, index);
    KbpsUp = encoding::ReadUint32(packet, index);
    KbpsDown = encoding::ReadUint32(packet, index);
    encoding::ReadAddress(packet, index, NextAddr);
    encoding::ReadBytes(packet, index, PrivateKey, crypto_box_SECRETKEYBYTES);

    assert(index - start == RouteToken::ByteSize);
  }

  bool RouteToken::encrypt(GenericPacket& packet,
   const size_t& index,
   const crypto::GenericKey& senderPrivateKey,
   const crypto::GenericKey& receiverPublicKey,
   const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce)
  {
    assert(packet.size() >= RouteToken::EncryptionLength);

    if (crypto_box_easy(packet.data() + index,
         packet.data() + index,
         RouteToken::ByteSize,
         nonce.data(),
         receiverPublicKey.data(),
         senderPrivateKey.data()) != 0) {
      return false;
    }

    return true;
  }

  bool RouteToken::decrypt(GenericPacket& packet,
   const size_t& index,
   const crypto::GenericKey& senderPublicKey,
   const crypto::GenericKey& receiverPrivateKey,
   const size_t nonceIndex)
  {
    if (crypto_box_open_easy(packet.data() + index,
         packet.data() + index,
         RouteToken::EncryptionLength,
         packet.data() + nonceIndex,
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

    encoding::write_uint64(&buffer, token->expire_timestamp);
    encoding::write_uint64(&buffer, token->session_id);
    encoding::write_uint8(&buffer, token->session_version);
    encoding::write_uint8(&buffer, token->session_flags);
    encoding::write_uint32(&buffer, token->kbps_up);
    encoding::write_uint32(&buffer, token->kbps_down);
    encoding::write_address(&buffer, &token->next_address);
    encoding::write_bytes(&buffer, token->private_key, crypto_box_SECRETKEYBYTES);

    assert(buffer - start == core::RouteToken::ByteSize);
  }

  void relay_read_route_token(relay_route_token_t* token, const uint8_t* buffer)
  {
    assert(token);
    assert(buffer);

    const uint8_t* start = buffer;

    (void)start;

    token->expire_timestamp = encoding::read_uint64(&buffer);
    token->session_id = encoding::read_uint64(&buffer);
    token->session_version = encoding::read_uint8(&buffer);
    token->session_flags = encoding::read_uint8(&buffer);
    token->kbps_up = encoding::read_uint32(&buffer);
    token->kbps_down = encoding::read_uint32(&buffer);
    encoding::read_address(&buffer, &token->next_address);
    encoding::read_bytes(&buffer, token->private_key, crypto_box_SECRETKEYBYTES);
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

    if (crypto_box_open_easy(
         buffer, buffer, core::RouteToken::ByteSize + crypto_box_MACBYTES, nonce, sender_public_key, receiver_private_key) !=
        0) {
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

    encoding::write_bytes(buffer, nonce, crypto_box_NONCEBYTES);

    relay_write_route_token(token, *buffer, core::RouteToken::ByteSize);

    if (relay_encrypt_route_token(
         sender_private_key, receiver_public_key, nonce, *buffer, core::RouteToken::ByteSize + crypto_box_NONCEBYTES) !=
        RELAY_OK)
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