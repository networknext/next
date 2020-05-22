#ifndef CORE_SESSION_HPP
#define CORE_SESSION_HPP

#include "expireable.hpp"
#include "net/address.hpp"
#include "replay_protection.hpp"
#include "util/logger.hpp"

namespace core
{
  class Session: public Expireable
  {
   public:
    Session(const util::Clock& relayClock);
    virtual ~Session() override = default;

    uint64_t SessionID;
    uint8_t SessionVersion;
    int KbpsUp;
    int KbpsDown;
    net::Address PrevAddr;
    net::Address NextAddr;
    std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
    // Not tested or benchmarked yet, don't use
    // ReplayProtection ServerToClientProtection;
    // ReplayProtection ClientToServerProtection;
    legacy::relay_replay_protection_t ServerToClientProtection;
    legacy::relay_replay_protection_t ClientToServerProtection;

    void setServerToClientSeq(uint64_t seq);
    void setSessionPingSeq(uint64_t seq);
    void setSessionPongSeq(uint64_t seq);

    auto getServerToClientSeq() -> uint64_t;
    auto getSessionPingSeq() -> uint64_t;
    auto getSessionPongSeq() -> uint64_t;

   private:
    uint64_t mServerToClientSeq = 0;
    uint64_t mSessionPingSeq = 0;
    uint64_t mSessionPongSeq = 0;
  };

  inline Session::Session(const util::Clock& relayClock)
   : Expireable(relayClock), mServerToClientSeq(0), mSessionPingSeq(0), mSessionPongSeq(0)
  {}

  inline void Session::setServerToClientSeq(uint64_t seq)
  {
    LogDebug("setting server to client seq: ", seq);
    mServerToClientSeq = seq;
  }

  inline void Session::setSessionPingSeq(uint64_t seq)
  {
    LogDebug("setting session ping seq: ", seq);
    mSessionPingSeq = seq;
  }

  inline void Session::setSessionPongSeq(uint64_t seq)
  {
    LogDebug("setting session pong seq: ", seq);
    mSessionPongSeq = seq;
  }

  inline auto Session::getServerToClientSeq() -> uint64_t
  {
    return mServerToClientSeq;
  }

  inline auto Session::getSessionPingSeq() -> uint64_t
  {
    return mSessionPingSeq;
  }

  inline auto Session::getSessionPongSeq() -> uint64_t
  {
    return mSessionPongSeq;
  }

  using SessionPtr = std::shared_ptr<Session>;
}  // namespace core
#endif
