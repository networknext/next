# General Info

The relay is responsible for forwarding packets between the game client and servers and reporting various statistics to the relay backend.

# Setup

## Dependencies

`sudo apt install rapidjson-dev libcurl4-openssl-dev`

- `RapidJSON`: Fast JSON parsing header only library.
- `cURL`: For HTTP communication.

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

### Optional

- `RELAY_SOCKET_BUFFER_SIZE`: In bytes, lets you set the amount of memory to use for each socket's send & receive buffer.
  - Example `RELAY_SOCKET_BUFFER_SIZE="4000000"`
  - Note, Macs apparently have issues with values above a million
- `RELAY_PROCESSOR_COUNT`: Number of processors to allocate to the relay. Each relay thread is assigned affinity starting at core 0 to n. If unset the relay will attempt to auto detect the number of processors on the system.
  - Example `RELAY_PROCESSOR_COUNT='1'` or `RELAY_PROCESSOR_COUNT="$(( $(nproc) / 4 ))"`

# Building

Several makefiles are available for building. If you are not developing the relay you probably want the release build.

## Release

`make release`

- Only enables the Log() macro.
- Only enables the -O3 flag for optimization, the remaining performance flags are commented out, using them will produce a non-portable binary. Unless the performance gain is necessary, these should remain unused.
- Defines `NDEBUG` to disable assertions and other debug related things.
- Use `make run-release` to run the exe. Alternatively execute directly with `bin/relay`

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
