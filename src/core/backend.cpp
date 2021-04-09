#include "includes.h"
#include "backend.hpp"

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "net/http.hpp"
#include "os/system.hpp"

using core::RelayStats;
using crypto::KEY_SIZE;
using util::Second;

namespace core
{
  using namespace std::chrono_literals;

  const char* RELAY_VERSION = "2.0.1";

  const char* const UPDATE_ENDPOINT = "/relay_update";

  const double CLEAN_SHUTDOWN_TIMEOUT_SECS = 60.0;

  auto UpdateRequest::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::read_uint32(v, index, this->version)) {
      return false;
    }
    if (!encoding::read_string(v, index, this->address)) {
      return false;
    }
    if (!encoding::read_bytes(v, index, public_key, public_key.size())) {
      return false;
    }
    if (!encoding::read_uint32(v, index, this->ping_stats.num_relays)) {
      return false;
    }

    for (size_t i = 0; i < ping_stats.num_relays; i++) {
      if (!encoding::read_uint64(v, index, this->ping_stats.ids[i])) {
        return false;
      }
      if (!encoding::read_bytes(
           v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&ping_stats.rtt[i]), sizeof(float), sizeof(float))) {
        return false;
      }
      if (!encoding::read_bytes(
           v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&ping_stats.jitter[i]), sizeof(float), sizeof(float))) {
        return false;
      }
      if (!encoding::read_bytes(
           v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&ping_stats.packet_loss[i]), sizeof(float), sizeof(float))) {
        return false;
      }
    }

    if (!encoding::read_uint64(v, index, this->session_count)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->envelope_up)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->envelope_down)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->outbound_ping_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->route_request_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->route_request_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->route_response_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->route_response_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->client_to_server_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->client_to_server_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->server_to_client_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->server_to_client_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->inbound_ping_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->inbound_ping_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->pong_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->session_ping_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->session_ping_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->session_pong_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->session_pong_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->continue_request_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->continue_request_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->continue_response_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->continue_response_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->near_ping_rx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->near_ping_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->unknown_rx)) {
      return false;
    }
    uint8_t shutdown_flag;
    if (!encoding::read_uint8(v, index, shutdown_flag)) {
      return false;
    }
    this->shutting_down = static_cast<bool>(shutdown_flag);

    if (!encoding::read_double(v, index, this->cpu_usage)) {
      return false;
    }

    if (!encoding::read_double(v, index, this->mem_usage)) {
      return false;
    }

    return true;
  }

  auto UpdateResponse::size() -> size_t
  {
    size_t size = 10 * 1024;

    for (size_t i = 0; i < this->num_relays; i++) {
      const auto& relay = relays[i];
      size += relay.address.to_string().length();
    }

    return size;
  }

  // only used in tests
  auto UpdateResponse::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::write_uint32(v, index, this->version)) {
      LOG(TRACE, "could not write version");
      return false;
    }

    if (!encoding::write_uint64(v, index, this->timestamp)) {
      LOG(TRACE, "could not write timestamp");
      return false;
    }

    if (!encoding::write_uint32(v, index, this->num_relays)) {
      LOG(TRACE, "could not write num relays");
      return false;
    }

    for (size_t i = 0; i < this->num_relays; i++) {
      const auto& relay = relays[i];

      if (!encoding::write_uint64(v, index, relay.id)) {
        LOG(TRACE, "could not write relay id");
        return false;
      }

      if (!encoding::write_string(v, index, relay.address.to_string())) {
        LOG(TRACE, "could not write relay address");
        return false;
      }
    }

    return true;
  }

  auto UpdateResponse::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::read_uint32(v, index, version)) {
      LOG(ERROR, "unable to read update response version");
      return false;
    }

    if (!encoding::read_uint64(v, index, this->timestamp)) {
      LOG(ERROR, "unable to read update response timestamp");
      return false;
    }

    if (!encoding::read_uint32(v, index, this->num_relays)) {
      LOG(ERROR, "unable to read update response relay count");
      return false;
    }

    for (size_t i = 0; i < this->num_relays; i++) {
      auto& relay = relays[i];
      if (!encoding::read_uint64(v, index, relay.id)) {
        LOG(ERROR, "unable to read update response relay id #", i);
        return false;
      }

      std::string addr;
      if (!encoding::read_string(v, index, addr)) {
        LOG(ERROR, "unable to read update response relay address #", i);
        return false;
      }

      if (!relay.address.parse(addr)) {
        LOG(ERROR, "unable to parse relay address: ", addr);
        return false;
      }
    }

    std::string target_version;
    if (!encoding::read_string(v, index, target_version)) {
      LOG(ERROR, "unable to read update response target version");
      return false;
    }

    // printf( "target_version = %s\n", target_version.c_str());

    return true;
  }

  Backend::Backend(
   std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& router_info,
   RelayManager& relay_manager,
   const core::SessionMap& sessions,
   net::CurlWrapper& client)
   : hostname(hostname),
     relay_address(address),
     keychain(keychain),
     router_info(router_info),
     relay_manager(relay_manager),
     session_map(sessions),
     http_client(client)
  {}

  bool Backend::update_loop(
   volatile bool& should_loop,
   const volatile bool& should_clean_shutdown,
   util::ThroughputRecorder& recorder,
   core::SessionMap& sessions)
  {
    bool success = true;

    while (should_loop) {
      LOG(DEBUG, "should loop = ", should_loop ? "true" : "false");
      switch (update(recorder, false, should_loop)) {
        case UpdateResult::Failure: {
          LOG(ERROR, "could not update relay");
          success = should_loop = false;
        } break;
        default: {
          sessions.purge(this->router_info.current_time<uint64_t>());
          std::this_thread::sleep_for(1s);
        }
      }
    }

    LOG(DEBUG, "exiting update loop");

    Clock backend_timeout;
    if (should_clean_shutdown) {
      LOG(INFO, "clean shutdown");
      for (int i = 0; i < CLEAN_SHUTDOWN_TIMEOUT_SECS; i++) {
        LOG(INFO, CLEAN_SHUTDOWN_TIMEOUT_SECS - i);
        update(recorder, true, should_loop);
        std::this_thread::sleep_for(1s);
      }
    }

    return success;
  }

  auto Backend::update(util::ThroughputRecorder& recorder, bool shutdown, const volatile bool& should_retry) -> UpdateResult
  {
    std::vector<uint8_t> req, res;

    static bool first_update = true;

    // serialize request
    {
      RelayStats stats;
      this->relay_manager.get_stats(stats);

      req.resize(100*1024);

      size_t index = 0;

      encoding::write_uint32(req, index, UPDATE_REQUEST_VERSION);
      encoding::write_string(req, index, this->relay_address);
      encoding::write_bytes(req, index, this->update_token, this->update_token.size());
      encoding::write_uint32(req, index, stats.num_relays);

      for (unsigned int i = 0; i < stats.num_relays; ++i) {
        encoding::write_uint64(req, index, stats.ids[i]);
        encoding::write_bytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.rtt[i]), sizeof(float));
        encoding::write_bytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.jitter[i]), sizeof(float));
        encoding::write_bytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.packet_loss[i]), sizeof(float));
      }

      encoding::write_uint64(req, index, this->session_map.size());

      encoding::write_uint8(req, index, shutdown);

      encoding::write_string(req, index, RELAY_VERSION);
    }

    // LOG(DEBUG, "sending request");
    util::Clock timeout;
    double elapsed_seconds = timeout.elapsed<Second>();
    size_t num_retries = 0;
    bool request_success = false;
    while (!(request_success = this->http_client.send_request(this->hostname, UPDATE_ENDPOINT, req, res)) && should_retry &&
           num_retries < MAX_UPDATE_ATTEMPTS) {
      LOG(ERROR, "relay update failed ", num_retries);
      num_retries++;
      std::this_thread::sleep_for(1s);
    }

    if (num_retries >= MAX_UPDATE_ATTEMPTS) {
      return UpdateResult::Failure;
    }

    if (!request_success) {
      return UpdateResult::Failure;
    }

    // early return if shutting down since the response won't be valid
    if (shutdown) {
      return UpdateResult::Success;
    }

    LOG(DEBUG, "parsing response");

    // parse response
    {
      UpdateResponse response;
      if (!response.from(res)) {
        LOG(ERROR, "could not deserialize update response, response size = ", res.size());
        return UpdateResult::Failure;
      }

      if (response.version != UPDATE_RESPONSE_VERSION) {
        LOG(ERROR, "bad relay version response version. expected ", UPDATE_RESPONSE_VERSION, ", got ", response.version);
        return UpdateResult::Failure;
      }

      this->router_info.set_timestamp(response.timestamp);

      if (response.num_relays > MAX_RELAYS) {
        LOG(ERROR, "too many relays to ping. max is ", MAX_RELAYS, ", got ", response.num_relays, '\n');
        return UpdateResult::Failure;
      }

      if (!this->relay_manager.update(response.num_relays, response.relays)) {
        LOG(ERROR, "could not update relay manager");
        return UpdateResult::Failure;
      }
    }

    LOG(DEBUG, "updated relay");

    if (first_update)
    {
      LOG(INFO, "relay initialized");
      first_update = false;
    }

    return UpdateResult::Success;
  }
}  // namespace core
