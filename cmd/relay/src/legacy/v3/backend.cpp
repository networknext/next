#include "includes.h"
#include "backend.hpp"

#include "core/packets/types.hpp"
#include "core/relay_stats.hpp"
#include "crypto/hash.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "packet_recv.hpp"
#include "packet_send.hpp"

using namespace std::chrono_literals;

namespace
{
  const uint8_t PacketType = 123;
  const uint64_t OneSecInNanos = 1000000000ULL;
}  // namespace

namespace legacy
{
  namespace v3
  {
    Backend::Backend(
     volatile bool& shouldComm,
     util::Receiver<core::GenericPacket<>>& receiver,
     util::Env& env,
     const uint64_t relayID,
     os::Socket& socket,
     const util::Clock& relayClock,
     TrafficStats& stats,
     core::RelayManager<core::V3Relay>& manager,
     const size_t speed,
     std::atomic<ResponseState>& state,
     const crypto::Keychain& keychain,
     const core::SessionMap& sessions)
     : mShouldCommunicate(shouldComm),
       mReceiver(receiver),
       mEnv(env),
       mSocket(socket),
       mClock(relayClock),
       mStats(stats),
       mRelayManager(manager),
       mSpeed(speed),
       mRelayID(relayID),
       mState(state),
       mKeychain(keychain),
       mSessions(sessions)
    {
      std::array<uint8_t, PingKeySize> key;
      crypto_auth_keygen(key.data());
      mPingKey.resize(PingKeySize * 2);                          // allocate enough space for the encoding
      mPingKey.resize(encoding::base64::Encode(key, mPingKey));  // truncate the length
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

      std::vector<uint8_t> reqData(std::begin(InitKey), std::end(InitKey));

      BackendRequest request;
      {
        request.Type = core::packets::Type::V3InitRequest;
      }
      BackendResponse response;
      util::JSON respDoc;

      mState = ResponseState::Init;
      auto [ok, err] = sendBinRecvJSON(request, reqData, response, respDoc);
      if (!ok) {
        Log(err);
        return false;
      }

      // process response

      if (!respDoc.memberExists("Timestamp")) {
        Log("v3 backend json does not contain 'Timestamp' member");
        return false;
      }

      if (!respDoc.memberExists("Token")) {
        Log("v3 backend json does not contain 'Token' member");
        return false;
      }

      // "Timestamp" is in milliseconds, convert to nanos
      mInitTimestamp = respDoc.get<uint64_t>("Timestamp") * 1000000 + (response.At - request.At) / 2;
      auto b64TokenBuff = respDoc.get<std::string>("Token");
      std::array<uint8_t, TokenBytes> tokenBuff;
      if (encoding::base64::Decode(b64TokenBuff, tokenBuff) != 51) {
        Log("failed to decode master token: ", b64TokenBuff);
        return false;
      }

      size_t index = 0;
      encoding::ReadAddress(tokenBuff, index, mToken.Address);
      encoding::ReadBytes(tokenBuff, index, mToken.HMAC, mToken.HMAC.size());

      mInitReceived = mClock.unixTime<util::Nanosecond>();

      return true;
    }

    auto Backend::config() -> bool
    {
      util::JSON reqDoc;
      {
        if (!buildConfigJSON(reqDoc)) {
          Log("failed to build config json");
          return false;
        }
      }

      BackendRequest request;
      {
        request.Type = core::packets::Type::V3ConfigRequest;
      }
      BackendResponse response;
      util::JSON respDoc;

      mState = ResponseState::Config;
      auto [ok, err] = sendJSONRecvJSON(request, reqDoc, response, respDoc);
      if (!ok) {
        Log(err);
        return false;
      }

      if (!respDoc.memberExists("Group")) {
        Log("v3 backend json does not contain 'Group' member");
        return false;
      }

      mGroup = respDoc.get<std::string>("Group");

      mGroupID = crypto::FNV(mGroup);

      return true;
    }

