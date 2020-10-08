/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "includes.h"

#include "bench/bench.hpp"
#include "core/backend.hpp"
#include "core/packet_processing.hpp"
#include "core/packet_header.hpp"
#include "core/router_info.hpp"
#include "crypto/bytes.hpp"
#include "crypto/hash.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "net/http.hpp"
#include "os/socket.hpp"
#include "testing/test.hpp"
#include "util/env.hpp"

using namespace std::chrono_literals;
using core::Backend;
using crypto::KEY_SIZE;
using crypto::Keychain;
using net::Address;
using net::BeastWrapper;
using os::Socket;
using os::SocketConfig;
using os::SocketPtr;
using os::SocketType;
using util::Env;

namespace base64 = encoding::base64;

struct
{
  volatile bool alive;
  volatile bool should_clean_shutdown;
} Globals = {
 .alive = true,
 .should_clean_shutdown = false,
};

namespace
{
  INLINE void set_thread_affinity(std::thread& thread, int core_id)
  {
    cpu_set_t cpuset;
    CPU_ZERO(&cpuset);
    CPU_SET(core_id, &cpuset);
    auto res = pthread_setaffinity_np(thread.native_handle(), sizeof(cpuset), &cpuset);
    if (res != 0) {
      LOG(ERROR, "error setting thread affinity: ", res);
    }
  }

  INLINE void set_thread_sched_max(std::thread& thread)
  {
    struct sched_param param;
    param.sched_priority = sched_get_priority_max(SCHED_FIFO);
    int ret = pthread_setschedparam(thread.native_handle(), SCHED_FIFO, &param);
    if (ret) {
      LOG(ERROR, "unable to increase server thread priority: ", strerror(ret));
    }
  }

  INLINE void get_crypto_keys(const Env& env, Keychain& keychain)
  {
    // relay private key
    {
      auto len = base64::decode(env.relay_private_key, keychain.relay_private_key);
      if (len != KEY_SIZE) {
        LOG(FATAL, "invalid relay private key");
      }

      LOG(INFO, "relay private key is '", env.relay_private_key, '\'');
    }

    // relay public key
    {
      auto len = base64::decode(env.relay_public_key, keychain.relay_public_key);
      if (len != KEY_SIZE) {
        LOG(FATAL, "invalid relay public key");
      }

      LOG(INFO, "relay public key is '", env.relay_public_key, '\'');
    }

    // router public key
    {
      auto len = base64::decode(env.relay_router_public_key, keychain.backend_public_key);
      if (len != KEY_SIZE) {
        LOG(FATAL, "invalid router public key");
      }

      LOG(INFO, "router public key is '", env.relay_router_public_key, '\'');
    }
  }

  INLINE auto get_num_cpus(const std::optional<std::string>& envvar) -> size_t
  {
    size_t num_cpus = 0;
    if (envvar.has_value()) {
      try {
        num_cpus = std::stoull(*envvar);
      } catch (std::exception& e) {
        LOG(FATAL, "could not parse RELAY_MAX_CORES to a number, value: ", *envvar);
      }
    } else {
      num_cpus = std::thread::hardware_concurrency();  // first core reserved for updates/outgoing pings
      if (num_cpus > 0) {
        LOG(INFO, "RELAY_MAX_CORES not set, autodetected number of processors available: ", num_cpus);
      } else {
        LOG(FATAL, "RELAY_MAX_CORES not set, could not detect processor count, please set the env var");
      }
    }
    return num_cpus;
  }

  INLINE auto get_buffer_size(const std::optional<std::string>& envvar) -> size_t
  {
    size_t socket_buffer_size = 1000000;

    if (envvar.has_value()) {
      try {
        socket_buffer_size = std::stoull(*envvar);
      } catch (std::exception& e) {
        LOG(ERROR, "Could not parse ", *envvar, " env var to a number: ", e.what());
      }
    }

    return socket_buffer_size;
  }

  INLINE void setup_signal_handlers()
  {
    auto graceful_shutdown_handler = [](int) {
      if (Globals.alive) {
        LOG(INFO, "gracefully shutting down");
        Globals.alive = false;
      } else {
        LOG(INFO, "hard shutdown");
        std::exit(1);
      }
    };

    auto clean_shutdown_handler = [](int) {
      if (Globals.alive) {
        LOG(INFO, "clean shutdown flagged");
        Globals.should_clean_shutdown = true;
        Globals.alive = false;
      } else {
        LOG(INFO, "hard shutdown");
        std::exit(1);
      }
    };

    std::signal(SIGINT, graceful_shutdown_handler);  // ctrl-c
    std::signal(SIGTERM, clean_shutdown_handler);    // systemd stop
    std::signal(SIGHUP, clean_shutdown_handler);     // terminal session ends
  }
}  // namespace

