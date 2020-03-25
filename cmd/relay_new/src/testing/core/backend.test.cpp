#include "includes.h"
#include "testing/test.hpp"
#include "core/backend.hpp"

namespace
{
  const unsigned int Base64NonceLength = 32;
  const unsigned int Base64EncryptedTokenLength = 64;
}  // namespace

Test(core_backend_init_valid)
{
  std::string backendHostname = "http://totally-real-backend.com";
  std::string relayAddr = "127.0.0.1:12345";
  crypto::Keychain keychain;
  core::RouterInfo routerInfo;
  util::Clock clock;
  core::RelayManager manager(clock);
  std::string base64RelayPublicKey = "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=";
  std::string base64RelayPrivateKey = "lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=";
  std::string base64RouterPublicKey = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=";

  check(keychain.parse(base64RelayPublicKey, base64RelayPrivateKey, base64RouterPublicKey));

  core::Backend<testing::StubbedCurlWrapper> backend(
   backendHostname, relayAddr, keychain, routerInfo, manager, base64RelayPublicKey);

  testing::StubbedCurlWrapper::Response = R"({
    "version": 0,
    "timestamp": 123456789
  })";

  check(backend.init());

  check(testing::StubbedCurlWrapper::Hostname == backendHostname);
  check(testing::StubbedCurlWrapper::Endpoint == "/relay_init");
  check(routerInfo.InitalizeTimeInSeconds == 123456789 / 1000);

  util::JSON doc;

  check(doc.parse(testing::StubbedCurlWrapper::Request));

  check(doc.get<uint32_t>("magic_request_protection") == core::InitRequestMagic);
  check(doc.get<uint32_t>("version") == core::InitRequestVersion);
  check(doc.get<std::string>("relay_address") == relayAddr);

  // gonna be random, so all that can be done is asserting the length
  check(doc.get<std::string>("nonce").length() == Base64NonceLength);
  check(doc.get<std::string>("encrypted_token").length() == Base64EncryptedTokenLength);
}

Test(core_backend_update_valid)
{
  std::string backendHostname = "http://totally-real-backend.com";
  std::string relayAddr = "127.0.0.1:12345";
  crypto::Keychain keychain;
  core::RouterInfo routerInfo;
  util::Clock clock;
  core::RelayManager manager(clock);
  std::string base64RelayPublicKey = "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=";
  std::string base64RelayPrivateKey = "lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=";
  std::string base64RouterPublicKey = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=";

  // seed relay manager
  {
    const size_t numRelays = 1;
    std::array<uint64_t, MAX_RELAYS> ids;
    std::array<net::Address, MAX_RELAYS> addrs;
    ids[0] = 987654321;
    net::Address addr;
    addr.parse("127.0.0.1:12345");
    addrs[0] = addr;
    manager.update(numRelays, ids, addrs);

    for (auto i = 1; i <= 6; i++) {
      manager.processPong(addr, i);
    }
  }

  check(keychain.parse(base64RelayPublicKey, base64RelayPrivateKey, base64RouterPublicKey));

  core::Backend<testing::StubbedCurlWrapper> backend(
   backendHostname, relayAddr, keychain, routerInfo, manager, base64RelayPublicKey);

  testing::StubbedCurlWrapper::Response = R"({
     "version": 0,
     "ping_data": [
       {
         "relay_id": 135792468,
         "relay_address": "127.0.0.1:54321"
       },
       {
         "relay_id": 246813579,
         "relay_address": "127.0.0.1:13524"
       }
     ]
   })";

  const uint64_t bytesReceived = 10000000000;

  check(backend.update(bytesReceived));

  util::JSON doc;

  check(doc.parse(testing::StubbedCurlWrapper::Request));

  check(doc.get<uint32_t>("version") == 0);
  check(doc.get<std::string>("relay_address") == relayAddr);
  check(doc.get<std::string>("Metadata", "PublicKey") == base64RelayPublicKey);
  check(doc.get<uint64_t>("TrafficStats", "BytesMeasurementRx") == bytesReceived);

  auto pingStats = doc.get<util::JSON>("PingStats");

  check(pingStats.isArray());

  auto& value = pingStats[0];

  check(value.HasMember("RelayId"));
  check(value.HasMember("RTT"));
  check(value.HasMember("Jitter"));
  check(value.HasMember("PacketLoss"));

  auto& relayID = value["RelayId"];
  auto& rtt = value["RTT"];
  auto& jitter = value["Jitter"];
  auto& packetLoss = value["PacketLoss"];

  check(relayID.Get<uint64_t>() == 987654321);
  check(rtt.Get<float>() > 0.0f);
  check(jitter.Get<float>() == 0.0f);
  check(packetLoss.Get<float>() == 0.0f);

  std::array<core::PingData, MAX_RELAYS> pingData;
  auto count = manager.getPingData(pingData);

  check(count == 2);
  check(pingData[0].Addr.toString() == "127.0.0.1:54321");
  check(pingData[1].Addr.toString() == "127.0.0.1:13524");
}
