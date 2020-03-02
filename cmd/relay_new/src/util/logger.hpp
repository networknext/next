#ifndef UTIL_LOGGER_HPP
#define UTIL_LOGGER_HPP

#include "console.hpp"

extern util::Console _console_;

// LogDebug() is for debug logging, be as verbose as you'd like, it's turned off for release builds
// so there's no performance impact

// Log() is for things you always want logged, like setup or error handling, enabled for release builds
// unlike the debug version of Log(), the release does not record the file and line as to not expose that

#ifndef NDEBUG
#define LogDebug(...)                                          \
  _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); \
  std::cout << std::flush
// Define regular logging
#define Log(...) LogDebug(__VA_ARGS__)
#elif defined BENCH_BUILD
// disable all logging to not clutter output
#define LogDebug(...)
#define Log(...)
#else
// only disable debug logging
#define LogDebug(...)
#define Log(...) _console_.log(__VA_ARGS__)
#endif

#define LogError(...) Log(__VA_ARGS__); perror("\tOS Msg")

#endif