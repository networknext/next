# Setup

## Environment Variables

### Required
- `RELAY_ADDRESS`: The address the other relays and sdk should talk to
- `RELAY_PUBLIC_KEY`: The public key of the relay
- `RELAY_PRIVATE_KEY`: The corresponding private key
- `RELAY_ROUTER_PUBLIC_KEY`: The router's public key, used to encrypt data for relay verification

### Optional
- `RELAY_SOCKET_BUFFER_SIZE`: In bytes, lets you set the amount of memory to use for each socket's send & receive buffer
- `RELAY_PROCESSOR_COUNT`: Number of processors to allocate to the relay. Each relay thread is assigned affinity starting at core 0 to n. If unset the relay will attempt to auto detect the number of processors on the system

# Development
Haphazard hell

- Before requesting a pull request it's advised to spin up the relay backend, server backend, 9-10 relays, and the game server and client. Then ensure the relay works successfully, not just relying on the test build.

# Building

Several makefiles are available for building

## Release
- For production use
- Only enables the -O3 flag for optimization, the remaining performance flags are commented out, using them will produce a non-portable binary. Unless the performance gain is necessary, these should remain unused.
- Defines `NDEBUG` to disable assertions and other debug related things.

## Debug
- For debugging
- Compiles with no optimization and the -g flag for debug information.

## Test
- For executing tests only.
- Enables the LogTest() macro. Functions just like the others but is only enabled for test builds as to not clutter output.
- Disables all forms of optimization and enables the -g flag in gcc.

## Benchmarks
- For running benchmarks only
- Disables all log macros. Ensure the code is working before benchmarking.
- Enables many optimizations to ensure fair benchmarks.