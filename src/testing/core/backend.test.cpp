#include "includes.h"
#include "testing/test.hpp"
#include "core/backend.hpp"
#include "testing/mocks.hpp"

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
using core::UpdateRequest;
using core::UpdateResponse;
using crypto::KEY_SIZE;
using crypto::Keychain;
using net::Address;
using util::Clock;
using util::Second;
using util::ThroughputRecorder;

namespace
{
  const std::string BACKEND_HOSTNAME = "http://totally-real-backend.com";
  const auto RELAY_ADDR = "127.0.0.1:12345";
  const Keychain KEYCHAIN = testing::make_keychain();
  const std::vector<uint8_t> BASIC_VALID_UPDATE_RESPONSE = [] {
    InitResponse response = {
     .version = 0,
     .timestamp = 0,
     .public_key = {},
    };

    std::vector<uint8_t> buff(InitResponse::SIZE_OF);
    CHECK(response.into(buff));
    return buff;
  }();

  auto make_init_response(uint32_t version, uint64_t timestamp, std::array<uint8_t, KEY_SIZE>& pk) -> std::vector<uint8_t>
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

TEST(core_backend_init_valid)
{
  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  std::array<uint8_t, crypto::KEY_SIZE> pk{};
  testing::MockHttpClient client;
  client.Response = make_init_response(0, 123456789, pk);
  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);

  CHECK(backend.init());

  CHECK(client.Hostname == BACKEND_HOSTNAME);
  CHECK(client.Endpoint == "/relay_init");
  CHECK(router_info.current_time() >= 123456789 / 1000);

  InitRequest request;
  CHECK(request.from(client.Request));

  CHECK(request.magic == INIT_REQUEST_MAGIC);
  CHECK(request.version == INIT_REQUEST_VERSION);
  CHECK(request.address == RELAY_ADDR);

  // can't check nonce or encrypted token since they're random
}

// Update the backend for 2 seconds, then proceed to switch the should_loop to false.
// The relay should then attempt to ack the backend.
// It won't receive a success response from the backend so instead it will
// live for 60 seconds and skip the ack
TEST(core_Backend_updateCycle_shutdown_60s)
{
  Clock test_clock;

  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  volatile bool should_loop = true;
  volatile bool should_clean_shutdown = false;
  ThroughputRecorder logger;
  testing::MockHttpClient client;

  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);

  client.Success = true;
  client.Response = BASIC_VALID_UPDATE_RESPONSE;

  test_clock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    client.Success = false;
    should_clean_shutdown = true;  // just to mimic actual behavior
    should_loop = false;
  });

  CHECK(backend.update_loop(should_loop, should_clean_shutdown, logger, sessions));
  auto elapsed = test_clock.elapsed<Second>();
  CHECK(elapsed >= 62.0);
}

// Update the backend for 2 seconds, then proceed to switch the should_loop to false.
// The relay should then attempt to ack the backend and shutdown for 30 seconds.
// It will receive a success response and then live for another 30 seconds.
// The 60 second timeout will not apply here
TEST(core_Backend_updateCycle_ack_and_30s)
{
  Clock test_clock;

  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  volatile bool should_loop = true;
  volatile bool should_clean_shutdown = false;
  ThroughputRecorder logger;
  testing::MockHttpClient client;

  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);

  client.Success = true;
  client.Response = BASIC_VALID_UPDATE_RESPONSE;

  test_clock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    should_clean_shutdown = true;
    should_loop = false;
  });

  CHECK(backend.update_loop(should_loop, should_clean_shutdown, logger, sessions));
  auto elapsed = test_clock.elapsed<Second>();
  CHECK(elapsed >= 32.0);
}

// Update the backend for 2 seconds, then proceed to switch the should_loop to false.
// The relay will not get a success response for 31 seconds after the should_loop is set.
// After which it will get a success and then proceed with the normal routine of waiting 30 seconds
// The amount of time waited will be greater than 60 seconds
// This is to assert the update_cycle will ignore the 60 second timeout if the backend gets an update
TEST(core_Backend_updateCycle_no_ack_for_40s_then_ack_then_wait)
{
  Clock test_clock;

  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  volatile bool should_loop = true;
  volatile bool should_clean_shutdown = false;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);

  client.Success = true;
  client.Response = BASIC_VALID_UPDATE_RESPONSE;

  test_clock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    should_clean_shutdown = true;
    client.Success = false;
    should_loop = false;
    std::this_thread::sleep_for(31s);
    client.Success = true;
  });

  CHECK(backend.update_loop(should_loop, should_clean_shutdown, recorder, sessions));
  auto elapsed = test_clock.elapsed<Second>();
  CHECK(elapsed >= 63.0);
}

