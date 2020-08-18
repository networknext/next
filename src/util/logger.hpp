#ifndef UTIL_LOGGER_HPP
#define UTIL_LOGGER_HPP

#include "console.hpp"

extern util::Console _console_;
extern util::Console _error_console_;

// Log() is for things you always want logged, like setup or error handling, enabled for release builds
// unlike the debug version of Log(), the release does not record the file and line as to not expose that

// LogDebug() is for debug logging, be as verbose as you'd like, it's turned off for release builds
// so there's no performance impact

// LogTest() is for logging during tests. It becomes hard to read test output if it's cluttered by all the debug output
// but if you need to debug a test and want logging to not be a mess, this should be used

// LogError() is for when there's libc error and errno is set to something but you also want to include a custom message

#define __LOG(...) _console_.log(__VA_ARGS__)
#define __LOG_ERROR(...) _error_console_.log(__VA_ARGS__)
#define __LOG_DEBUG(...)                               \
  _log_(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); \
  _console_.flush()
#define __LOG_SYSERR(...)   \
  __LOG_ERROR(__VA_ARGS__); \
  perror("\tOS Msg")

#ifdef LOG_ALL
#define LOG(...) __LOG_EXTRA(__VA_ARGS__)
#define LOG_DEBUG(...) __LOG_EXTRA(__VA_ARGS__)
#define LOG_ERROR(...) __LOG_ERROR(__VA_ARGS__)
#define LOG_SYS(...) __LOG_SYSERR(__VA_ARGS__)
#define LOG_TEST(...) __LOG_EXTRA(__VA_ARGS__)
#elif not defined NDEBUG and not defined TEST_BUILD
// redirect log to debug logging,
// log debug exposes file and line to help figure out where the log came from
// and error redirects to log debug too
#define LOG(...) __LOG_DEBUG(__VA_ARGS__)
#define LOG_DEBUG(...) __LOG_DEBUG(__VA_ARGS__)
#define LOG_TEST(...)
#define LOG_SYS(...) __LOG_SYSERR(__VA_ARGS__)
#elif defined TEST_BUILD
// enable test & system error logging and disable all else
#define LOG(...)
#define LOG_ERROR(...)
#define LOG_SYS(...) __LOG_SYSERR(__VA_ARGS__)
#define LOG_DEBUG(...)
#define LOG_TEST(...) __LOG_EXTRA(__VA_ARGS__)
#elif defined BENCH_BUILD
// disable all logging to not clutter output
#define LOG(...)
#define LOG_DEBUG(...)
#define LOG_TEST(...)
#define LOG_SYS(...)
#else
// only disable debug and test logging,
// error logs redirect to regular Log()
#define LOG(...) __LOG(__VA_ARGS__)
#define LOG_DEBUG(...)
#define LOG_TEST(...)
#define LOG_SYS(...) __LOG_ERRNO(__VA_ARGS__)
#endif

#endif