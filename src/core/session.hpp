#pragma once

#include "expireable.hpp"
#include "net/address.hpp"
#include "replay_protection.hpp"
#include "router_info.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

namespace core
{
  class Session: public Expireable
  {
   public:
    Session() = default;
    virtual ~Session() override = default;

    uint64_t session_id;
    uint8_t session_version;
    uint64_t client_to_server_sequence;
    uint64_t server_to_client_sequence;
    uint32_t kbps_up;
    uint32_t kbps_down;
    net::Address prev_addr;
    net::Address next_addr;
    std::array<uint8_t, crypto_box_SECRETKEYBYTES> private_key;
    legacy::relay_replay_protection_t client_to_server_protection;
    legacy::relay_replay_protection_t server_to_client_protection;
    // Not tested or benchmarked yet, don't use
    // ReplayProtection ClientToServerProtection;
    // ReplayProtection ServerToClientProtection;
  };

  using SessionPtr = std::shared_ptr<Session>;

  INLINE std::ostream& operator<<(std::ostream& os, const Session& session)
  {
    return os << std::hex << session.session_id << '.' << std::dec << static_cast<unsigned int>(session.session_version);
  }
}  // namespace core
