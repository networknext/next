#ifndef UTIL_LOGGER_HPP
#define UTIL_LOGGER_HPP

#include "console.hpp"

extern util::Console _console_;

// Log() is for things you always want logged, like setup or error handling, enabled for release builds
// unlike the debug version of Log(), the release does not record the file and line as to not expose that

// LogDebug() is for debug logging, be as verbose as you'd like, it's turned off for release builds
// so there's no performance impact

// LogTest() is for logging during tests. It becomes hard to read test output if it's cluttered by all the debug output
// but if you need to debug a test and want logging to not be a mess, this should be used

// LogError() is for when there's libc error and errno is set to something but you also want to include a custom message

#define __LOG(...) _console_.log(__VA_ARGS__)
#define __LOG_EXTRA(...)                                       \
  _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); \
  _console_.flush()
#define __LOG_ERRNO(...) \
  __LOG(__VA_ARGS__);    \
  perror("\tOS Msg")

#ifdef LOG_ALL
#define Log(...) __LOG_EXTRA(__VA_ARGS__)
#define LogDebug(...) __LOG_EXTRA(__VA_ARGS__)
#define LogTest(...) __LOG_EXTRA(__VA_ARGS__)
#define LogError(...) __LOG_ERRNO(__VA_ARGS__)
#elif not defined NDEBUG and not defined TEST_BUILD
// redirect log to debug logging,
// log debug exposes file and line to help figure out where the log came from
// and error redirects to log debug too
#define Log(...) __LOG_EXTRA(__VA_ARGS__)
#define LogDebug(...) __LOG_EXTRA(__VA_ARGS__)
#define LogTest(...)
#define LogError(...) __LOG_ERRNO(__VA_ARGS__)
#elif defined TEST_BUILD
// enable test logging and disable all else
#define Log(...)
#define LogDebug(...)
#define LogTest(...) __LOG_EXTRA(__VA_ARGS__)
#define LogError(...)
#elif defined BENCH_BUILD
// disable all logging to not clutter output
#define Log(...)
#define LogDebug(...)
#define LogTest(...)
#define LogError(...)
#else
// only disable debug and test logging,
// error logs redirect to regular Log()
#define Log(...) __LOG(__VA_ARGS__)
#define LogDebug(...)
#define LogTest(...)
#define LogError(...) __LOG_ERRNO(__VA_ARGS__)
#endif

#endif