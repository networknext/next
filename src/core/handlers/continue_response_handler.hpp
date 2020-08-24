#pragma once

#include "core/session_map.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "os/socket.hpp"

using core::packets::Header;
using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;
namespace core
{
  namespace handlers
  {
    inline void continue_response_handler(
     GenericPacket<>& packet, SessionMap& session_map, ThroughputRecorder& recorder, const Socket& socket, bool is_signed)
    {
      size_t index = 0;
      size_t length = packet.Len;

      if (is_signed) {
        index = crypto::PacketHashLength;
        length = packet.Len - crypto::PacketHashLength;
      }

      if (length != Header::ByteSize) {
        return;
      }

      Header header = {
       .direction = Direction::ServerToClient,
      };

      {
        size_t i = index;
        if (!header.read(packet.Buffer, i)) {
          LOG(ERROR, "ignoring continue response, relay header could not be read");
          return;
        }
      }

      uint64_t hash = header.hash();

      auto session = session_map.get(hash);

      if (!session) {
        LOG(ERROR, "ignoring continue response, could not find session: session = ", header);
        return;
      }

      if (session->expired()) {
        LOG(INFO, "ignoring continue response, session expired: session = ", *session);
        session_map.erase(hash);
        return;
      }

      uint64_t clean_sequence = header.clean_sequence();

      if (clean_sequence <= session->ServerToClientSeq) {
        LOG(
         ERROR,
         "ignoring continue response, packet already received: session = ",
         *session,
         ", ",
         clean_sequence,
         " <= ",
         session->ServerToClientSeq);
        return;
      }

      {
        size_t i = index;
        if (!header.verify(packet.Buffer, i, session->PrivateKey)) {
          LOG(ERROR, "ignoring continue response, could not verify header: session = ", *session);
          return;
        }
      }

      session->ServerToClientSeq = clean_sequence;

      recorder.ContinueResponseTx.add(packet.Len);

      if (!socket.send(session->PrevAddr, packet.Buffer.data(), packet.Len)) {
        LOG(ERROR, "failed to forward continue response to ", session->PrevAddr);
      }
    }
  }  // namespace handlers
}  // namespace core
