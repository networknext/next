#include "includes.h"
#include "backend.hpp"

#include "crypto/bytes.hpp"
#include "util/logger.hpp"
#include "util/json.hpp"
#include "encoding/base64.hpp"
#include "net/curl.hpp"

namespace
{
  const uint32_t InitRequestMagic = 0x9083708f;

  const uint32_t InitRequestVersion = 0;
  const uint32_t InitResponseVersion = 0;

  const uint32_t UpdateRequestVersion = 0;
  const uint32_t UpdateResponseVersion = 0;
}  // namespace

namespace core
{
  Backend::Backend(
   std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& routerInfo,
   RelayManager& relayManager,
   std::string base64RelayPublicKey)
   : mHostname(hostname), mAddressStr(address), mKeychain(keychain), mRouterInfo(routerInfo), mRelayManager(relayManager), mBase64RelayPublicKey(base64RelayPublicKey)
  {}

  bool Backend::init()
  {
    std::string base64NonceStr;
    std::string base64TokenStr;
    {
      // Nonce
      std::array<uint8_t, crypto_box_NONCEBYTES> nonce = {};
      {
        crypto::CreateNonceBytes(nonce);
        std::vector<char> b64Nonce(nonce.size() * 2);

        auto len = encoding::base64::Encode(nonce, b64Nonce);
        if (len < nonce.size()) {
          Log("failed to encode base64 nonce for init");
          return false;
        }
        base64NonceStr = std::string(b64Nonce.begin(), b64Nonce.begin() + len);
      }

      // Token
      {
        std::array<uint8_t, RELAY_TOKEN_BYTES> token = {};
        crypto::RandomBytes(token, token.size());
        std::array<uint8_t, RELAY_TOKEN_BYTES + crypto_box_MACBYTES> encryptedToken = {};
        std::vector<char> b64EncryptedToken(encryptedToken.size() * 2);

        if (
         crypto_box_easy(
          encryptedToken.data(),
          token.data(),
          token.size(),
          nonce.data(),
          mKeychain.RouterPublicKey.data(),
          mKeychain.RelayPrivateKey.data()) != 0) {
          Log("failed to encrypt init token");
          return false;
        }

        auto len = encoding::base64::Encode(encryptedToken, b64EncryptedToken);
        if (len < encryptedToken.size()) {
          Log("failed to encode base64 token for init");
          return false;
        }

        base64TokenStr = std::string(b64EncryptedToken.begin(), b64EncryptedToken.begin() + len);
      }
    }

    util::JSON doc;
    doc.set(InitRequestMagic, "magic_request_protection");
    doc.set(InitRequestVersion, "version");
    doc.set(mAddressStr, "relay_address");
    doc.set(base64NonceStr, "nonce");
    doc.set(base64TokenStr, "encrypted_token");

    std::string jsonStr = doc.toString();
    std::vector<uint8_t> msg(jsonStr.begin(), jsonStr.end());
    std::vector<uint8_t> resp;

    LogDebug("init request: ", doc.toPrettyString());

    if (!net::CurlWrapper::SendTo(mHostname, "/relay_init", msg, resp)) {
      Log("curl request failed in init");
      return false;
    }

    jsonStr = std::string(resp.begin(), resp.end());
    if (!doc.parse(jsonStr)) {
      Log("could not parse json response in init");
      return false;
    }

    LogDebug("init response: ", doc.toPrettyString());

    if (!doc.memberExists("version")) {
      Log("resposne json missing member 'version'");
      return false;
    }

    if (!doc.memberExists("timestamp")) {
      Log("response json missing member 'timestamp'");
      return false;
    }

    uint32_t version;
    if (doc.memberType("version") == rapidjson::Type::kNumberType) {
      version = doc.get<uint32_t>("version");
    } else {
      Log("init version response not a number");
      return false;
    }

    if (version != InitResponseVersion) {
      Log("error: bad relay init response version. expected ", InitResponseVersion, ", got ", version);
      return false;
    }

    if (doc.memberType("timestamp") == rapidjson::Type::kNumberType) {
      // for old relay compat the router sends this back in millis, so turn back to seconds
      mRouterInfo.InitalizeTimeInSeconds = doc.get<uint64_t>("timestamp") / 1000;
    } else {
      Log("init timestamp not a number");
      return false;
    }

    return true;
  }

