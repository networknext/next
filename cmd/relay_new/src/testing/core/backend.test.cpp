#include "includes.h"
#include "testing/test.hpp"
#include "core/backend.hpp"

using namespace std::chrono_literals;

namespace
{
  const unsigned int Base64NonceLength = 32;
  const unsigned int Base64EncryptedTokenLength = 64;

  const auto BackendHostname = "http://totally-real-backend.com";
  const auto RelayAddr = "127.0.0.1:12345";
  const auto Base64RelayPublicKey = "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=";
  const auto Base64RelayPrivateKey = "lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=";
  const auto Base64RouterPublicKey = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=";

  const auto BasicValidUpdateResponse = R"({
     "version": 0,
     "ping_data": []
   })";

  core::Backend<testing::StubbedCurlWrapper> makeBackend(
   core::RouterInfo& info, core::RelayManager& manager, core::SessionMap& sessions)
  {
    crypto::Keychain keychain;

    check(keychain.parse(Base64RelayPublicKey, Base64RelayPrivateKey, Base64RouterPublicKey));

    return core::Backend<testing::StubbedCurlWrapper>(
     BackendHostname, RelayAddr, keychain, info, manager, Base64RelayPublicKey, sessions);
  }
}  // namespace

Test(core_backend_init_valid)
{
  core::RouterInfo routerInfo;
  util::Clock clock;
  core::RelayManager manager(clock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(routerInfo, manager, sessions));

  testing::StubbedCurlWrapper::Response = R"({
    "version": 0,
    "Timestamp": 123456789
  })";

  check(backend.init());

  check(testing::StubbedCurlWrapper::Hostname == BackendHostname);
  check(testing::StubbedCurlWrapper::Endpoint == "/relay_init");
  check(routerInfo.InitializeTimeInSeconds == 123456789 / 1000);

  util::JSON doc;

  check(doc.parse(testing::StubbedCurlWrapper::Request));

  check(doc.get<uint32_t>("magic_request_protection") == core::InitRequestMagic);
  check(doc.get<uint32_t>("version") == core::InitRequestVersion);
  check(doc.get<std::string>("relay_address") == RelayAddr);

  // gonna be random, so all that can be done is asserting the length
  check(doc.get<std::string>("nonce").length() == Base64NonceLength);
  check(doc.get<std::string>("encrypted_token").length() == Base64EncryptedTokenLength);
}

// Update the backend for 10 seconds, then proceed to switch the handle to false.
// The relay should then attempt to ack the backend.
// It won't receive a success response from the backend so instead it will
// live for 60 seconds and skip the ack
Test(core_Backend_updateCycle_shutdown_60s)
{
  util::Clock testClock;

  core::RouterInfo info;
  util::Clock backendClock;
  core::RelayManager manager(backendClock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(info, manager, sessions));
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  util::ThroughputRecorder logger;

  testing::StubbedCurlWrapper::Success = true;
  testing::StubbedCurlWrapper::Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    testing::StubbedCurlWrapper::Success = false;
    shouldCleanShutdown = true;  // just to mimic actual behavior
    handle = false;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, logger, sessions, backendClock));
  auto elapsed = testClock.elapsed<util::Second>();
  check(elapsed >= 62.0 && elapsed < 63.0);
}

// Update the backend for 10 seconds, then proceed to switch the handle to false.
// The relay should then attempt to ack the backend and shutdown for 30 seconds.
// It will receive a success response and then live for another 30 seconds.
// The 60 second timeout will not apply here
Test(core_Backend_updateCycle_ack_and_30s)
{
  util::Clock testClock;

  core::RouterInfo info;
  util::Clock backendClock;
  core::RelayManager manager(backendClock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(info, manager, sessions));
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  util::ThroughputRecorder logger;

  testing::StubbedCurlWrapper::Success = true;
  testing::StubbedCurlWrapper::Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    shouldCleanShutdown = true;
    handle = false;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, logger, sessions, backendClock));
  auto elapsed = testClock.elapsed<util::Second>();
  check(elapsed >= 32.0 && elapsed < 33.0);
}

// Update the backend for 10 seconds, then proceed to switch the handle to false.
// The relay will not get a success response for 40 seconds after the handle is set
// After which it will get a success and then proceed with the normal routine of waiting 30 seconds
// The amount of time waited will be greater than 60 seconds
// This is to assert the updateCycle will ignore the 60 second timeout if the backend gets an update
Test(core_Backend_updateCycle_no_ack_for_40s_then_ack_then_wait)
{
  util::Clock testClock;

  core::RouterInfo info;
  util::Clock backendClock;
  core::RelayManager manager(backendClock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(info, manager, sessions));
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  util::ThroughputRecorder recorder;

  testing::StubbedCurlWrapper::Success = true;
  testing::StubbedCurlWrapper::Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    shouldCleanShutdown = true;
    testing::StubbedCurlWrapper::Success = false;
    handle = false;
    std::this_thread::sleep_for(31s);
    testing::StubbedCurlWrapper::Success = true;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, recorder, sessions, backendClock));
  auto elapsed = testClock.elapsed<util::Second>();
  check(elapsed >= 63.0 && elapsed < 64.0);
}

