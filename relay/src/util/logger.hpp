#pragma once

#include "console.hpp"

extern util::Console _console_;

// #define LOG_TRACE(...) 	_console_.log("trace ", __FILE__, " (", __LINE__, "): ", __VA_ARGS__)
#define LOG_TRACE(...)

// #define LOG_DEBUG(...) 	_console_.log("", __VA_ARGS__)
#define LOG_DEBUG(...)

#define LOG_INFO(...)   _console_.log("", __VA_ARGS__); _console_.flush();

#define LOG_WARN(...)   _console_.log("warning: ", __VA_ARGS__); _console_.flush();

#define LOG_ERROR(...) 	_console_.log("error: ", __VA_ARGS__); _console_.flush();

#define LOG_FATAL(...) 	_console_.log("error: ", __VA_ARGS__); _console_.flush(); std::exit(1)

#define LOG(level, ...) LOG_##level(__VA_ARGS__)
