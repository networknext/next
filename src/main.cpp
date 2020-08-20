/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "includes.h"

#include "bench/bench.hpp"
#include "core/backend.hpp"
#include "core/packet_processor.hpp"
#include "core/ping_processor.hpp"
#include "core/router_info.hpp"
#include "crypto/bytes.hpp"
#include "crypto/hash.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "os/socket.hpp"
#include "os/thread.hpp"
#include "relay/relay.hpp"
#include "testing/test.hpp"
#include "util/env.hpp"
#include "core/packets/header.hpp"

using namespace std::chrono_literals;
using util::Env;
using crypto::Keychain;
using crypto::KeySize;

namespace base64 = encoding::base64;

volatile bool gAlive = true;
volatile bool gShouldCleanShutdown = false;

namespace
{
  INLINE void get_crypto_keys(const Env& env, Keychain& keychain)
  {
    // relay private key
    {
      auto len = base64::decode(env.relay_private_key, keychain.RelayPrivateKey);
      if (len != crypto::KeySize) {
        LOG(FATAL, "invalid relay private key");
      }

      LOG(INFO, "relay private key is '", env.relay_private_key, "'\n");
    }

    // relay public key
    {
      auto len = base64::decode(env.relay_public_key, keychain.RelayPublicKey);
      if (len != crypto::KeySize) {
        LOG(FATAL, "invalid relay public key");
      }

      LOG(INFO, "relay public key is '", env.relay_public_key, "'\n");
    }

    // router public key
    {
      auto len = base64::decode(env.relay_router_public_key, keychain.RouterPublicKey);
      if (len != crypto::KeySize) {
        LOG(FATAL, "invalid router public key");
      }

      LOG(INFO, "router public key is '", env.relay_router_public_key, "'\n");
    }
  }

  INLINE auto get_num_cpus(const std::string& env) -> size_t
  {
    size_t num_cpus;
    if (env.empty()) {
      num_cpus = std::thread::hardware_concurrency();  // first core reserved for updates/outgoing pings
      if (num_cpus > 0) {
        LOG(INFO, "RELAY_MAX_CORES not set, autodetected number of processors available: ", num_cpus);
      } else {
        LOG(FATAL, "RELAY_MAX_CORES not set, could not detect processor count, please set the env var");
      }
    } else {
      try {
        num_cpus = std::stoull(env);
      } catch (std::exception& e) {
        LOG(FATAL, "could not parse RELAY_MAX_CORES to a number, value: ", env.max_cpus);
      }
    }
    return num_cpus;
  }

  INLINE auto get_buffer_size(const std::string& envvar) -> size_t
  {
    size_t socketBufferSize = 1000000;

    if (!envvar.empty()) {
      try {
        socketBufferSize = std::stoull(envvar);
      } catch (std::exception& e) {
        LOG(ERROR, "Could not parse ", envvar, " env var to a number: ", e.what());
      }
    }

    return socketBufferSize;
  }

