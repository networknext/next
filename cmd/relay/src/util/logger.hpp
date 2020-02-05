#ifndef UTIL_LOGGER_HPP
#define UTIL_LOGGER_HPP

#include "console.hpp"

extern util::Console _console_;

// always on

#define LogRelease(...) _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__)

// a macro to disable any debug logging done for development purposes

#ifndef NDEBUG
#define LogDebug(...) _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); std::cout << std::flush
#else
#define LogDebug(...)
#endif

// this behaves like LogDebug, except it will always be off when not running tests, safeguard against forgetting to remove something when benchmarking

#if defined TESTING and not defined BENCHMARKING
#define LogTest(...) _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__); std::cout << std::flush
#else
#define LogTest(...)
#endif

#endif