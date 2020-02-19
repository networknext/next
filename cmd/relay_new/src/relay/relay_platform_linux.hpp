/*
    Network Next Relay.
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#ifndef RELAY_RELAY_PLATFORM_LINUX_HPP
#define RELAY_RELAY_PLATFORM_LINUX_HPP

#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#define RELAY_PLATFORM_HAS_IPV6 1
#define RELAY_PLATFORM_SOCKET_NON_BLOCKING 0
#define RELAY_PLATFORM_SOCKET_BLOCKING 1

namespace relay
{
  // -------------------------------------
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
}  // namespace relay

#endif  // #if RELAY_PLATFORM == RELAY_PLATFORM_LINUX
#endif  // #ifndef RELAY_LINUX_H
