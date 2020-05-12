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
      RouteResponseHandler(
       GenericPacket<>& packet, const int packetSize, core::SessionMap& sessions, util::ThroughputRecorder& recorder);

      template <size_t Size>
      void handle(core::GenericPacketBuffer<Size>& buff);

     private:
      core::SessionMap& mSessionMap;
      util::ThroughputRecorder& mRecorder;
    };

    inline RouteResponseHandler::RouteResponseHandler(
     GenericPacket<>& packet, const int packetSize, core::SessionMap& sessions, util::ThroughputRecorder& recorder)
     : BaseHandler(packet, packetSize), mSessionMap(sessions), mRecorder(recorder)
    {}

    template <size_t Size>
    inline void RouteResponseHandler::handle(core::GenericPacketBuffer<Size>& buff)
    {
      if (mPacketSize != RELAY_HEADER_BYTES) {
        Log("ignoring route response, header byte count invalid: ", mPacketSize, " != ", RELAY_HEADER_BYTES);
        return;
      }

      uint8_t type;
      uint64_t sequence;
      uint64_t session_id;
      uint8_t session_version;
      if (
       relay::relay_peek_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT,
        &type,
        &sequence,
        &session_id,
        &session_version,
        mPacket.Buffer.data(),
        mPacketSize) != RELAY_OK) {
        Log("ignoring route response, relay header could not be read");
        return;
      }

      uint64_t hash = session_id ^ session_version;

      if (!mSessionMap.exists(hash)) {
        Log("ignoring route response, could not find session");
        return;
      }

      auto session = mSessionMap.get(hash);

      if (session->expired()) {
        mSessionMap.erase(hash);
        return;
      }

      uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
      if (clean_sequence <= session->ServerToClientSeq) {
        return;
      }

      session->ServerToClientSeq = clean_sequence;
      if (
       relay::relay_verify_header(
        RELAY_DIRECTION_SERVER_TO_CLIENT, session->PrivateKey.data(), mPacket.Buffer.data(), mPacketSize) != RELAY_OK) {
        Log("ignoring route response, header is invalid");
        return;
      }

      LogDebug("sending route response to ", session->PrevAddr);

      mRecorder.addToSent(mPacketSize);
      buff.push(session->PrevAddr, mPacket.Buffer.data(), mPacketSize);
    }
  }  // namespace handlers
}  // namespace core
#endif