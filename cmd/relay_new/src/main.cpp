/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "includes.h"

#include "crypto/keychain.hpp"

#include "encoding/base64.hpp"

#include "bench/bench.hpp"
#include "testing/test.hpp"

#include "relay/relay.hpp"
#include "relay/relay_platform.hpp"

#include "core/router_info.hpp"
#include "core/packet_processor.hpp"
#include "core/ping_processor.hpp"

using namespace std::chrono_literals;

namespace
{
  volatile bool gAlive;

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
    fprintf(stderr, "Error: signal %d:\n", sig);
    backtrace_symbols_fd(arr, size, STDERR_FILENO);
    exit(1);
  }

  inline void updateLoop(CURL* curl,
   const char* backend_hostname,
   const uint8_t* relay_token,
   const char* relay_address_string,
   core::RelayManager& relayManager,
   util::ThroughputLogger& logger)
  {
    std::vector<uint8_t> update_response_memory;
    update_response_memory.resize(RESPONSE_MAX_BYTES);
    while (gAlive) {
      auto bytesReceived = logger.print();
      bool updated = false;

      for (int i = 0; i < 10; ++i) {
        if (relay::relay_update(curl,
             backend_hostname,
             relay_token,
             relay_address_string,
             update_response_memory.data(),
             relayManager,
             bytesReceived) == RELAY_OK) {
          updated = true;
          break;
        }
      }

      if (!updated) {
        printf("error: could not update relay\n\n");
        gAlive = false;
        break;
      }

      std::this_thread::sleep_for(1s);
    }
  }

  inline bool getCryptoKeys(crypto::Keychain& keychain)
  {
    const char* relay_private_key_env = relay::relay_platform_getenv("RELAY_PRIVATE_KEY");
    if (!relay_private_key_env) {
      printf("\nerror: RELAY_PRIVATE_KEY not set\n\n");
      return false;
    }

    if (encoding::base64_decode_data(relay_private_key_env, keychain.RelayPrivateKey.data(), RELAY_PRIVATE_KEY_BYTES) !=
        RELAY_PRIVATE_KEY_BYTES) {
      printf("\nerror: invalid relay private key\n\n");
      return false;
    }

    printf("    relay private key is '%s'\n", relay_private_key_env);

    const char* relay_public_key_env = relay::relay_platform_getenv("RELAY_PUBLIC_KEY");
    if (!relay_public_key_env) {
      printf("\nerror: RELAY_PUBLIC_KEY not set\n\n");
      return false;
    }

    if (encoding::base64_decode_data(relay_public_key_env, keychain.RelayPublicKey.data(), RELAY_PUBLIC_KEY_BYTES) !=
        RELAY_PUBLIC_KEY_BYTES) {
      printf("\nerror: invalid relay public key\n\n");
      return false;
    }

    printf("    relay public key is '%s'\n", relay_public_key_env);

    const char* router_public_key_env = relay::relay_platform_getenv("RELAY_ROUTER_PUBLIC_KEY");
    if (!router_public_key_env) {
      printf("\nerror: RELAY_ROUTER_PUBLIC_KEY not set\n\n");
      return false;
    }

    if (encoding::base64_decode_data(router_public_key_env, keychain.RouterPublicKey.data(), crypto_sign_PUBLICKEYBYTES) !=
        crypto_sign_PUBLICKEYBYTES) {
      printf("\nerror: invalid router public key\n\n");
      return false;
    }

    printf("    router public key is '%s'\n", router_public_key_env);

    return true;
  }

  inline bool getNumProcessors(unsigned int& numProcs)
  {
    const char* nproc = relay::relay_platform_getenv("RELAY_PROCESSOR_COUNT");
    if (nproc == nullptr) {
      numProcs = std::thread::hardware_concurrency();
      if (numProcs > 0) {
        Log("RELAY_PROCESSOR_COUNT not set, autodetected number of processors available: ", numProcs, "\n\n");
      } else {
        Log("error: RELAY_PROCESSOR_COUNT not set, could not detect processor count, please set the env var\n\n");
        return false;
      }
    } else {
      numProcs = std::atoi(nproc);
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

  inline int getBufferSize(const char* envvar)
  {
    int socketBufferSize = 1000000;

    auto env = std::getenv(envvar);
    if (env != nullptr) {
      int num = -1;
      try {
        num = std::stoi(std::string(env));  // to cause an exception to be thrown if not a number
      } catch (std::exception& e) {
        Log("Could not parse ", envvar, " env var to a number: ", e.what());
        std::exit(1);
      }

      if (num < 0) {
        Log("", envvar, " is less than 0");
        std::exit(1);
      }

      socketBufferSize = num;
    }

    return socketBufferSize;
  }
}  // namespace

int main()
{
  gAlive = true;
  signal(SIGSEGV, segfaultHandler);

#ifdef TEST_BUILD
  return testing::SpecTest::Run() ? 0 : 1;
#endif

#ifdef BENCH_BUILD
  benchmarking::Benchmark::Run();
  return 0;
#endif

  const util::Clock relayClock;

  printf("\nNetwork Next Relay\n");

  printf("\nEnvironment:\n\n");

  // relay address - the address other devices should use to talk to this
  // sent to the relay backend and is the addr everything communicates with
  net::Address relayAddr;
  {
    auto env = std::getenv("RELAY_ADDRESS");
    if (env == nullptr) {
      Log("error: RELAY_ADDRESS not set\n");
      return 1;
    }

    if (!relayAddr.parse(env)) {
      Log("error: invalid relay address '", env, "'\n");
      return 1;
    }

    std::cout << "    relay address is '" << relayAddr << "'\n";
  }

  crypto::Keychain keychain;
  if (!getCryptoKeys(keychain)) {
    return 1;
  }

  std::string backendHostname;
  {
    backendHostname = std::getenv("RELAY_BACKEND_HOSTNAME");
    if (backendHostname.empty()) {
      Log("error: RELAY_BACKEND_HOSTNAME not set\n");
      return 1;
    }

    std::cout << "    backend hostname is '" << backendHostname << "'\n";
  }

  unsigned int numProcessors = 0;
  if (!getNumProcessors(numProcessors)) {
    return 1;
  }

  int socketRecvBuffSize = getBufferSize("RELAY_RECV_BUFFER_SIZE");
  int socketSendBuffSize = getBufferSize("RELAY_SEND_BUFFER_SIZE");

  LogDebug("Socket recv buffer size is ", socketRecvBuffSize, " bytes");
  LogDebug("Socket send buffer size is ", socketSendBuffSize, " bytes");

  std::unique_ptr<std::ofstream> output;
  std::unique_ptr<util::ThroughputLogger> logger;
  {
    auto logfile = std::getenv("RELAY_LOG_FILE");
    if (logfile != nullptr && strlen(logfile) != 0) {
      auto file = std::make_unique<std::ofstream>();
      file->open(logfile);

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

  CURL* curl = curl_easy_init();
  if (!curl) {
    Log("error: could not initialize curl\n\n");
    curl_easy_cleanup(curl);
    relay::relay_term();
    return 1;
  }

  uint8_t relay_token[RELAY_TOKEN_BYTES];

  Log("Initializing relay\n");

  core::RouterInfo routerInfo;
  core::RelayManager relayManager(relayClock);

  LogDebug("creating sockets and threads");

  std::atomic<bool> socketAndThreadReady(false);
  std::mutex lock;
  std::unique_lock<std::mutex> waitLock(lock);
  std::condition_variable waitVar;

  std::vector<os::SocketPtr> sockets;
  std::unique_ptr<std::thread> pingThread;
  std::vector<std::unique_ptr<std::thread>> packetThreads;
  std::string relayAddrString;

  // session map to be shared across packet processors
  core::SessionMap sessions;

  // helpful lambdas

  // wait until a processor is set, serializes the blocking io
  // so the relay doesn't commuicate with the backend until it
  // is fully ready to receive packets
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
          gAlive = false;
          closeSockets();
          joinThreads();
          relay::relay_term();
          return 1;
        }
      }

      sockets.push_back(packetSocket);

      packetThreads[i] = std::make_unique<std::thread>(
       [&waitVar, &socketAndThreadReady, packetSocket, &relayClock, &keychain, &routerInfo, &sessions, &relayManager, &logger] {
         core::PacketProcessor processor(
          *packetSocket, relayClock, keychain, routerInfo, sessions, relayManager, gAlive, *logger);
         processor.process(waitVar, socketAndThreadReady);
       });

      wait();  // wait the the processor is ready to receive

      int error;
      if (!os::SetThreadAffinity(*packetThreads[i], i, error)) {
        Log("Error setting thread affinity: ", error);
      }
    }
  }

  // if using port 0, it is discovered in ping socket's create(). That being said sockets
  // must be created before communicating with the backend otherwise port 0 will be reused
  LogDebug("Actual address: ", relayAddr);

  relayAddr.toString(relayAddrString);

  /* ping processing setup
   * pings are sent out on a different port number than received
   * if they are the same the relay behaves weird, it'll sometimes behave right
   * othertimes it'll just ignore everything coming to it
   */
  {
    net::Address bindAddr = bindAddr;
    {
      bindAddr.Port = 0;  // make sure the port is dynamically assigned
    }

    auto pingSocket = makeSocket(bindAddr.Port);
    if (!pingSocket) {
      Log("could not create pingSocket");
      gAlive = false;
      relay::relay_term();
      closeSockets();
      joinThreads();
      return 1;
    }

    sockets.push_back(pingSocket);

    // setup the ping processor to use the external address
    // relays use it to know where the receving port of other relays are
    pingThread = std::make_unique<std::thread>([&waitVar, &socketAndThreadReady, pingSocket, &relayManager, &relayAddr] {
      core::PingProcessor pingProcessor(*pingSocket, relayManager, gAlive, relayAddr);
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

  for (int i = 0; i < 60; ++i) {
    if (relay::relay_init(curl,
         backendHostname.c_str(),
         relay_token,
         relayAddrString.c_str(),
         keychain.RouterPublicKey.data(),
         keychain.RelayPrivateKey.data(),
         &routerInfo.InitalizeTimeInSeconds) == RELAY_OK) {
      printf("\n");
      relay_initialized = true;
      break;
    }

    std::cout << '.' << std::flush;

    std::this_thread::sleep_for(1s);
  }

  if (!relay_initialized) {
    Log("error: could not initialize relay\n\n");
    curl_easy_cleanup(curl);
    gAlive = false;
    joinThreads();
    closeSockets();
    relay::relay_term();
    return 1;
  }

  Log("Relay initialized\n\n");

  signal(SIGINT, interrupt_handler);  // ctrl c shuts down gracefully

  // g++ complains that updateLoop is ambiguous without the scope resolution op
  ::updateLoop(curl, backendHostname.c_str(), relay_token, relayAddrString.c_str(), relayManager, *logger);

  Log("Cleaning up\n");

  LogDebug("Closing sockets");
  closeSockets();

  LogDebug("Joining threads");
  joinThreads();

  LogDebug("Closing log file");
  if (output) {
    output->close();
  }

  LogDebug("Cleaning up curl");
  curl_easy_cleanup(curl);

  LogDebug("Terminating relay");
  relay::relay_term();

  LogDebug("Relay terminated. Address: ", relayAddr);

  return 0;
}