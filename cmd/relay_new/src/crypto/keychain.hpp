#ifndef CRYPTO_KEYCHAIN_HPP
#define CRYPTO_KEYCHAIN_HPP
namespace crypto
{
  struct Keychain
  {
    std::array<uint8_t, RELAY_PUBLIC_KEY_BYTES> RelayPublicKey;
    std::array<uint8_t, RELAY_PRIVATE_KEY_BYTES> RelayPrivateKey;

    std::array<uint8_t, crypto_sign_PUBLICKEYBYTES> RouterPublicKey;
  };
}  // namespace crypto
#endif