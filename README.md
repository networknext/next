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

### Installing Shared Libraries

```sh
$ sudo apt install golang libsodium
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

### Relay (C++)

This is the service that suppliers install on their hardware to become part of the Network Next relay network.

- Command: [`cmd/relay`](./cmd/relay)

## Relay Backend (Go)

Manages the database of connected relays and tells them which other relays to ping. Collates ping statistics received from relays into a cost matrix which is used by the optimizer to calculate the route matrix.

- Command: [`cmd/relay_backend`](./cmd/relay_backend)

## Server (C++)

Reference implentation of a server using the Network Next SDK.

- Command: [`cmd/server`](./cmd/server)
- Dependencies: [`sdk`](./sdk)

## Server Backend (Go)

Pulls the route matrix from the optimizer and uses this to serve up routes between clients to servers across the relay network.

- Command: [`cmd/server_backend`](./cmd/server_backend)

## Client (C++)

Reference implentation of a client using the Network Next SDK. 

- Command: [`cmd/client`](./cmd/client)
- Dependencies: [`sdk`](./sdk)

## SDK (C++)

This is the SDK we ship to customers.
