#include "includes.h"
#include "packet_send.hpp"

#include "crypto/bytes.hpp"
#include "encoding/write.hpp"

namespace legacy
{
  namespace v3
  {
    // 1 byte packet type -- enum class PacketType
    // <encrypted> -- by crypto_box_seal()
    //   <master token> -- filled in by init request
    //     19 byte IP address -- is the ip of [?]
    //     8 byte timestamp -- not sure of resolution
    //     32 byte MAC -- doesn't appear to be mutable, just copy from response
    //   </master token>
    //   8 byte GUID -- Random uint of this packet
    //   1 byte fragment index -- Index of the fragment
    //   1 byte fragment total -- Total number of fragments
    //   <zipped> -- with standard zlib
    //     data -- normally a JSON string
    //   </zipped>
    // </encrypted>
    // 64 byte MAC -- handled automatically by sodium

    auto build_udp_fragment(

     PacketType packet_type,
     const BackendToken& token,
     uint64_t id,
     uint8_t fragmentIndex,
     uint8_t fragmentTotal,
     const core::Packet<std::vector<uint8_t>>& packet,
     core::Packet<std::vector<uint8_t>>& out) -> bool
    {
      assert(fragmentTotal > 0 && fragmentIndex < fragmentTotal);

      int total_bytes = HeaderBytes + packet.Len + crypto_box_SEALBYTES;

      std::vector<uint8_t> buffer(total_bytes - 1);

      size_t index = 0;
      encoding::WriteAddress(buffer, index, token.Address);
      encoding::WriteBytes(buffer, index, token.HMAC, token.HMAC.size());
      encoding::WriteUint64(buffer, index, id);
      encoding::WriteUint8(buffer, index, fragmentIndex);
      encoding::WriteUint8(buffer, index, fragmentTotal);
      encoding::WriteBytes(buffer, index, packet.Buffer, packet.Len);

      out.Buffer.resize(total_bytes);
      out.Len = total_bytes;
      out.Addr = packet.Addr;
      out.Buffer[0] = (uint8_t)packet_type;

      if (crypto_box_seal(&out.Buffer[1], buffer.data(), HeaderBytes - 1 + packet.Len, UDPSealKey) != 0) {
        Log("failed to seal v3 udp packet");
        return false;
      }

      return true;
    }

    auto packet_send(
     const os::Socket& socket,
     const BackendToken& master_token,
     PacketType packet_type,
     core::GenericPacket<>& packet,
     BackendRequest& request) -> bool
    {
      if (packet.Addr.Type == net::AddressType::None) {
        LogDebug("can't send master UDP packet: address has not resolved yet");  // should not happen in this repo
        return false;
      }

      request = {};
      request.id = crypto::Random<uint64_t>();

      size_t compressed_bytes_available = packet.Len + 32;
      LogDebug("creating buffer for compressed data, size: ", compressed_bytes_available);
      std::vector<uint8_t> compressed_buffer(compressed_bytes_available);

      z_stream z = {};
      z.next_out = compressed_buffer.data();
      z.avail_out = compressed_bytes_available;
      z.next_in = packet.Buffer.data();
      z.avail_in = packet.Buffer.size();

      LogDebug("deflate init");
      int result = deflateInit(&z, Z_DEFAULT_COMPRESSION);
      if (result != Z_OK) {
        Log("failed to compress master UDP packet: deflateInit failed");
        return false;
      }

      LogDebug("deflating");
      result = deflate(&z, Z_FINISH);
      if (result != Z_STREAM_END || z.avail_in > 0) {
        Log("failed to compress master UDP packet: deflate failed, result: ", result, ", avail in: ", z.avail_in);
        return false;
      }

      LogDebug("end deflate");
      result = deflateEnd(&z);
      if (result != Z_OK) {
        Log("failed to compress master UDP packet: deflateEnd failed");
        return false;
      }

      size_t compressed_bytes = compressed_bytes_available - z.avail_out;
      LogDebug("overall compressed size: ", compressed_bytes);

      size_t fragment_total = compressed_bytes / FragmentSize;
      if (compressed_bytes % FragmentSize != 0) {
        fragment_total += 1;
      }

      if (fragment_total > FragmentMax) {
        Log(compressed_bytes, " byte master packet is too large even for ", FragmentMax, " fragments!");
        return false;
      }

      LogDebug("sending ", fragment_total, " fragments");
      for (size_t i = 0; i < fragment_total; i++) {
        int fragment_bytes;
        if (i == fragment_total - 1) {
          // last fragment
          fragment_bytes = compressed_bytes - ((fragment_total - 1) * FragmentSize);
        } else {
          fragment_bytes = FragmentSize;
        }

        core::Packet<std::vector<uint8_t>> pkt;
        core::Packet<std::vector<uint8_t>> out;
        pkt.Buffer.resize(fragment_bytes);
        pkt.Len = fragment_bytes;
        pkt.Addr = packet.Addr;
        std::copy(
         compressed_buffer.begin() + i * FragmentSize,
         compressed_buffer.begin() + i * FragmentSize + fragment_bytes,
         pkt.Buffer.begin());

        if (!build_udp_fragment(packet_type, master_token, request.id, i, fragment_total, pkt, out)) {
          Log("failed to build v3 packet");
          return false;
        }

        if (!socket.send(out)) {
          Log("failed to send v3 packet");
          return false;
        }
      }

      return true;
    }
  }  // namespace v3
}  // namespace legacy