#ifndef RELAY_RELAY_PLATFORM
#define RELAY_RELAY_PLATFORM
#if RELAY_PLATFORM == RELAY_PLATFORM_WINDOWS
#include "relay_platform_windows.hpp"
#elif RELAY_PLATFORM == RELAY_PLATFORM_MAC
#include "relay_platform_mac.hpp"
#elif RELAY_PLATFORM == RELAY_PLATFORM_LINUX
#include "relay_platform_linux.hpp"
#endif
#endif
