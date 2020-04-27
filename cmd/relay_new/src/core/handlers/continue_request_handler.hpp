#ifndef CORE_HANDLERS_CONTINUE_REQUEST_HANDLER
#define CORE_HANDLERS_CONTINUE_REQUEST_HANDLER

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "crypto/keychain.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class ContinueRequestHandler: public BaseHandler
    {
     public:
      ContinueRequestHandler(
       const util::Clock& relayClock,
       GenericPacket<>& packet,
       const int packetSize,
       core::SessionMap& sessions,
       const crypto::Keychain& keychain,
       util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff);

     private:
      const util::Clock& mRelayClock;
      core::SessionMap& mSessionMap;
      const crypto::Keychain& mKeychain;
      util::ThroughputRecorder& mRecorder;
    };

    inline ContinueRequestHandler::ContinueRequestHandler(
     const util::Clock& relayClock,
     GenericPacket<>& packet,
     const int packetSize,
     core::SessionMap& sessions,
     const crypto::Keychain& keychain,
     util::ThroughputRecorder& recorder)
     : BaseHandler(packet, packetSize), mRelayClock(relayClock), mSessionMap(sessions), mKeychain(keychain), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void ContinueRequestHandler::handle(core::GenericPacketBuffer<Size>& buff)
    {
      if (mPacketSize < int(1 + ContinueToken::EncryptedByteSize * 2)) {
        Log("ignoring continue request. bad packet size (", mPacketSize, ")");
        return;
      }

      size_t index = 1;
      core::ContinueToken token(mRelayClock);
      if (!token.readEncrypted(mPacket, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
        Log("ignoring continue request. could not read continue token");
        return;
      }

      if (token.expired()) {
        Log("ignoring continue request. token is expired");
        return;
      }

      auto hash = token.key();

      if (!mSessionMap.exists(hash)) {
        Log("ignoring continue request. session does not exist");
        return;
      }

      auto session = mSessionMap.get(hash);

      if (session->expired()) {
        Log("ignoring continue request. session is expired");
        mSessionMap.erase(hash);
        return;
      }

      if (session->ExpireTimestamp != token.ExpireTimestamp) {
        Log("session continued: ", std::hex, token.SessionID, '.', std::dec, static_cast<unsigned int>(token.SessionVersion));
      }

      session->ExpireTimestamp = token.ExpireTimestamp;
      mPacket.Buffer[ContinueToken::EncryptedByteSize] = RELAY_CONTINUE_REQUEST_PACKET;

      auto length = mPacketSize - ContinueToken::EncryptedByteSize;
      mRecorder.addToSent(length);
      buff.push(session->NextAddr, &mPacket.Buffer[ContinueToken::EncryptedByteSize], length);
    }
  }  // namespace handlers
}  // namespace core
#endif