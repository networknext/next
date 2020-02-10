<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This is a monorepo that contains a WIP migration/refactor of the Network Next backend.

## Development

The tool chain used for development is kept simple to make it easy for any operating system to install and use and work out of the box for POSIX Linux distributions.

- [Redis](https://redis.io)
- [Docker](https://docs.docker.com/install/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [make](http://man7.org/linux/man-pages/man1/make.1.html)
- [sh](https://linux.die.net/man/1/sh)
- [Go](https://golang.org/dl/#stable) (at least Go 1.13)
- [g++](http://man7.org/linux/man-pages/man1/g++.1.html)
    - [libcurl](https://curl.haxx.se/libcurl/)
    - [libsodium](https://libsodium.gitbook.io)
    - [libpthread](https://www.gnu.org/software/hurd/libpthread.html)

Developers should install these requirements however they need to be installed based on your operating system. Windows users can leverage WSL to get all of these.

## Docker and Docker Compose

While all of the components can be run locally either independently or collectively it can be tedious to run multiple relays to get a true test of everything. We can leverage Docker and Docker Compose to easily stand everything up as a system. There is a [`./cmd/docker-compose.yaml`](./cmd/docker-compose.yaml) along with all required `Dockerfile`s in each of the binary directories to create the system of backend services (`relay`, `relay_backend`, and `server_backend`).

### First Time

The first time you run the system with Docker Compose it will build all the required `Dockerfile`s and run them.

From the root of the project you can run `docker-compose` and specify the configuration file to stand everything up.

```bash
$ docker-compose -f ./cmd/docker-compose.yaml up
```

### Second, Third, Forth, and N Times

After all of the `Dockerfile`s have been built subsequent runs of `docker-compose up` will ONLY run them. If you make changes to any of the code you need to rebuild all of the services.

```bash
$ docker-compose -f ./cmd/docker-compose.yaml build
```

Once everything is rebuilt you can run everything again to see your changes.

```bash
$ docker-compose -f ./cmd/docker-compose.yaml up
```

### One Service at a Time

Some instances you only want to run some instances at a time and you would use `docker-compose run SERVICE_NAME`. Read the `./cmd/docker-compose.yaml` for all the service names.

```bash
$ docker-compose -f ./cmd/docker-compose.yaml run relay_backend
```

### Scaling a Service

Docker Compose makes is very trivial to scale up the number of instances of a service. Currently we can only scale the `relay` service because port numbers will not conflict. Scaling any other service will not work since port numbers are hard coded. For our purposes this is fine. To develop locally we really want to specify any number of relays to run.

Here we can run everything again, but this time it will run 10 instances of the relay service.

```bash
$ docker-compose -f ./cmd/docker-compose.yaml up --scale relay=10
```

## Components

```
                       Relays init and update
        +---------------------------------------------------+   Relay Backend
        |                                                   |   builds Cost &
        |        +----------------------------------------+ |   Route Matrices
        |        |                                        | |
        |   +----+----+       +---------+                +V-V-----------------+
        |   | Relay 2 |       | Relay 4 +----------------> Relay Backend (Go) |
        |   +---------+       +---------+                +^-------+---+---+---+
        |   ||       ||                                   |       |   |   |
   +----+----+       +---------+                          |       |   |   |
   | Relay 1 |       | Relay 3 +--------------------------+       |   |   |
   +---------+       +---------+                                  |   |   |
        ||                ||                  +-------------------V-+ |   |
        ||                ||                +-> Server Backend (Go) | |   |
        ||                ||                | +---------------------+ |   |
        ||          +-------------------+   |     +-------------------V-+ |
        ||          | Game Server (SDK) <---------> Server Backend (Go) | |
        ||          +----------^--------+   |     +---------------------+ |
        ||                     |            |         +-------------------V-+
        ||                     |            +---------> Server Backend (Go) |
+-------------------+          |                      +---------------------+
| Game Client (SDK) <----------+
+-------------------+                                  Server Backends pull
                         Game Server gets              copy of Route Matrix
                         routes  and tells
                         Game Client
```

Made with [asciiflow](http://asciiflow.com/). This text can be imported, changed, and exported to update if needed.

## Relay (C++)

This is the service that suppliers run on their hardware to become part of Network Next.

- Command: `dist/relay` or `make dev-relay`

### To Run

Run `make dev-relay`

### Environment Variables

#### Required

- `RELAY_ADDRESS`: The address of this relay, defaults to "127.0.0.1" when run using make.
- `RELAY_BACKEND_HOSTNAME`: The address of the relay backend, defaults to "http://localhost:40000" when run using make.

#### To get values for the following three variables, see [Generating Key Pairs](#generating-key-pairs)

- `RELAY_PRIVATE_KEY`: The private key of each relay.
- `RELAY_PUBLIC_KEY`: The public key of each relay generated with the private key.
- `RELAY_ROUTER_PUBLIC_KEY`: The public key of the router.

## Relay Backend (Go)

See [cmd/relay_backend](cmd/relay_backend)

## Server (C++)

Reference implementation of a server using the Network Next SDK.

- Command: [`cmd/server`](./cmd/server)
- Dependencies: [`sdk`](./sdk)

## Server Backend (Go)

See [cmd/server_backend](cmd/server_backend)

## Client (C++)

Reference implementation of a client using the Network Next SDK.

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
- `-h` to display all options

## Generating Key Pairs

Keys must be generated with Go's `box.GenerateKey()` function call. The parameter to pass is `rand.Reader`.

Generate two sets. One for relays, and one for the router. Set the following environment variables appropriately:

### Relay Environment Variables

- `RELAY_PRIVATE_KEY`
- `RELAY_PUBLIC_KEY`
- `RELAY_ROUTER_PUBLIC_KEY`

### Relay Backend Environment Variables

- `RELAY_PUBLIC_KEY`
- `RELAY_PRIVATE_KEY`
- `ROUTER_PUBLIC_KEY`
- `ROUTER_PRIVATE_KEY`

Likely the values for the three relay environment variables will be reused for the relay backend.
