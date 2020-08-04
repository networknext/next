#ifndef OS_PLATFORM_HPP
#define OS_PLATFORM_HPP
#if RELAY_PLATFORM == RELAY_PLATFORM_WINDOWS
#include "platform_windows.hpp"
#elif RELAY_PLATFORM == RELAY_PLATFORM_MAC
#include "platform_mac.hpp"
#elif RELAY_PLATFORM == RELAY_PLATFORM_LINUX
#include "linux/platform.hpp"
#endif
#endif