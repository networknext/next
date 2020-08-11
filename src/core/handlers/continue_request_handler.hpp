#ifndef CORE_HANDLERS_CONTINUE_REQUEST_HANDLER
#define CORE_HANDLERS_CONTINUE_REQUEST_HANDLER

#include "base_handler.hpp"
#include "core/packets/types.hpp"
#include "core/session_map.hpp"
#include "crypto/keychain.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"
#include "core/router_info.hpp"

namespace core
{
  namespace handlers
  {
    class ContinueRequestHandler: public BaseHandler
    {
     public:
      ContinueRequestHandler(
       GenericPacket<>& packet,
       core::SessionMap& sessions,
       const crypto::Keychain& keychain,
       util::ThroughputRecorder& recorder,
       const RouterInfo& routerInfo);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned);

     private:
      core::SessionMap& mSessionMap;
      const crypto::Keychain& mKeychain;
      util::ThroughputRecorder& mRecorder;
      const RouterInfo& mRouterInfo;
    };

    inline ContinueRequestHandler::ContinueRequestHandler(
     GenericPacket<>& packet,
     core::SessionMap& sessions,
     const crypto::Keychain& keychain,
     util::ThroughputRecorder& recorder,
     const RouterInfo& routerInfo)
     : BaseHandler(packet),
       mSessionMap(sessions),
       mKeychain(keychain),
       mRecorder(recorder),
       mRouterInfo(routerInfo)
    {}

    template <size_t Size>
    inline void ContinueRequestHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
    {
      (void)buff;
      (void)socket;

      uint8_t* data;
      size_t length;

      if (isSigned) {
        data = &mPacket.Buffer[crypto::PacketHashLength];
        length = mPacket.Len - crypto::PacketHashLength;
      } else {
        data = &mPacket.Buffer[0];
        length = mPacket.Len;
      }

      if (length < int(1 + ContinueToken::EncryptedByteSize * 2)) {
        Log("ignoring continue request. bad packet size (", length, ")");
        return;
      }

      size_t index = 1;
      core::ContinueToken token(mRouterInfo);
      if (!token.readEncrypted(data, length, index, mKeychain.RouterPublicKey, mKeychain.RelayPrivateKey)) {
        Log("ignoring continue request. could not read continue token");
        return;
      }

      if (token.expired()) {
        Log("ignoring continue request. token is expired");
        return;
      }

      auto hash = token.key();

      auto session = mSessionMap.get(hash);

      if (!session) {
        Log("ignoring continue request. session does not exist");
        return;
      }

      if (session->expired()) {
        Log("ignoring continue request. session is expired");
        mSessionMap.erase(hash);
        return;
      }

      if (session->ExpireTimestamp != token.ExpireTimestamp) {
        Log("session continued: ", token);
      }

      session->ExpireTimestamp = token.ExpireTimestamp;

      length = mPacket.Len - ContinueToken::EncryptedByteSize;

      if (isSigned) {
        mPacket.Buffer[ContinueToken::EncryptedByteSize + crypto::PacketHashLength] =
         static_cast<uint8_t>(packets::Type::ContinueRequest);
        legacy::relay_sign_network_next_packet(&mPacket.Buffer[ContinueToken::EncryptedByteSize], length);
      } else {
        mPacket.Buffer[ContinueToken::EncryptedByteSize] = static_cast<uint8_t>(packets::Type::ContinueRequest);
      }

      mRecorder.ContinueRequestTx(length);

#ifdef RELAY_MULTISEND
      buff.push(session->NextAddr, &mPacket.Buffer[ContinueToken::EncryptedByteSize], length);
#else
      if (!socket.send(session->NextAddr, &mPacket.Buffer[ContinueToken::EncryptedByteSize], length)) {
        Log("failed to forward continue request");
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif