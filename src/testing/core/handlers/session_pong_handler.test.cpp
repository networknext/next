#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/session_pong_handler.hpp"

#define CRYPTO_HELPERS
#define OS_HELPERS
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
using os::SocketConfig;
using util::ThroughputRecorder;

Test(core_handlers_session_pong_handler_unsigned)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.setTimestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.Len = Header::ByteSize + 32;
  packet.Addr = addr;

  Header header = {
   .type = Type::ServerToClient,
   .sequence = 123123130131LL | (1ULL << 63),
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>();
  session->ClientToServerSeq = 0;
  session->ExpireTimestamp = 10;
  session->NextAddr = addr;
  session->PrevAddr = addr;
  session->PrivateKey = private_key;
  session->SessionID = header.session_id;
  session->SessionVersion = header.session_version;

  map.set(header.hash(), session);

  size_t index = 0;

  check(header.write(packet, index, Direction::ServerToClient, private_key));
  check(index == Header::ByteSize);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, false);

  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(prev_len == packet.Len);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, false);
  check(!socket.recv(packet));
}

Test(core_handlers_session_pong_handler_signed)
{
  Packet packet;
  SessionMap map;
  ThroughputRecorder recorder;
  RouterInfo router_info;
  Socket socket;

  const GenericKey private_key = random_private_key();

  router_info.setTimestamp(0);

  Address addr;
  SocketConfig config = default_socket_config();

  check(addr.parse("127.0.0.1"));
  check(socket.create(addr, config));

  packet.Len = crypto::PacketHashLength + Header::ByteSize + 32;
  packet.Addr = addr;

  Header header = {
   .type = Type::ServerToClient,
   .sequence = 123123130131LL | (1ULL << 63),
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>();
  session->ClientToServerSeq = 0;
  session->ExpireTimestamp = 10;
  session->NextAddr = addr;
  session->PrevAddr = addr;
  session->PrivateKey = private_key;
  session->SessionID = header.session_id;
  session->SessionVersion = header.session_version;

  map.set(header.hash(), session);

  size_t index = crypto::PacketHashLength;

  check(header.write(packet, index, Direction::ServerToClient, private_key));
  check(index == crypto::PacketHashLength + Header::ByteSize);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, true);

  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(prev_len == packet.Len);

  core::handlers::session_pong_handler(packet, map, recorder, router_info, socket, true);
  check(!socket.recv(packet));
}
