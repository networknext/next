#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/relay_pong_handler.hpp"
#include "encoding/write.hpp"

using core::Packet;
using core::PingData;
using core::RelayManager;
using core::RelayPingInfo;
using core::packets::RELAY_PING_PACKET_SIZE;
using core::packets::Type;
using net::Address;

Test(core_handlers_relay_pong_handler)
{
  Packet packet;
  RelayManager manager;

  Address addr;
  check(addr.parse("127.0.0.1"));

  packet.Addr = addr;
  packet.Len = RELAY_PING_PACKET_SIZE;

  std::array<RelayPingInfo, 1024> relays;
  relays[0].ID = 0;
  relays[0].Addr = addr;
  manager.update(1, relays);

  check(manager.mRelays[0].Addr == addr);

  auto& ping_sent = (*manager.mRelays[0].History)[0].TimePingSent;

  check(ping_sent == -1);

  check(manager.mNumRelays == 1);

  std::array<PingData, 1024> ping_data;
  // just to increment the sequence
  manager.mRelays[0].LastPingTime = RELAY_PING_TIME * -1.0;
  check(manager.getPingData(ping_data) == 1);

  check(ping_data[0].Seq == 0);
  check(ping_data[0].Addr == addr);

  check(ping_sent > 0);

  auto& pong_received = (*manager.mRelays[0].History)[0].TimePongReceived;

  check(pong_received == -1).onFail([&] {
    std::cout << "pong received == " << pong_received << '\n';
  });

  size_t index = crypto::PacketHashLength;
  encoding::WriteUint8(packet.Buffer, index, static_cast<uint8_t>(Type::RelayPong));
  encoding::WriteUint64(packet.Buffer, index, ping_data[0].Seq);

  core::handlers::relay_pong_handler(packet, manager, true);

  check(pong_received > 0).onFail([&] {
    std::cout << "pong received == " << pong_received << '\n';
  });
}