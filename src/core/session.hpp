#pragma once

#include "expireable.hpp"
#include "net/address.hpp"
#include "replay_protection.hpp"
#include "router_info.hpp"
#include "util/logger.hpp"
#include "util/macros.hpp"

namespace core
{
  struct Session: public Expireable
  {
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
    ReplayProtection client_to_server_protection;
    ReplayProtection server_to_client_protection;
  };

  using SessionPtr = std::shared_ptr<Session>;

  INLINE std::ostream& operator<<(std::ostream& os, const Session& session)
  {
    return os << std::hex << session.session_id << '.' << std::dec << static_cast<unsigned int>(session.session_version);
  }

  class SessionHasher
  {
   public:
    SessionHasher() = default;
    virtual ~SessionHasher() = default;
    // session id (8) +
    // session version (1) +
    static const size_t SIZE_OF = 9;

    uint64_t session_id;
    uint8_t session_version;

    auto hash() -> uint64_t;
  };

  INLINE auto SessionHasher::hash() -> uint64_t
  {
    return this->session_id ^ this->session_version;
  }
}  // namespace core
