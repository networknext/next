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
   core::RelayManager& relayManager)
  {
    std::vector<uint8_t> update_response_memory;
    update_response_memory.resize(RESPONSE_MAX_BYTES);
    while (gAlive) {
      bool updated = false;

      for (int i = 0; i < 10; ++i) {
        if (relay::relay_update(
             curl, backend_hostname, relay_token, relay_address_string, update_response_memory.data(), relayManager) ==
            RELAY_OK) {
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

  // external relay address - exposed to the internet
  // sent to the relay backend and is the addr everything communicates with
  net::Address externalAddr;
  {
    auto env = std::getenv("RELAY_PUBLIC_ADDRESS");
    if (env == nullptr) {
      Log("error: RELAY_PUBLIC_ADDRESS not set\n");
      return 1;
    }

    if (!externalAddr.parse(env)) {
      Log("error: invalid relay address '", env, "'\n");
      return 1;
    }

    std::cout << "    external address is '" << externalAddr << "'\n";
  }

  // internal relay address - from your router
  // used to bind and allow the external address to work
  net::Address internalAddr;
  {
    auto env = std::getenv("RELAY_BIND_ADDRESS");
    if (env == nullptr) {
      Log("error: RELAY_BIND_ADDRESS not set, set it to your internal ip\n");
      return 1;
    }

    if (!internalAddr.parse(env)) {
      Log("error: invalid relay bind address '", env, "'\n");
      return 1;
    }

    std::cout << "    internal address is '" << externalAddr << "'\n";
  }

  if (externalAddr.Port != internalAddr.Port) {
    Log("warning: bind & public ports are not the same, this may cause issues on some providers (ex. google cloud)");
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

  int socketBufferSize = 1000000;
  {
    auto env = std::getenv("RELAY_SOCKET_BUFFER_SIZE");
    if (env != nullptr) {
      int num = -1;
      try {
        num = std::stoi(std::string(env));  // to cause an exception to be thrown if not a number
      } catch (std::exception& e) {
        Log("Could not parse RELAY_SOCKET_BUFFER_SIZE env var to a number: ", e.what());
        std::exit(1);
      }

      if (num < 0) {
        Log("RELAY_SOCKET_BUFFER_SIZE is less than 0");
        std::exit(1);
      }

      socketBufferSize = num;
    }
  }

  LogDebug("Socket buffer size is ", socketBufferSize, " bytes");

  std::ofstream* output = nullptr;
  util::ThroughputLogger* logger = nullptr;
  {
    std::string relayThroughputLog = std::getenv("RELAY_LOG_FILE");
    if (!relayThroughputLog.empty()) {
      auto file = new std::ofstream;
      file->open(relayThroughputLog);

      if (*file) {
        output = file;
      } else {
        delete file;
      }
    }

    if (output != nullptr) {
      logger = new util::ThroughputLogger(*output);
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
  auto makeSocket = [&sockets, &socketBufferSize](net::Address& addr) -> os::SocketPtr {
    auto socket = std::make_shared<os::Socket>(os::SocketType::Blocking);
    if (!socket->create(addr, socketBufferSize, socketBufferSize, 0.0f, true, 0)) {
      return nullptr;
    }

    sockets.push_back(socket);

    return socket;
  };

  /* packet processing setup
   * must come before ping setup
   * otherwise ping may take the port that is reserved for packet processing, usually 40000
   * odds are slim but it may happen
   */
  {
    packetThreads.resize(numProcessors);

    for (unsigned int i = 0; i < numProcessors; i++) {
      auto packetSocket = makeSocket(internalAddr);
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
          *packetSocket, relayClock, keychain, routerInfo, sessions, relayManager, gAlive, logger);
         processor.process(waitVar, socketAndThreadReady);
       });

      wait();  // wait the the processor is ready to receive
    }
  }

  // if using port 0, it is discovered in ping socket's create(). That being said sockets
  // must be created before communicating with the backend otherwise port 0 will be reused
  LogDebug("Actual address: ", internalAddr);

  if (externalAddr.Port == 0) {
    Log("external port is 0, setting to same as internal (", internalAddr.Port, ')');
    externalAddr.Port = internalAddr.Port;
  }

  externalAddr.toString(relayAddrString);

  /* ping processing setup
   * pings are sent out on a different port number than received
   * if they are the same the relay behaves weird, it'll sometimes behave right
   * othertimes it'll just ignore everything coming to it
   */
  {
    net::Address bindAddr = internalAddr;
    {
      bindAddr.Port = 0;  // make sure the port is dynamically assigned
    }

    auto pingSocket = makeSocket(bindAddr);
    if (!pingSocket) {
      Log("could not create pingSocket");
      relay::relay_term();
      closeSockets();
      joinThreads();
      return 1;
    }

    sockets.push_back(pingSocket);

    // setup the ping processor to use the external address
    // relays use it to know where the receving port of other relays are
    pingThread = std::make_unique<std::thread>([&waitVar, &socketAndThreadReady, pingSocket, &relayManager, &externalAddr] {
      core::PingProcessor pingProcessor(*pingSocket, relayManager, gAlive, externalAddr);
      pingProcessor.process(waitVar, socketAndThreadReady);
    });

    wait();
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
  ::updateLoop(curl, backendHostname.c_str(), relay_token, relayAddrString.c_str(), relayManager);

  Log("Cleaning up\n");

  LogDebug("Closing sockets");
  closeSockets();

  LogDebug("Joining threads");
  joinThreads();

  LogDebug("Stopping throughput logger");
  if (logger != nullptr) {
    logger->stop();
    delete logger;
  }

  LogDebug("Closing log file");
  if (output != nullptr) {
    output->close();
    delete output;
  }

  LogDebug("Cleaning up curl");
  curl_easy_cleanup(curl);

  LogDebug("Terminating relay");
  relay::relay_term();

  LogDebug("Relay terminated. Address: ", internalAddr);

  return 0;
}