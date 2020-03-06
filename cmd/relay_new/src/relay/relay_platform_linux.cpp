/*
    Network Next Relay.
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/
#include "includes.h"
#include "relay_platform_linux.hpp"

#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#include "util/logger.hpp"

#include "net/net.hpp"

namespace
{
  double time_start;
}

namespace relay
{
  // TODO doesn't this depend on processor arch and not platform (Linux, Mac, Windows)?
  // even if not it's hardcoded so the bytes are always swapped regardless of if they're
  // already in the proper endianess

  uint16_t relay_platform_ntohs(uint16_t in)
  {
    return (uint16_t)(((in << 8) & 0xFF00) | ((in >> 8) & 0x00FF));
  }

  uint16_t relay_platform_htons(uint16_t in)
  {
    return (uint16_t)(((in << 8) & 0xFF00) | ((in >> 8) & 0x00FF));
  }

  int relay_platform_init()
  {
    timespec ts;
    clock_gettime(CLOCK_MONOTONIC_RAW, &ts);
    time_start = ts.tv_sec + ((double)(ts.tv_nsec)) / 1000000000.0;
    return RELAY_OK;
  }

  void relay_platform_term()
  {
    // ...
  }

  const char* relay_platform_getenv(const char* var)
  {
    return getenv(var);
  }

  // ---------------------------------------------------

  int relay_platform_inet_pton4(const char* address_string, uint32_t* address_out)
  {
    sockaddr_in sockaddr4;
    bool success = inet_pton(AF_INET, address_string, &sockaddr4.sin_addr) == 1;
    *address_out = sockaddr4.sin_addr.s_addr;
    return success ? RELAY_OK : RELAY_ERROR;
  }

  int relay_platform_inet_pton6(const char* address_string, uint16_t* address_out)
  {
    return inet_pton(AF_INET6, address_string, address_out) == 1 ? RELAY_OK : RELAY_ERROR;
  }

  int relay_platform_inet_ntop6(const uint16_t* address, char* address_string, size_t address_string_size)
  {
    return inet_ntop(AF_INET6, (void*)address, address_string, socklen_t(address_string_size)) == NULL ? RELAY_ERROR : RELAY_OK;
  }

  // time in seconds
  double relay_platform_time()
  {
    timespec ts;
    clock_gettime(CLOCK_MONOTONIC_RAW, &ts);
    double current = ts.tv_sec + ((double)(ts.tv_nsec)) / 1000000000.0;
    return current - time_start;
  }

  void relay_platform_sleep(double time)
  {
    usleep((int)(time * 1000000));
  }

  // ---------------------------------------------------

  relay_platform_thread_t* relay_platform_thread_create(relay_platform_thread_func_t* thread_function, void* arg)
  {
    relay_platform_thread_t* thread = (relay_platform_thread_t*)malloc(sizeof(relay_platform_thread_t));

    assert(thread);

    if (pthread_create(&thread->handle, NULL, thread_function, arg) != 0) {
      free(thread);
      return NULL;
    }

    return thread;
  }

  void relay_platform_thread_join(relay_platform_thread_t* thread)
  {
    assert(thread);
    pthread_join(thread->handle, NULL);
  }

  void relay_platform_thread_destroy(relay_platform_thread_t* thread)
  {
    assert(thread);
    free(thread);
  }

  void relay_platform_thread_set_sched_max(relay_platform_thread_t* thread)
  {
    struct sched_param param;
    param.sched_priority = sched_get_priority_max(SCHED_FIFO);
    int ret = pthread_setschedparam(thread->handle, SCHED_FIFO, &param);
    if (ret) {
      LogError("unable to increase server thread priority");
    }
  }

  // ---------------------------------------------------

  relay_platform_mutex_t* relay_platform_mutex_create()
  {
    relay_platform_mutex_t* mutex = (relay_platform_mutex_t*)malloc(sizeof(relay_platform_mutex_t));
    assert(mutex);

    assert(mutex);

    pthread_mutexattr_t attr;
    pthread_mutexattr_init(&attr);
    pthread_mutexattr_settype(&attr, 0);
    int result = pthread_mutex_init(&mutex->handle, &attr);
    pthread_mutexattr_destroy(&attr);

    if (result != 0) {
      free(mutex);
      return NULL;
    }

    return mutex;
  }

  relay_mutex_helper_t::relay_mutex_helper_t(relay_platform_mutex_t* mutex): mutex(mutex)
  {
    assert(mutex);
    relay_platform_mutex_acquire(mutex);
  }

  relay_mutex_helper_t::~relay_mutex_helper_t()
  {
    assert(mutex);
    relay_platform_mutex_release(mutex);
    mutex = NULL;
  }

  void relay_platform_mutex_acquire(relay_platform_mutex_t* mutex)
  {
    assert(mutex);
    pthread_mutex_lock(&mutex->handle);
  }

  void relay_platform_mutex_release(relay_platform_mutex_t* mutex)
  {
    assert(mutex);
    pthread_mutex_unlock(&mutex->handle);
  }

  void relay_platform_mutex_destroy(relay_platform_mutex_t* mutex)
  {
    assert(mutex);
    pthread_mutex_destroy(&mutex->handle);
    free(mutex);
  }
}  // namespace relay

#else  // #if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

int relay_linux_dummy_symbol = 0;

#endif  // #if RELAY_PLATFORM == RELAY_PLATFORM_LINUX
