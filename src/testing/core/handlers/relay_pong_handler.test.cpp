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
using core::PING_RATE;

TEST(core_handlers_relay_pong_handler)
{
  Packet packet;
  RelayManager manager;

  Address addr;
  CHECK(addr.parse("127.0.0.1"));

  packet.addr = addr;
  packet.length = PACKET_HASH_LENGTH + RELAY_PING_PACKET_SIZE;

  std::array<RelayPingInfo, 1024> relays;
  relays[0].id = 0;
  relays[0].address = addr;
  manager.update(1, relays);

  CHECK(manager.relays[0].address == addr);

  auto& ping_sent = (*manager.relays[0].history)[0].time_ping_sent;

  CHECK(ping_sent == -1);

  CHECK(manager.num_relays == 1);

  std::array<PingData, 1024> ping_data;
  // just to increment the sequence
  manager.relays[0].last_ping_time = PING_RATE * -1.0;
  CHECK(manager.get_ping_targets(ping_data) == 1);

  CHECK(ping_data[0].sequence == 0);
  CHECK(ping_data[0].address == addr);

  CHECK(ping_sent > 0);

  auto& pong_received = (*manager.relays[0].history)[0].time_pong_received;

  CHECK(pong_received == -1).on_fail([&] {
    std::cout << "pong received == " << pong_received << '\n';
  });

  size_t index = crypto::PACKET_HASH_LENGTH;
  encoding::write_uint8(packet.buffer, index, static_cast<uint8_t>(PacketType::RelayPong));
  encoding::write_uint64(packet.buffer, index, ping_data[0].sequence);

  core::handlers::relay_pong_handler(packet, manager, true);

  CHECK(pong_received > 0).on_fail([&] {
    std::cout << "pong received == " << pong_received << '\n';
  });
}