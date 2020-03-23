#ifndef CORE_BACKEND_HPP
#define CORE_BACKEND_HPP

#include "router_info.hpp"
#include "relay_manager.hpp"

#include "crypto/keychain.hpp"

namespace core
{
  class Backend
  {
   public:
    Backend(std::string hostname, std::string address, const crypto::Keychain& keychain, RouterInfo& routerInfo, RelayManager& relayManager, std::string base64RelayPublicKey);
    ~Backend() = default;

    bool init();
    bool update(uint64_t bytesReceived);

   private:
    const std::string mHostname;
    const std::string mAddressStr;
    const crypto::Keychain& mKeychain;
    RouterInfo& mRouterInfo;
    RelayManager& mRelayManager;
    const std::string mBase64RelayPublicKey;
  };
}  // namespace core
#endif