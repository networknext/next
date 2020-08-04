/*
    Network Next Relay.
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#ifndef RELAY_RELAY_PLATFORM_MAC_HPP
#define RELAY_RELAY_PLATFORM_MAC_HPP

#if RELAY_PLATFORM == RELAY_PLATFORM_MAC

#include "relay/relay_address.hpp"

#define RELAY_PLATFORM_HAS_IPV6 1
#define RELAY_PLATFORM_SOCKET_NON_BLOCKING 0
#define RELAY_PLATFORM_SOCKET_BLOCKING 1

namespace relay
{
  typedef int relay_platform_socket_handle_t;

  struct relay_platform_socket_t
  {
    int type;
    relay_platform_socket_handle_t handle;
  };

  struct relay_platform_thread_t
  {
    pthread_t handle;
  };

  typedef void* relay_platform_thread_return_t;

#define RELAY_PLATFORM_THREAD_RETURN() \
  do {                                 \
    return NULL;                       \
  } while (0)

#define RELAY_PLATFORM_THREAD_FUNC

  typedef relay_platform_thread_return_t(RELAY_PLATFORM_THREAD_FUNC relay_platform_thread_func_t)(void*);

  // -------------------------------------

  struct relay_platform_mutex_t
  {
    pthread_mutex_t handle;
  };

  int relay_platform_init();

  void relay_platform_term();

  const char* relay_platform_getenv(const char*);

  int relay_platform_inet_pton4(const char* address_string, uint32_t* address_out);

  int relay_platform_inet_pton6(const char* address_string, uint16_t* address_out);

  int relay_platform_inet_ntop6(const uint16_t* address, char* address_string, size_t address_string_size);

  uint16_t relay_platform_ntohs(uint16_t in);

  uint16_t relay_platform_htons(uint16_t in);

  double relay_platform_time();

  void relay_platform_sleep(double time);

  relay_platform_socket_t* relay_platform_socket_create(
   struct relay_address_t* address, int socket_type, float timeout_seconds, int send_buffer_size, int receive_buffer_size);

  void relay_platform_socket_destroy(relay_platform_socket_t* socket);

  void relay_platform_socket_send_packet(
   relay_platform_socket_t* socket, const relay_address_t* to, const void* packet_data, int packet_bytes);

  int relay_platform_socket_receive_packet(
   relay_platform_socket_t* socket, relay_address_t* from, void* packet_data, int max_packet_size);

  relay_platform_thread_t* relay_platform_thread_create(relay_platform_thread_func_t* func, void* arg);

  void relay_platform_thread_join(relay_platform_thread_t* thread);

  void relay_platform_thread_destroy(relay_platform_thread_t* thread);

  void relay_platform_thread_set_sched_max(relay_platform_thread_t* thread);

  relay_platform_mutex_t* relay_platform_mutex_create();

  void relay_platform_mutex_acquire(relay_platform_mutex_t* mutex);

  void relay_platform_mutex_release(relay_platform_mutex_t* mutex);

  void relay_platform_mutex_destroy(relay_platform_mutex_t* mutex);

  struct relay_mutex_helper_t
  {
    relay_mutex_helper_t(relay_platform_mutex_t* mutex);
    ~relay_mutex_helper_t();
    relay_platform_mutex_t* mutex;
  };

#define relay_mutex_guard(_mutex) relay_mutex_helper_t __mutex_helper(_mutex)

  // -------------------------------------
}  // namespace relay
#endif  // #if RELAY_PLATFORM == RELAY_PLATFORM_MAC
#endif  // #ifndef RELAY_MAC_H
