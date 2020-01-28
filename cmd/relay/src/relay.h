/*
    Network Next Relay.
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#include <stdint.h>

#include "sysinfo.hpp"

#ifndef RELAY_H
#define RELAY_H

#if RELAY_PLATFORM == RELAY_PLATFORM_WINDOWS
#include "relay_windows.h"
#elif RELAY_PLATFORM == RELAY_PLATFORM_MAC
#include "relay_mac.h"
#elif RELAY_PLATFORM == RELAY_PLATFORM_LINUX
#include "relay_linux.h"
#endif

#include "relay/relay_address.hpp"

void relay_printf(const char* format, ...);
#endif  // #ifndef RELAY_H
