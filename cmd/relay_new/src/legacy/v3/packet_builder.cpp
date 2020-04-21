#include "includes.h"
#include "packet_builder.hpp"

#include "encoding/write.hpp"

namespace
{
  const size_t TokenBytes = net::Address::ByteSize + 32;
  const size_t FragmentMax = 255;
  const size_t HeaderBytes = 1 + TokenBytes + sizeof(uint64_t) + 2;

  const uint8_t UDPSealKey[] = {0x77, 0x9f, 0xf2, 0xeb, 0x45, 0xfb, 0xe8, 0x25, 0x7a, 0xf3, 0x78, 0xf9, 0x26, 0x22, 0x29, 0xc0,
                                0xa8, 0xd0, 0x66, 0x92, 0x8b, 0xf9, 0x47, 0xcc, 0x8b, 0x93, 0x62, 0xbe, 0xb3, 0x88, 0xf9, 0x6f};
}  // namespace

namespace legacy
{
  namespace v3
  {
    // 1 byte packet type
    // <encrypted>
    //   <master token>
    //     19 byte IP address
    //     8 byte timestamp
    //     32 byte MAC
    //   </master token>
    //   8 byte GUID
    //   1 byte fragment index
    //   1 byte fragment count
    //   <zipped>
    //     data (normally a JSON string)
    //   </zipped>
    // </encrypted>
    // 64 byte MAC (handled automatically by sodium)

    auto build_udp_fragment(
     uint8_t packet_type,
     const BackendToken& token,
     uint64_t id,
     uint8_t fragmentTotal,
     uint8_t fragmentCount,
     const core::GenericPacket<>& packet,
     core::Packet<std::vector<uint8_t>>& out) -> bool
    {
      assert(fragmentCount > 0 && fragmentTotal >= 0 && fragmentTotal < fragmentCount && fragmentCount <= FragmentMax);

      int total_bytes = HeaderBytes + packet.Len + crypto_box_SEALBYTES;

      std::vector<uint8_t> buffer(total_bytes - 1);

      size_t index = 0;
      encoding::WriteAddress(buffer, index, token.Address);
      encoding::WriteBytes(buffer, index, token.HMAC, token.HMAC.size());
      encoding::WriteUint64(buffer, index, id);
      encoding::WriteUint8(buffer, index, fragmentTotal);
      encoding::WriteUint8(buffer, index, fragmentCount);
      encoding::WriteBytes(buffer, index, packet.Buffer, packet.Len);

      out.Buffer.resize(total_bytes);
      out.Len = total_bytes;
      out.Addr = packet.Addr;
      out.Buffer[0] = packet_type;

      if (crypto_box_seal(&out.Buffer[1], buffer.data(), HeaderBytes - 1 + packet.Len, UDPSealKey) != 0) {
        Log("failed to seal v3 udp packet");
        return false;
      }

      return true;
    }
  }  // namespace v3
}  // namespace legacy