  bool Backend::update(uint64_t bytesReceived)
  {
    // TODO once the other stats are finally added, pull out the json parts that are alwasy the same, no sense rebuilding those parts of the document
    util::JSON doc;
    {
      doc.set(UpdateRequestVersion, "version");
      doc.set(mAddressStr, "relay_address");
      doc.set(mBase64RelayPublicKey, "Metadata", "PublicKey");

      // Traffic stats
      {
        util::JSON trafficStats;
        trafficStats.set(bytesReceived, "BytesMeasurementRx");

        doc.set(trafficStats, "TrafficStats");
      }

      // Ping stats
      {
        core::RelayStats stats;
        mRelayManager.getStats(stats);
        util::JSON pingStats;
        pingStats.setArray();

        auto& allocator = doc.internal().GetAllocator();
        for (unsigned int i = 0; i < stats.NumRelays; ++i) {
          rapidjson::Value obj;
          rapidjson::Value stat;
          obj.SetObject();

          stat.Set(stats.IDs[i]);
          obj.AddMember("RelayId", stat, allocator);

          stat.Set(stats.RTT[i]);
          obj.AddMember("RTT", stat, allocator);

          stat.Set(stats.Jitter[i]);
          obj.AddMember("Jitter", stat, allocator);

          stat.Set(stats.PacketLoss[i]);
          obj.AddMember("PacketLoss", stat, allocator);

          if (!pingStats.push(obj)) {
            Log("ping stats not array! can't update!");
            return false;
          }
        }

        doc.set(pingStats, "PingStats");
      }
    }

    std::string jsonStr = doc.toString();

    std::vector<uint8_t> msg(jsonStr.begin(), jsonStr.end());
    std::vector<uint8_t> resp;

    LogDebug("update request: ", doc.toPrettyString());
    if (!net::CurlWrapper::SendTo(mHostname, "/relay_update", msg, resp)) {
      Log("curl request failed in update");
      return false;
    }

    jsonStr = std::string(resp.begin(), resp.end());
    if (!doc.parse(jsonStr)) {
      Log("could not parse json response in update");
      return false;
    }

    LogDebug("update response: ", doc.toPrettyString());

    auto version = doc.get<uint32_t>("version");
    if (doc.memberType("version") == rapidjson::Type::kNumberType) {
      if (version != UpdateResponseVersion) {
        Log("error: bad relay version response version. expected ", UpdateResponseVersion, ", got ", version);
        return false;
      }
    } else {
      Log("update version not number");
      return false;
    }

    bool allValid = true;
    auto relays = doc.get<util::JSON>("ping_data");
    if (relays.isArray()) {
      size_t count = 0;
      std::array<uint64_t, MAX_RELAYS> relayIDs;
      std::array<net::Address, MAX_RELAYS> relayAddresses;
      relays.foreach([&allValid, &count, &relayIDs, &relayAddresses](rapidjson::Value& relayData) {
        if (!relayData.HasMember("relay_id")) {
          Log("ping data missing 'relay_id'");
          allValid = false;
          return;
        }

        auto idMember = std::move(relayData["relay_id"]);
        if (idMember.GetType() != rapidjson::Type::kNumberType) {
          Log("id from ping not number type");
          allValid = false;
          return;
        }

        auto id = idMember.GetUint64();

        if (!relayData.HasMember("relay_address")) {
          Log("ping data missing member 'relay_address' for relay id: ", id);
          allValid = false;
          return;
        }

        auto addrMember = std::move(relayData["relay_address"]);
        if (addrMember.GetType() != rapidjson::Type::kStringType) {
          Log("relay address is not a string in ping data for relay with id: ", id);
          allValid = false;
          return;
        }

        std::string address = addrMember.GetString();

        relayIDs[count] = id;
        if (!relayAddresses[count].parse(address)) {
          Log("failed to parse address for relay '", id, "': ", address);
          allValid = false;
          return;
        }

        LogDebug("got id: ", id);
        LogDebug("got address ", relayAddresses[count]);

        count++;
      });

      if (count > MAX_RELAYS) {
        Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", count, '\n');
        return false;
      }

      mRelayManager.update(count, relayIDs, relayAddresses);
    } else if (relays.memberType() == rapidjson::Type::kNullType) {
      Log("no relays received from backend, ping data is null");
    } else {
      Log("update ping data not array: rapidjson type value = ", relays.memberType());
      // TODO how to handle
    }

    if (!allValid) {
      Log("some or all of the update ping data was invalid");
    }

    return true;
  }
}  // namespace core
