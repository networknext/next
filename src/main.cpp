/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "includes.h"

#include "bench/bench.hpp"
#include "core/backend.hpp"
#include "core/packet_header.hpp"
#include "core/packet_processing.hpp"
#include "core/router_info.hpp"
#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "net/http.hpp"
#include "os/socket.hpp"
#include "testing/test.hpp"
#include "util/env.hpp"

using namespace std::chrono_literals;

using core::Backend;
using core::RelayManager;
using core::RouterInfo;
using core::SessionMap;
using crypto::KEY_SIZE;
using crypto::Keychain;
using net::Address;
using net::CurlWrapper;
using os::Socket;
using os::SocketConfig;
using os::SocketPtr;
using os::SocketType;
using util::Env;
using util::ThroughputRecorder;

volatile bool alive = true;
volatile bool upgrading = false;
volatile bool should_clean_shutdown = false;

INLINE void set_thread_affinity(std::thread& thread, int core_id)
{
  (void) thread;
  (void) core_id;
#if defined(linux) || defined(__linux) || defined(__linux__)
  cpu_set_t cpuset;
  CPU_ZERO(&cpuset);
  CPU_SET(core_id, &cpuset);
  auto res = pthread_setaffinity_np(thread.native_handle(), sizeof(cpuset), &cpuset);
  if (res != 0) {
    LOG(ERROR, "error setting thread affinity: ", res);
  }
#endif // #if defined(linux) || defined(__linux) || defined(__linux__)
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
  // todo: rework this to be non-stupid
  /*
  namespace base64 = encoding::base64;

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
  */
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
      LOG(INFO, "number of processors available: ", num_cpus);
    } else {
      LOG(FATAL, "cannot autodetect number of processors available. please set RELAY_MAX_CORES");
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
  auto hard_shutdown_handler = [](int) {
    if (alive) {
      alive = false;
    } else {
      std::exit(1);
    }
  };

  auto clean_shutdown_handler = [](int) {
    if (alive) {
      should_clean_shutdown = true;
      alive = false;
    } else {
      std::exit(1);
    }
  };

  std::signal(SIGINT, hard_shutdown_handler);      // ctrl-c
  std::signal(SIGTERM, clean_shutdown_handler);    // systemd stop
  std::signal(SIGHUP, clean_shutdown_handler);     // terminal session ends
}

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

  if (argc == 2 && strcmp(argv[1], "version")==0) {
    printf("%s\n", core::RELAY_VERSION);
    fflush(stdout);
    exit(0);
  }

  LOG(INFO, "Network Next Relay");

  LOG(INFO, "relay version is ", core::RELAY_VERSION);

  Env env;

  Address relay_addr;
  if (!relay_addr.parse(env.relay_address)) {
    LOG(FATAL, "invalid relay address: ", env.relay_address);
    return 1;
  }

  LOG(INFO, "relay address is '", relay_addr, "'");

  Keychain keychain;
  get_crypto_keys(env, keychain);

  LOG(INFO, "backend hostname is '", env.backend_hostname, "'");

  size_t num_cpus = get_num_cpus(env.max_cpus);
  int socket_recv_buff_size = get_buffer_size(env.recv_buffer_size);
  int socket_send_buff_size = get_buffer_size(env.send_buffer_size);

  if (sodium_init() == -1) {
    LOG(FATAL, "failed to initialize sodium");
  }

  bool success = false;

  RouterInfo router_info;
  RelayManager relay_manager;
  ThroughputRecorder recorder;

  std::vector<SocketPtr> sockets;
  std::vector<std::shared_ptr<std::thread>> threads;

  // decides if the relay should receive packets
  std::atomic<bool> should_receive(true);

  // session map to be shared across packet processors
  SessionMap sessions;

  // makes a shared ptr to a socket object
  auto make_socket = [&](Address& addr_in, SocketConfig config) -> SocketPtr {
    Address addr;
    addr.port = addr_in.port;
    addr.type = addr_in.type;
    auto socket = std::make_shared<Socket>();
    if (!socket->create(addr, config)) {
      return nullptr;
    }

    // if port was 0, this will set the reference parameter to what it changed to
    addr_in.port = addr.port;

    sockets.push_back(socket);

    return socket;
  };

  SocketConfig config;
  config.socket_type = SocketType::Blocking;
  config.reuse_port = true;
  config.send_buffer_size = socket_send_buff_size;
  config.recv_buffer_size = socket_recv_buff_size;
  config.recv_timeout = 0.1;

  size_t last_packet_processing_core_id = num_cpus - 1;

  if (last_packet_processing_core_id == 0) {
    //  create 1 thread and assign it to core 0
    auto socket = make_socket(relay_addr, config);
    if (!socket) {
      LOG(ERROR, "could not create socket");
      alive = false;
    }

    auto thread = std::make_shared<std::thread>([&, socket] {
      core::recv_loop(should_receive, *socket, keychain, sessions, relay_manager, alive, recorder, router_info);
    });

    set_thread_affinity(*thread, 0);

    set_thread_sched_max(*thread);

    sockets.push_back(socket);
    threads.push_back(thread);

    LOG(DEBUG, "created 1 packet processing thread assigned to core 0");
  } else {
    // create num_cpus - 1 threads and assign it to cores 1-n
    for (size_t i = 0; i < last_packet_processing_core_id && alive; i++) {
      auto socket = make_socket(relay_addr, config);
      if (!socket) {
        LOG(ERROR, "could not create socket");
        alive = false;
      }

      auto thread = std::make_shared<std::thread>([&, socket, i] {
        core::recv_loop(should_receive, *socket, keychain, sessions, relay_manager, alive, recorder, router_info);
        LOG(DEBUG, "exiting recv loop #", i);
      });

      set_thread_affinity(*thread, i + 1);

      set_thread_sched_max(*thread);

      sockets.push_back(socket);
      threads.push_back(thread);

      LOG(DEBUG, "created packet processing thread ", i, " assigned to core ", i + 1);
    }
  }

  // ping processing setup
  if (alive) {
    auto socket = sockets[0];  // will always have at least 1
    auto thread = std::make_shared<std::thread>([&, socket] {
      core::ping_loop(*socket, relay_manager, alive, recorder);
      LOG(DEBUG, "exiting ping loop");
    });

    set_thread_affinity(*thread, 0);

    set_thread_sched_max(*thread);

    sockets.push_back(socket);
    threads.push_back(thread);
  }

  // new backend setup
  if (alive) {
    std::thread thread([&] {
      CurlWrapper wrapper;
      core::Backend backend(
       env.backend_hostname, relay_addr.to_string(), keychain, router_info, relay_manager, sessions, wrapper);

      if (alive) {
        setup_signal_handlers();
        success = backend.update_loop(alive, should_clean_shutdown, recorder, sessions);
      }
    });

    set_thread_affinity(thread, 0);

    set_thread_sched_max(thread);

    thread.join();
  }

  // IMPORTANT: If an error occurs, shut down hard! We don't need to wait and join.
  if (!should_clean_shutdown)
  {
    LOG(ERROR, "hard shutdown!");
    exit(1);
  }

  LOG(INFO, "clean shutdown...");

  should_receive = false;

  LOG(INFO, "shutting down sockets");

  for (auto& socket : sockets) {
    socket->close();
  }

  int i = 0;
  for (auto& thread : threads) {
    LOG(INFO, "joining thread ", i);
    thread->join();
    i++;
  }

  LOG(INFO, "all done");

  return 1; // IMPORTANT: Always return failure. Otherwise some relays won't restart the service.
}
