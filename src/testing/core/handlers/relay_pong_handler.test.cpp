#include "includes.h"
#include "testing/test.hpp"

#include "core/handlers/relay_pong_handler.hpp"
#include "encoding/write.hpp"

using core::Packet;
using core::PingData;
using core::RELAY_PING_PACKET_SIZE;
using core::RelayManager;
using core::RelayPingInfo;
using core::PacketType;
using crypto::PACKET_HASH_LENGTH;
using net::Address;

Test(core_handlers_relay_pong_handler)
{
  Packet packet;
  RelayManager manager;

  Address addr;
  check(addr.parse("127.0.0.1"));

  packet.addr = addr;
  packet.length = PACKET_HASH_LENGTH + RELAY_PING_PACKET_SIZE;

  std::array<RelayPingInfo, 1024> relays;
  relays[0].id = 0;
  relays[0].address = addr;
  manager.update(1, relays);

  check(manager.relays[0].address == addr);

  auto& ping_sent = (*manager.relays[0].history)[0].TimePingSent;

  check(ping_sent == -1);

  check(manager.num_relays == 1);

  std::array<PingData, 1024> ping_data;
  // just to increment the sequence
  manager.relays[0].last_ping_time = RELAY_PING_TIME * -1.0;
  check(manager.get_ping_targets(ping_data) == 1);

  check(ping_data[0].sequence == 0);
  check(ping_data[0].address == addr);

  check(ping_sent > 0);

  auto& pong_received = (*manager.relays[0].history)[0].TimePongReceived;

  check(pong_received == -1).onFail([&] {
    std::cout << "pong received == " << pong_received << '\n';
  });

  size_t index = crypto::PACKET_HASH_LENGTH;
  encoding::write_uint8(packet.buffer, index, static_cast<uint8_t>(PacketType::RelayPong));
  encoding::write_uint64(packet.buffer, index, ping_data[0].sequence);

  core::handlers::relay_pong_handler(packet, manager, true);

  check(pong_received > 0).onFail([&] {
    std::cout << "pong received == " << pong_received << '\n';
  });
}