  INLINE void setup_signal_handlers()
  {
    auto gracefulShutdownHandler = [](int) {
      if (gAlive) {
        gAlive = false;
      } else {
        std::exit(1);
      }
    };

    auto cleanShutdownHandler = [](int) {
      if (gAlive) {
        gShouldCleanShutdown = true;
        gAlive = false;
      } else {
        std::exit(1);
      }
    };

    signal(SIGINT, gracefulShutdownHandler);  // ctrl-c
    signal(SIGTERM, cleanShutdownHandler);    // systemd stop
    signal(SIGHUP, cleanShutdownHandler);     // terminal session ends
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

  net::Address relay_addr;
  if (!relay_addr.parse(env.relay_address)) {
    LOG(FATAL, "invalid relay address: ", env.relay_address);
    return 1;
  }

  LOG(INFO, "relay address is '", relay_addr, "'\n");

  Keychain keychain;
  get_crypto_keys(env, keychain);

  LOG(INFO, "backend hostname is '", env.backend_hostname, "'\n");

  unsigned int num_cpus = get_num_cpus(env.max_cpus);
  int socketRecvBuffSize = get_buffer_size(env.recv_buffer_size);
  int socketSendBuffSize = get_buffer_size(env.send_buffer_size);

  if (sodium_init() == -1) {
    LOG(FATAL, "failed to initialize sodium");
  }

  LOG(DEBUG, "initializing relay");

  bool success = false;

  core::RouterInfo router_info;
  core::RelayManager<core::Relay> relay_manager;
  util::ThroughputRecorder recorder;

  std::vector<os::SocketPtr> sockets;
  std::vector<std::shared_ptr<std::thread>> threads;

  // decides if the relay should receive packets
  std::atomic<bool> should_receive(true);

  // session map to be shared across packet processors
  core::SessionMap sessions;

  auto close_sockets = [&sockets] {
    for (auto& socket : sockets) {
      socket->close();
    }
  };

  auto join_threads = [&threads] {
    for (auto& thread : threads) {
      thread->join();
    }
  };

  auto cleanup = [&] {
    gAlive = false;
    close_sockets();
  };

  // makes a shared ptr to a socket object
  auto make_socket = [&](net::Address& addr_in) -> os::SocketPtr {
    // don't set addr, so that it's 0.0.0.0:some-port
    net::Address addr;
    addr.Port = addr_in.Port;
    addr.Type = addr_in.Type;
    auto socket = std::make_shared<os::Socket>();
    if (!socket->create(os::SocketType::Blocking, addr, socketSendBuffSize, socketRecvBuffSize, 0.0f, true)) {
      return nullptr;
    }

    // if port was 0, this will set the reference parameter to what it changed to
    addr_in.Port = addr.Port;

    sockets.push_back(socket);

    return socket;
  };

  size_t num_threads = (num_cpus == 1) ? num_cpus : num_cpus - 1;
  LOG(INFO, "creating ", num_cpus, " packet processing thread", (num_cpus != 1) ? "s" : "");

  // packet processing setup
  for (unsigned int i = (num_cpus == 1) ? 0 : 1; i < num_cpus && gAlive; i++) {
    auto socket = make_socket(relay_addr);
    if (!socket) {
      LOG(ERROR, "could not create socket");
      cleanup();
    }

    auto thread = std::make_shared<std::thread>([socket, &] {
      core::PacketProcessor processor(
       should_receive, *socket, keychain, sessions, relay_manager, gAlive, recorder, router_info);
      processor.process(socketAndThreadReady);
    });

    {
      auto [ok, err] = os::set_thread_affinity(*thread, (std::thread::hardware_concurrency() == 1) ? 0 : i);
      if (!ok) {
        LOG(ERROR, err);
      }
    }

    {
      auto [ok, err] = os::set_thread_sched_max(*thread);
      if (!ok) {
        LOG(ERROR, err);
      }
    }

    sockets.push_back(socket);
    threads.push_back(thread);
  }

  // gets a socket from those available round robbin style
  auto next_socket = [&] {
    static size_t socketChooser = 0;
    return sockets[socketChooser++ % sockets.size()];
  };

  // ping processing setup
  if (gAlive) {
    auto socket = next_socket();
    auto thread = std::make_shared<std::thread>([socket, &] {
      core::PingProcessor pinger(*socket, relay_manager, gAlive, recorder);
      pinger.process(socketAndThreadReady);
    });

    {
      auto [ok, err] = os::set_thread_affinity(*thread, 0);
      if (!ok) {
        LOG(ERROR, err);
      }
    }

    {
      auto [ok, err] = os::set_thread_sched_max(*thread);
      if (!ok) {
        LOG(ERROR, err);
      }
    }

    sockets.push_back(socket);
    threads.push_back(thread);
  }

  // new backend setup
  {
    std::thread thread([&] {
      bool relayInitialized = false;

      net::BeastWrapper wrapper;
      core::Backend backend(
       env.backend_hostname, relay_addr.toString(), keychain, router_info, relay_manager, b64RelayPubKey, sessions, wrapper);

      for (int i = 0; i < 60; ++i) {
        if (backend.init()) {
          std::cout << '\n';
          relayInitialized = true;
          break;
        }

        std::this_thread::sleep_for(1s);
      }

      if (!relayInitialized) {
        LOG("error: could not initialize relay");
        cleanup();
      }

      LOG("relay initialized with new backend");

      if (gAlive) {
        setup_signal_handlers();

        success = backend.updateCycle(gAlive, gShouldCleanShutdown, recorder, sessions);
      }
    });

    {
      auto [ok, err] = os::set_thread_affinity(thread, 0);
      if (!ok) {
        LOG(err);
      }
    }

    {
      auto [ok, err] = os::set_thread_sched_max(thread);
      if (!ok) {
        LOG(err);
      }
    }

    thread.join();
  }

  should_receive = false;

  cleanup();

  join_threads();

  LOG(DEBUG, "Receiving Address: ", relay_addr);

  return success ? 0 : 1;
}