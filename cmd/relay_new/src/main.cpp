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
#include "core/v3_backend.hpp"
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

    return true;
  }

  inline bool getNumProcessors(const util::Env& env, unsigned int& numProcs)
  {
    if (env.ProcessorCount.empty()) {
      numProcs = std::thread::hardware_concurrency();
      if (numProcs > 0) {
        Log("RELAY_PROCESSOR_COUNT not set, autodetected number of processors available: ", numProcs, "\n\n");
      } else {
        Log("error: RELAY_PROCESSOR_COUNT not set, could not detect processor count, please set the env var\n\n");
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
      gAlive = false;
    };

    auto cleanShutdownHandler = [](int) {
      gShouldCleanShutdown = true;
      gAlive = false;
    };

    signal(SIGINT, gracefulShutdownHandler);
    signal(SIGTERM, gracefulShutdownHandler);
    signal(SIGHUP, cleanShutdownHandler);
#endif
  }
}  // namespace

int main()
{
  setupSignalHandlers();

#ifdef TEST_BUILD
  return testing::SpecTest::Run() ? 0 : 1;
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
      Log("error: invalid relay address '", env.RelayAddress, "'\n");
      return 1;
    }

    std::cout << "    relay address is '" << relayAddr << "'\n";
  }

  crypto::Keychain keychain;
  std::string b64RelayPubKey;
  if (!getCryptoKeys(env, keychain, b64RelayPubKey)) {
    return 1;
  }

  std::string backendHostname = env.BackendHostname;
  std::cout << "    backend hostname is '" << backendHostname << "'\n";

  // v3 backend hostname
  net::Address v3BackendAddr;
  {
    if (!v3BackendAddr.resolve(env.RelayV3BackendHostname, env.RelayV3BackendPort)) {
      Log("Could not resolve the v3 backend hostname to an ip address");
      return 1;
    }
  }

  std::cout << "    v3 ip is '" << v3BackendAddr.toString() << "'\n";

  unsigned int numProcessors = 0;
  if (!getNumProcessors(env, numProcessors)) {
    return 1;
  }

  int socketRecvBuffSize = getBufferSize(env.RecvBufferSize);
  int socketSendBuffSize = getBufferSize(env.SendBufferSize);

  LogDebug("Socket recv buffer size is ", socketRecvBuffSize, " bytes");
  LogDebug("Socket send buffer size is ", socketSendBuffSize, " bytes");

  if (relay::relay_initialize() != RELAY_OK) {
    Log("error: failed to initialize relay\n\n");
    return 1;
  }

  Log("Initializing relay\n");

  core::RouterInfo routerInfo;
  core::RelayManager relayManager(relayClock);
  util::ThroughputRecorder recorder;

  LogDebug("creating sockets and threads");

  // these next four variables are used to force the threads to be
  // created serially
  std::atomic<bool> socketAndThreadReady(false);
  std::mutex lock;
  std::unique_lock<std::mutex> waitLock(lock);
  std::condition_variable waitVar;

  std::vector<os::SocketPtr> sockets;
  std::vector<std::shared_ptr<std::thread>> threads;

  // the relay address that should be exposed to other relays.
  // do not set this using the value from the env var.
  // if using port 0, another port will be selected byt the os
  // which will cause a mismatch between what this value contains
  // and the port selected by the os
  std::string relayAddrString;

  // decides if the relay should receive packets
  std::atomic<bool> shouldReceive = true;

  // session map to be shared across packet processors
  core::SessionMap sessions;

  // wait until a thread is ready to do its job.
  // serializes the thread spawning so the relay doesn't
  // communicate with the backend until it is fully ready
  // to receive packets
  auto wait = [&waitVar, &waitLock, &socketAndThreadReady] {
    waitVar.wait(waitLock, [&socketAndThreadReady]() -> bool {
      return socketAndThreadReady;
    });
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
    net::Address addr;
    addr.Port = portNumber;
    addr.Type = net::AddressType::IPv4;
    auto socket = std::make_shared<os::Socket>(os::SocketType::Blocking);
    if (!socket->create(addr, socketSendBuffSize, socketRecvBuffSize, 0.0f, true, 0)) {
      return nullptr;
    }
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
      auto packetSocket = makeSocket(relayAddr.Port);
      {
        if (!packetSocket) {
          Log("could not create packetSocket");
          cleanup();
          return 1;
        }
      }

      auto thread = std::make_shared<std::thread>([&waitVar,
                                                   &socketAndThreadReady,
                                                   &shouldReceive,
                                                   packetSocket,
                                                   &relayClock,
                                                   &keychain,
                                                   &sessions,
                                                   &relayManager,
                                                   &recorder,
                                                   &relayAddr] {
        core::PacketProcessor processor(
         shouldReceive, *packetSocket, relayClock, keychain, sessions, relayManager, gAlive, recorder, relayAddr);
        processor.process(waitVar, socketAndThreadReady);
      });

      wait();  // wait the the packet processor is ready to receive

      sockets.push_back(packetSocket);
      threads.push_back(thread);

      int error;
      if (!os::SetThreadAffinity(*thread, i, error)) {
        Log("Error setting thread affinity: ", error);
      }
    }
  }

  // if using port 0, it is discovered in the first socket's create() function.
  // That being said packet sockets must be created before communicating with the
  // backend otherwise port 0 will be sent, and this string must be set afterwards
  // too for that same reason
  relayAddr.toString(relayAddrString);

  LogDebug("Actual address: ", relayAddrString);

  // ping processing setup
  // pings are sent out on a different port number than received
  // if they are the same the relay behaves weird, it'll sometimes behave right
  // othertimes it'll just ignore everything coming to it
  {
    net::Address bindAddr = relayAddr;
    {
      bindAddr.Port = 0;  // make sure the port is dynamically assigned
    }

    auto socket = makeSocket(bindAddr.Port);
    if (!socket) {
      Log("could not create pingSocket");
      cleanup();
      return 1;
    }

    // setup the ping processor to use the external address
    // relays use it to know where the receiving port of other relays are
    auto thread = std::make_shared<std::thread>([&waitVar, &socketAndThreadReady, socket, &relayManager, &relayAddr] {
      core::PingProcessor pingProcessor(*socket, relayManager, gAlive, relayAddr);
      pingProcessor.process(waitVar, socketAndThreadReady);
    });

    wait();

    sockets.push_back(socket);
    threads.push_back(thread);

    int error;
    if (!os::SetThreadAffinity(*thread, getPingProcNum(numProcessors), error)) {
      Log("Error setting thread affinity: ", error);
    }
  }

  LogDebug("communicating with backend");
  bool relay_initialized = false;
  bool v3BackendSuccess = false;

  // v3 backend compatability setup
  {
    net::Address bindAddr = relayAddr;
    {
      bindAddr.Port = 0;
    }

    auto socket = makeSocket(bindAddr.Port);
    if (!socket) {
      Log("could not create v3 backend socket");
      cleanup();
      return 1;
    }

    auto thread = std::make_shared<std::thread>([&v3BackendAddr, socket, &cleanup, &v3BackendSuccess] {
      core::V3Backend backend(v3BackendAddr, *socket);

      if (!backend.init()) {
        cleanup();
        return;
      }

      v3BackendSuccess = backend.updateCycle(gAlive);

      gAlive = false;
    });

    sockets.push_back(socket);
    threads.push_back(thread);
  }

  core::Backend<net::CurlWrapper> backend(
   backendHostname, relayAddrString, keychain, routerInfo, relayManager, b64RelayPubKey, sessions);

  for (int i = 0; i < 60; ++i) {
    if (backend.init()) {
      std::cout << '\n';
      relay_initialized = true;
      break;
    }

    std::cout << '.' << std::flush;

    std::this_thread::sleep_for(1s);
  }

  if (!relay_initialized) {
    Log("error: could not initialize relay\n\n");
    cleanup();
    return 1;
  }

  Log("Relay initialized\n\n");

  bool success = backend.updateCycle(gAlive, gShouldCleanShutdown, recorder, sessions, relayClock);

  Log("Cleaning up");

  shouldReceive = false;

  cleanup();

  LogDebug("Relay terminated. Address: ", relayAddr);

  return (success && v3BackendSuccess) ? 0 : 1;
}