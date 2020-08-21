#pragma once

#include "crypto/keychain.hpp"
#include "packet.hpp"
#include "router_info.hpp"
#include "token.hpp"

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

    bool writeEncrypted(
     const GenericPacket<>& packet,
     size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey);

    bool readEncrypted(
     const GenericPacket<>& packet,
     size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey);

   private:
    void write(const uint8_t* packetData, size_t packetLength, size_t& index);

    void read(const uint8_t* packetData, size_t packetLength, size_t& index);

    bool encrypt(
     const uint8_t* packetData,
     size_t packetLength,
     const size_t& index,
     const crypto::GenericKey& senderPrivateKey,
     const crypto::GenericKey& receiverPublicKey,
     const std::array<uint8_t, crypto_box_NONCEBYTES>& nonce);

    bool decrypt(
     const uint8_t* packetData,
     size_t packetLength,
     const size_t& index,
     const crypto::GenericKey& senderPublicKey,
     const crypto::GenericKey& receiverPrivateKey,
     const size_t nonceIndex);
  };

  inline ContinueToken::ContinueToken(const RouterInfo& routerInfo): Token(routerInfo) {}
}  // namespace core