    auto Backend::updateCycle(const volatile bool& handle) -> bool
    {
      while (handle) {
        if (!update(false)) {
          Log("failed to update relay");
          return false;
        }

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
      util::JSON reqDoc;
      if (!buildUpdateJSON(reqDoc, shuttingDown)) {
        Log("could not build v3 update json");
        return false;
      }

      BackendRequest request;
      {
        request.Type = core::packets::Type::V3UpdateRequest;
      }
      BackendResponse response;
      util::JSON respDoc;

      mState = ResponseState::Update;
      auto [ok, err] = sendJSONRecvJSON(request, reqDoc, response, respDoc);
      if (!ok) {
        Log(err);
      }

      size_t count = 0;
      std::array<core::V3Relay, MAX_RELAYS> incoming{};

      bool allValid = true;
      auto relays = respDoc.get<util::JSON>("PingTargets");
      if (relays.isArray()) {
        // 'return' functions like 'continue' within the lambda
        relays.foreach([this, &allValid, &count, &incoming](rapidjson::Value& relayData) {
          if (!relayData.HasMember("Id")) {
            Log("ping targets missing 'Id'");
            allValid = false;
            return;
          }

          auto idMember = std::move(relayData["Id"]);
          if (idMember.GetType() != rapidjson::Type::kNumberType) {
            Log("id from ping not number type");
            allValid = false;
            return;
          }

          auto id = idMember.GetUint64();

          // mustn't ping thyself
          if (id == this->mRelayID) {
            return;
          }

          if (!relayData.HasMember("Address")) {
            Log("ping data missing member 'Address' for relay id: ", id);
            allValid = false;
            return;
          }

          auto addrMember = std::move(relayData["Address"]);
          if (addrMember.GetType() != rapidjson::Type::kStringType) {
            Log("relay address is not a string in ping data for relay with id: ", id);
            allValid = false;
            return;
          }

          std::string b64Address = addrMember.GetString();

          auto tokenMember = std::move(relayData["PingToken"]);
          if (tokenMember.GetType() != rapidjson::Type::kStringType) {
            Log("ping token not string type");
            allValid = false;
            return;
          }

          std::string b64Token = tokenMember.GetString();

          std::array<uint8_t, 64> addrBuff{};
          size_t len = encoding::base64::Decode(b64Address, addrBuff);
          std::string address(addrBuff.begin(), addrBuff.begin() + len);

          std::array<uint8_t, 48> token{};
          encoding::base64::Decode(b64Token, token);

          incoming[count].ID = id;
          if (!incoming[count].Addr.parse(address)) {
            Log("failed to parse address for relay '", id, "': ", address);
            allValid = false;
            return;
          }
          incoming[count].PingToken = token;

          count++;
        });

        if (count > MAX_RELAYS) {
          Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", count, '\n');
          return false;
        }

      } else if (relays.memberIs(util::JSON::Type::Null)) {
        LogDebug("no relays received from v3 backend, ping data is null");
      } else {
        Log("update ping data not array");
        // TODO how to handle
      }

      if (!allValid) {
        Log("some or all of the update ping data was invalid");
      }

      mRelayManager.update(count, incoming);

      return true;
    }

    /*
     *  {
     *    "RelayId": uint64,
     *    "Timestamp": uint64,
     *    "Signature": string,
     *  }
     */
    auto Backend::buildConfigJSON(util::JSON& doc) -> bool
    {
      doc.set(mRelayID, "RelayId");
      return signRequest(doc);
    }

    // clang-format off
    /*
     *  {
     *    "Timestamp": uint64,
     *    "Signature": string,
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
     *      "PublicKey": string, // base64 of the public key in firestore. Old relay generates it, need to resuse it here to make things compatable for route/continue tokens
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
        trafficStats.set(mSessions.size(), "SessionCount");

        doc.set(trafficStats, "TrafficStats");

        auto total = bytesPerSecInvalidRx + bytesPerSecPaidRx + bytesPerSecManagementRx + bytesPerSecMeasurementRx +
                     bytesPerSecPaidTx + bytesPerSecManagementTx + bytesPerSecManagementTx;

        double usage = 100.0;

        if (mSpeed > 0) {
          usage *= 8.0 * total / mSpeed;
        }

        doc.set(usage, "Usage");
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

      return signRequest(doc);
    }

    auto Backend::sendBinRecvJSON(
     BackendRequest& request, std::vector<uint8_t>& reqData, BackendResponse& response, util::JSON& respBuff)
     -> std::tuple<bool, std::string>
    {
      std::vector<uint8_t> completeResponse;
      bool done = false;
      unsigned int attempts = 0;

      do {
        net::Address masterAddr;
        if (!masterAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
          return {false, "Could not resolve the v3 backend hostname to an ip address"};
        }

        LogDebug("sending ", request.Type, " request, attempts ", attempts);
        request.At = mClock.elapsed<util::Second>();
        bool sendSuccess = packet_send(mSocket, masterAddr, mToken, reqData, request);

        attempts++;

        // wait a second for the response to come in or if send failed
        std::this_thread::sleep_for(1s);

        if (!sendSuccess) {
          Log("failed to send v3 packet");
          continue;
        }

        // receive response(s) if it exists, if not resend
        while (mReceiver.hasItems()) {
          core::GenericPacket<> pkt;
          mReceiver.recv(pkt);
          // this will return true once all the fragments have been received
          done = packet_recv(pkt, request, response, completeResponse);
        }
      } while (!done && !mSocket.closed() && !mReceiver.closed() && attempts < 60);

      response.At = mClock.elapsed<util::Second>();

      if (mSocket.closed() || mReceiver.closed() || attempts == 60) {
        std::stringstream ss;
        ss << "could not send request, attempts: " << attempts;
        return {false, ss.str()};
      }

      auto [ok, err] = this->buildCompleteResponse(completeResponse, respBuff);
      if (!ok) {
        return {false, err};
      }

      return {true, ""};
    }

    auto Backend::sendJSONRecvJSON(
     BackendRequest& request, util::JSON& reqData, BackendResponse& response, util::JSON& respBuff)
     -> std::tuple<bool, std::string>
    {
      std::vector<uint8_t> completeResponse(0);

      bool done = false;
      unsigned int attempts = 0;
      do {
        net::Address masterAddr;
        if (!masterAddr.resolve(mEnv.RelayV3BackendHostname, mEnv.RelayV3BackendPort)) {
          return {false, "Could not resolve the v3 backend hostname to an ip address"};
        }
        signRequest(reqData);

        auto jsonStr = reqData.toString();
        std::vector<uint8_t> requestBuffer(jsonStr.begin(), jsonStr.end());

        LogDebug("sending a ", request.Type, ", attempts ", attempts, ", json: ", reqData.toPrettyString());
        request.At = mClock.elapsed<util::Second>();
        bool sendSuccess = packet_send(mSocket, masterAddr, mToken, requestBuffer, request);

        attempts++;

        // wait a second for the response to come in or if send failed
        std::this_thread::sleep_for(1s);

        if (!sendSuccess) {
          Log("failed to send v3 packet");
          continue;
        }

        // receive response(s) if it exists, if not resend
        while (mReceiver.hasItems()) {
          core::GenericPacket<> pkt;
          mReceiver.recv(pkt);
          // this will return true once all the fragments have been received
          done = packet_recv(pkt, request, response, completeResponse);
        }
      } while (!done && !mSocket.closed() && !mReceiver.closed() && mShouldCommunicate && attempts < 60);

      response.At = mClock.elapsed<util::Second>();

      if (mSocket.closed() || mReceiver.closed() || !mShouldCommunicate || attempts == 60) {
        std::stringstream ss;
        ss << "could not send request, attempts: " << attempts;
        return {false, ss.str()};
      }

      auto [ok, err] = this->buildCompleteResponse(completeResponse, respBuff);
      if (!ok) {
        return {false, err};
      }

      LogDebug("received v3: ", respBuff.toPrettyString());

      return {true, ""};
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

      int result = inflateInit(&z);
      if (result != Z_OK) {
        return {false, "failed to decompress master UDP packet: inflateInit failed"};
      }

      result = inflate(&z, Z_NO_FLUSH);
      if (result != Z_STREAM_END) {
        std::stringstream ss;
        ss << "failed to decompress master UDP packet: inflate failed, result is " << result;
        return {false, ss.str()};
      }

      result = inflateEnd(&z);
      if (result != Z_OK) {
        return {false, "failed to decompress master UDP packet: inflateEnd failed"};
      }

      int bytes = int(MaxPayload - z.avail_out);
      if (bytes == 0) {
        return {false, "failed to decompress master UDP packet: not enough buffer space"};
      }

      if (!doc.parse(buffer)) {
        std::stringstream ss;
        std::string strbuff(buffer.begin(), buffer.end());
        ss << "failed to parse json response, looks like: " << strbuff;
        return {false, ss.str()};
      }

      return {true, ""};
    }

    auto Backend::signRequest(util::JSON& doc) -> bool
    {
      // timestamp and signature

      auto ts = timestamp();

      doc.set(ts, "Timestamp");

      std::array<uint8_t, crypto_sign_BYTES> signature{};
      unsigned long long len = 0;
      crypto_sign_detached(signature.data(), &len, reinterpret_cast<uint8_t*>(&ts), sizeof(ts), mKeychain.UpdateKey.data());

      if (len != crypto_sign_BYTES) {
        LogDebug("failed to sign packet, length is ", len);
        return false;
      }

      std::array<char, signature.size() * 2> b64Sig{};
      size_t sigLen = encoding::base64::Encode(signature, b64Sig);

      doc.set(std::string(b64Sig.begin(), b64Sig.begin() + sigLen), "Signature");
      return true;
    }

    // one second resolution
    auto Backend::timestamp() -> uint64_t
    {
      return (mInitTimestamp + mClock.unixTime<util::Nanosecond>() - mInitReceived) / OneSecInNanos;
    }
  }  // namespace v3
}  // namespace legacy
