<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This is a monorepo that contains the Network Next backend.

## Monitoring

[![GCP Metrics](https://img.shields.io/badge/GCP-metrics-lightgray?style=for-the-badge&logo=google-cloud)](https://console.cloud.google.com/monitoring/dashboards/custom/17676944979741730633?organizationId=434699063105&project=network-next-v3-stackdriver-ws&timeDomain=1h)

[![GCP Logs](https://img.shields.io/badge/GCP-logs-lightgray?style=for-the-badge&logo=google-cloud)](https://console.cloud.google.com/logs)

## Git Workflow

### Development Release

1. Branch from `master` to a properly named branch for your bug/feature
2. Do your work in your bug/feature branch
3. Issue a pull request into `master` and mark it according to what you need
	- **Draft**: mark a PR as a draft to expose you have started work and have questions/comments in order to complete the work
	- **Ready for Review**: mark a PR as ready for review and include the appropriate reviewers when unit tests for your bug/feature are all passing
4. Once your pull request has been reviewed merge it into `master`
5. Semaphore will build your PR and copy artifacts to the google cloud gs://dev_artifacts bucket automatically.
6. Manually trigger a rolling update in google cloud on each managed instance group you want to update to latest code.

### Production Release

1. Ensure tests pass locally as a sanity check
2. Create a PR to push your changes to the "prod" branch
5. Semaphore will build your PR and copy artifacts to the google cloud gs://prod_artifacts bucket automatically.
5. Manually trigger a rolling update in google cloud on each managed instance group you want to update to latest code.

## Development

IMPORTANT: This repo uses [Git Submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules) to link in the [SDK](https://github.com/networknext/console). In order for this to work you need clone and interact with this repo over [SSH](https://help.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh).

```bash
git clone git@github.com:networknext/backend.git
git submodule init
git submodule update
```

The tool chain used for development is kept simple to make it easy for any operating system to install and use and work out of the box for POSIX Linux distributions.

- [GCP Cloud SDK](https://cloud.google.com/sdk/docs/quickstarts): needed for the `gsutil` command to publish artifacts
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

## Recommended Setup

The following steps outline the setup process on a standard Linux Debian/Ubuntu installation. Dependencies are aquired through the package manager for ease of use where possible.
For Mac OSX, use the corresponding `brew` command with the equivalent package name.
For Windows, use WSL or WSL 2 to install a Linux environment and follow the steps accordingly.

NOTE: This is NOT the only way to set up the project, this is just ONE way. Feel free to set up in whatever way is easiest for you.

1. Update package manager
	`sudo apt update`

2. Install build-essential -- This will install make, gcc, and g++
	Linux:
	`sudo apt install build-essential`
	Mac:
	`xcode-select --install`

3. Install pkg-config
	Linux:
	`sudo apt install pkg-config`
	Mac:
	`brew install pkg-config`

4. Install libsodium
	Linux:
	`sudo apt install libsodium-dev`
	Mac:
	`brew install libsodium`

5. Install libcurl
	Linux:
	`sudo apt install libcurl4-openssl-dev`
	Mac:
	`brew install openssl`

6. Install RapidJSON
	Linux:
  `sudo apt install rapidjson-dev`
	Mac:
	`brew install rapidjson`

7. Install g++ version 8
	Linux:
  `sudo apt install g++-8`
	Mac:
	`brew install gcc@8`

8. Install Go (must be 1.13+)
	`cd /usr/local/`
	`sudo curl https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz | sudo tar -zxv`
	Add Go to PATH:
		`echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.profile`
	NOTE: For changes to your `.profile` to reflect in the terminal, sign out and sign back in.
	If you're running WSL, you can stop it by typing `wsl -t <distro>` in Powershell and start it again.

9. Install Redis
	`sudo apt install redis-server`

10. Install Docker
	NOTE: If you're running Windows with WSL 1, follow this article: https://nickjanetakis.com/blog/setting-up-docker-for-windows-and-wsl-to-work-flawlessly
	For WSL 2 these steps aren't necessary, installing Docker Desktop by itself is sufficient.

11. Clone the repo with an SSH key
	Instructions from `https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent`

	`ssh-keygen -t rsa -b 4096 -C "your_email@example.com"`
	`eval $(ssh-agent -s)`
	`ssh-add <filepath_priv>` Replace <filepath_priv> with the path to your SSH private key (ex. ~/.ssh/id_rsa)
	Copy the contents of your SSH public key (in same directory as public key, ex. ~/.ssh/id_rsa.pub)
	Add the SSH public key to your Github account
		- Login and go to Settings > SSH and GPG Keys > New SSH Key and paste in your key
  `git clone git@github.com:networknext/backend.git`
  `cd <clone_path>` where `<clone_path>` is the directory you cloned the repo to (usually `~/backend`)

12. Init and update git submodules
	`git submodule init`
	`git submodule update`

13. Install Google Cloud SDK
	Instructions from `https://cloud.google.com/sdk/docs/quickstart-debian-ubuntu`
	For other platforms, see `https://cloud.google.com/sdk/docs/quickstarts`

	`echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list`
	`curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -`
	`sudo apt update && sudo apt install google-cloud-sdk`
	`gcloud init`
	When asked to choose a cloud project, choose `network-next-v3-dev`

14. Install the Firestore emulator
	`sudo apt install google-cloud-sdk-firestore-emulator`

15. Install the Pub/Sub emulator
	`sudo apt install google-cloud-sdk-pubsub-emulator`

16. Run tests to confirm everything is working properly
	`make test`
	`make test-func-parallel`

## Running the "Happy Path"

A good test to see if everything works and is installed is to run the "Happy Path". For this you will need to run the following commands **in separate terminal sessions**.

1. `redis-cli flushall && make BACKEND_LOG_LEVEL=info dev-relay-backend`: this will clear your local redis completely to start fresh and then run the relay backend
2. `make dev-multi-relays`: this will run 10 instances of a relay and each will register themselves with the relay backend
3. `make BACKEND_LOG_LEVEL=info dev-server-backend`: this will run the server backend and start pulling route information from the relay backend every 10 seconds
4. `make dev-server`: this will run a fake game server and register itself with the server backend
5. `make dev-client`: this will run a fake game client and request a route from the server which will ask the server backend for a new route for the game client. You can also run `make dev-multi-clients` to create 20 client sessions.
6. `make JWT_AUDIENCE="oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n" dev-portal`: this will run the Portal RPC API and Portal UI. You can visit https://localhost:20000 to view currently connected sessions.

You should see the fake game server upgrade the clients session and get `(next route)` and `(continue route)` from the server backend which it sends to the fake game client.

Simultaneously you will see the terminal with the relays logging `session created` indicating traffic is passing through relays.

## Backend

All of these services are controlled and deployed by us.

- [`cmd/portal`](cmd/portal)
- [`cmd/relay`](cmd/relay)
- [`cmd/relay_backend`](cmd/relay_backend)
- [`cmd/server_backend`](cmd/server_backend)

## SDK

The [`SDK`](./sdk) is shipped to customers to use in their game client and server implementations. The client and server here are slim reference implementations so we can use the SDK locally.

- [`cmd/server`](./cmd/server)
- [`cmd/client`](./cmd/client)

## High-Level Flow Diagram

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

## Testing

Unit tests and functional tests are used in order to test code before it ships.

## Unit Tests

To run the unit tests, run `make test`. This will run unit tests for the SDK, relay, and all backend components.
Because there are some remote services such as GCP that the backend components talk to, not all unit tests can be run without gcloud emulators or certain environment variables set. If the requirements for each of unit tests aren't met, they will be skipped.
Here are the requirements to run each of the GCP related unit tests:

Firestore:
Install the gcloud firestore emulator: (Note that the emulator needs a Java Runtime Environment version 1.8 or higher installed and added to PATH)
`gcloud components install beta`
`gcloud components install cloud-firestore-emulator`
or
`sudo apt install google-cloud-sdk-firestore-emulator`

    Add the environment variable `FIRESTORE_EMULATOR_HOST` to your makefile with the local address of the emulator (ex. `localhost:8000`).

Stackdriver Metrics:
Add the environment variable `GOOGLE_PROJECT_ID` to your makefile. Set it to a GCP project you have credentials to (ex. `network-next-v3-dev`).
Add the environment variable `GOOGLE_APPLICATION_CREDENTIALS` to your makefile. Set it to the file path of your credentials file (ex. `$(CURRENT_DIR)/testdata/v3-dev-creds.json`).

Pub/Sub:
Install the gcloud pubsub emulator: (Note that the emulator needs a Java Runtime Environment version 1.7 or higher installed and added to PATH)
`gcloud components install beta`
`gcloud components install pubsub-emulator`
or
`sudo apt install google-cloud-sdk-pubsub-emulator`

    Add the environment variable `FIRESTORE_PUBSUB_HOST` to your makefile with the local address of the emulator (ex. `localhost:9000`).

## Functional Tests

In addition to unit tests, the system also take advantage of functional tests that run real world scenarios to make sure that all of the components are working properly.
To run the functional tests, run `make test-func`, or more preferably, `make test-func-parallel`, since the func tests take a long time to run in series.

## Docker and Docker Compose

While all of the components can be run locally either independently or collectively it can be tedious to run multiple relays to get a true test of everything. We can leverage Docker and Docker Compose to easily stand everything up as a system. There is a [`./cmd/docker-compose.yaml`](./cmd/docker-compose.yaml) along with all required `Dockerfile`s in each of the binary directories to create the system of backend services (`relay_backend` and `server_backend`).

### First Time

The first time you run the system with Docker Compose it will build all the required `Dockerfile`s and run them.

From the root of the project you can run `docker-compose` and specify the configuration file to stand everything up.

```bash
$ docker-compose -f ./cmd/docker-compose.yaml up
```

### Second, Third, Fourth, and N Times

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
