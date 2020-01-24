<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This is a monorepo that contains a WIP migration/refactor of the Network Next backend.

## Development

The toolchain used for development is kept simple to make it easy for any operating system to install and use and work out of the box for POSIX Linux distributions.

- [Docker](https://www.docker.com)
- [make](http://man7.org/linux/man-pages/man1/make.1.html)
- [sh](https://linux.die.net/man/1/sh)
- [Go](https://golang.org/dl/#stable)
- [g++](http://man7.org/linux/man-pages/man1/g++.1.html)
    - [libcurl](https://curl.haxx.se/libcurl/)
    - [libsodium](https://libsodium.gitbook.io)
    - [libpthread](https://www.gnu.org/software/hurd/libpthread.html)

### Linux

```sh
$ sudo apt install golang libsodium libcurl4-openssl-dev
```

### macOS

Install [brew](https://brew.sh)

Then:

```sh
brew install golang libsodium
```

### Windows

Using the Windows Subsystem for Linux (WSL) with Ubuntu makes it easy to work with this repo provided all the tools are installed above.

## Components

## Relay (C++)

This is the service that suppliers run on their hardware to become part of Network Next.

- Command: [`cmd/relay`](./cmd/relay)

### To Run
You must have the following environment variables set:
- `RELAY_ADDRESS`: The address of this relay. Example being "127.0.0.1:1234"
- `RELAY_PRIVATE_KEY`: The relay's private key. To generate see [generating keys](#generating-keys)
- `RELAY_PUBLIC_KEY`:
- `RELAY_ROUTER_PUBLIC_KEY`:
- `RELAY_BACKEND_HOSTNAME`:

## Relay Backend (Go)

Manages the database (Redis) of connected relays and tells them which other relays to ping. Collates ping statistics received from relays into a cost matrix.

- Command: [`cmd/relay_backend`](./cmd/relay_backend)

### To Run
You must have the following environment variables set:
- `RELAY_KEY_PUBLIC`
- `ROUTER_KEY_PRIVATE`

### What it does
- Opens a couple endpoints for communication with the relays and server backend
- Takes data from the in memory StatsDatabase and generates a CostMatrix, then takes the CostMatrix and generates a RouteMatrix for the server backend once every 10 seconds
### Endpoints
- `/relay_init`: When a relay starts up, the first thing it will do is call this endpoint with its information
  - Gathers geolocation information
  - Generates the relay public keys, which is sent back in the body of the response
  - Stores the relay in redis in binary format
- `/relay_update`: After a relay has confirmation of successful initialization, it will keep sending this endpoint stats about its network timings
  - Within the response body will be a list of all other relays to ping and gather stats on


## Server (C++)

Reference implentation of a server using the Network Next SDK.

- Command: [`cmd/server`](./cmd/server)
- Dependencies: [`sdk`](./sdk)

## Server Backend (Go)

Pulls the route matrix from the relay backend and uses it to serve up routes across the relay network.

- Command: [`cmd/server_backend`](./cmd/server_backend)

### Environment Variables

- `MAXMIND_DB_URI`: local path to a `.mmdb` file for IP lookups. Defaults to `./GeoLite2-City.mmdb`. You can ask a dev for a copy of one or register and download one from https://www.maxmind.com.

## Client (C++)

Reference implentation of a client using the Network Next SDK.

- Command: [`cmd/client`](./cmd/client)
- Dependencies: [`sdk`](./sdk)

## SDK (C++)

This is the SDK we ship to customers.

## Tools

## relay-spawner.sh

Uses the env var `RUNNING_RELAYS` to keep track of spawned relays
- Because of that, the script must be sourced to work properly
- It will exit if you try to run the script directly

Usage: `source relay-spawner.sh` [`options`] [`starting port number`] [`ending port number`]
- Spawns relays with the port numbers between the given arguments, inclusively
- If the ending port number is not specified, one relay with the starting port number will be spawned
- `-k` is used to kill all relays spawned with this script, any spawned without the script will be left alive

## Generating Key Pairs

Keys must be generated with Go's `box.GenerateKey()` function call. The parameter to pass is `rand.Reader`.