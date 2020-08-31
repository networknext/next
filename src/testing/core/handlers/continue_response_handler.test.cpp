#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/continue_response_handler.hpp"

#define CRYPTO_HELPERS
#include "testing/helpers.hpp"

using core::Packet;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::packets::Direction;
using core::packets::Header;
using core::packets::Type;
using net::Address;
using os::Socket;
using util::ThroughputRecorder;

Test(core_handlers_continue_response_handler) {
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  Socket socket;

  const GenericKey private_key = random_private_key();
  RouterInfo info;
  info.setTimestamp(0);

  Address addr;
  check(addr.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, addr, 64 * 1024, 64 * 1024, 0.0, false));

  packet.Len = Header::ByteSize + 100;

  Header header = {
   .type = Type::ContinueResponse,
   .sequence = 123123130131LL | (1ULL << 63) | (1ULL << 62),
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>(info);
  session->SessionID = header.session_id;
  session->SessionVersion = header.session_version;
  session->PrivateKey = private_key;
  map.set(header.hash(), session);

  size_t index = 0;
  check(header.write(packet, index, Direction::ServerToClient, private_key));
}
