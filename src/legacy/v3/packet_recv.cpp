#include "includes.h"
#include "packet_recv.hpp"

#include "encoding/read.hpp"

namespace legacy
{
  namespace v3
  {
    // 1 byte packet type
    // 64 byte signature
    // <signed>
    //   8 byte GUID
    //   1 byte fragment index
    //   1 byte fragment count
    //   2 byte status code
    //   <zipped>
    //     JSON string
    //   </zipped>
    // </signed>
    auto packet_recv(
     core::GenericPacket<>& packet, BackendRequest& request, BackendResponse& response, std::vector<uint8_t>& completeBuffer)
     -> bool
    {
      size_t zip_start = (size_t)(1 + crypto_sign_BYTES + sizeof(uint64_t) + sizeof(uint16_t) + sizeof(uint16_t));

      if (packet.Len < zip_start || packet.Len > zip_start + FragmentSize) {
        Log(
         "invalid master UDP packet. expected between ",
         zip_start,
         " and ",
         zip_start + FragmentSize,
         " bytes, got ",
         packet.Len);
        return false;
      }

      if (
       crypto_sign_verify_detached(
        &packet.Buffer[1], &packet.Buffer[1 + crypto_sign_BYTES], packet.Len - (1 + crypto_sign_BYTES), UDPSignKey) != 0) {
        Log("invalid master UDP packet. bad cryptographic signature.");
        return false;
      }

      size_t index = 1 + crypto_sign_BYTES;
      uint64_t packet_id = encoding::ReadUint64(packet.Buffer, index);
      if (packet_id != request.ID) {
        Log("discarding unexpected master UDP packet, expected ID ", request.ID, ", got ", packet_id);
        return false;
      }

      response.FragIndex = encoding::ReadUint8(packet.Buffer, index);
      response.FragCount = encoding::ReadUint8(packet.Buffer, index);
      response.StatusCode = encoding::ReadUint16(packet.Buffer, index);

      if (response.FragCount == 0) {
        Log("invalid master fragment count (", static_cast<uint32_t>(response.FragCount), "), discarding packet");
        return false;
      }

      if (response.FragIndex >= response.FragCount) {
        Log(
         "invalid master fragment index (",
         static_cast<uint32_t>(response.FragIndex + 1),
         "/",
         static_cast<uint32_t>(response.FragCount),
         "), discarding packet");
        return false;
      }

      response.Type = static_cast<core::packets::Type>(packet.Buffer[0]);

      if (request.FragmentTotal == 0) {
        request.Type = response.Type;
        request.FragmentTotal = response.FragCount;
      }

      if (response.Type != request.Type) {
        Log("expected packet type ", request.Type, ", got ", static_cast<uint32_t>(packet.Buffer[0]), ", discarding packet");
        return false;
      }

      if (response.FragCount != request.FragmentTotal) {
        Log(
         "expected ",
         request.FragmentTotal,
         " fragments, got fragment ",
         static_cast<uint32_t>(response.FragIndex + 1),
         "/",
         static_cast<uint32_t>(response.FragCount),
         ", discarding packet");
        return false;
      }

      if (request.Fragments[response.FragIndex].Received) {
        Log(
         "already received master fragment ",
         static_cast<uint32_t>(response.FragIndex + 1),
         "/",
         static_cast<uint32_t>(response.FragCount),
         ", ignoring packet");
        return false;
      }

      // save this fragment
      {
        auto& fragment = request.Fragments[response.FragIndex];
        fragment.Length = static_cast<uint16_t>(packet.Len - zip_start);
        std::copy(
         packet.Buffer.begin() + zip_start, packet.Buffer.begin() + zip_start + fragment.Length, fragment.Data.begin());
        fragment.Received = true;
      }

      // check received fragments

      int complete_bytes = 0;

      for (int i = 0; i < request.FragmentTotal; i++) {
        auto& fragment = request.Fragments[i];
        if (fragment.Received) {
          complete_bytes += fragment.Length;
        } else {
          return false;  // not all fragments have been received yet
        }
      }

      // all fragments have been received

      request.ID = 0;  // reset request

      completeBuffer.resize(complete_bytes);

      int bytes = 0;
      for (int i = 0; i < request.FragmentTotal; i++) {
        auto& fragment = request.Fragments[i];
        std::copy(fragment.Data.begin(), fragment.Data.begin() + fragment.Length, completeBuffer.begin() + bytes);
        bytes += fragment.Length;
      }

      assert(bytes == complete_bytes);

      return true;
    }
  }  // namespace v3
}  // namespace legacy
