#pragma once

// ifdef's guard against circular dependencies, only enable what you need for a particular test

namespace testing
{
  template <typename T>
  inline std::enable_if_t<std::is_floating_point<T>::value, T> RandomFloat()
  {
    static auto rand = std::bind(std::uniform_real_distribution<T>(), std::default_random_engine());
    return static_cast<T>(rand());
  }

  template <typename T>
  inline std::enable_if_t<std::numeric_limits<T>::is_integer, T> RandomWhole()
  {
    static auto rand = std::bind(std::uniform_int_distribution<T>(), std::default_random_engine());
    return rand();
  }

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

#ifdef OS_HELPERS
#include "os/socket.hpp"
  inline auto default_socket_config() -> os::SocketConfig
  {
    os::SocketConfig config;
    config.socket_type = os::SocketType::NonBlocking;
    config.send_buffer_size = 1000000;
    config.recv_buffer_size = 1000000;
    config.reuse_port = false;

    return config;
  }
#endif

#ifdef NET_HELPERS
#include "net/address.hpp"
  inline auto RandomAddress() -> net::Address
  {
    net::Address retval;
    if (RandomWhole<uint8_t>() & 1) {
      retval.Type = net::AddressType::IPv4;
      for (auto& ip : retval.IPv4) {
        ip = RandomWhole<uint8_t>();
      }
      retval.Port = crypto::RandomWhole<uint16_t>();
    } else {
      retval.Type = net::AddressType::IPv6;
      for (auto& ip : retval.IPv6) {
        ip = crypto::RandomWhole<uint16_t>();
      }
      retval.Port = crypto::RandomWhole<uint16_t>();
    }
    return retval;
  }
#endif
}  // namespace testing