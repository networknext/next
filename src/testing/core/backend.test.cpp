#include "includes.h"
#include "testing/test.hpp"
#include "core/backend.hpp"
#include "testing/mocks.hpp"
#include "util/misc.hpp"

#define CRYPTO_HELPERS
#include "testing/helpers.hpp"

using namespace std::chrono_literals;

using core::Backend;
using core::INIT_REQUEST_MAGIC;
using core::INIT_REQUEST_VERSION;
using core::InitRequest;
using core::InitResponse;
using core::PingData;
using core::RelayManager;
using core::RelayPingInfo;
using core::RouterInfo;
using core::Session;
using core::SessionMap;
using core::UPDATE_REQUEST_VERSION;
using core::UpdateRequest;
using core::UpdateResponse;
using net::Address;
using util::Clock;
using util::Second;
using util::ThroughputRecorder;
using util::ThroughputStats;

namespace
{
  const unsigned int Base64NonceLength = 32;
  const unsigned int Base64EncryptedTokenLength = 64;

  const std::string BackendHostname = "http://totally-real-backend.com";
  const auto RelayAddr = "127.0.0.1:12345";
  const crypto::Keychain Keychain = testing::make_keychain();

  const std::vector<uint8_t> BasicValidUpdateResponse = [] {
    InitResponse response = {
     .version = 0,
     .timestamp = 0,
     .public_key = {},
    };

    std::vector<uint8_t> buff(InitResponse::SIZE_OF);
    check(response.into(buff));
    return buff;
  }();

