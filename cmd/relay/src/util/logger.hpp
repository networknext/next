#ifndef UTIL_LOGGER_HPP
#define UTIL_LOGGER_HPP

#include "console.hpp"

extern util::Console _console_;

#ifndef NDEBUG
#define LogDebug(...) _console_.log(__FILE__, " (", __LINE__, "): ", __VA_ARGS__)
#else
#define LogDebug(...)
#endif

#endif