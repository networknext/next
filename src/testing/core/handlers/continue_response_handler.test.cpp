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

Test(core_handlers_continue_response_handler_unsigned)
{
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

  packet.Len = Header::ByteSize;

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
  session->PrevAddr = addr;
  session->ExpireTimestamp = 10;

  map.set(header.hash(), session);

  size_t index = 0;
  check(header.write(packet, index, Direction::ServerToClient, private_key));

  core::handlers::continue_response_handler(packet, map, recorder, socket, false);
  size_t prev_len = packet.Len;
  check(socket.recv(packet)).onFail([&] {
    std::cout << "unable to receive packet\n";
  });
  check(prev_len == packet.Len);

  check(recorder.ContinueResponseTx.PacketCount == 1);
  check(recorder.ContinueResponseTx.ByteCount == packet.Len).onFail([&] {
    std::cout << "packet len = " << packet.Len << '\n';
    std::cout << "byte count = " << recorder.ContinueResponseRx.ByteCount << '\n';
  });

  core::handlers::continue_response_handler(packet, map, recorder, socket, false);
  check(!socket.recv(packet));
}
