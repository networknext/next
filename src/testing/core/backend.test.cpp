#include "includes.h"
#include "testing/test.hpp"
#include "core/backend.hpp"
#include "testing/mocks.hpp"

#define CRYPTO_HELPERS
#include "testing/helpers.hpp"

using namespace std::chrono_literals;

using core::Backend;
using core::InitRequest;
using core::InitRequestMagic;
using core::InitRequestVersion;
using core::InitResponse;
using core::PingData;
using core::RelayManager;
using core::RelayPingInfo;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::UpdateRequest;
using core::UpdateResponse;
using net::Address;
using util::Clock;
using util::Second;
using util::ThroughputRecorder;

namespace
{
  const unsigned int Base64NonceLength = 32;
  const unsigned int Base64EncryptedTokenLength = 64;

  const std::string BackendHostname = "http://totally-real-backend.com";
  const auto RelayAddr = "127.0.0.1:12345";
  const crypto::Keychain Keychain = testing::make_keychain();

  const std::vector<uint8_t> BasicValidUpdateResponse = [] {
    InitResponse response = {
     .Version = 0,
     .Timestamp = 0,
     .PublicKey = {},
    };

    std::vector<uint8_t> buff(InitResponse::ByteSize);
    check(response.into(buff));
    return buff;
  }();

  auto makeInitResponse(uint32_t version, uint64_t timestamp, std::array<uint8_t, crypto::KEY_SIZE>& pk) -> std::vector<uint8_t>
  {
    std::vector<uint8_t> buff(InitResponse::ByteSize);
    InitResponse resp{
     .Version = version,
     .Timestamp = timestamp,
     .PublicKey = pk,
    };

    resp.into(buff);

    return buff;
  }
}  // namespace

Test(core_backend_init_valid)
{
  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  std::array<uint8_t, crypto::KEY_SIZE> pk{};
  testing::MockHttpClient client;
  client.Response = makeInitResponse(0, 123456789, pk);
  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  check(backend.init());

  check(client.Hostname == BackendHostname);
  check(client.Endpoint == "/relay_init");
  check(routerInfo.currentTime() >= 123456789 / 1000);

  InitRequest request;
  check(request.from(client.Request));

  check(request.Magic == InitRequestMagic);
  check(request.Version == InitRequestVersion);
  check(request.Address == RelayAddr);

  // can't check nonce or encrypted token since they're random
}

// Update the backend for 2 seconds, then proceed to switch the handle to false.
// The relay should then attempt to ack the backend.
// It won't receive a success response from the backend so instead it will
// live for 60 seconds and skip the ack
Test(core_Backend_updateCycle_shutdown_60s)
{
  Clock testClock;

  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  ThroughputRecorder logger;
  testing::MockHttpClient client;

  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  client.Success = true;
  client.Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    client.Success = false;
    shouldCleanShutdown = true;  // just to mimic actual behavior
    handle = false;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, logger, sessions));
  auto elapsed = testClock.elapsed<Second>();
  check(elapsed >= 62.0);
}

// Update the backend for 2 seconds, then proceed to switch the handle to false.
// The relay should then attempt to ack the backend and shutdown for 30 seconds.
// It will receive a success response and then live for another 30 seconds.
// The 60 second timeout will not apply here
Test(core_Backend_updateCycle_ack_and_30s)
{
  Clock testClock;

  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  ThroughputRecorder logger;
  testing::MockHttpClient client;

  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  client.Success = true;
  client.Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    shouldCleanShutdown = true;
    handle = false;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, logger, sessions));
  auto elapsed = testClock.elapsed<Second>();
  check(elapsed >= 32.0);
}

// Update the backend for 2 seconds, then proceed to switch the handle to false.
// The relay will not get a success response for 31 seconds after the handle is set.
// After which it will get a success and then proceed with the normal routine of waiting 30 seconds
// The amount of time waited will be greater than 60 seconds
// This is to assert the updateCycle will ignore the 60 second timeout if the backend gets an update
Test(core_Backend_updateCycle_no_ack_for_40s_then_ack_then_wait)
{
  Clock testClock;

  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  client.Success = true;
  client.Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    shouldCleanShutdown = true;
    client.Success = false;
    handle = false;
    std::this_thread::sleep_for(31s);
    client.Success = true;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, recorder, sessions));
  auto elapsed = testClock.elapsed<Second>();
  check(elapsed >= 63.0);
}

// Update the backend for 2 seconds, then switch the success of the request to false.
// That will trigger the failure attempts which the number of is controlled by the MaxUpdateAttempts constant.
// After the max attempts is reached it will shutdown.
// But because the success value is never reset to true, the cleanshutdown ack will never succeed
// so the final duration should be 2 seconds of success and (MaxUpdateAttempts - 1) seconds of failure.
Test(core_Backend_updateCycle_update_fails_for_max_number_of_attempts)
{
  Clock testClock;

  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  client.Success = true;
  client.Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    client.Success = false;  // set to false here to trigger failed updates
  });

  check(!backend.updateCycle(handle, shouldCleanShutdown, recorder, sessions));
  auto elapsed = testClock.elapsed<Second>();
  // time will be 2 seconds of good updates and
  // 10 seconds of bad updates, which will cause
  // the relay to abort with no clean shutdown
  check(elapsed >= 12.0).onFail([&elapsed] {
    std::cout << "elapsed: " << elapsed << '\n';
  });
}

