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

    std::string RelayV3Enabled;          // whether or not the old backend compatability is turned on
    std::string RelayV3Name;             // Firestore ID
    std::string RelayV3BackendHostname;  // just the hostname, no http prefix
    std::string RelayV3BackendPort;      // port that should be used to talk to the backend
    std::string RelayV3UpdateKey;        // From firestore
    std::string RelayV3Speed;            // May take away the v3 when it comes time to implement utilization
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
     {"RELAY_MAX_CORES", &ProcessorCount},
     {"RELAY_V3_ENABLED", &RelayV3Enabled},
     {"RELAY_V3_NAME", &RelayV3Name},
     {"RELAY_V3_BACKEND_HOSTNAME", &RelayV3BackendHostname},
     {"RELAY_V3_BACKEND_PORT", &RelayV3BackendPort},
     {"RELAY_V3_UPDATE_KEY", &RelayV3UpdateKey},
     {"RELAY_V3_SPEED", &RelayV3Speed},
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