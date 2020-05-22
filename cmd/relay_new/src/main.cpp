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
#include "legacy/v3/backend.hpp"
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

    return true;
  }

  inline bool getNumProcessors(const util::Env& env, unsigned int& numProcs)
  {
    if (env.ProcessorCount.empty()) {
      numProcs = std::thread::hardware_concurrency();
      if (numProcs > 0) {
        Log("RELAY_PROCESSOR_COUNT not set, autodetected number of processors available: ", numProcs);
      } else {
        Log("error: RELAY_PROCESSOR_COUNT not set, could not detect processor count, please set the env var");
        return false;
      }
    } else {
      try {
        numProcs = std::stoi(env.ProcessorCount);
      } catch (std::exception& e) {
        Log("could not parse RELAY_PROCESSOR_COUNT to a number, value: ", env.ProcessorCount);
      }
    }

    return true;
  }

  inline bool getPingProcNum(unsigned int numProcs)
  {
    auto actualProcCount = std::thread::hardware_concurrency();

    // if already using all available procs, just use the first
    // else use the next one
    return actualProcCount > 0 && numProcs == actualProcCount ? 0 : numProcs + 1;
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

  const util::Clock relayClock;

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

  legacy::v3::TrafficStats v3TrafficStats;
  core::RouterInfo routerInfo;
  core::RelayManager<core::Relay> relayManager(relayClock);
  core::RelayManager<core::V3Relay> v3RelayManager(relayClock);
  util::ThroughputRecorder recorder;
  auto chan = util::makeChannel<core::GenericPacket<>>();
  auto sender = std::get<0>(chan);
  auto receiver = std::get<1>(chan);

  // used to make sockets and threads serially
  std::atomic<bool> socketAndThreadReady(false);

  std::vector<os::SocketPtr> sockets;
  std::vector<std::shared_ptr<std::thread>> threads;

  // only used for v3 compatability
  const auto relayID = crypto::FNV(env.RelayV3Name);

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
    joinThreads();
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

  /* packet processing setup
   * must come before ping setup
   * otherwise ping may take the port that is reserved for packet processing, usually 40000
   * odds are slim but it may happen
   */
  Log("creating ", numProcessors, " packet processing threads");
  {
    for (unsigned int i = 0; i < numProcessors; i++) {
      auto socket = makeSocket(relayAddr.Port);
      if (!socket) {
        Log("could not create socket");
        cleanup();
        return 1;
      }

      auto thread = std::make_shared<std::thread>([&socketAndThreadReady,
                                                   &shouldReceive,
                                                   socket,
                                                   &relayClock,
                                                   &keychain,
                                                   &sessions,
                                                   &relayManager,
                                                   &v3RelayManager,
                                                   &recorder,
                                                   &sender,
                                                   &v3TrafficStats,
                                                   relayID] {
        core::PacketProcessor processor(
         shouldReceive,
         *socket,
         relayClock,
         keychain,
         sessions,
         relayManager,
         v3RelayManager,
         gAlive,
         recorder,
         sender,
         v3TrafficStats,
         relayID);
        processor.process(socketAndThreadReady);
      });

      wait();  // wait the the packet processor is ready to receive

      sockets.push_back(socket);
      threads.push_back(thread);

      int error;
      if (!os::SetThreadAffinity(*thread, i, error)) {
        Log("Error setting thread affinity: ", error);
      }
    }
  }

  // ping processing setup
  {
    auto socket = nextSocket();
    auto thread =
     std::make_shared<std::thread>([&socketAndThreadReady, socket, &relayManager, &recorder, &v3TrafficStats, &relayID] {
       core::PingProcessor pingProcessor(*socket, relayManager, gAlive, recorder, v3TrafficStats, relayID);
       pingProcessor.process(socketAndThreadReady);
     });

    wait();

    sockets.push_back(socket);
    threads.push_back(thread);

    int error;
    if (!os::SetThreadAffinity(*thread, getPingProcNum(numProcessors), error)) {
      Log("error setting thread affinity: ", error);
    }
  }

  bool v3BackendSuccess = true;

  // v3 backend compatability setup
  if (env.RelayV3Enabled == "1") {
    v3BackendSuccess = false;
    // ping proc setup
    {
      auto socket = nextSocket();
      auto thread =
       std::make_shared<std::thread>([&socketAndThreadReady, socket, &v3RelayManager, &recorder, &v3TrafficStats, &relayID] {
         core::PingProcessor pingProcessor(*socket, v3RelayManager, gAlive, recorder, v3TrafficStats, relayID);
         pingProcessor.process(socketAndThreadReady);
       });

      wait();

      sockets.push_back(socket);
      threads.push_back(thread);

      int error;
      if (!os::SetThreadAffinity(*thread, getPingProcNum(numProcessors), error)) {
        Log("error setting thread affinity: ", error);
      }
    }

    // backend setup
    {
      auto socket = nextSocket();
      auto thread = std::make_shared<std::thread>(
       [&receiver, &env, socket, &cleanup, &v3BackendSuccess, &relayClock, &v3TrafficStats, &v3RelayManager, &relayID] {
         size_t speed = std::stoi(env.RelayV3Speed) * 1000000;
         legacy::v3::Backend backend(receiver, env, relayID, *socket, relayClock, v3TrafficStats, v3RelayManager, speed);

         if (!backend.init()) {
           Log("could not initialize relay with old backend");
           cleanup();
           return;
         }

         Log("relay initialized with old backend");

         if (!backend.config()) {
           Log("could not configure relay with old backend");
           cleanup();
           return;
         }

         Log("old backend entering update cycle");

         v3BackendSuccess = backend.updateCycle(gAlive);

         gAlive = false;
       });

      sockets.push_back(socket);
      threads.push_back(thread);

      Log("relay configured with old backend using address ", relayAddr);
    }
  }

  core::Backend<net::CurlWrapper> backend(
   env.BackendHostname, relayAddr.toString(), keychain, routerInfo, relayManager, b64RelayPubKey, sessions, v3TrafficStats);

  bool relayInitialized = false;

  for (int i = 0; i < 60; ++i) {
    if (backend.init()) {
      std::cout << '\n';
      relayInitialized = true;
      break;
    }

    std::cout << '.' << std::flush;

    std::this_thread::sleep_for(1s);
  }

  if (!relayInitialized) {
    Log("error: could not initialize relay");
    cleanup();
    return 1;
  }

  Log("relay initialized with new backend");

  setupSignalHandlers();

  bool success = backend.updateCycle(gAlive, gShouldCleanShutdown, recorder, sessions, relayClock);

  Log("cleaning up");

  receiver.close();
  sender.close();  // redundant

  shouldReceive = false;

  cleanup();

  LogDebug("Receiving Address: ", relayAddr);

  return (success && v3BackendSuccess) ? 0 : 1;
}