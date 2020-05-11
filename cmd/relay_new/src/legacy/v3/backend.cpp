#include "includes.h"
#include "backend.hpp"
#include "packet_send.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "crypto/hash.hpp"
#include "core/relay_stats.hpp"

using namespace std::chrono_literals;

namespace
{
  const uint8_t PacketType = 123;
}

namespace legacy
{
  namespace v3
  {
    Backend::Backend(
     util::Receiver<core::GenericPacket<>>& receiver,
     util::Env& env,
     os::Socket& socket,
     const util::Clock& relayClock,
     TrafficStats& stats,
     core::RelayManager& manager)
     : mReceiver(receiver),
       mEnv(env),
       mSocket(socket),
       mClock(relayClock),
       mStats(stats),
       mRelayManager(manager),
       mRelayID(crypto::FNV(mEnv.RelayV3Name).Value)
    {
      LogDebug("generating ping key");
      std::array<uint8_t, PingKeySize> key;
      crypto_auth_keygen(key.data());
      mPingKey.resize(PingKeySize * 2);                          // allocate enough space for the encoding
      mPingKey.resize(encoding::base64::Encode(key, mPingKey));  // truncate the length
      LogDebug("done");
    }

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

      core::GenericPacket<> packet;
      {
        std::copy(std::begin(InitKey), std::end(InitKey), packet.Buffer.begin());
        packet.Len = sizeof(InitKey);
        packet.Addr = backendAddr;
      }
      BackendRequest request = {};
      {
        request.type = PacketType::InitRequest;
      }
      BackendResponse response;
      util::JSON doc;

      auto [ok, err] = sendAndRecv(packet, request, response, doc);
      if (!ok) {
        Log(err);
        return false;
      }

      // process response

      LogDebug("processing init response");

      if (!doc.memberExists("Timestamp")) {
        Log("v3 backend json does not contain 'Timestamp' member");
        return false;
      }

      if (!doc.memberExists("Token")) {
        Log("v3 backend json does not contain 'Token' member");
        return false;
      }

      mInitTimestamp = doc.get<uint64_t>("Timestamp") / 1000000 - (response.At - request.At) / 2;
      auto b64TokenBuff = doc.get<std::string>("Token");
      std::array<uint8_t, TokenBytes> tokenBuff;
      if (encoding::base64::Decode(b64TokenBuff, tokenBuff) == 0) {
        Log("failed to decode master token: ", b64TokenBuff);
        return false;
      }

      size_t index = 0;
      encoding::ReadAddress(tokenBuff, index, mToken.Address);
      encoding::ReadBytes(tokenBuff, index, mToken.HMAC, mToken.HMAC.size());

      return true;
    }

    auto Backend::config() -> bool
    {
      LogDebug("configuring relay");

      net::Address backendAddr;
      {
        LogDebug("resolving backend addr");
        if (!backendAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
          Log("Could not resolve the v3 backend hostname to an ip address");
          return 1;
        }
      }

      core::GenericPacket<> packet;
      {
        util::JSON doc;
        {
          if (!buildConfigJSON(doc)) {
            Log("failed to build config json");
            return false;
          }
        }

        auto jsonStr = doc.toString();
        std::copy(jsonStr.begin(), jsonStr.end(), packet.Buffer.begin());
        packet.Len = jsonStr.length();
        packet.Addr = backendAddr;
      }
      BackendRequest request = {};
      {
        request.type = PacketType::ConfigRequest;
      }
      BackendResponse response;
      util::JSON doc;

      auto [ok, err] = sendAndRecv(packet, request, response, doc);
      if (!ok) {
        Log(err);
        return false;
      }

      if (!doc.memberExists("Group")) {
        Log("v3 backend json does not contain 'Group' member");
        return false;
      }

      mGroup = doc.get<std::string>("Group");

      mGroupID = crypto::FNV(mGroup).Value;

      return true;
    }

