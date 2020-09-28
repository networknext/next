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

namespace
{
  // TODO move this out of main and somewhere else to allow for test coverage
  inline bool getCryptoKeys(const util::Env& env, crypto::Keychain& keychain, std::string& b64RelayPubKey)
  {
    // relay private key
    {
      std::string b64RelayPrivateKey = env.RelayPrivateKey;
      auto len = encoding::base64::Decode(b64RelayPrivateKey, keychain.RelayPrivateKey);
      if (len != crypto::KeySize) {
        std::cout << "error: invalid relay private key\n";
        return false;
      }
      std::cout << "    relay private key is '" << env.RelayPrivateKey << "'\n";
    }

    // relay public key
    {
      b64RelayPubKey = env.RelayPublicKey;
      auto len = encoding::base64::Decode(b64RelayPubKey, keychain.RelayPublicKey);
      if (len != crypto::KeySize) {
        std::cout << "error: invalid relay public key\n";
        return false;
      }

      std::cout << "    relay public key is '" << env.RelayPublicKey << "'\n";
    }

    // router public key
    {
      std::string b64RouterPublicKey = env.RelayRouterPublicKey;
      auto len = encoding::base64::Decode(b64RouterPublicKey, keychain.RouterPublicKey);
      if (len != crypto::KeySize) {
        std::cout << "error: invalid router public key\n";
        return false;
      }

      std::cout << "    router public key is '" << env.RelayRouterPublicKey << "'\n";
    }

    // update key

    if (env.RelayV3Enabled == "1") {
      std::string b64UpdateKey = env.RelayV3UpdateKey;
      auto len = encoding::base64::Decode(b64UpdateKey, keychain.UpdateKey);
      if (len != crypto_sign_SECRETKEYBYTES) {
        std::cout << "error: invalid update key\n";
        return false;
      }

      std::cout << "    update key is '" << env.RelayV3UpdateKey << "'\n";
    }

    return true;
  }

  inline bool getNumProcessors(const util::Env& env, unsigned int& numProcs)
  {
    if (env.ProcessorCount.empty()) {
      numProcs = std::thread::hardware_concurrency();  // first core reserved for updates/outgoing pings
      if (numProcs > 0) {
        Log("RELAY_MAX_CORES not set, autodetected number of processors available: ", numProcs);
      } else {
        Log("error: RELAY_MAX_CORES not set, could not detect processor count, please set the env var");
        return false;
      }
    } else {
      try {
        numProcs = std::stoi(env.ProcessorCount);
      } catch (std::exception& e) {
        Log("could not parse RELAY_MAX_CORES to a number, value: ", env.ProcessorCount);
      }
    }

    return true;
  }

  inline int getBufferSize(const std::string& envvar)
  {
    int socketBufferSize = 1000000;

    if (!envvar.empty()) {
      try {
        socketBufferSize = std::stoi(envvar);
      } catch (std::exception& e) {
        Log("Could not parse ", envvar, " env var to a number: ", e.what());
      }
    }

    return socketBufferSize;
  }

  inline void setupSignalHandlers()
  {
#ifndef NDEBUG
    signal(SIGSEGV, [](int) {
      gAlive = false;
      const auto StacktraceDepth = 13;
      void* arr[StacktraceDepth];

      // get stack frames
      size_t size = backtrace(arr, StacktraceDepth);

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
  net::Address relayAddr;
  {
    if (!relayAddr.parse(env.RelayAddress)) {
      Log("error: invalid relay address: ", env.RelayAddress);
      return 1;
    }

    std::cout << "    relay address is '" << relayAddr << "'\n";
  }

  crypto::Keychain keychain;
  std::string b64RelayPubKey;
  if (!getCryptoKeys(env, keychain, b64RelayPubKey)) {
    return 1;
  }

  std::cout << "    backend hostname is '" << env.BackendHostname << "'\n";
  std::cout << "    v3 backend hostname is '" << env.RelayV3BackendHostname << ':' << env.RelayV3BackendPort << "'\n";

  unsigned int numProcessors = 0;
  if (!getNumProcessors(env, numProcessors)) {
    return 1;
  }

  int socketRecvBuffSize = getBufferSize(env.RecvBufferSize);
  int socketSendBuffSize = getBufferSize(env.SendBufferSize);

  if (relay::relay_initialize() != RELAY_OK) {
    Log("error: failed to initialize relay\n\n");
    return 1;
  }

  Log("Initializing relay");

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
  Log("creating ", (numProcessors == 1) ? 1 : numProcessors - 1, " packet processing threads");
  {
    for (unsigned int i = ((numProcessors == 1) ? 0 : 1); i < numProcessors && gAlive; i++) {
      auto socket = makeSocket(relayAddr.Port);
      if (!socket) {
        Log("could not create socket");
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
          Log(err);
        }
      }

      {
        auto [ok, err] = os::SetThreadSchedMax(*thread);
        if (!ok) {
          Log(err);
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
        Log(err);
      }
    }

    {
      auto [ok, err] = os::SetThreadSchedMax(*thread);
      if (!ok) {
        Log(err);
      }
    }
  }

  // new backend setup
  {
    std::thread updateThread(
     [&env, &relayAddr, &keychain, &routerInfo, &relayManager, &b64RelayPubKey, &sessions, &cleanup, &recorder, &success] {
       bool relayInitialized = false;

       net::BeastWrapper wrapper;
       core::Backend backend(
        env.BackendHostname, relayAddr.toString(), keychain, routerInfo, relayManager, b64RelayPubKey, sessions, wrapper);

       for (int i = 0; i < 60; ++i) {
         if (backend.init()) {
           std::cout << '\n';
           relayInitialized = true;
           break;
         }

         std::this_thread::sleep_for(1s);
       }

       if (!relayInitialized) {
         Log("error: could not initialize relay");
         cleanup();
       }

       Log("relay initialized with new backend");

       if (gAlive) {
         setupSignalHandlers();

         success = backend.updateCycle(gAlive, gShouldCleanShutdown, recorder, sessions);
       }
     });

    {
      auto [ok, err] = os::SetThreadAffinity(updateThread, 0);
      if (!ok) {
        Log(err);
      }
    }

    {
      auto [ok, err] = os::SetThreadSchedMax(updateThread);
      if (!ok) {
        Log(err);
      }
    }

    updateThread.join();
  }

  Log("cleaning up");

  shouldReceive = false;

  cleanup();
  joinThreads();

  LogDebug("Receiving Address: ", relayAddr);

  return success ? 0 : 1;
}