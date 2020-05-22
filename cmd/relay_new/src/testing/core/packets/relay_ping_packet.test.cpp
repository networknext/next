#include "includes.h"
#include "testing/test.hpp"

#include "core/packets/new_relay_ping_packet.hpp"
#include "core/packets/types.hpp"
#include "crypto/bytes.hpp"
#include "encoding/write.hpp"

Test(core_packets_RelayPingPacket_general)
{
  uint8_t type = static_cast<uint8_t>(core::packets::Type::NewRelayPing);
  auto seqnum = crypto::Random<uint64_t>();
  net::Address addr = testing::RandomAddress();
  core::GenericPacket<> packet;

  // get test data into buffer
  size_t index = crypto::PacketHashLength;
  encoding::WriteUint8(packet.Buffer, index, type);
  encoding::WriteUint64(packet.Buffer, index, seqnum);

  packet.Len = core::packets::NewRelayPingPacket::ByteSize;
  core::packets::NewRelayPingPacket pingPacket(packet);

  // make sure packet is passed values correctly
  check(pingPacket.Internal.Buffer[crypto::PacketHashLength] == static_cast<uint8_t>(core::packets::Type::NewRelayPing));

  // ensure getters work
  check(pingPacket.getSeqNum() == seqnum);
}