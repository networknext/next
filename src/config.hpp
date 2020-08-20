#ifndef CONFIG_HPP
#define CONFIG_HPP

#define RELAY_VERSION "1.1.2"

#define RELAY_MTU 1300

#define RELAY_ADDRESS_BUFFER_SAFETY 32

#define RELAY_REPLAY_PROTECTION_BUFFER_SIZE 256UL

#define RELAY_BANDWIDTH_LIMITER_INTERVAL 1.0

#define RELAY_PING_HISTORY_ENTRY_COUNT 256

#define RELAY_PING_TIME 0.1

#define RELAY_STATS_WINDOW 10.0

// how many seconds before a packet is considered as lost
#define RELAY_PING_SAFETY 1.0

#define RELAY_MAX_PACKET_BYTES 1500

#define RELAY_PUBLIC_KEY_BYTES 32UL
#define RELAY_PRIVATE_KEY_BYTES 32UL

#define RELAY_MAX_ADDRESS_STRING_LENGTH 256

#define MAX_RELAYS 1024

#define RELAY_TOKEN_BYTES 32
#define RESPONSE_MAX_BYTES 1024 * 1024

#define INVALID_SEQUENCE_NUMBER 0xFFFFFFFFFFFFFFFFULL

/* This will prevent GCC from optimizing out useless function calls, for benchmarking */
#ifdef BENCH_BUILD
#define GCC_NO_OPT_OUT asm("")
#else
#define GCC_NO_OPT_OUT
#endif

// #define RELAY_MULTISEND

#endif