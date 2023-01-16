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
#include "net/http.hpp"
#include "os/socket.hpp"
#include "testing/test.hpp"
#include "util/env.hpp"

#include "version.h"

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

static const unsigned char base64_table_encode[65] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

static const int base64_table_decode[256] =
{
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,   0,  0,  0,  0,  0, 62, 63, 62, 62, 63, 52, 53, 54, 55,
    56, 57, 58, 59, 60, 61,  0,  0,  0,  0,  0,  0,  0,  0,  1,  2,  3,  4,  5,  6,
    7,  8,  9,  10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,  0,
    0,  0,  0,  63,  0, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
};

int relay_base64_decode( const char * input, uint8_t * output, size_t output_size )
{
    assert( input );
    assert( output );
    assert( output_size > 0 );

    size_t input_length = strlen( input );
    int pad = input_length > 0 && ( input_length % 4 || input[input_length - 1] == '=' );
    size_t L = ( ( input_length + 3 ) / 4 - pad ) * 4;
    size_t output_length = L / 4 * 3 + pad;

    if ( output_length > output_size )
    {
        printf( "not enough room\n" );
        return 0;
    }

    for ( size_t i = 0, j = 0; i < L; i += 4 )
    {
        int n = base64_table_decode[int( input[i] )] << 18 | base64_table_decode[int( input[i + 1] )] << 12 | base64_table_decode[int( input[i + 2] )] << 6 | base64_table_decode[int( input[i + 3] )];
        output[j++] = uint8_t( n >> 16 );
        output[j++] = uint8_t( n >> 8 & 0xFF );
        output[j++] = uint8_t( n & 0xFF );
    }

    if ( pad )
    {
        int n = base64_table_decode[int( input[L] )] << 18 | base64_table_decode[int( input[L + 1] )] << 12;
        output[output_length - 1] = uint8_t( n >> 16 );

        if (input_length > L + 2 && input[L + 2] != '=')
        {
            n |= base64_table_decode[int( input[L + 2] )] << 6;
            output_length += 1;
            if ( output_length > output_size )
            {
                return 0;
            }
            output[output_length - 1] = uint8_t( n >> 8 & 0xFF );
        }
    }

    return int( output_length );
}

INLINE void get_crypto_keys(const Env& env, Keychain& keychain)
{
  // relay private key
  {
    LOG(INFO, "relay private key is '", env.relay_private_key, '\'');
    int len = relay_base64_decode( env.relay_private_key.c_str(), &keychain.relay_private_key[0], crypto::RELAY_PRIVATE_KEY_SIZE);
    printf( "len = %d\n", len );
    if (len != KEY_SIZE) {
      LOG(FATAL, "invalid relay private key");
    }
  }

  // relay public key
  {
    LOG(INFO, "relay public key is '", env.relay_public_key, '\'');
    int len = relay_base64_decode( env.relay_public_key.c_str(), &keychain.relay_public_key[0], crypto::RELAY_PUBLIC_KEY_SIZE);
    if (len != KEY_SIZE) {
      LOG(FATAL, "invalid relay public key");
    }
  }

  // router public key
  {
    LOG(INFO, "router public key is '", env.router_public_key, '\'');
    int len = relay_base64_decode( env.router_public_key.c_str(), &keychain.backend_public_key[0], crypto_sign_PUBLICKEYBYTES);
    if (len != KEY_SIZE) {
      LOG(FATAL, "invalid router public key");
    }
  }
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
    printf("%s\n", RELAY_VERSION);
    fflush(stdout);
    exit(0);
  }

  LOG(INFO, "Network Next Relay");

  LOG(INFO, "relay version is ", RELAY_VERSION);

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