int main(int argc, const char* argv[])
{
  (void)argc;
  (void)argv;
#ifdef TEST_BUILD
  return testing::SpecTest::Run(argc, argv) ? 0 : 1;
#endif

#ifdef BENCH_BUILD
  benchmarking::Benchmark::Run();
  return 0;
#endif

  LOG(INFO, "Network Next Relay");

  Env env;

  Address relay_addr;
  if (!relay_addr.parse(env.relay_address)) {
    LOG(FATAL, "invalid relay address: ", env.relay_address);
    return 1;
  }

  LOG(INFO, "relay address is '", relay_addr, '\'');

  Keychain keychain;
  get_crypto_keys(env, keychain);

  LOG(INFO, "backend hostname is '", env.backend_hostname, '\'');

  unsigned int num_cpus = get_num_cpus(env.max_cpus);
  int socket_recv_buff_size = get_buffer_size(env.recv_buffer_size);
  int socket_send_buff_size = get_buffer_size(env.send_buffer_size);

  if (sodium_init() == -1) {
    LOG(FATAL, "failed to initialize sodium");
  }

  LOG(DEBUG, "initializing relay");

  bool success = false;

  core::RouterInfo router_info;
  core::RelayManager relay_manager;
  util::ThroughputRecorder recorder;

  std::vector<SocketPtr> sockets;
  std::vector<std::shared_ptr<std::thread>> threads;

  // decides if the relay should receive packets
  std::atomic<bool> should_receive(true);

  // session map to be shared across packet processors
  core::SessionMap sessions;

  // makes a shared ptr to a socket object
  auto make_socket = [&](Address& addr_in, SocketConfig config) -> SocketPtr {
    Address addr;
    addr.Port = addr_in.Port;
    addr.Type = addr_in.Type;
    auto socket = std::make_shared<Socket>();
    if (!socket->create(addr, config)) {
      return nullptr;
    }

    // if port was 0, this will set the reference parameter to what it changed to
    addr_in.Port = addr.Port;

    sockets.push_back(socket);

    return socket;
  };

  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.reuse_port = true;
  config.send_buffer_size = socket_send_buff_size;
  config.recv_buffer_size = socket_recv_buff_size;

  size_t num_threads = (num_cpus == 1) ? num_cpus : num_cpus - 1;
  LOG(DEBUG, "creating ", num_cpus, " packet processing thread", (num_cpus != 1) ? "s" : "");

  // packet processing setup
  for (unsigned int i = (num_cpus == 1) ? 0 : 1; i < num_threads && Globals.alive; i++) {
    auto socket = make_socket(relay_addr, config);
    if (!socket) {
      LOG(ERROR, "could not create socket");
      Globals.alive = false;
    }

    auto thread = std::make_shared<std::thread>([&, socket, i] {
      core::recv_loop(should_receive, *socket, keychain, sessions, relay_manager, Globals.alive, recorder, router_info);
      LOG(DEBUG, "exiting recv loop #", i);
    });

    set_thread_affinity(*thread, (std::thread::hardware_concurrency() == 1) ? 0 : i);

    set_thread_sched_max(*thread);

    sockets.push_back(socket);
    threads.push_back(thread);
  }

  // ping processing setup
  if (Globals.alive) {
    auto socket = sockets[0];  // will always have at least 1
    auto thread = std::make_shared<std::thread>([&, socket] {
      core::ping_loop(*socket, relay_manager, Globals.alive, recorder);
      LOG(DEBUG, "exiting ping loop");
    });

    set_thread_affinity(*thread, 0);

    set_thread_sched_max(*thread);

    sockets.push_back(socket);
    threads.push_back(thread);
  }

  LOG(DEBUG, "Globals.alive = ", Globals.alive ? "true" : "false");

  // new backend setup
  if (Globals.alive) {
    std::thread thread([&] {
      BeastWrapper wrapper;
      core::Backend backend(
       env.backend_hostname,
       relay_addr.to_string(),
       keychain,
       router_info,
       relay_manager,
       env.relay_public_key,
       sessions,
       wrapper);

      size_t attempts = 0;
      while (attempts < 60) {
        if (backend.init()) {
          std::cout << '\n';
          break;
        }

        std::this_thread::sleep_for(1s);
        attempts++;
      }

      if (attempts < 60) {
        LOG(INFO, "relay initialized with new backend");
      } else {
        LOG(ERROR, "could not initialize relay");
        Globals.alive = false;
      }

      if (Globals.alive) {
        setup_signal_handlers();

        success = backend.update_loop(Globals.alive, Globals.should_clean_shutdown, recorder, sessions);
      }
    });

    set_thread_affinity(thread, 0);

    set_thread_sched_max(thread);

    thread.join();
  }

  LOG(DEBUG, "clean shutdown = ", Globals.should_clean_shutdown ? "true" : "false");

  should_receive = false;

  for (auto& socket : sockets) {
    socket->close();
  }

  for (auto& thread : threads) {
    thread->join();
  }

  LOG(DEBUG, "Receiving Address: ", relay_addr);

  return success ? 0 : 1;
}