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
  Backend::Backend(std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& routerInfo,
   RelayManager& relayManager)
   : mHostname(hostname), mAddress(address), mKeychain(keychain), mRouterInfo(routerInfo), mRelayManager(relayManager)
  {}

  bool Backend::init()
  {
    std::string base64NonceStr;
    std::string base64TokenStr;
    {
      std::array<uint8_t, crypto_box_NONCEBYTES> nonce = {};
      crypto::CreateNonceBytes(nonce);
      base64NonceStr.resize(nonce.size() * 2);
      if (!encoding::base64::EncodeToString(nonce, base64NonceStr)) {
        Log("failed to encode base64 nonce for init");
        return false;
      }

      std::array<uint8_t, RELAY_TOKEN_BYTES> token = {};
      std::array<uint8_t, RELAY_TOKEN_BYTES + crypto_box_MACBYTES> encryptedToken = {};
      if (crypto_box_easy(encryptedToken.data(),
           token.data(),
           token.size(),
           nonce.data(),
           mKeychain.RouterPublicKey.data(),
           mKeychain.RelayPrivateKey.data()) != 0) {
        Log("failed to encrypt init token");
        return false;
      }

      base64TokenStr.resize(token.size() * 2);
      if (!encoding::base64::EncodeToString(encryptedToken, base64TokenStr)) {
        Log("failed to encode base64 token for init");
        return false;
      }
    }

    util::JSON doc;
    doc.set(InitRequestMagic, "magic_request_protection");
    doc.set(InitRequestVersion, "version");
    doc.set(mAddress, "relay_address");
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

    if (!doc.memberExists("version")) {
      Log("resposne json missing member 'version'");
      return false;
    }

    if (!doc.memberExists("timestamp")) {
      Log("response json missing member 'timestamp'");
      return false;
    }

    uint32_t version = doc.get<uint32_t>("version");

    if (version != InitResponseVersion) {
      Log("error: bad relay init response version. expected ", InitResponseVersion, ", got ", version);
      return false;
    }

    // for old relay compat the router sends this back in millis, so turn back to seconds
    mRouterInfo.InitalizeTimeInSeconds = doc.get<unsigned long>("timestamp") / 1000;

    // TODO for the sake of getting this done, putting it here for now but this should be done elsewhere
    if (!encoding::base64::EncodeToString(mKeychain.RelayPublicKey, mRelayPublicKeyBase64)) {
      Log("failed to cache relay public key");
      return false;
    }

    return true;
  }

  bool Backend::update(uint64_t bytesReceived)
  {
    util::JSON doc;
    {
      doc.set(UpdateRequestVersion, "version");
      doc.set(mAddress, "relay_address");
      doc.set(mRelayPublicKeyBase64, "Metadata", "PublicKey");

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

        for (unsigned int i = 0; i < stats.NumRelays; ++i) {
          util::JSON stat;
          stat.set(stats.IDs[i], "RelayId");
          stat.set(stats.RTT[i], "RTT");
          stat.set(stats.Jitter[i], "Jitter");
          stat.set(stats.PacketLoss[i], "PacketLoss");
          pingStats.push(stat);
        }

        doc.set(pingStats, "PingStats");
      }
    }

    std::string jsonStr = doc.toString();
    std::vector<uint8_t> msg(jsonStr.begin(), jsonStr.end());
    std::vector<uint8_t> resp;

    if (!net::CurlWrapper::SendTo(mHostname, "/relay_update", msg, resp)) {
      Log("curl request failed in update");
      return false;
    }

    jsonStr = std::string(resp.begin(), resp.end());
    if (!doc.parse(jsonStr)) {
      Log("could not parse json response in update");
      return false;
    }

    auto version = doc.get<uint32_t>("version");
    {
      if (version != UpdateResponseVersion) {
        Log("error: bad relay version response version. expected ", UpdateResponseVersion, ", got ", version);
        return false;
      }
    }

    auto relays = doc.get<util::JSON>("ping_data");
    {
      size_t count = 0;
      std::array<uint64_t, MAX_RELAYS> relayIDs;
      std::array<net::Address, MAX_RELAYS> relayAddresses;
      relays.foreach ([&count, &relayIDs, &relayAddresses](rapidjson::Value& relayData) {
        if (!relayData.HasMember("relay_id")) {
          Log("ping data missing 'relay_id'");
          return;
        }

        auto id = relayData["relay_id"].GetUint64();

        if (!relayData.HasMember("relay_address")) {
          Log("ping data missing member 'relay_address' for relay id: ", id);
          return;
        }

        std::string address = relayData["relay_address"].GetString();

        relayIDs[count] = id;
        if (!relayAddresses[count].parse(address)) {
          Log("failed to parse address for relay '", id, "': ", address);
          return;
        }

        count++;
      });

      if (count > MAX_RELAYS) {
        Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", count, '\n');
        return RELAY_ERROR;
      }

      mRelayManager.update(count, relayIDs, relayAddresses);
    }

    return true;
  }
}  // namespace core
