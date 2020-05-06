#include "includes.h"
#include "backend.hpp"
#include "packet_send.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"

using namespace std::chrono_literals;

namespace
{
  const uint8_t PacketType = 123;
}

namespace legacy
{
  namespace v3
  {
    Backend::Backend(util::Receiver<core::GenericPacket<>>& receiver, util::Env& env, os::Socket& socket)
     : mReceiver(receiver), mEnv(env), mSocket(socket)
    {}

    // clang-format off
    /*
     * Init response appears to be
     *
     *{
     *  "Timestamp": number, // unsigned, millisecond resolution, needs to be converted to nanos and have RTT compensation (t = backend_timestamp / 1000000 - (received - requested) / 2)
     *  "Token": string // base64 composed of an address (which one?) and hmac (of the address?)
     *}
     *
     * Config response appears to be
     *
     *{
     *  "Group": string // read-only once it's saved
     *}
     */
    // clang-format on

    auto Backend::init() -> bool
    {
      // prep

      LogDebug("resolving backend addr");
      net::Address backendAddr;
      if (!backendAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
        Log("Could not resolve the v3 backend hostname to an ip address");
        return 1;
      }

      BackendRequest request;
      BackendResponse response;
      std::vector<uint8_t> completeResponse;

      core::GenericPacket<> packet;
      std::copy(std::begin(InitKey), std::end(InitKey), packet.Buffer.begin());
      packet.Len = sizeof(InitKey);
      packet.Addr = backendAddr;

      uint8_t attempts = 0;
      bool done = false;
      do {
        // send request
        LogDebug("sending init packet");
        if (!packet_send(mSocket, mToken, PacketType::InitRequest, packet, request)) {
          Log("failed to send init packet");
          return false;
        }

        // wait a second for the response to come in
        std::this_thread::sleep_for(1s);

        LogDebug("checking for response");

        // receive response(s) if it exists, if not resend
        if (mReceiver.hasItems()) {
          while (mReceiver.hasItems()) {
            LogDebug("got packet data");
            mReceiver.recv(packet);

            // this will return true once all the fragments have been received
            if (readResponse(packet, request, response, completeResponse)) {
              done = true;
            }
          }
        } else {
          LogDebug("no received packets yet");
        }

        attempts++;
      } while (!done && !mSocket.closed() && !mReceiver.closed() && attempts < 60);

      if (mSocket.closed() || mReceiver.closed() || attempts == 60) {
        LogDebug("could not init relay");
        return false;
      }

      util::JSON doc;
      if (!this->buildCompleteResponse(completeResponse, doc)) {
        return false;
      }

      // process response

      auto timestamp = doc.get<uint64_t>("Timestamp");
      auto b64TokenBuff = doc.get<std::string>("Token");
      std::array<uint8_t, TokenBytes> tokenBuff;
      if (encoding::base64::Decode(b64TokenBuff, tokenBuff) == 0) {
        Log("failed to decode master token");
        return false;
      }

      size_t index = 0;
      encoding::ReadAddress(tokenBuff, index, mToken.Address);
      encoding::ReadBytes(tokenBuff, index, mToken.HMAC, mToken.HMAC.size());

      return true;
    }

    auto Backend::config() -> bool
    {
      return true;
    }

    auto Backend::updateCycle(const volatile bool& handle) -> bool
    {
      while (handle) {
        if (!update()) {
          return false;
        }

        LogDebug("updated with old backend");

        std::this_thread::sleep_for(10s);
      }

      return true;
    }

    /*
     * Update response appears to be
     *
     *{
     *  "PingTargets": [
     *  {
     *    "Address": string, // address of the relay
     *    "Id": number, // id of the relay -- not going to be the same ids as the new backend
     *    "Group": string, // not entirely sure
     *    "PingToken": string // base64 token composed of ??
     *  },
     *  ...
     * ]
     *}
     */

    auto Backend::update() -> bool
    {
      return true;
      util::JSON doc;

      if (!buildUpdateJSON(doc)) {
        Log("could not build v3 update json");
        return false;
      }

      std::string data = doc.toString();
      std::vector<uint8_t> buff(data.begin(), data.end());

      core::GenericPacket<> packet;
      mSocket.recv(packet);

      std::string resp(packet.Buffer.begin() + 1, packet.Buffer.begin() + packet.Len);
      if (!doc.parse(resp)) {
        Log("v3 update resp parse error: ", doc.err());
        return false;
      }

      return true;
    }

    /*
     *  {
     *    "RelayId": uint64,
     *  }
     */
    auto Backend::buildConfigJSON(util::JSON& doc) -> bool
    {
      doc.set("config", "value");
      return true;
    }

