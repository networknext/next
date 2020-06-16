#ifndef CORE_ROUTE_TOKEN_HPP
#define CORE_ROUTE_TOKEN_HPP

#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"

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

    bool writeEncrypted(
     uint8_t* packet,
     size_t packetLength,
     size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey);

    bool readEncrypted(
     uint8_t* packet,
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
     const size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey,
     const size_t nonceIndex);
  };

  inline RouteToken::RouteToken(const RouterInfo& routerInfo): Token(routerInfo) {}
}  // namespace core

namespace legacy
{
  struct relay_route_token_t
  {
    uint64_t expire_timestamp;
    uint64_t session_id;
    uint8_t session_version;
    uint8_t session_flags;
    int kbps_up;
    int kbps_down;
    legacy::relay_address_t next_address;
    uint8_t private_key[crypto_box_SECRETKEYBYTES];
  };

  void relay_write_route_token(relay_route_token_t* token, uint8_t* buffer, int buffer_length);

  void relay_read_route_token(relay_route_token_t* token, const uint8_t* buffer);

  int relay_encrypt_route_token(
   uint8_t* sender_private_key, uint8_t* receiver_public_key, uint8_t* nonce, uint8_t* buffer, int buffer_length);

  int relay_decrypt_route_token(
   const uint8_t* sender_public_key, const uint8_t* receiver_private_key, const uint8_t* nonce, uint8_t* buffer);

  int relay_write_encrypted_route_token(
   uint8_t** buffer, relay_route_token_t* token, uint8_t* sender_private_key, uint8_t* receiver_public_key);

  int relay_read_encrypted_route_token(
   uint8_t** buffer, relay_route_token_t* token, const uint8_t* sender_public_key, const uint8_t* receiver_private_key);
}  // namespace legacy
#endif
