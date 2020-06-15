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
    //     32 byte MAC -- hash
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
     core::packets::Type packet_type,
     const BackendToken& master_token,
     uint64_t id,
     uint8_t fragmentIndex,
     uint8_t fragmentTotal,
     const std::vector<uint8_t>& frag,
     std::vector<uint8_t>& fragOut) -> bool
    {
      assert(fragmentTotal > 0 && fragmentIndex < fragmentTotal);

      size_t total_bytes = HeaderBytes + frag.size() + crypto_box_SEALBYTES;

      std::vector<uint8_t> buffer(total_bytes - 1);

      {
        size_t index = 0;
        if (!encoding::WriteAddress(buffer, index, master_token.Address)) {
          LogDebug("could not write addr");
          return false;
        }

        if (!encoding::WriteBytes(buffer, index, master_token.HMAC, master_token.HMAC.size())) {
          LogDebug("could not write master hmac");
          return false;
        }

        if (!encoding::WriteUint64(buffer, index, id)) {
          LogDebug("could not write request id");
          return false;
        }

        if (!encoding::WriteUint8(buffer, index, fragmentIndex)) {
          LogDebug("could not write frag index");
          return false;
        }

        if (!encoding::WriteUint8(buffer, index, fragmentTotal)) {
          LogDebug("could not write could not frag total");
          return false;
        }

        if (!encoding::WriteBytes(buffer, index, frag, frag.size())) {
          LogDebug("could not write packet data");
          return false;
        }
      }

      fragOut.resize(total_bytes);
      fragOut[0] = (uint8_t)packet_type;

      if (crypto_box_seal(&fragOut[1], buffer.data(), HeaderBytes + frag.size() - 1, UDPSealKey) != 0) {
        Log("failed to seal v3 udp packet");
        return false;
      }

      return true;
    }

    auto packet_send(
     const os::Socket& socket,
     const net::Address& masterAddr,
     const BackendToken& master_token,
     std::vector<uint8_t>& data,
     BackendRequest& request) -> bool
    {
      // resolving is not async in this codebase, this should be a debug only check
      assert(masterAddr.Type != net::AddressType::None && "can't send master UDP packet: address has not resolved");

      request.ID = crypto::Random<uint64_t>();

      size_t compressed_bytes_available = data.size() + 32;
      std::vector<uint8_t> compressed_buffer(compressed_bytes_available);

      z_stream z = {};
      z.next_out = compressed_buffer.data();
      z.avail_out = compressed_bytes_available;
      z.next_in = data.data();
      z.avail_in = data.size();

      int result = deflateInit(&z, Z_DEFAULT_COMPRESSION);
      if (result != Z_OK) {
        Log("failed to compress master UDP packet: deflateInit failed");
        return false;
      }

      result = deflate(&z, Z_FINISH);
      if (result != Z_STREAM_END || z.avail_in > 0) {
        Log("failed to compress master UDP packet: deflate failed, result: ", result, ", avail in: ", z.avail_in);
        return false;
      }

      result = deflateEnd(&z);
      if (result != Z_OK) {
        Log("failed to compress master UDP packet: deflateEnd failed");
        return false;
      }

      size_t compressed_bytes = compressed_bytes_available - z.avail_out;

      size_t fragment_total = compressed_bytes / FragmentSize;
      if (compressed_bytes % FragmentSize != 0) {
        fragment_total += 1;
      }

      if (fragment_total > FragmentMax) {
        Log(
         compressed_bytes,
         " byte master packet is too large even for ",
         FragmentMax,
         " fragments!: ",
         fragment_total,
         " > ",
         FragmentMax);
        return false;
      }

      LogDebug("sending ", fragment_total, " packets");
      for (size_t i = 0; i < fragment_total; i++) {
        size_t fragment_bytes;
        if (i == fragment_total - 1) {
          // last fragment
          fragment_bytes = compressed_bytes - (fragment_total - 1) * FragmentSize;
        } else {
          fragment_bytes = FragmentSize;
        }

        std::vector<uint8_t> frag(fragment_bytes);
        std::vector<uint8_t> fragOut;
        frag.resize(fragment_bytes);
        std::copy(
         compressed_buffer.begin() + i * FragmentSize,
         compressed_buffer.begin() + i * FragmentSize + fragment_bytes,
         frag.begin());

        if (!build_udp_fragment(request.Type, master_token, request.ID, i, fragment_total, frag, fragOut)) {
          Log("failed to build v3 packet");
          return false;
        }

        if (!socket.send(masterAddr, fragOut.data(), fragOut.size())) {
          Log("failed to send v3 packet");
          return false;
        }
      }

      return true;
    }
  }  // namespace v3
}  // namespace legacy