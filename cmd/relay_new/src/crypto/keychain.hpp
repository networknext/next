#ifndef CRYPTO_KEYCHAIN_HPP
#define CRYPTO_KEYCHAIN_HPP

#include "encoding/base64.hpp"
namespace crypto
{
  const size_t KeySize = 32UL;

  using GenericKey = std::array<uint8_t, KeySize>;

  struct Keychain
  {
    // at the time of writing this these are equal,
    // just to throw something in the event an unintentional typo is made
    static_assert(KeySize == RELAY_PUBLIC_KEY_BYTES);
    static_assert(KeySize == RELAY_PRIVATE_KEY_BYTES);
    static_assert(KeySize == crypto_sign_PUBLICKEYBYTES);

    std::array<uint8_t, RELAY_PUBLIC_KEY_BYTES> RelayPublicKey;
    std::array<uint8_t, RELAY_PRIVATE_KEY_BYTES> RelayPrivateKey;
    std::array<uint8_t, crypto_sign_PUBLICKEYBYTES> RouterPublicKey;
    std::array<uint8_t, crypto_sign_SECRETKEYBYTES> UpdateKey;

    auto parse(std::string relayPublicKey, std::string relayPrivateKey, std::string routerPublicKey, std::string updateKey) -> bool;
  };

  inline auto Keychain::parse(std::string relayPublicKey, std::string relayPrivateKey, std::string routerPublicKey, std::string updateKey) -> bool
  {
    return encoding::base64::Decode(relayPublicKey, RelayPublicKey) &&
           encoding::base64::Decode(relayPrivateKey, RelayPrivateKey) &&
           encoding::base64::Decode(routerPublicKey, RouterPublicKey) &&
           encoding::base64::Decode(updateKey, UpdateKey);
  }
}  // namespace crypto
#endif