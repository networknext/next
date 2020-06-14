#ifndef CORE_CONTINUE_TOKEN_HPP
#define CORE_CONTINUE_TOKEN_HPP

#include "crypto/keychain.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"

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

    bool writeEncrypted(
     uint8_t* packetData,
     size_t packetLength,
     size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey);

    bool readEncrypted(
     uint8_t* packetData,
     size_t packetLength,
     size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey);

   private:
    void write(uint8_t* packetData, size_t packetLength, size_t& index);

    void read(uint8_t* packetData, size_t packetLength, size_t& index);

    bool encrypt(
     uint8_t* packetData,
     size_t packetLength,
     const size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce);

    bool decrypt(
     uint8_t* packetData,
     size_t packetLength,
     const size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey,
     const size_t nonceIndex);
  };

  inline ContinueToken::ContinueToken(const RouterInfo& routerInfo): Token(routerInfo) {}
}  // namespace core

namespace legacy
{
  struct relay_continue_token_t
  {
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;
  };

  void relay_write_continue_token(relay_continue_token_t* token, uint8_t* buffer, int buffer_length);

  void relay_read_continue_token(relay_continue_token_t* token, const uint8_t* buffer);

  int relay_encrypt_continue_token(
   uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length);

  int relay_decrypt_continue_token(
   const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer);

  int relay_write_encrypted_continue_token(
   uint8_t** buffer, relay_continue_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key);

  int relay_read_encrypted_continue_token(
   uint8_t** buffer, relay_continue_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key);
}  // namespace legacy
#endif