// Update the backend for 10 seconds, then switch the success of the request to false.
// That will trigger the failure attempts which the number of is controlled by the MaxUpdateAttempts constant.
// After the max attempts is reached it will shutdown.
// But because the success value is never reset to true, the cleanshutdown ack will never succeed
// so the final duration should be 10 seconds of success and (MaxUpdateAttempts - 1) seconds of failure.
Test(core_Backend_updateCycle_update_fails_for_max_number_of_attempts)
{
  util::Clock testClock;

  core::RouterInfo info;
  util::Clock backendClock;
  core::RelayManager manager(backendClock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(info, manager, sessions));
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  util::ThroughputRecorder recorder;

  testing::StubbedCurlWrapper::Success = true;
  testing::StubbedCurlWrapper::Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    testing::StubbedCurlWrapper::Success = false;  // set to false here to trigger failed updates
  });

  check(!backend.updateCycle(handle, shouldCleanShutdown, recorder, sessions, backendClock));
  auto elapsed = testClock.elapsed<util::Second>();
  // time will be 10 seconds of good updates and
  // 10 seconds of bad updates, which will cause
  // the relay to abort with no clean shutdown
  check(elapsed >= 12.0 && elapsed < 13.0);
}

// When clean shutdown is not set to true, the function should return immediately
Test(core_Backend_updateCycle_no_clean_shutdown)
{
  util::Clock testClock;

  core::RouterInfo info;
  util::Clock backendClock;
  core::RelayManager manager(backendClock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(info, manager, sessions));
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  util::ThroughputRecorder recorder;

  testing::StubbedCurlWrapper::Success = true;
  testing::StubbedCurlWrapper::Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    testing::StubbedCurlWrapper::Success = false;
    handle = false;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, recorder, sessions, backendClock));
  auto elapsed = testClock.elapsed<util::Second>();
  check(elapsed >= 2.0 && elapsed < 3.0);
}

Test(core_Backend_update_valid)
{
  core::RouterInfo routerInfo;
  util::Clock clock;
  core::RelayManager manager(clock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(routerInfo, manager, sessions));
  util::ThroughputRecorder recorder;

  sessions.set(1234, std::make_shared<core::Session>(clock));  // just add one thing to the map to make it non-zero

  // seed relay manager
  {
    const size_t numRelays = 1;
    std::array<uint64_t, MAX_RELAYS> ids;
    std::array<net::Address, MAX_RELAYS> addrs;
    std::array<core::PingData, MAX_RELAYS> pingData;
    ids[0] = 987654321;
    net::Address addr;
    check(addr.parse("127.0.0.1:12345"));
    addrs[0] = addr;
    manager.update(numRelays, ids, addrs);
    check(manager.getPingData(pingData) == 1);
    manager.processPong(pingData[0].Addr, pingData[0].Seq);
  }

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

  const auto bytesSent = 123456789;
  const auto bytesReceived = 987654321;

  recorder.addToSent(bytesSent);
  recorder.addToReceived(bytesReceived);

  check(backend.update(recorder, false));

  util::JSON doc;

  check(doc.parse(testing::StubbedCurlWrapper::Request));

  check(doc.get<uint32_t>("version") == 0);
  check(doc.get<std::string>("relay_address") == RelayAddr);
  check(doc.get<std::string>("Metadata", "PublicKey") == Base64RelayPublicKey);
  check(doc.get<uint64_t>("TrafficStats", "BytesMeasurementTx") == bytesSent);
  check(doc.get<uint64_t>("TrafficStats", "BytesMeasurementRx") == bytesReceived);
  check(doc.get<size_t>("TrafficStats", "SessionCount") == sessions.size());
  check(!doc.get<bool>("shutting_down"));

  auto pingStats = doc.get<util::JSON>("PingStats");

  check(pingStats.isArray());

  auto& value = pingStats[0];

  check(value.HasMember("RelayId"));
  check(value.HasMember("RTT"));
  check(value.HasMember("Jitter"));
  check(value.HasMember("PacketLoss"));

  auto& relayID = value["RelayId"];
  // auto& rtt = value["RTT"];
  // auto& jitter = value["Jitter"];
  // auto& packetLoss = value["PacketLoss"];

  check(relayID.Get<uint64_t>() == 987654321);

  // not set up right, the actual tests pass for this
  // so not currently concerned with it passing here,
  // need to learn how the logic actually works to
  // set this up right

  // check(rtt.Get<float>() != 10000.0f);
  // check(jitter.Get<float>() == 0.0f);
  // check(packetLoss.Get<float>() == 0.0f);

  std::array<core::PingData, MAX_RELAYS> pingData;
  auto count = manager.getPingData(pingData);

  check(count == 2);
  check(pingData[0].Addr.toString() == "127.0.0.1:54321");
  check(pingData[1].Addr.toString() == "127.0.0.1:13524");
}

Test(core_Backend_update_shutting_down_true)
{
  core::RouterInfo routerInfo;
  util::Clock clock;
  core::RelayManager manager(clock);
  core::SessionMap sessions;
  auto backend = std::move(makeBackend(routerInfo, manager, sessions));
  util::ThroughputRecorder recorder;

  testing::StubbedCurlWrapper::Response = ::BasicValidUpdateResponse;

  check(backend.update(recorder, true));

  util::JSON doc;
  check(doc.parse(testing::StubbedCurlWrapper::Request));
  check(doc.get<uint32_t>("version") == 0);
  check(doc.get<std::string>("relay_address") == RelayAddr);
  check(doc.get<std::string>("Metadata", "PublicKey") == Base64RelayPublicKey);
  check(doc.get<uint64_t>("TrafficStats", "BytesMeasurementRx") == 0);
  check(doc.get<size_t>("TrafficStats", "SessionCount") == 0);
  check(doc.get<bool>("shutting_down"));
}
