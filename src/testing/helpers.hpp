#pragma once

// ifdef's guard against circular dependencies, only enable what you need for a particular test

namespace testing
{
  template <typename T>
  INLINE std::enable_if_t<std::is_floating_point<T>::value, T> random_decimal()
  {
    static auto rand = std::bind(std::uniform_real_distribution<T>(0.0, 1.0), std::default_random_engine());
    return static_cast<T>(rand());
  }

  template <typename T>
  INLINE std::enable_if_t<std::numeric_limits<T>::is_integer, T> random_whole()
  {
    static auto rand = std::bind(std::uniform_int_distribution<T>(), std::default_random_engine());
    return rand();
  }

#ifdef CRYPTO_HELPERS
#include "crypto/keychain.hpp"
  const auto BASE64_RELAY_PUBLIC_KEY = "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=";
  const auto BASE64_RELAY_PRIVATE_KEY = "lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=";
  const auto BASE64_ROUTER_PUBLIC_KEY = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=";
  const auto BASE64_ROUTER_PRIVATE_KEY = "ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=";

  INLINE auto random_private_key() -> crypto::GenericKey
  {
    crypto::GenericKey private_key;
    crypto::random_bytes(private_key, private_key.size());
    return private_key;
  }

  INLINE auto gen_keypair() -> std::tuple<crypto::GenericKey, crypto::GenericKey>
  {
    GenericKey public_key;
    GenericKey private_key;
    crypto_box_keypair(public_key.data(), private_key.data());
    return {public_key, private_key};
  }
#endif

#ifdef OS_HELPERS
#include "os/socket.hpp"
  INLINE auto default_socket_config() -> os::SocketConfig
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
  INLINE auto random_address() -> net::Address
  {
    net::Address retval;
    if (random_whole<uint8_t>() & 1) {
      retval.type = net::AddressType::IPv4;
      for (auto& ip : retval.ipv4) {
        ip = random_whole<uint8_t>();
      }
      retval.port = random_whole<uint16_t>();
    }
    return retval;
  }
#endif
}  // namespace testing