    auto Backend::updateCycle(const volatile bool& handle) -> bool
    {
      while (handle) {
        if (!update(false)) {
          Log("failed to update relay");
          return false;
        }

        LogDebug("updated with old backend");

        std::this_thread::sleep_for(10s);
      }

      update(true);

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

    auto Backend::update(bool shuttingDown) -> bool
    {
      net::Address backendAddr;
      {
        LogDebug("resolving backend addr");
        if (!backendAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
          Log("Could not resolve the v3 backend hostname to an ip address");
          return 1;
        }
      }

      core::GenericPacket<> packet;
      {
        util::JSON doc;
        if (!buildUpdateJSON(doc, shuttingDown)) {
          Log("could not build v3 update json");
          return false;
        }
        LogDebug("sending update v3: ", doc.toPrettyString());
        auto jsonStr = doc.toString();
        std::copy(jsonStr.begin(), jsonStr.end(), packet.Buffer.begin());
        packet.Len = jsonStr.length();
        packet.Addr = backendAddr;
      }
      BackendRequest request = {};
      {
        request.type = PacketType::UpdateRequest;
      }
      BackendResponse response;
      util::JSON doc;

      auto [ok, err] = sendAndRecv(packet, request, response, doc);
      if (!ok) {
        Log(err);
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
      doc.set(mRelayID, "RelayId");
      return true;
    }

    // clang-format off
    /*
     *  {
     *    "Usage": double, // 100.0 * 8.0 * (bytes_per_sec_total_tx + bytes_per_sec_total_rx) / relaySpeed

     *    "TrafficStats": {
     *      "BytesPaidTx": uint64, // number of bytes sent between game client <-> server
     *      "BytesPaidRx": uint64, // ditto but received
     *      "BytesManagementTx": uint64, // number of bytes sent from management things, like pings and backend communication
     *      "BytesManagementRx": uint64, // ditto but received
     *      "BytesMeasurementTx": uint64, // number of bytes sent for all remaining cases
     *      "BytesMeasurementRx": uint64, // ditto but received
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
    auto Backend::buildUpdateJSON(util::JSON& doc, bool shuttingDown) -> bool
    {
      // traffic stats
      {
        util::JSON trafficStats;

        size_t bytesPerSecPaidTx = mStats.BytesPerSecPaidTx;
        size_t bytesPerSecPaidRx = mStats.BytesPerSecPaidRx;
        size_t bytesPerSecManagementTx = mStats.BytesPerSecManagementTx;
        size_t bytesPerSecManagementRx = mStats.BytesPerSecManagementRx;
        size_t bytesPerSecMeasurementTx = mStats.BytesPerSecMeasurementTx;
        size_t bytesPerSecMeasurementRx = mStats.BytesPerSecMeasurementRx;
        size_t bytesPerSecInvalidRx = mStats.BytesPerSecInvalidRx;

        mStats.BytesPerSecPaidTx -= bytesPerSecPaidTx;
        mStats.BytesPerSecPaidRx -= bytesPerSecPaidRx;
        mStats.BytesPerSecManagementTx -= bytesPerSecManagementTx;
        mStats.BytesPerSecManagementRx -= bytesPerSecManagementRx;
        mStats.BytesPerSecMeasurementTx -= bytesPerSecMeasurementTx;
        mStats.BytesPerSecMeasurementRx -= bytesPerSecMeasurementRx;
        mStats.BytesPerSecInvalidRx -= bytesPerSecInvalidRx;

        trafficStats.set(bytesPerSecPaidTx, "BytesPaidTx");
        trafficStats.set(bytesPerSecPaidRx, "BytesPaidRx");
        trafficStats.set(bytesPerSecManagementTx, "BytesManagementTx");
        trafficStats.set(bytesPerSecManagementRx, "BytesManagementRx");
        trafficStats.set(bytesPerSecMeasurementTx, "BytesMeasurementTx");
        trafficStats.set(bytesPerSecMeasurementRx, "BytesMeasurementRx");
        trafficStats.set(bytesPerSecInvalidRx, "BytesInvalidRx");

        doc.set(trafficStats, "TrafficStats");
      }

      // metadata
      {
        util::JSON metadata;

        metadata.set(mRelayID, "Id");
        metadata.set(mEnv.RelayPublicKey, "PublicKey");
        metadata.set(mPingKey, "PingKey");
        metadata.set(mGroup, "Group");
        metadata.set(shuttingDown, "Shutdown");

        doc.set(metadata, "Metadata");
      }

      // ping stats
      {
        util::JSON pingStats;

        pingStats.setArray();

        core::RelayStats stats;
        mRelayManager.getStats(stats);

        for (unsigned int i = 0; i < stats.NumRelays; ++i) {
          util::JSON pingStat;
          pingStat.set(stats.IDs[i], "RelayId");
          pingStat.set(stats.RTT[i], "RTT");
          pingStat.set(stats.Jitter[i], "Jitter");
          pingStat.set(stats.PacketLoss[i], "PacketLoss");

          if (!pingStats.push(pingStat)) {
            return false;
          }
        }

        doc.set(pingStats, "PingStats");
      }

      return true;
    }

    auto Backend::sendAndRecv(
     core::GenericPacket<>& packet, BackendRequest& request, BackendResponse& response, util::JSON& doc)
     -> std::tuple<bool, std::string>
    {
      std::vector<uint8_t> completeResponse;
      bool done = false;
      unsigned int attempts = 0;
      do {
        LogDebug("sending ", request.type, " packet");
        request.At = mClock.elapsed<util::Nanosecond>();
        bool sendSuccess = packet_send(mSocket, mToken, packet, request);

        attempts++;

        // wait a second for the response to come in or if send failed
        std::this_thread::sleep_for(1s);

        if (!sendSuccess) {
          Log("failed to send init packet");
          continue;
        }

        LogDebug("checking for response");

        // receive response(s) if it exists, if not resend
        // if + while so that I can log something for the sake of debugging
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
      } while (!done && !mSocket.closed() && !mReceiver.closed() && attempts < 60);

      response.At = mClock.elapsed<util::Nanosecond>();

      if (mSocket.closed() || mReceiver.closed() || attempts == 60) {
        std::stringstream ss;
        ss << "could not send request, attempts: " << attempts;
        return {false, ss.str()};
      }

      LogDebug("building complete response");
      auto [ok, err] = this->buildCompleteResponse(completeResponse, doc);
      if (!ok) {
        return {false, err};
      }

      LogDebug("received v3: ", doc.toPrettyString());

      return {true, ""};
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

      LogDebug("verifing signature");
      if (
       crypto_sign_verify_detached(
        &packet.Buffer[1], &packet.Buffer[1 + crypto_sign_BYTES], packet.Len - (1 + crypto_sign_BYTES), UDPSignKey) != 0) {
        Log("invalid master UDP packet. bad cryptographic signature.");
        return false;
      }

      LogDebug("getting packet id");
      size_t index = 1 + crypto_sign_BYTES;
      uint64_t packet_id = encoding::ReadUint64(packet.Buffer, index);
      if (packet_id != request.id) {
        Log("discarding unexpected master UDP packet, expected ID ", request.id, ", got ", packet_id);
        return false;
      }
      LogDebug("id is ", packet_id);

      response.FragIndex = encoding::ReadUint8(packet.Buffer, index);
      response.FragCount = encoding::ReadUint8(packet.Buffer, index);
      response.StatusCode = encoding::ReadUint16(packet.Buffer, index);

      LogDebug("fragment index is ", (int)response.FragIndex);
      LogDebug("fragment count is ", (int)response.FragCount);
      LogDebug("status code is ", (int)response.StatusCode);

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

      LogDebug("packet is of type: ", response.Type);

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
        LogDebug("saving fragment");
        auto& fragment = request.fragments[response.FragIndex];
        fragment.length = static_cast<uint16_t>(packet.Len - zip_start);
        std::copy(
         packet.Buffer.begin() + zip_start, packet.Buffer.begin() + zip_start + fragment.length, fragment.data.begin());
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
      LogDebug("all fragments received");

      request.id = 0;  // reset request

      LogDebug("total size of fragment: ", complete_bytes);
      completeBuffer.resize(complete_bytes);

      int bytes = 0;
      for (int i = 0; i < request.fragment_total; i++) {
        auto& fragment = request.fragments[i];
        std::copy(fragment.data.begin(), fragment.data.begin() + fragment.length, completeBuffer.begin() + bytes);
        bytes += fragment.length;
      }

      LogDebug("built complete packet");

      assert(bytes == complete_bytes);

      return true;
    }

    auto Backend::buildCompleteResponse(std::vector<uint8_t>& completeBuffer, util::JSON& doc) -> std::tuple<bool, std::string>
    {
      const int MaxPayload = 2 * FragmentSize * FragmentMax;
      std::vector<char> buffer(MaxPayload + 1);

      z_stream z = {};
      z.next_in = (Bytef*)(completeBuffer.data());
      z.avail_in = completeBuffer.size();
      z.next_out = (Bytef*)(buffer.data());
      z.avail_out = MaxPayload;

      LogDebug("initalizing inflate");
      int result = inflateInit(&z);
      if (result != Z_OK) {
        return {false, "failed to decompress master UDP packet: inflateInit failed"};
      }

      LogDebug("inflating");
      result = inflate(&z, Z_NO_FLUSH);
      if (result != Z_STREAM_END) {
        std::stringstream ss;
        ss << "failed to decompress master UDP packet: inflate failed, result is " << result;
        return {false, ss.str()};
      }

      LogDebug("ending inflate");
      result = inflateEnd(&z);
      if (result != Z_OK) {
        return {false, "failed to decompress master UDP packet: inflateEnd failed"};
      }

      int bytes = int(MaxPayload - z.avail_out);
      if (bytes == 0) {
        return {false, "failed to decompress master UDP packet: not enough buffer space"};
      }

      LogDebug("parsing buffer");
      if (!doc.parse(buffer)) {
        std::stringstream ss;
        std::string strbuff(buffer.begin(), buffer.end());
        ss << "failed to parse json response, looks like: " << strbuff;
        return {false, ss.str()};
      }
      LogDebug("done");

      return {true, ""};
    }
  }  // namespace v3
}  // namespace legacy
