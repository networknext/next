#ifndef CORE_HANDLERS_ROUTE_RESPONSE_HANDLER_HPP
#define CORE_HANDLERS_ROUTE_RESPONSE_HANDLER_HPP

#include "base_handler.hpp"
#include "core/session_map.hpp"
#include "os/platform.hpp"
#include "util/throughput_recorder.hpp"

namespace core
{
  namespace handlers
  {
    class RouteResponseHandler: public BaseHandler
    {
     public:
      RouteResponseHandler(GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
    };

    inline RouteResponseHandler::RouteResponseHandler(
     GenericPacket<>& packet, core::SessionMap& sessions, util::ThroughputRecorder& recorder)
     : BaseHandler(packet), mSessionMap(sessions), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void RouteResponseHandler::handle(core::GenericPacketBuffer<Size>& buff, const os::Socket& socket, bool isSigned)
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

      if (length != RELAY_HEADER_BYTES) {
        Log("ignoring route response, header byte count invalid: ", length, " != ", RELAY_HEADER_BYTES);
        return;
      }

      packets::Type type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;
      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT, &type, &sequence, &session_id, &session_version, data, length) != RELAY_OK) {
        Log("ignoring route response, relay header could not be read");
        return;
      }

      uint64_t hash = session_id ^ session_version;

      auto session = mSessionMap.get(hash);

      if (!session) {
        Log(
         "ignoring route response, could not find session: session = ",
         std::hex,
         session_id,
         '.',
         std::dec,
         static_cast<unsigned int>(session_version));
        return;
      }

      if (session->expired()) {
        Log("ignoring route response, session expired: session = ", *session);
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);

      if (clean_sequence <= session->ServerToClientSeq) {
        Log(
         "ignoring route response, packet already received: session = ",
         *session,
         ", ",
         clean_sequence,
         " <= ",
         session->ServerToClientSeq);
        return;
      }

      if (relay::relay_verify_header(RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), data, length) != RELAY_OK) {
        Log("ignoring route response, header is invalid: session = ", *session);
        return;
      }

      session->ServerToClientSeq = clean_sequence;

      mRecorder.RouteResponseTx(mPacket.Len);

#ifdef RELAY_MULTISEND
      buff.push(session->PrevAddr, mPacket.Buffer.data(), mPacket.Len);
#else
      if (!socket.send(session->PrevAddr, mPacket.Buffer.data(), mPacket.Len)) {
        Log("failed to forward route response to ", session->PrevAddr);
      }
#endif
    }
  }  // namespace handlers
}  // namespace core
#endif