// Update the backend for 2 seconds, then switch the success of the request to false.
// That will trigger the failure attempts which the number of is controlled by the MaxUpdateAttempts constant.
// After the max attempts is reached it will shutdown.
// But because the success value is never reset to true, the cleanshutdown ack will never succeed
// so the final duration should be 2 seconds of success and (MaxUpdateAttempts - 1) seconds of failure.
TEST(core_Backend_updateCycle_update_fails_for_max_number_of_attempts)
{
  Clock test_clock;

  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  volatile bool should_loop = true;
  volatile bool should_clean_shutdown = false;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);

  client.Success = true;
  client.Response = BASIC_VALID_UPDATE_RESPONSE;

  test_clock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    client.Success = false;  // set to false here to trigger failed updates
  });

  CHECK(!backend.update_loop(should_loop, should_clean_shutdown, recorder, sessions));
  auto elapsed = test_clock.elapsed<Second>();
  // time will be 2 seconds of good updates and
  // 10 seconds of bad updates, which will cause
  // the relay to abort with no clean shutdown
  CHECK(elapsed >= 12.0).on_fail([&elapsed] {
    std::cout << "elapsed: " << elapsed << '\n';
  });
}

// When clean shutdown is not set to true, the function should return immediately
TEST(core_Backend_updateCycle_no_clean_shutdown)
{
  Clock test_clock;

  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  volatile bool should_loop = true;
  volatile bool should_clean_shutdown = false;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);

  client.Success = true;
  client.Response = BASIC_VALID_UPDATE_RESPONSE;

  test_clock.reset();
  auto fut = std::async(std::launch::async, [&] {
    std::this_thread::sleep_for(2s);
    client.Success = false;
    should_loop = false;
  });

  CHECK(backend.update_loop(should_loop, should_clean_shutdown, recorder, sessions));
  auto elapsed = test_clock.elapsed<Second>();
  CHECK(elapsed >= 2.0);
}

