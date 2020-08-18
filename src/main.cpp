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
#include "relay/relay.hpp"
#include "relay/relay_platform.hpp"
#include "testing/test.hpp"
#include "util/env.hpp"

using namespace std::chrono_literals;

volatile bool gAlive = true;
volatile bool gShouldCleanShutdown = false;

namespace base64 = encoding::base64;

namespace
{
  // TODO move this out of main and somewhere else to allow for test coverage
  inline bool get_crypto_keys(const util::Env& env, crypto::Keychain& keychain, std::string& b64_relay_pub_key)
  {
    // relay private key
    {
      std::string b64_relay_priv_key = env.relay_private_key;
      auto len = base64::decode(b64_relay_priv_key, keychain.RelayPrivateKey);
      if (len != crypto::KeySize) {
        LOG("error: invalid relay private key");
        return false;
      }

      LOG("relay private key is '", env.relay_private_key, "'\n");
    }

    // relay public key
    {
      b64_relay_pub_key = env.relay_public_key;
      auto len = base64::decode(b64_relay_pub_key, keychain.RelayPublicKey);
      if (len != crypto::KeySize) {
        LOG("error: invalid relay public key");
        return false;
      }

      LOG("relay public key is '", env.relay_public_key, "'\n");
    }

    // router public key
    {
      std::string b64RouterPublicKey = env.relay_router_public_key;
      auto len = base64::decode(b64RouterPublicKey, keychain.RouterPublicKey);
      if (len != crypto::KeySize) {
        LOG("error: invalid router public key");
        return false;
      }

      LOG("router public key is '", env.relay_router_public_key, "'\n");
    }

    return true;
  }

  inline bool get_num_procs(const util::Env& env, unsigned int& num_procs)
  {
    if (env.max_cpus.empty()) {
      num_procs = std::thread::hardware_concurrency();  // first core reserved for updates/outgoing pings
      if (num_procs > 0) {
        LOG("RELAY_MAX_CORES not set, autodetected number of processors available: ", num_procs);
      } else {
        LOG("error: RELAY_MAX_CORES not set, could not detect processor count, please set the env var");
        return false;
      }
    } else {
      try {
        num_procs = std::stoi(env.max_cpus);
      } catch (std::exception& e) {
        LOG("could not parse RELAY_MAX_CORES to a number, value: ", env.max_cpus);
      }
    }

    return true;
  }

  inline int get_buffer_size(const std::string& envvar)
  {
    int socketBufferSize = 1000000;

    if (!envvar.empty()) {
      try {
        socketBufferSize = std::stoi(envvar);
      } catch (std::exception& e) {
        LOG("Could not parse ", envvar, " env var to a number: ", e.what());
      }
    }

    return socketBufferSize;
  }

