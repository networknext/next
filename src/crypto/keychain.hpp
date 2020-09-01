#pragma once

#include "encoding/base64.hpp"
#include "bytes.hpp"

namespace base64 = encoding::base64;

namespace crypto
{
  const size_t KEY_SIZE = 32UL;
  const size_t RELAY_PUBLIC_KEY_SIZE = KEY_SIZE;
  const size_t RELAY_PRIVATE_KEY_SIZE = KEY_SIZE;

  using GenericKey = std::array<uint8_t, KEY_SIZE>;

  struct Keychain
  {
    // at the time of writing this these are equal,
    // just to throw something in the event an unintentional typo is made
    static_assert(KEY_SIZE == RELAY_PUBLIC_KEY_SIZE);
    static_assert(KEY_SIZE == RELAY_PRIVATE_KEY_SIZE);
    static_assert(KEY_SIZE == crypto_sign_PUBLICKEYBYTES);

    std::array<uint8_t, RELAY_PUBLIC_KEY_SIZE> relay_public_key;
    std::array<uint8_t, RELAY_PRIVATE_KEY_SIZE> relay_private_key;
    std::array<uint8_t, crypto_sign_PUBLICKEYBYTES> backend_public_key;

    auto parse(std::string relay_public_key, std::string relay_private_key, std::string backend_public_key) -> bool;
  };

  inline auto Keychain::parse(std::string relay_public_key, std::string relay_private_key, std::string backend_public_key)
   -> bool
  {
    return base64::decode(relay_public_key, this->relay_public_key) &&
           base64::decode(relay_private_key, this->relay_private_key) &&
           base64::decode(backend_public_key, this->backend_public_key);
  }
}  // namespace crypto
