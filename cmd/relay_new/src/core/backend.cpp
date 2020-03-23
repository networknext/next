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
   RelayManager& relayManager)
   : mHostname(hostname), mAddressStr(address), mKeychain(keychain), mRouterInfo(routerInfo), mRelayManager(relayManager)
  {}

  bool Backend::init()
  {
    // Cache the base64 version of the relay public key for updating
    // TODO pass this in instead from the env var after init is verified, here for debugging reasons
    {
      std::vector<char> b64RelayPublicKey(mKeychain.RelayPublicKey.size() * 2);
      auto len = encoding::base64::Encode(mKeychain.RelayPublicKey, b64RelayPublicKey);
      if (len < mKeychain.RelayPublicKey.size()) {
        Log("failed to cache relay public key to base64");
        return false;
      }
      mRelayPublicKeyBase64 = std::string(b64RelayPublicKey.begin(), b64RelayPublicKey.begin() + len);
    }

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

    if (!net::CurlWrapper::SendTo(mHostname, "/relay_init", msg, resp)) {
      Log("curl request failed in init");
      return false;
    }

    jsonStr = std::string(resp.begin(), resp.end());
    if (!doc.parse(jsonStr)) {
      Log("could not parse json response in init");
      return false;
    }

    LogDebug("init response: \n", doc.toPrettyString());

    if (!doc.memberExists("version")) {
      Log("resposne json missing member 'version'");
      return false;
    }

    if (!doc.memberExists("timestamp")) {
      Log("response json missing member 'timestamp'");
      return false;
    }

    LogDebug("extracting version");
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

    LogDebug("extracting timestamp");
    if (doc.memberType("timestamp") == rapidjson::Type::kNumberType) {
      // for old relay compat the router sends this back in millis, so turn back to seconds
      mRouterInfo.InitalizeTimeInSeconds = doc.get<uint64_t>("timestamp") / 1000;
    } else {
      Log("init timestamp not a number");
      return false;
    }

    LogDebug("timestamp extracted");

    return true;
  }

  bool Backend::update(uint64_t bytesReceived)
  {
    LogDebug("updating relay");
    util::JSON doc;
    LogDebug("current doc: ", doc.toPrettyString());
    {
      doc.set(UpdateRequestVersion, "version");
      doc.set(mAddressStr, "relay_address");
      doc.set(mRelayPublicKeyBase64, "Metadata", "PublicKey");

      // Traffic stats
      {
        LogDebug("setting traffic stats");
        util::JSON trafficStats;
        trafficStats.set(bytesReceived, "BytesMeasurementRx");
        LogDebug("current ts: ", trafficStats.toPrettyString());

        doc.set(trafficStats, "TrafficStats");
      }
      LogDebug("current doc: ", doc.toPrettyString());

      // Ping stats
      LogDebug("setting ping stats");
      {
        core::RelayStats stats;
        mRelayManager.getStats(stats);
        util::JSON pingStats;
        pingStats.setArray();

        for (unsigned int i = 0; i < stats.NumRelays; ++i) {
          util::JSON stat;
          stat.set(stats.IDs[i], "RelayId");
          stat.set(stats.RTT[i], "RTT");
          stat.set(stats.Jitter[i], "Jitter");
          stat.set(stats.PacketLoss[i], "PacketLoss");
          if (!pingStats.push(stat)) {
            Log("ping stats not array! can't update!");
            return false;
          }
        }

        doc.set(pingStats, "PingStats");
      }
      LogDebug("current doc: ", doc.toPrettyString());
    }

    LogDebug("building msg for backend");

    std::string jsonStr = doc.toString();

    LogDebug("to string'ed");

    std::vector<uint8_t> msg(jsonStr.begin(), jsonStr.end());
    std::vector<uint8_t> resp;

    LogDebug("sending msg");
    if (!net::CurlWrapper::SendTo(mHostname, "/relay_update", msg, resp)) {
      Log("curl request failed in update");
      return false;
    }

    LogDebug("parsing response");

    jsonStr = std::string(resp.begin(), resp.end());
    if (!doc.parse(jsonStr)) {
      Log("could not parse json response in update");
      return false;
    }

    LogDebug("extracting version");
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

    LogDebug("extracting ping data");
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

        count++;
      });

      if (count > MAX_RELAYS) {
        Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", count, '\n');
        return RELAY_ERROR;
      }

      LogDebug("Updating relay manager");
      mRelayManager.update(count, relayIDs, relayAddresses);
      LogDebug("Updated relay manager");
    } else if (relays.memberType() == rapidjson::Type::kNullType) {
      Log("no relays received from backend, ping data is null");
    } else {
      Log("update ping data not array: rapidjson type value = ", relays.memberType());
      return false;
    }

    if (!allValid) {
      Log("some or all of the update ping data was invalid");
    }

    return true;
  }
}  // namespace core