  inline void setup_signal_handlers()
  {
#ifndef NDEBUG
    signal(SIGSEGV, [](int) {
      gAlive = false;
      const auto STACKTRACE_DEPTH = 13;
      void* arr[STACKTRACE_DEPTH];

      // get stack frames
      size_t size = backtrace(arr, STACKTRACE_DEPTH);

      // print the stack trace
      std::cerr << "stacktrace\n";
      backtrace_symbols_fd(arr, size, STDERR_FILENO);
      exit(1);
    });
#endif

#if not defined TEST_BUILD and not defined BENCH_BUILD
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
#endif
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

  std::cout << "\nNetwork Next Relay\n";

  std::cout << "\nEnvironment:\n\n";

  util::Env env;

  // relay address - the address other devices should use to talk to this
  // sent to the relay backend and is the addr everything communicates with
  net::Address relay_addr;
  {
    if (!relay_addr.parse(env.relay_address)) {
      LOG("error: invalid relay address: ", env.relay_address);
      return 1;
    }

    std::cout << "    relay address is '" << relay_addr << "'\n";
  }

  crypto::Keychain keychain;
  std::string b64RelayPubKey;
  if (!get_crypto_keys(env, keychain, b64RelayPubKey)) {
    return 1;
  }

  std::cout << "    backend hostname is '" << env.backend_hostname << "'\n";

  unsigned int numProcessors = 0;
  if (!get_num_procs(env, numProcessors)) {
    return 1;
  }

  int socketRecvBuffSize = get_buffer_size(env.recv_buffer_size);
  int socketSendBuffSize = get_buffer_size(env.send_buffer_size);

  if (relay::relay_initialize() != RELAY_OK) {
    LOG("error: failed to initialize relay\n\n");
    return 1;
  }

  LOG("Initializing relay");

  bool success = false;

  core::RouterInfo routerInfo;
  core::RelayManager<core::Relay> relayManager;
  util::ThroughputRecorder recorder;

  // used to make sockets and threads serially
  std::atomic<bool> socketAndThreadReady(false);

  std::vector<os::SocketPtr> sockets;
  std::vector<std::shared_ptr<std::thread>> threads;

  // decides if the relay should receive packets
  std::atomic<bool> shouldReceive(true);

  // session map to be shared across packet processors
  core::SessionMap sessions;

  auto nextSocket = [&sockets] {
    static size_t socketChooser = 0;
    return sockets[socketChooser++ % sockets.size()];
  };

  // wait until a thread is ready to do its job.
  // serializes the thread spawning so the relay doesn't
  // communicate with the backend until it is fully ready
  // to receive packets
  auto wait = [&socketAndThreadReady] {
    while (!socketAndThreadReady) {
      std::this_thread::sleep_for(10ms);
    }

    socketAndThreadReady = false;
  };

  // closes all opened sockets in the vector
  auto closeSockets = [&sockets] {
    for (auto& socket : sockets) {
      socket->close();
    }
  };

  // joins all threads that were placed in the vector
  auto joinThreads = [&threads] {
    for (auto& thread : threads) {
      thread->join();
    }
  };

  auto cleanup = [&closeSockets, &joinThreads] {
    gAlive = false;
    closeSockets();
    relay::relay_term();
  };

  // makes a shared ptr to a socket object
  auto makeSocket = [&sockets, socketSendBuffSize, socketRecvBuffSize](uint16_t& portNumber) -> os::SocketPtr {
    // don't set addr, so that it's 0.0.0.0:some-port
    net::Address addr;
    addr.Port = portNumber;
    addr.Type = net::AddressType::IPv4;
    auto socket = std::make_shared<os::Socket>(os::SocketType::Blocking);
    if (!socket->create(addr, socketSendBuffSize, socketRecvBuffSize, 0.0f, true)) {
      return nullptr;
    }

    // if port was 0, this will set the reference parameter to what it changed to
    portNumber = addr.Port;

    sockets.push_back(socket);

    return socket;
  };

  // packet processing setup
  LOG("creating ", (numProcessors == 1) ? 1 : numProcessors - 1, " packet processing threads");
  {
    for (unsigned int i = ((numProcessors == 1) ? 0 : 1); i < numProcessors && gAlive; i++) {
      auto socket = makeSocket(relay_addr.Port);
      if (!socket) {
        LOG("could not create socket");
        cleanup();
      }

      auto thread = std::make_shared<std::thread>(
       [&socketAndThreadReady, &shouldReceive, socket, &keychain, &sessions, &relayManager, &recorder, &routerInfo] {
         core::PacketProcessor processor(
          shouldReceive, *socket, keychain, sessions, relayManager, gAlive, recorder, routerInfo);
         processor.process(socketAndThreadReady);
       });

      wait();  // wait the the packet processor is ready to receive

      sockets.push_back(socket);
      threads.push_back(thread);

      {
        auto [ok, err] = os::SetThreadAffinity(*thread, (std::thread::hardware_concurrency() == 1) ? 0 : i);
        if (!ok) {
          LOG(err);
        }
      }

      {
        auto [ok, err] = os::SetThreadSchedMax(*thread);
        if (!ok) {
          LOG(err);
        }
      }
    }
  }

  // ping processing setup
  if (gAlive) {
    auto socket = nextSocket();
    auto thread = std::make_shared<std::thread>([&socketAndThreadReady, socket, &relayManager, &recorder] {
      core::PingProcessor pingProcessor(*socket, relayManager, gAlive, recorder);
      pingProcessor.process(socketAndThreadReady);
    });

    wait();

    sockets.push_back(socket);
    threads.push_back(thread);

    {
      auto [ok, err] = os::SetThreadAffinity(*thread, 0);
      if (!ok) {
        LOG(err);
      }
    }

    {
      auto [ok, err] = os::SetThreadSchedMax(*thread);
      if (!ok) {
        LOG(err);
      }
    }
  }

  // new backend setup
  {
    std::thread updateThread(
     [&env, &relay_addr, &keychain, &routerInfo, &relayManager, &b64RelayPubKey, &sessions, &cleanup, &recorder, &success] {
       bool relayInitialized = false;

       net::BeastWrapper wrapper;
       core::Backend backend(
        env.backend_hostname, relay_addr.toString(), keychain, routerInfo, relayManager, b64RelayPubKey, sessions, wrapper);

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
      auto [ok, err] = os::SetThreadAffinity(updateThread, 0);
      if (!ok) {
        LOG(err);
      }
    }

    {
      auto [ok, err] = os::SetThreadSchedMax(updateThread);
      if (!ok) {
        LOG(err);
      }
    }

    updateThread.join();
  }

  LOG("cleaning up");

  shouldReceive = false;

  cleanup();
  joinThreads();

  LOG_DEBUG("Receiving Address: ", relay_addr);

  return success ? 0 : 1;
}