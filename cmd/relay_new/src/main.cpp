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
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "relay/relay.hpp"
#include "relay/relay_platform.hpp"
#include "testing/test.hpp"
#include "util/env.hpp"

using namespace std::chrono_literals;

namespace
{
  volatile bool gAlive = true;

  void interrupt_handler(int signal)
  {
    (void)signal;
    gAlive = false;
  }

  void segfaultHandler(int sig)
  {
    gAlive = false;
    const auto StacktraceDepth = 13;
    void* arr[StacktraceDepth];

    // get stack frames
    size_t size = backtrace(arr, StacktraceDepth);

    // print the stack trace
    std::cerr << "Error: signal " << sig << ":\n";
    backtrace_symbols_fd(arr, size, STDERR_FILENO);
    exit(1);
  }

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
}  // namespace

int main()
{
#ifndef NDEBUG
  signal(SIGSEGV, segfaultHandler);
#endif

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

  unsigned int numProcessors = 0;
  if (!getNumProcessors(env, numProcessors)) {
    return 1;
  }

  int socketRecvBuffSize = getBufferSize(env.RecvBufferSize);
  int socketSendBuffSize = getBufferSize(env.SendBufferSize);

  LogDebug("Socket recv buffer size is ", socketRecvBuffSize, " bytes");
  LogDebug("Socket send buffer size is ", socketSendBuffSize, " bytes");

  std::unique_ptr<std::ofstream> output;
  std::unique_ptr<util::ThroughputLogger> logger;
  {
    if (!env.LogFile.empty()) {
      auto file = std::make_unique<std::ofstream>();
      file->open(env.LogFile);

      if (*file) {
        output = std::move(file);
      }

      if (output) {
        logger = std::make_unique<util::ThroughputLogger>(*output, true);
      } else {
        logger = std::make_unique<util::ThroughputLogger>(std::cout, false);
      }
    }
  }

  if (relay::relay_initialize() != RELAY_OK) {
    Log("error: failed to initialize relay\n\n");
    return 1;
  }

  Log("Initializing relay\n");

  core::RouterInfo routerInfo;
  core::RelayManager relayManager(relayClock);

  LogDebug("creating sockets and threads");

  // these next four variables are used to force the threads to be
  // created serially
  std::atomic<bool> socketAndThreadReady(false);
  std::mutex lock;
  std::unique_lock<std::mutex> waitLock(lock);
  std::condition_variable waitVar;

  std::vector<os::SocketPtr> sockets;
  std::unique_ptr<std::thread> pingThread;
  std::vector<std::unique_ptr<std::thread>> packetThreads;

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
  auto joinThreads = [&pingThread, &packetThreads] {
    pingThread->join();
    for (auto& thread : packetThreads) {
      thread->join();
    }
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
    packetThreads.resize(numProcessors);

    for (unsigned int i = 0; i < numProcessors; i++) {
      auto packetSocket = makeSocket(relayAddr.Port);
      {
        if (!packetSocket) {
          Log("could not create packetSocket");
          ::gAlive = false;
          closeSockets();
          joinThreads();
          relay::relay_term();
          return 1;
        }
      }

      sockets.push_back(packetSocket);

      // clang-format did this
      packetThreads[i] = std::make_unique<std::thread>([&waitVar,
                                                        &socketAndThreadReady,
                                                        &shouldReceive,
                                                        packetSocket,
                                                        &relayClock,
                                                        &keychain,
                                                        &sessions,
                                                        &relayManager,
                                                        &logger,
                                                        &relayAddr] {
        core::PacketProcessor processor(
         shouldReceive, *packetSocket, relayClock, keychain, sessions, relayManager, ::gAlive, *logger, relayAddr);
        processor.process(waitVar, socketAndThreadReady);
      });

      wait();  // wait the the packet processor is ready to receive

      int error;
      if (!os::SetThreadAffinity(*packetThreads[i], i, error)) {
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
    net::Address bindAddr = bindAddr;
    {
      bindAddr.Port = 0;  // make sure the port is dynamically assigned
    }

    auto pingSocket = makeSocket(bindAddr.Port);
    if (!pingSocket) {
      Log("could not create pingSocket");
      ::gAlive = false;
      relay::relay_term();
      closeSockets();
      joinThreads();
      return 1;
    }

    sockets.push_back(pingSocket);

    // setup the ping processor to use the external address
    // relays use it to know where the receiving port of other relays are
    pingThread = std::make_unique<std::thread>([&waitVar, &socketAndThreadReady, pingSocket, &relayManager, &relayAddr] {
      core::PingProcessor pingProcessor(*pingSocket, relayManager, ::gAlive, relayAddr);
      pingProcessor.process(waitVar, socketAndThreadReady);
    });

    wait();

    int error;
    if (!os::SetThreadAffinity(*pingThread, getPingProcNum(numProcessors), error)) {
      Log("Error setting thread affinity: ", error);
    }
  }

  LogDebug("communicating with backend");
  bool relay_initialized = false;

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
    ::gAlive = false;
    joinThreads();
    closeSockets();
    relay::relay_term();
    return 1;
  }

  Log("Relay initialized\n\n");

  // ctrl c shuts down gracefully
  signal(SIGINT, interrupt_handler);

  // and so do these
  signal(SIGTERM, interrupt_handler);
  signal(SIGHUP, interrupt_handler);

  backend.updateCycle(gAlive, *logger, sessions, relayClock);

  gAlive = false;

  Log("Cleaning up\n");

  shouldReceive = false;

  LogDebug("Closing sockets");
  closeSockets();

  LogDebug("Joining threads");
  joinThreads();

  LogDebug("Closing log file");
  if (output) {
    output->close();
  }

  LogDebug("Terminating relay");
  relay::relay_term();

  LogDebug("Relay terminated. Address: ", relayAddr);

  return 0;
}