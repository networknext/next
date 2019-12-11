<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This is a monorepo that contains a WIP migration/refactor of the Network Next services.

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

### Installing Shared Libraries

```sh
$ sudo apt install libsodium-dev libcurl4-openssl-dev
```

### macOS

TBD any notes needed to require things to run on macOS.

### Windows

Using the Windows Subsystem for Linux (WSL) with Ubuntu makes it easy to work with this repo provided all the tools are installed above.

## Components

### Relay (C++)

This is the service that suppliers install on their hardware to become part of the Network Next relay network.

- Command: [`cmd/relay`](./cmd/relay)

## Relay Backend (Go)

This ingests information from all online Relays to feed into the Optimizer to calculate the Relay matrix of optimal routes.

- Command: [`cmd/relay_backend`](./cmd/relay_backend)

## Server (aka Game Server, C++)

This is a reference implentation of a **game server** using the Network Next SDK. This game server gets optimal route information from the Server Backend.

- Command: [`cmd/server`](./cmd/server)
- Dependancies: [`sdk`](./sdk)

## Server Backend (Go)

This pulls information from the Optimizer and distrubutes the optimal routes to the proper game servers.

- Command: [`cmd/server_backend`](./cmd/server_backend)

## Client (aka Game Client, C++)

This is a reference implentation of a **game client** using the Network Next SDK. This game client talks to the game server and is instructed which relays it should start sending traffic through.

- Command: [`cmd/client`](./cmd/client)
- Dependancies: [`sdk`](./sdk)

## SDK (C++)

This is the C++ SDK which is used by game developers when building their game clients and servers that handles talking to the Relay network.
