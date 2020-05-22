# General Info

The Relay is responsible for:

1. Routing packets across other relays in the network to and from the game client and server
2. Sending stats updates to the relay_backend that are optimized for route generation

This can be thought of as version 2 of the relay. Version 1 being the relay in the `relay_old` directory. This is not currently deployed and is under active development.

# Pre-build setup

## Dependencies

`sudo apt install g++-8 rapidjson-dev libcurl4 libcurl4-openssl-dev libsodium23 libsodum-dev`

- `g++-8`: Specifically version 8. This is because newer versions of Ubuntu come with g++-9 as a default and compiling with that doesn't let the Ubuntu 18.04 servers run the relay.
- `RapidJSON`: Fast JSON parsing header only library.
- `cURL`: For HTTP communication.
- `libsodium`: Encryption/decryption/signature/etc... library.

## Environment Variables

### Required

- `RELAY_ADDRESS`: The address the other relays and sdk should talk to.
  - Example `RELAY_ADDRESS='127.0.0.1:1234'`
- `RELAY_PUBLIC_KEY`: The public key of the relay encoded in base64.
  - Example `RELAY_PUBLIC_KEY='9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14='`
- `RELAY_PRIVATE_KEY`: The corresponding private key encoded in base64.
  - Example `RELAY_PRIVATE_KEY='lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8='`
- `RELAY_ROUTER_PUBLIC_KEY`: The router's public key encoded in base64, used to encrypt data for relay verification.
  - Example `RELAY_ROUTER_PUBLIC_KEY='SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y='`
- `RELAY_BACKEND_HOSTNAME`: The backend's hostname
  - Example `RELAY_BACKEND_HOSTNAME='http://localhost:30000'`
- `RELAY_V3_NAME`: On servers this must be the firestore id, for local development this can be anything.
  - Example `RELAY_V3_NAME='fake-relay-name'`
- `RELAY_V3_BACKEND_HOSTNAME`: The old relay backend, no http prefix, just the hostname.
  - Example `RELAY_V3_BACKEND_HOSTNAME='localhost'`
- `RELAY_V3_BACKEND_PORT`: The port the old relay backend is running on. This must be separate from the hostname.
  - Example `RELAY_V3_BACKEND_PORT='40000'`
- `RELAY_V3_UPDATE_KEY`: The key in firestore that allows a relay to update with the old backend. For local dev this can be any base64 decryptable string
  - Example `RELAY_V3_UPDATE_KEY='fake-update-key'`
- `RELAY_V3_SPEED`: The speed in mega bits per second of the relay's NIC. Needed to calculate utilization. Must be a number (not 10Gb or anything like that).
  - Example `RELAY_V3_SPEED='10000000000'`



### Optional
- `RELAY_V3_ENABLED`: Controls whether or not the old backend communication is enabled or not. In the makefile it is defaulted to "0". Set to "1", and only that, to enable.
  - Example `RELAY_V3_ENABLED="1"`
- `RELAY_PROCESSOR_COUNT`: Number of processors to allocate to the relay. Each relay thread is assigned affinity starting at core 0 to n. If unset the relay will attempt to auto detect the number of processors on the system.
  - Example `RELAY_PROCESSOR_COUNT='1'` or `RELAY_PROCESSOR_COUNT="$(( $(nproc) / 4 ))"`
- `RELAY_SEND_BUFFER_SIZE` & `RELAY_RECV_BUFFER_SIZE`: In bytes, lets you set the amount of memory to use for each socket's send & receive buffers.
  - Example `RELAY_SOCKET_BUFFER_SIZE="4000000"`
  - Note, Macs apparently have issues with values above a million, thus "1000000" is the default value if not set

# Building

Several makefiles are available for building. If you are not developing the relay you probably want the release build.

## Release

`make release`

- Only enables the Log() macro.
- Only enables the -O3 flag for optimization, the remaining performance flags are commented out, using them will produce a non-portable binary. Unless the performance gain is necessary, these should remain unused.
- Defines `NDEBUG` to disable assertions and other debug related things.
- Use `make run-release` to run the exe. Alternatively execute directly with `bin/relay` or from the repo root dir `dist/relay`

## Debug

`make debug`

- Enables the LogDebug() macro and redirects all Log() macros to LogDebug().
- Compiles with no optimization and the -g flag for debug information.
- Disables all forms of optimization and enables the -g flag in gcc.
- Use `make run-debug` to run the exe. Alternatively execute directly with `bin/relay.debug`

## Test

`make test`

- For executing tests only.
- Enables the LogTest() macro. Functions just like the others but is only enabled for test builds as to not clutter output. All others are disabled.
- Disables all forms of optimization and enables the -g flag in gcc.
- Use `make run-tests` to run them. Alternatively execute directly with `bin/relay.test`

## Benchmarks

`make benchmark`

- For running benchmarks only
- Disables all log macros. Ensure the code is working before benchmarking.
- Enables many optimizations to ensure fair benchmarks.
- Use `make run-benchmarks` to run them. Alternatively execute directly with `bin/relay.benchmark`

## Shutdown procedures

The relay can be shutdown in a few ways depending on the signals you give it.

### Graceful shutdown

Signals: SIGINT, SIGTERM

This is just a standard graceful shutdown. The relay will join threads, clean up memory, and log it's shutting down.

### Clean shutdown

Signal: SIGHUP

This shutdown performs like the graceful shutdown, but additionally it will do one of two things:

  1. Tell the backend it is shutting down, and then upon getting a response wait an additional 30 seconds.
  2. Wait 60 seconds and then proceed with shutting down if there is no communication with the backend.

This is so the relay can keep serving game clients connected to it until it is removed from any routes.

### Regular process termination

All other signals will result in killing the relay process in the standard way.