    // clang-format off
    /*
     *  {
     *    "Usage": double, // from manage_get_usage_from_samples()
     *    "TrafficStats": {
     *      "BytesPaidTx": uint64, // ? what is this, look up usage
     *      "BytesPaidRx": uint64, // ? ditto
     *      "BytesManagementTx": uint64, // number of bytes sent from management things, like pings and backend communication
     *      "BytesManagementRx": uint64, // ditto except for received bytes
     *      "BytesMeasurementTx": uint64, // number of bytes sent for all remaining cases
     *      "BytesMeasurementRx": uint64, // number of bytes received for all remaining cases
     *      "BytesInvalidRx": uint64, // bytes received that the relay can't/shouldn't process
     *      "SessionCount": uint64, // number of sessions this relay is currently handling
     *    },
     *    "PingStats": [
     *      {
     *        "RelayId": uint64,
     *        "RTT": double,
     *        "Jitter": double,
     *        "PacketLoss": double
     *      },
     *      ...
     *    ],
     *    "Metadata": {
     *      "Id": uint64, // this relay's id
     *      "PublicKey": string, // base64 of the public key
     *      "PingKey": string, // base64 of the ping key, relay.cpp (4362) crypto_auth_keygen
     *      "Group": string, // from config response
     *      "Shutdown": bool, // false until shutdown handle is true
     *    }
     *  }
     */
    // clang-format on
    auto Backend::buildUpdateJSON(util::JSON& doc) -> bool
    {
      doc.set("update", "value");
      return true;
    }

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
    auto Backend::readResponse(
     core::GenericPacket<>& packet, BackendRequest& request, BackendResponse& response, std::vector<uint8_t>& completeBuffer)
     -> bool
    {
      LogDebug("reading backend response");
      int zip_start = int(1 + crypto_sign_BYTES + sizeof(uint64_t) + sizeof(uint16_t) + sizeof(uint16_t));

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
      uint64_t packet_id = encoding::ReadUint8(packet.Buffer, index);
      if (packet_id != request.id) {
        Log("discarding unexpected master UDP packet, expected ID ", request.id, ", got ", packet_id);
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

      response.Type = static_cast<PacketType>(packet.Buffer[0]);

      if (request.fragment_total == 0) {
        request.type = response.Type;
        request.fragment_total = response.FragCount;
      }

      if (response.Type != request.type) {
        Log("expected packet type ", request.type, ", got ", static_cast<uint32_t>(packet.Buffer[0]), ", discarding packet");
        return false;
      }

      if (response.FragCount != request.fragment_total) {
        Log(
         "expected ",
         request.fragment_total,
         " fragments, got fragment ",
         static_cast<uint32_t>(response.FragIndex + 1),
         "/",
         static_cast<uint32_t>(response.FragCount),
         ", discarding packet");
        return false;
      }

      if (request.fragments[response.FragIndex].received) {
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
        auto& fragment = request.fragments[response.FragIndex];
        fragment.length = uint16_t(packet.Len - zip_start);
        std::copy(packet.Buffer.begin() + zip_start, packet.Buffer.begin() + fragment.length, fragment.data.begin());
        fragment.received = true;
      }

      // check received fragments

      int complete_bytes = 0;

      for (int i = 0; i < request.fragment_total; i++) {
        auto& fragment = request.fragments[i];
        if (fragment.received) {
          complete_bytes += fragment.length;
        } else {
          return false;  // not all fragments have been received yet
        }
      }

      // all fragments have been received

      request.id = 0;  // reset request

      completeBuffer.resize(complete_bytes);

      int bytes = 0;
      for (int i = 0; i < request.fragment_total; i++) {
        auto& fragment = request.fragments[i];
        std::copy(fragment.data.begin(), fragment.data.begin() + fragment.length, completeBuffer.begin() + bytes);
        bytes += fragment.length;
      }

      assert(bytes == complete_bytes);

      return true;
    }

    auto Backend::buildCompleteResponse(std::vector<uint8_t>& completeBuffer, util::JSON& doc) -> bool
    {
      const int MaxPayload = 2 * FragmentSize * FragmentMax;
      std::vector<char> buffer(MaxPayload + 1);

      z_stream z = {};
      z.next_in = (Bytef*)(completeBuffer.data());
      z.avail_in = completeBuffer.size();
      z.next_out = (Bytef*)(buffer.data());
      z.avail_out = MaxPayload;

      int result = inflateInit(&z);
      if (result != Z_OK) {
        Log("failed to decompress master UDP packet: inflateInit failed");
        return false;
      }

      result = inflate(&z, Z_NO_FLUSH);
      if (result != Z_STREAM_END) {
        Log("failed to decompress master UDP packet: inflate failed");
        return false;
      }

      result = inflateEnd(&z);
      if (result != Z_OK) {
        Log("failed to decompress master UDP packet: inflateEnd failed");
        return false;
      }

      int bytes = int(MaxPayload - z.avail_out);
      if (bytes == 0) {
        Log("failed to decompress master UDP packet: not enough buffer space");
        return false;
      }

      doc.parse(buffer);

      return true;
    }
  }  // namespace v3
}  // namespace legacy
