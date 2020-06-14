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

#if not defined NDEBUG and not defined TEST_BUILD
// redirect log to debug logging,
// log debug exposes file and line to help figure out where the log came from
// and error redirects to log debug too
#define Log(...) LogDebug(__VA_ARGS__)
#define LogDebug(...)                                          \
  _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); \
  _console_.flush()
#define LogTest(...)
#define LogError(...)    \
  LogDebug(__VA_ARGS__); \
  perror("\tOS Msg")
#elif defined TEST_BUILD
// enable test logging and disable all else
#define Log(...)
#define LogDebug(...)
#define LogTest(...)                                           \
  _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); \
  _console_.flush()
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
#define Log(...) _console_.log(__VA_ARGS__)
#define LogDebug(...)
#define LogTest(...)
#define LogError(...) \
  Log(__VA_ARGS__);   \
  perror("\tOS Msg")
#endif

#endif