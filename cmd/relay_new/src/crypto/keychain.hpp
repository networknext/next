#ifndef CRYPTO_KEYCHAIN_HPP
#define CRYPTO_KEYCHAIN_HPP
namespace crypto
{
  const size_t KeySize = 32UL;

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
  };

  using GenericKey = std::array<uint8_t, KeySize>;
}  // namespace crypto
#endif