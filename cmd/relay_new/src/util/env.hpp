#pragma once

#include "logger.hpp"

namespace util
{
  struct Env
  {
    Env();

    std::string RelayAddress;
    std::string RelayPrivateKey;
    std::string RelayPublicKey;
    std::string RelayRouterPublicKey;
    std::string BackendHostname;
    std::string SendBufferSize;
    std::string RecvBufferSize;
    std::string ProcessorCount;
    std::string LogFile;
  };

  inline Env::Env()
  {
    std::unordered_map<const char*, std::string*> requiredVars = {
     {"RELAY_ADDRESS", &RelayAddress},
     {"RELAY_PRIVATE_KEY", &RelayPrivateKey},
     {"RELAY_PUBLIC_KEY", &RelayPublicKey},
     {"RELAY_ROUTER_PUBLIC_KEY", &RelayRouterPublicKey},
     {"RELAY_BACKEND_HOSTNAME", &BackendHostname},
    };

    std::unordered_map<const char*, std::string*> optionalVars = {
     {"RELAY_SEND_BUFFER_SIZE", &SendBufferSize},
     {"RELAY_RECV_BUFFER_SIZE", &RecvBufferSize},
     {"RELAY_PROCESSOR_COUNT", &ProcessorCount},
     {"RELAY_LOG_FILE", &LogFile},
    };

    for (auto& pair : requiredVars) {
      auto env = std::getenv(pair.first);

      if (env == nullptr) {
        Log("Error: ", pair.first, " not set");
        std::exit(1);
      }

      *pair.second = env;
    }

    for (auto& pair : optionalVars) {
      auto env = std::getenv(pair.first);
      if (env != nullptr) {
        *pair.second = env;
      }
    }
  }
}  // namespace util