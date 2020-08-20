#ifndef CRYPTO_KEYCHAIN_HPP
#define CRYPTO_KEYCHAIN_HPP

#include "encoding/base64.hpp"


namespace crypto
{
  const size_t KEY_SIZE = 32UL;
  const size_t RELAY_PUBLIC_KEY_SIZE = KEY_SIZE;
  const size_t RELAY_PRIVATE_KEY_SIZE = KEY_SIZE;

  using GenericKey = std::array<uint8_t, KeySize>;

  struct Keychain
  {
    // at the time of writing this these are equal,
    // just to throw something in the event an unintentional typo is made
    static_assert(KEY_SIZE == RELAY_PUBLIC_KEY_SIZE);
    static_assert(KEY_SIZE == RELAY_PRIVATE_KEY_SIZE);
    static_assert(KEY_SIZE == crypto_sign_PUBLICKEYBYTES);

    std::array<uint8_t, RELAY_PUBLIC_KEY_SIZE> RelayPublicKey;
    std::array<uint8_t, RELAY_PRIVATE_KEY_SIZE> RelayPrivateKey;
    std::array<uint8_t, crypto_sign_PUBLICKEYBYTES> RouterPublicKey;
    std::array<uint8_t, crypto_sign_SECRETKEYBYTES> UpdateKey;

    auto parse(std::string relayPublicKey, std::string relayPrivateKey, std::string routerPublicKey, std::string updateKey) -> bool;
  };

  inline auto Keychain::parse(std::string relayPublicKey, std::string relayPrivateKey, std::string routerPublicKey, std::string updateKey) -> bool
  {
    return encoding::base64::decode(relayPublicKey, RelayPublicKey) &&
           encoding::base64::decode(relayPrivateKey, RelayPrivateKey) &&
           encoding::base64::decode(routerPublicKey, RouterPublicKey) &&
           encoding::base64::decode(updateKey, UpdateKey);
  }
}  // namespace crypto
#endif