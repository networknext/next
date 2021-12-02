#include "includes.h"
#include "backend.hpp"

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "net/http.hpp"
#include "os/system.hpp"
#include <string.h>

using core::RelayStats;
using crypto::KEY_SIZE;
using util::Second;

extern volatile bool alive;
extern volatile bool upgrading;
extern volatile bool should_clean_shutdown;

using namespace std::chrono_literals;

void * upgrade_thread_function( void * data )
{
  const char * version = (const char*) data;

  LOG(INFO, "upgrading from ", core::RELAY_VERSION, " -> ", version);

  char command[1024];
  sprintf( command, "rm -f relay-%s", version );
  system( command );

  sprintf( command, "wget https://storage.googleapis.com/relay_artifacts/relay-%s", version );
  if ( system( command ) != 0 ) {
    LOG(ERROR, "failed to download relay version ", version);
    std::this_thread::sleep_for(60s);
    upgrading = false;
    return NULL;
  }

  LOG(INFO, "successfully downloaded relay-", version);

  sprintf( command, "chmod +x relay-%s", version );
  if ( system( command ) != 0 ) {
    LOG(ERROR, "failed to chmod +x relay-", version);
    std::this_thread::sleep_for(60s);
    upgrading = false;
    return NULL;
  }

  LOG(INFO, "chmod +x relay-", version, " succeeded");

  sprintf( command, "./relay-%s version", version );
  FILE * file = popen( command, "r" );
  char buffer[1024];
  if ( fgets( buffer, sizeof(buffer), file ) == NULL || strstr(buffer, version) == NULL )
  {
    pclose( file );
    LOG(ERROR, "relay binary is bad");
    std::this_thread::sleep_for(60s);
    upgrading = false;
    return NULL;
  }
  pclose( file );

  LOG(INFO, "relay binary is good");

  system( "rm -f relay 2>/dev/null" );

  sprintf( command, "mv relay-%s relay", version );
  if ( system( command ) != 0 )
  {
    LOG(ERROR, "could not install new relay binary");
    std::this_thread::sleep_for(60s);
    upgrading = false;
    return NULL;
  }

  LOG(INFO, "new relay binary is installed");

  should_clean_shutdown = true;
  alive = false;
  upgrading = false;

  return NULL;
}

namespace core
{
  using namespace std::chrono_literals;

  const char* RELAY_VERSION = "2.0.9";

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
    uint8_t shutdown_flag;
    if (!encoding::read_uint8(v, index, shutdown_flag)) {
      return false;
    }
    this->shutting_down = static_cast<bool>(shutdown_flag);

    if (!encoding::read_string(v, index, this->relay_version)) {
      return false;
    }
    if (!encoding::read_uint8(v, index, this->cpu_usage)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->envelope_up)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->envelope_down)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->bandwidth_tx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->bandwidth_rx)) {
      return false;
    }

    return true;
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
      if (update(recorder, false, should_loop) == UpdateResult::Failure) {
        LOG(ERROR, "could not update relay");
        success = should_loop = false;
      }
      sessions.purge(this->router_info.current_time<uint64_t>());
      std::this_thread::sleep_for(1s);
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

      uint8_t cpu = 0;
      #if defined(linux) || defined(__linux) || defined(__linux__)
      double cpu_percent = os::GetCPU();
      if ( cpu_percent > 100.0 )
        cpu_percent = 100.0;
      cpu = (uint8_t) floor(cpu_percent + 0.5f);
      #endif
      encoding::write_uint8(req, index, cpu);

      encoding::write_uint64(req, index, this->session_map.envelope_up_total());
      encoding::write_uint64(req, index, this->session_map.envelope_down_total());

      util::ThroughputRecorder traffic_stats(std::move(recorder));

      uint64_t bandwidth_tx = traffic_stats.outbound_ping_tx.num_bytes.load()
                              + traffic_stats.route_request_tx.num_bytes.load()
                              + traffic_stats.route_response_tx.num_bytes.load()
                              + traffic_stats.client_to_server_tx.num_bytes.load()
                              + traffic_stats.server_to_client_tx.num_bytes.load()
                              + traffic_stats.inbound_ping_tx.num_bytes.load()
                              + traffic_stats.session_ping_tx.num_bytes.load()
                              + traffic_stats.session_pong_tx.num_bytes.load()
                              + traffic_stats.continue_request_tx.num_bytes.load()
                              + traffic_stats.continue_response_tx.num_bytes.load()
                              + traffic_stats.near_ping_tx.num_bytes.load();

      uint64_t bandwidth_rx = traffic_stats.route_request_rx.num_bytes.load()
                              + traffic_stats.route_request_rx.num_bytes.load()
                              + traffic_stats.route_response_rx.num_bytes.load()
                              + traffic_stats.client_to_server_rx.num_bytes.load()
                              + traffic_stats.server_to_client_rx.num_bytes.load()
                              + traffic_stats.inbound_ping_rx.num_bytes.load()
                              + traffic_stats.pong_rx.num_bytes.load()
                              + traffic_stats.session_ping_rx.num_bytes.load()
                              + traffic_stats.session_pong_rx.num_bytes.load()
                              + traffic_stats.continue_request_rx.num_bytes.load()
                              + traffic_stats.continue_response_rx.num_bytes.load()
                              + traffic_stats.near_ping_rx.num_bytes.load()
                              + traffic_stats.unknown_rx.num_bytes.load();

      encoding::write_uint64(req, index, bandwidth_tx);
      encoding::write_uint64(req, index, bandwidth_rx);
    }

    // todo: this whole loop here is naff... the retry loop shouldn't occur in place here, but in the regular, 1 sec delay updates!
    
    LOG(DEBUG, "sending relay update");
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

    // end naff section

    if (num_retries >= MAX_UPDATE_ATTEMPTS) {
      return UpdateResult::Failure;
    }

    if (!request_success) {
      return UpdateResult::Failure;
    }

    // early return if shutting down since the response won't be valid
    if (shutdown) {
      std::array<RelayPingInfo, MAX_RELAYS> relays;
      this->relay_manager.update(0, relays);
      return UpdateResult::Success;
    }

    LOG(DEBUG, "parsing response");

    // parse response
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

    LOG(DEBUG, "updated relay");

    if (first_update)
    {
      LOG(INFO, "relay initialized");
      first_update = false;
    }

    if (response.target_version[0] != '\0' && strcmp(core::RELAY_VERSION, response.target_version.c_str()) != 0) {

      if (!upgrading)
      {
        upgrading = true;

        static pthread_t upgrade_thread;
        static char target_version[1024];
        strcpy( target_version, response.target_version.c_str() );
        int err = pthread_create( &upgrade_thread, NULL, &upgrade_thread_function, target_version );
        if ( err )
        {
          LOG(ERROR, "could not create upgrade thread");
          upgrading = false;          
          return UpdateResult::Success;
        }
      }
    }

    return UpdateResult::Success;
  }
}  // namespace core
