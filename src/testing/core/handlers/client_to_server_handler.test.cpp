#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/client_to_server_handler.hpp"
#include "crypto/bytes.hpp"

using core::GenericPacket;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::packets::Direction;
using core::packets::Header;
using core::packets::Type;
using net::Address;
using os::Socket;
using util::ThroughputRecorder;

Test(core_handlers_client_to_server_handler_unsigned_packet)
{
  GenericPacket<> packet;
  SessionMap map;
  ThroughputRecorder recorder;
  Socket socket;

  Address addr;
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
    return private_key;
  }();
  RouterInfo info;
  info.setTimestamp(0);

  packet.Len = Header::ByteSize + 100;

  check(addr.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, addr, 64 * 1024, 64 * 1024, 0.0, false));

  Header header = {
   .type = Type::ClientToServer,
   .sequence = 123123130131LL,
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>(info);
  session->NextAddr = addr;
  session->ExpireTimestamp = 10;
  session->PrivateKey = private_key;
  legacy::relay_replay_protection_reset(&session->ClientToServerProtection);
  legacy::relay_replay_protection_reset(&session->ServerToClientProtection);

  map.set(header.hash(), session);

  size_t index = 0;

  check(header.write(packet.Buffer, index, Direction::ClientToServer, private_key));
  check(index == Header::ByteSize);

  core::handlers::client_to_server_handler(packet, map, recorder, socket, false);
  size_t prev_len = packet.Len;
  check(socket.recv(packet));
  check(prev_len == packet.Len);

  check(recorder.ClientToServerTx.PacketCount == 1);
  check(recorder.ClientToServerTx.ByteCount == packet.Len).onFail([&] {
    std::cout << "packet len = " << packet.Len << std::endl;
    std::cout << "byte count = " << recorder.ClientToServerTx.ByteCount << std::endl;
  });

  core::handlers::client_to_server_handler(packet, map, recorder, socket, false);
  // check already received
  check(!socket.recv(packet));
}

Test(core_handlers_client_to_server_handler_signed_packet)
{
  Socket socket;
  Address addr;
  GenericPacket<> packet;
  SessionMap map;
  ThroughputRecorder recorder;
  const GenericKey private_key = [] {
    GenericKey private_key;
    crypto::RandomBytes(private_key, private_key.size());
    return private_key;
  }();
  RouterInfo info;
  info.setTimestamp(0);

  packet.Len = crypto::PacketHashLength + Header::ByteSize + 100;

  check(addr.parse("127.0.0.1"));
  check(socket.create(os::SocketType::NonBlocking, addr, 64 * 1024, 64 * 1024, 0.0, false));

  Header header = {
   .type = Type::ClientToServer,
   .sequence = 123123130131LL,
   .session_id = 0x12313131,
   .session_version = 0x12,
  };

  auto session = std::make_shared<Session>(info);
  session->NextAddr = addr;
  session->ExpireTimestamp = 10;
  session->PrivateKey = private_key;
  session->SessionID = header.session_id;
  session->SessionVersion = header.session_version;
  legacy::relay_replay_protection_reset(&session->ClientToServerProtection);
  legacy::relay_replay_protection_reset(&session->ServerToClientProtection);

  map.set(header.hash(), session);

  size_t index = crypto::PacketHashLength;

  check(header.write(packet.Buffer, index, Direction::ClientToServer, private_key));
  check(index == crypto::PacketHashLength + Header::ByteSize);

  core::handlers::client_to_server_handler(packet, map, recorder, socket, true);
  check(socket.recv(packet));

  core::handlers::client_to_server_handler(packet, map, recorder, socket, true);
  // check already received
  check(!socket.recv(packet));
}
