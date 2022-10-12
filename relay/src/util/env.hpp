#pragma once

#include "logger.hpp"

namespace util
{
  struct Env
  {
    Env();

    std::string relay_address;
    std::string relay_private_key;
    std::string relay_public_key;
    std::string relay_router_public_key;
    std::string backend_hostname;
    std::optional<std::string> send_buffer_size;
    std::optional<std::string> recv_buffer_size;
    std::optional<std::string> max_cpus;
  };

  INLINE Env::Env()
  {
    std::unordered_map<const char*, std::string*> required_vars = {
     {"RELAY_ADDRESS", &relay_address},
     {"RELAY_PRIVATE_KEY", &relay_private_key},
     {"RELAY_PUBLIC_KEY", &relay_public_key},
     {"RELAY_ROUTER_PUBLIC_KEY", &relay_router_public_key},
     {"RELAY_BACKEND_HOSTNAME", &backend_hostname},
    };

    std::unordered_map<const char*, std::optional<std::string>*> optional_vars = {
     {"RELAY_SEND_BUFFER_SIZE", &send_buffer_size},
     {"RELAY_RECV_BUFFER_SIZE", &recv_buffer_size},
     {"RELAY_MAX_CORES", &max_cpus},
    };

    for (auto& pair : required_vars) {
      auto env = std::getenv(pair.first);

      if (env == nullptr) {
        LOG(FATAL, pair.first, " not set");
      }

      *pair.second = env;
    }

    for (auto& pair : optional_vars) {
      auto env = std::getenv(pair.first);
      if (env != nullptr) {
        *pair.second = env;
      }
    }
  }
}  // namespace util