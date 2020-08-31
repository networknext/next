#pragma once

// ifdef's guard against circular dependencies, only enable what you need for a particular test

namespace testing
{
#ifdef CRYPTO_HELPERS
#include "crypto/keychain.hpp"
  const auto Base64RelayPublicKey = "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=";
  const auto Base64RelayPrivateKey = "lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=";
  const auto Base64RouterPublicKey = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=";
  const auto Base64RouterPrivateKey = "ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=";

  inline auto make_keychain() -> crypto::Keychain
  {
    crypto::Keychain keychain;
    keychain.parse(Base64RelayPublicKey, Base64RelayPrivateKey, Base64RouterPublicKey);
    return keychain;
  }

  inline auto router_private_key() -> crypto::GenericKey
  {
    std::string key = Base64RouterPrivateKey;
    crypto::GenericKey buff;
    encoding::base64::decode(key, buff);
    return buff;
  }

  inline auto random_private_key() -> crypto::GenericKey
  {
    crypto::GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
    return private_key;
  }
#endif
}  // namespace testing