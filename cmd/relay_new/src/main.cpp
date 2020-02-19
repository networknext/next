/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include "includes.h"

#include "util.hpp"

#include "encoding/base64.hpp"

#include "bench/bench.hpp"
#include "testing/test.hpp"

#include "relay/relay.hpp"
#include "relay/relay_platform.hpp"

#include "core/packet_processor.hpp"
#include "core/ping_processor.hpp"

namespace
{
  volatile bool gAlive = true;  // TODO make atomic

  void interrupt_handler(int signal)
  {
    (void)signal;
    gAlive = false;
  }

  inline void update_loop(relay::relay_t& relay,
   CURL* curl,
   const char* backend_hostname,
   const uint8_t* relay_token,
   const char* relay_address_string)
  {
    std::vector<uint8_t> update_response_memory;
    update_response_memory.resize(RESPONSE_MAX_BYTES);
    while (gAlive) {
      bool updated = false;

      for (int i = 0; i < 10; ++i) {
        if (relay_update(curl, backend_hostname, relay_token, relay_address_string, update_response_memory.data(), &relay) ==
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

      relay::relay_platform_sleep(1.0);
    }
  }
}  // namespace

int main(int argc, const char** argv)
{
  if (argc == 2 && strcmp(argv[1], "test") == 0) {
    return testing::SpecTest::Run() ? 0 : 1;
  }

  if (argc == 2 && strcmp(argv[1], "bench") == 0) {
    benchmarking::Benchmark::Run();
    return 0;
  }

  printf("\nNetwork Next Relay\n");

  printf("\nEnvironment:\n\n");

  const char* relay_address_env = relay::relay_platform_getenv("RELAY_ADDRESS");
  if (!relay_address_env) {
    printf("\nerror: RELAY_ADDRESS not set\n\n");
    return 1;
  }

  legacy::relay_address_t relay_address;
  if (relay_address_parse(&relay_address, relay_address_env) != RELAY_OK) {
    printf("\nerror: invalid relay address '%s'\n\n", relay_address_env);
    return 1;
  }

  {
    legacy::relay_address_t address_without_port = relay_address;
    address_without_port.port = 0;
    char address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
    printf("    relay address is '%s'\n", legacy::relay_address_to_string(&address_without_port, address_buffer));
  }

  uint16_t relay_bind_port = relay_address.port;

  printf("    relay bind port is %d\n", relay_bind_port);

  const char* relay_private_key_env = relay::relay_platform_getenv("RELAY_PRIVATE_KEY");
  if (!relay_private_key_env) {
    printf("\nerror: RELAY_PRIVATE_KEY not set\n\n");
    return 1;
  }

  uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
  if (encoding::base64_decode_data(relay_private_key_env, relay_private_key, RELAY_PRIVATE_KEY_BYTES) !=
      RELAY_PRIVATE_KEY_BYTES) {
    printf("\nerror: invalid relay private key\n\n");
    return 1;
  }

  printf("    relay private key is '%s'\n", relay_private_key_env);

  const char* relay_public_key_env = relay::relay_platform_getenv("RELAY_PUBLIC_KEY");
  if (!relay_public_key_env) {
    printf("\nerror: RELAY_PUBLIC_KEY not set\n\n");
    return 1;
  }

  uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
  if (encoding::base64_decode_data(relay_public_key_env, relay_public_key, RELAY_PUBLIC_KEY_BYTES) != RELAY_PUBLIC_KEY_BYTES) {
    printf("\nerror: invalid relay public key\n\n");
    return 1;
  }

  printf("    relay public key is '%s'\n", relay_public_key_env);

  const char* router_public_key_env = relay::relay_platform_getenv("RELAY_ROUTER_PUBLIC_KEY");
  if (!router_public_key_env) {
    printf("\nerror: RELAY_ROUTER_PUBLIC_KEY not set\n\n");
    return 1;
  }

  uint8_t router_public_key[crypto_sign_PUBLICKEYBYTES];
  if (encoding::base64_decode_data(router_public_key_env, router_public_key, crypto_sign_PUBLICKEYBYTES) !=
      crypto_sign_PUBLICKEYBYTES) {
    printf("\nerror: invalid router public key\n\n");
    return 1;
  }

  printf("    router public key is '%s'\n", router_public_key_env);

  const char* backend_hostname = relay::relay_platform_getenv("RELAY_BACKEND_HOSTNAME");
  if (!backend_hostname) {
    printf("\nerror: RELAY_BACKEND_HOSTNAME not set\n\n");
    return 1;
  }

  printf("    backend hostname is '%s'\n", backend_hostname);

  unsigned int numProcessors = 0;
  const char* nproc = relay::relay_platform_getenv("RELAY_PROCESSOR_COUNT");
  if (nproc == nullptr) {
    numProcessors = std::thread::hardware_concurrency();
    if (numProcessors > 0) {
      Log("RELAY_PROCESSOR_COUNT not set, autodetected number of processors available: ", numProcessors, "\n\n");
    } else {
      Log("error: RELAY_PROCESSOR_COUNT not set\n\n");
      return 1;
    }
  } else {
    numProcessors = std::atoi(nproc);
  }

  std::ofstream* output = nullptr;
  util::ThroughputLogger* logger = nullptr;

  const char* relayThroughputLog = relay::relay_platform_getenv("RELAY_LOG_FILE");
  if (relayThroughputLog != nullptr) {
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

  if (relay::relay_initialize() != RELAY_OK) {
    Log("error: failed to initialize relay\n\n");
    return 1;
  }

  Log("Relay socket opened on port ", relay_address.port);
  char relay_address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
  const char* relay_address_string = relay_address_to_string(&relay_address, relay_address_buffer);

  CURL* curl = curl_easy_init();
  if (!curl) {
    Log("error: could not initialize curl\n\n");
    curl_easy_cleanup(curl);
    relay::relay_term();
    return 1;
  }

  uint8_t relay_token[RELAY_TOKEN_BYTES];

  Log("Initializing relay\n");

  bool relay_initialized = false;

  uint64_t router_timestamp = 0;

  for (int i = 0; i < 60; ++i) {
    if (relay::relay_init(
         curl, backend_hostname, relay_token, relay_address_string, router_public_key, relay_private_key, &router_timestamp) ==
        RELAY_OK) {
      printf("\n");
      relay_initialized = true;
      break;
    }

    printf(".");
    fflush(stdout);

    relay::relay_platform_sleep(1.0);
  }

  if (!relay_initialized) {
    Log("error: could not initialize relay\n\n");
    curl_easy_cleanup(curl);
    relay::relay_term();
    return 1;
  }

  relay::relay_t relay(router_timestamp, relay_public_key, relay_private_key, router_public_key);

  relay.mutex = relay::relay_platform_mutex_create();
  if (!relay.mutex) {
    Log("error: could not create ping thread\n\n");
    gAlive = false;
  }

  relay.relay_manager = legacy::relay_manager_create();
  if (!relay.relay_manager) {
    Log("error: could not create relay manager\n\n");
    gAlive = false;
  }

  std::vector<os::SocketPtr> sockets;

  auto pingSocket = std::make_shared<os::Socket>(os::SocketType::Blocking);
  if (!pingSocket->create(relay_address, 100 * 1024, 100 * 1024, 0.0f, true)) {
    Log("could not create pingSocket");
    relay::relay_term();
    return 1;
  }

  sockets.push_back(pingSocket);

  std::unique_ptr<std::thread> pingThread = std::make_unique<std::thread>([pingSocket, &relay] {
    core::PingProcessor processor(relay, gAlive);
    processor.listen(*pingSocket);
  });

  std::vector<std::unique_ptr<std::thread>> packetThreads;
  packetThreads.resize(numProcessors);
  core::SessionMap sessions;
  for (unsigned int i = 0; i < numProcessors; i++) {
    auto packetSocket = std::make_shared<os::Socket>(os::SocketType::Blocking);
    if (!packetSocket->create(relay_address, 100 * 1024, 100 * 1024, 0.0f, true)) {
      Log("could not create socket");
      relay::relay_term();
      return 1;
    }

    sockets.push_back(packetSocket);

    packetThreads[i] = std::make_unique<std::thread>([packetSocket, &sessions, &relay, &logger] {
      core::PacketProcessor processor(sessions, relay, gAlive, logger);
      processor.listen(*packetSocket);
    });
  }

  Log("Relay initialized\n\n");

  signal(SIGINT, interrupt_handler);

  ::update_loop(relay, curl, backend_hostname, relay_token, relay_address_string);

  Log("Cleaning up\n");

  for (auto socket : sockets) {
    socket->close();
  }

  pingThread->join();

  for (unsigned int i = 0; i < numProcessors; i++) {
    packetThreads[i]->join();
  }

  if (logger != nullptr) {
    logger->stop();
    delete logger;
  }

  if (output != nullptr) {
    output->close();
    delete output;
  }

  relay_manager_destroy(relay.relay_manager);

  relay_platform_mutex_destroy(relay.mutex);

  curl_easy_cleanup(curl);

  relay::relay_term();

  return 0;
}