// When clean shutdown is not set to true, the function should return immediately
Test(core_Backend_updateCycle_no_clean_shutdown)
{
  Clock testClock;

  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  volatile bool handle = true;
  volatile bool shouldCleanShutdown = false;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  client.Success = true;
  client.Response = BasicValidUpdateResponse;

  testClock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    client.Success = false;
    handle = false;
  });

  check(backend.updateCycle(handle, shouldCleanShutdown, recorder, sessions));
  auto elapsed = testClock.elapsed<Second>();
  check(elapsed >= 2.0);
}

Test(core_Backend_update_valid)
{
  Clock clock;
  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;
  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  sessions.set(1234, std::make_shared<Session>(routerInfo));  // just add one thing to the map to make it non-zero

  // seed relay manager
  {
    const size_t numRelays = 1;
    std::array<RelayPingInfo, MAX_RELAYS> incoming;
    std::array<PingData, MAX_RELAYS> pingData;
    incoming[0].ID = 987654321;
    Address addr;
    check(addr.parse("127.0.0.1:12345"));
    incoming[0].Addr = addr;
    manager.update(numRelays, incoming);
    check(manager.getPingData(pingData) == 1);
    manager.processPong(pingData[0].Addr, pingData[0].Seq);
  }

  recorder.UnknownRx.add(10);
  UpdateResponse response;
  response.Version = 0;
  response.Timestamp = 123456789;
  response.NumRelays = 2;

  {
    RelayPingInfo relay1, relay2;

    relay1.ID = 135792468;
    check(relay1.Addr.parse("127.0.0.1:54321"));

    relay2.ID = 246813579;
    check(relay2.Addr.parse("127.0.0.1:13524"));
    response.Relays = {
     relay1,
     relay2,
    };
  }

  client.Response.resize(response.size());
  response.into(client.Response);

  const auto outboundPing = 123456789;
  const auto pong = 987654321;

  recorder.OutboundPingTx.add(outboundPing);
  recorder.PongRx.add(pong);

  check(backend.update(recorder, false));

  // check the request
  {
    UpdateRequest request;
    check(request.from(client.Request));

    check(request.Version == 1);
    check(request.Address == RelayAddr);
    check(request.PublicKey == Keychain.RelayPublicKey);
    check(request.SessionCount == sessions.size());
    check(request.OutboundPingTx == outboundPing);
    check(request.RouteRequestRx == 0);
    check(request.RouteRequestTx == 0);
    check(request.RouteResponseRx == 0);
    check(request.RouteResponseTx == 0);
    check(request.ClientToServerRx == 0);
    check(request.ClientToServerTx == 0);
    check(request.ServerToClientRx == 0);
    check(request.ServerToClientTx == 0);
    check(request.InboundPingRx == 0);
    check(request.InboundPingTx == 0);
    check(request.PongRx == pong);
    check(request.SessionPingRx == 0);
    check(request.SessionPingTx == 0);
    check(request.SessionPongRx == 0);
    check(request.SessionPongTx == 0);
    check(request.ContinueRequestRx == 0);
    check(request.ContinueRequestTx == 0);
    check(request.ContinueResponseRx == 0);
    check(request.ContinueResponseTx == 0);
    check(request.NearPingRx == 0);
    check(request.NearPingTx == 0);
    check(request.UnknownRx == 10);
    check(request.ShuttingDown == false);
    check(request.PingStats.NumRelays == 1);
    check(request.RelayVersion == RELAY_VERSION);
  }

  // check that the response was processed
  {
    std::array<PingData, MAX_RELAYS> pingData;

    std::this_thread::sleep_for(1s);  // needed so that getPingData() will always return the right number
    auto count = manager.getPingData(pingData);

    check(count == 2).onFail([&] {
      std::cout << "count is " << count << '\n';
    });
    check(pingData[0].Addr.toString() == "127.0.0.1:54321");
    check(pingData[1].Addr.toString() == "127.0.0.1:13524");

    check(routerInfo.currentTime() >= 123456789).onFail([&] {
      std::cout << "info timestamp = " << routerInfo.currentTime() << '\n';
    });
  }
}

Test(core_Backend_update_shutting_down_true)
{
  Clock clock;
  RouterInfo routerInfo;
  RelayManager manager;
  SessionMap sessions;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BackendHostname, RelayAddr, Keychain, routerInfo, manager, Base64RelayPublicKey, sessions, client);

  client.Response = ::BasicValidUpdateResponse;

  check(backend.update(recorder, true));

  UpdateRequest request;
  check(request.from(client.Request));

  check(request.Version == 1);
  check(request.Address == RelayAddr);
  check(request.PublicKey == Keychain.RelayPublicKey);
  check(request.SessionCount == 0);
  check(request.OutboundPingTx == 0);
  check(request.RouteRequestRx == 0);
  check(request.RouteRequestTx == 0);
  check(request.RouteResponseRx == 0);
  check(request.RouteResponseTx == 0);
  check(request.ClientToServerRx == 0);
  check(request.ClientToServerTx == 0);
  check(request.ServerToClientRx == 0);
  check(request.ServerToClientTx == 0);
  check(request.InboundPingRx == 0);
  check(request.InboundPingTx == 0);
  check(request.PongRx == 0);
  check(request.SessionPingRx == 0);
  check(request.SessionPingTx == 0);
  check(request.SessionPongRx == 0);
  check(request.SessionPongTx == 0);
  check(request.ContinueRequestRx == 0);
  check(request.ContinueRequestTx == 0);
  check(request.ContinueResponseRx == 0);
  check(request.ContinueResponseTx == 0);
  check(request.NearPingRx == 0);
  check(request.NearPingTx == 0);
  check(request.UnknownRx == 0);
  check(request.ShuttingDown == true);
  check(request.PingStats.NumRelays == 0);
  check(request.RelayVersion == RELAY_VERSION);
}