  auto makeInitResponse(uint32_t version, uint64_t timestamp, std::array<uint8_t, crypto::KEY_SIZE>& pk) -> std::vector<uint8_t>
  {
    std::vector<uint8_t> buff(InitResponse::SIZE_OF);
    InitResponse resp{
     .version = version,
     .timestamp = timestamp,
     .public_key = pk,
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
  check(routerInfo.current_time() >= 123456789 / 1000);

  InitRequest request;
  check(request.from(client.Request));

  check(request.magic == INIT_REQUEST_MAGIC);
  check(request.version == INIT_REQUEST_VERSION);
  check(request.address == RelayAddr);
  check(request.relay_version == RELAY_VERSION);

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

  check(backend.update_loop(handle, shouldCleanShutdown, logger, sessions));
  auto elapsed = testClock.elapsed<Second>();
  check(elapsed >= 62.0).onFail([&] {
    std::cout << "elapsed time = " << elapsed << '\n';
  });
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

  check(backend.update_loop(handle, shouldCleanShutdown, logger, sessions));
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

  check(backend.update_loop(handle, shouldCleanShutdown, recorder, sessions));
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

  check(!backend.update_loop(handle, shouldCleanShutdown, recorder, sessions));
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

  check(backend.update_loop(handle, shouldCleanShutdown, recorder, sessions));
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
  crypto::RandomBytes(backend.update_token, backend.update_token.size());

  auto session = std::make_shared<Session>();
  session->kbps_up = 123;
  session->kbps_down = 456;
  sessions.set(1234, session);  // just add one thing to the map to make it non-zero

  // seed relay manager
  {
    const size_t numRelays = 1;
    std::array<RelayPingInfo, MAX_RELAYS> incoming;
    std::array<PingData, MAX_RELAYS> pingData;
    incoming[0].id = 987654321;
    Address addr;
    check(addr.parse("127.0.0.1:12345"));
    incoming[0].address = addr;
    manager.update(numRelays, incoming);
    check(manager.get_ping_targets(pingData) == 1);
    manager.process_pong(pingData[0].address, pingData[0].sequence);
  }

  ThroughputStats* stats[] = {
   &recorder.outbound_ping_tx,
   &recorder.route_request_rx,
   &recorder.route_request_tx,
   &recorder.route_response_rx,
   &recorder.route_response_tx,
   &recorder.client_to_server_rx,
   &recorder.client_to_server_tx,
   &recorder.server_to_client_rx,
   &recorder.server_to_client_tx,
   &recorder.inbound_ping_rx,
   &recorder.inbound_ping_tx,
   &recorder.pong_rx,
   &recorder.session_ping_rx,
   &recorder.session_ping_tx,
   &recorder.session_pong_rx,
   &recorder.session_pong_tx,
   &recorder.continue_request_rx,
   &recorder.continue_request_tx,
   &recorder.continue_response_rx,
   &recorder.continue_response_tx,
   &recorder.near_ping_rx,
   &recorder.near_ping_tx,
   &recorder.unknown_rx};

  for (auto& stat : stats) {
    static size_t counter = 0;
    stat->add(counter++);
  }

  UpdateResponse response;
  response.version = 0;
  response.timestamp = 123456789;
  response.num_relays = 2;

  {
    RelayPingInfo relay1, relay2;

    relay1.id = 135792468;
    check(relay1.address.parse("127.0.0.1:54321"));

    relay2.id = 246813579;
    check(relay2.address.parse("127.0.0.1:13524"));
    response.Relays = {
     relay1,
     relay2,
    };
  }

  client.Response.resize(response.size());
  response.into(client.Response);

  bool should_retry = false;
  check(backend.update(recorder, false, should_retry) == Backend::UpdateResult::Success);

  // check the request
  {
    UpdateRequest request;
    check(request.from(client.Request));

    check(request.version == UPDATE_REQUEST_VERSION);
    check(request.address == RelayAddr);
    check(request.public_key == backend.update_token);
    check(request.session_count == sessions.size());
    check(request.envelope_up == 123).onFail([&] {
      std::cout << "up = " << request.envelope_up << '\n';
    });
    check(request.envelope_down == 456).onFail([&] {
      std::cout << "down = " << request.envelope_down << '\n';
    });

    uint64_t request_stats[] = {
     request.outbound_ping_tx,
     request.route_request_rx,
     request.route_request_tx,
     request.route_response_rx,
     request.route_response_tx,
     request.client_to_server_rx,
     request.client_to_server_tx,
     request.server_to_client_rx,
     request.server_to_client_tx,
     request.inbound_ping_rx,
     request.inbound_ping_tx,
     request.pong_rx,
     request.session_ping_rx,
     request.session_ping_tx,
     request.session_pong_rx,
     request.session_pong_tx,
     request.continue_request_rx,
     request.continue_request_tx,
     request.continue_response_rx,
     request.continue_response_tx,
     request.near_ping_rx,
     request.near_ping_tx,
     request.unknown_rx};

    static_assert(util::array_length(stats) == util::array_length(request_stats));

    for (uint64_t i = 0; i < util::array_length(stats); i++) {
      check(stats[i]->num_bytes.load() == 0);
      check(request_stats[i] == i);
    }

    check(request.shutting_down == false);
    check(request.ping_stats.num_relays == 1);
  }

  // check that the response was processed
  {
    std::array<PingData, MAX_RELAYS> pingData;

    std::this_thread::sleep_for(1s);  // needed so that getPingData() will always return the right number
    auto count = manager.get_ping_targets(pingData);

    check(count == 2).onFail([&] {
      std::cout << "count is " << count << '\n';
    });
    check(pingData[0].address.to_string() == "127.0.0.1:54321");
    check(pingData[1].address.to_string() == "127.0.0.1:13524");

    check(routerInfo.current_time() >= 123456789).onFail([&] {
      std::cout << "info timestamp = " << routerInfo.current_time() << '\n';
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
  crypto::RandomBytes(backend.update_token, backend.update_token.size());

  client.Response = ::BasicValidUpdateResponse;

  bool should_retry = false;
  check(backend.update(recorder, true, should_retry) == Backend::UpdateResult::Success);

  UpdateRequest request;
  check(request.from(client.Request));

  check(request.version == UPDATE_REQUEST_VERSION);
  check(request.address == RelayAddr);
  check(request.public_key == backend.update_token);
  check(request.session_count == 0);
  check(request.outbound_ping_tx == 0);
  check(request.route_request_rx == 0);
  check(request.route_request_tx == 0);
  check(request.route_response_rx == 0);
  check(request.route_response_tx == 0);
  check(request.client_to_server_rx == 0);
  check(request.client_to_server_rx == 0);
  check(request.server_to_client_rx == 0);
  check(request.server_to_client_tx == 0);
  check(request.inbound_ping_rx == 0);
  check(request.inbound_ping_tx == 0);
  check(request.pong_rx == 0);
  check(request.session_ping_rx == 0);
  check(request.session_ping_tx == 0);
  check(request.session_pong_rx == 0);
  check(request.session_pong_tx == 0);
  check(request.continue_request_rx == 0);
  check(request.continue_request_tx == 0);
  check(request.continue_response_rx == 0);
  check(request.continue_response_tx == 0);
  check(request.near_ping_rx == 0);
  check(request.near_ping_tx == 0);
  check(request.unknown_rx == 0);
  check(request.shutting_down == true);
  check(request.ping_stats.num_relays == 0);
}