TEST(core_Backend_update_valid)
{
  Clock clock;
  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;
  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);
  crypto::RandomBytes(backend.update_token, backend.update_token.size());

  sessions.set(1234, std::make_shared<Session>());  // just add one thing to the map to make it non-zero

  // seed relay manager
  {
    const size_t num_relays = 1;
    std::array<RelayPingInfo, MAX_RELAYS> incoming;
    std::array<PingData, MAX_RELAYS> ping_data;
    incoming[0].id = 987654321;
    Address addr;
    CHECK(addr.parse("127.0.0.1:12345"));
    incoming[0].address = addr;
    manager.update(num_relays, incoming);
    CHECK(manager.get_ping_targets(ping_data) == 1);
    manager.process_pong(ping_data[0].address, ping_data[0].sequence);
  }

  recorder.unknown_rx.add(10);
  UpdateResponse response;
  response.version = 0;
  response.timestamp = 123456789;
  response.num_relays = 2;

  {
    RelayPingInfo relay1, relay2;

    relay1.id = 135792468;
    CHECK(relay1.address.parse("127.0.0.1:54321"));

    relay2.id = 246813579;
    CHECK(relay2.address.parse("127.0.0.1:13524"));
    response.Relays = {
     relay1,
     relay2,
    };
  }

  client.Response.resize(response.size());
  response.into(client.Response);

  const auto outbound_ping = 123456789;
  const auto pong = 987654321;

  recorder.outbound_ping_tx.add(outbound_ping);
  recorder.pong_rx.add(pong);

  CHECK(backend.update(recorder, false));

  // check the request
  {
    UpdateRequest request;
    CHECK(request.from(client.Request));

    CHECK(request.version == 1);
    CHECK(request.address == RELAY_ADDR);
    CHECK(request.public_key == backend.update_token);
    CHECK(request.session_count == sessions.size());
    CHECK(request.outbound_ping_tx == outbound_ping);
    CHECK(request.route_request_rx == 0);
    CHECK(request.route_request_tx == 0);
    CHECK(request.route_response_rx == 0);
    CHECK(request.route_response_tx == 0);
    CHECK(request.client_to_server_rx == 0);
    CHECK(request.client_to_server_tx == 0);
    CHECK(request.server_to_client_rx == 0);
    CHECK(request.server_to_client_tx == 0);
    CHECK(request.inbound_ping_rx == 0);
    CHECK(request.inbound_ping_tx == 0);
    CHECK(request.pong_rx == pong);
    CHECK(request.session_ping_rx == 0);
    CHECK(request.session_ping_tx == 0);
    CHECK(request.session_pong_rx == 0);
    CHECK(request.session_pong_tx == 0);
    CHECK(request.continue_request_rx == 0);
    CHECK(request.continue_request_tx == 0);
    CHECK(request.continue_response_rx == 0);
    CHECK(request.continue_response_tx == 0);
    CHECK(request.near_ping_rx == 0);
    CHECK(request.near_ping_tx == 0);
    CHECK(request.unknown_rx == 10);
    CHECK(request.shutting_down == false);
    CHECK(request.ping_stats.num_relays == 1);
    CHECK(request.relay_version == RELAY_VERSION);
  }

  // check that the response was processed
  {
    std::array<PingData, MAX_RELAYS> ping_data;

    std::this_thread::sleep_for(1s);  // needed so that get_ping_targets() will always return the right number
    auto count = manager.get_ping_targets(ping_data);

    CHECK(count == 2).on_fail([&] {
      std::cout << "count is " << count << '\n';
    });
    CHECK(ping_data[0].address.to_string() == "127.0.0.1:54321");
    CHECK(ping_data[1].address.to_string() == "127.0.0.1:13524");

    CHECK(router_info.current_time() >= 123456789).on_fail([&] {
      std::cout << "info timestamp = " << router_info.current_time() << '\n';
    });
  }
}

TEST(core_Backend_update_shutting_down_true)
{
  Clock clock;
  RouterInfo router_info;
  RelayManager manager;
  SessionMap sessions;
  ThroughputRecorder recorder;
  testing::MockHttpClient client;

  Backend backend(BACKEND_HOSTNAME, RELAY_ADDR, KEYCHAIN, router_info, manager, sessions, client);
  crypto::RandomBytes(backend.update_token, backend.update_token.size());

  client.Response = ::BASIC_VALID_UPDATE_RESPONSE;

  CHECK(backend.update(recorder, true));

  UpdateRequest request;
  CHECK(request.from(client.Request));

  CHECK(request.version == 1);
  CHECK(request.address == RELAY_ADDR);
  CHECK(request.public_key == backend.update_token);
  CHECK(request.session_count == 0);
  CHECK(request.outbound_ping_tx == 0);
  CHECK(request.route_request_rx == 0);
  CHECK(request.route_request_tx == 0);
  CHECK(request.route_response_rx == 0);
  CHECK(request.route_response_tx == 0);
  CHECK(request.client_to_server_rx == 0);
  CHECK(request.client_to_server_rx == 0);
  CHECK(request.server_to_client_rx == 0);
  CHECK(request.server_to_client_tx == 0);
  CHECK(request.inbound_ping_rx == 0);
  CHECK(request.inbound_ping_tx == 0);
  CHECK(request.pong_rx == 0);
  CHECK(request.session_ping_rx == 0);
  CHECK(request.session_ping_tx == 0);
  CHECK(request.session_pong_rx == 0);
  CHECK(request.session_pong_tx == 0);
  CHECK(request.continue_request_rx == 0);
  CHECK(request.continue_request_tx == 0);
  CHECK(request.continue_response_rx == 0);
  CHECK(request.continue_response_tx == 0);
  CHECK(request.near_ping_rx == 0);
  CHECK(request.near_ping_tx == 0);
  CHECK(request.unknown_rx == 0);
  CHECK(request.shutting_down == true);
  CHECK(request.ping_stats.num_relays == 0);
  CHECK(request.relay_version == RELAY_VERSION);
}
