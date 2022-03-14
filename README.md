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
5. To deploy to dev, merge either your branch or `master` into the `dev` branch
6. Semaphore will build your PR and copy artifacts to the google cloud gs://dev_artifacts bucket automatically.
7. Manually trigger a rolling update in google cloud on each managed instance group you want to update to latest code. Follow the instructions and deployment process in [Production Release](#Production-Release) to ensure nothing goes wrong when deploying your changes to prod.
8. The relay backend VMs are deployed with the make tool via `make deploy-relay-backend-dev-1` and `make deploy-relay-backend-dev-2`. Wait 5 - 10 minutes between each deploy to avoid sessions falling back to direct.
9. The portal crunchers VMs are deployed with the make tool via `make deploy-portal-crunchers-dev`.

### Production Release

1. Ensure tests pass locally as a sanity check
2. Create a PR to push your changes to the "prod" branch
3. Semaphore will build your PR and copy artifacts to the google cloud gs://prod_artifacts bucket automatically.
4. Deploy the Server Backend half of the backend first, since it only relies on the route matrix. Deploy supporting services in order of consumer to producer, and roll the Server Backend MIG as the last step:
	
	----
	1. Portal Backend (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 0)
	2. Portal Cruncher (`make deploy-portal-crunchers-prod`)
	3. Ghost Army (`make deploy-ghost-army-prod`)
	----
	1. Billing (Rolling Replace, Maximum Surge 5, Maximum Unavailable 1, Minimum Wait Time 0)
	---- 
	1. Server Backend 4 (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 0)
		- Note: there is a 1 hour connection drain on server backend instances to reduce fallbacks to direct.
		- Once the new server backend instances are healthy and running, force the UDP load balancer to stop sending traffic to the old instance by setting the metadata field of each **old** server backend instances. Use the template below to set the metadata value per **old** instance before the connection drain ends. You will need the prod credentials to do this (i.e. setting `GOOGLE_APPLICATION_CREDENTIALS` to the prod credentials file downloaded from GCP).
			- Example: `gcloud compute instances add-metadata server-backend4-mig-6mr0 --metadata connection-drain=true --project=network-next-v3-prod`
5. Deploy the Relay Backend half of the backend next because it provides the route matrix. If any changes were made to the route matrix, wait for the new Server Backend instances to become healthy and the old ones to become unhealthy (no need to wait for the connection drain to fully finish).
	1. Analytics (Rolling Replace, Maximum Surge 5, Maximum Unavailable 1, Minimum Wait Time 0)
	2. Analytics Pusher (`make deploy-analytics-pusher-prod`)
	3. Relay Frontend (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 30s)
	4. Relay Backend (`make deploy-relay-backend-prod-1`, and after 10 minutes, `make deploy-relay-backend-prod-2`)
		- Check the `/status` endpoint of Relay Backend 1 to ensure a similar amount of routes are generated compared to Relay Backend 2 before deploying Relay Backend 2
	5. Relay Gateway (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 30s)
	6. Relay Forwarder (`make deploy-relay-forwarder-prod`)
6. The following services can be deployed at any time:
	1. Relay Pusher (`make deploy-relay-pusher-prod`)
	2. Pingdom (`make deploy-pingdom-prod`)


## Development

IMPORTANT: This repo uses [Git Submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules) to link in [SDK4](https://github.com/networknext/sdk4) and [SDK5](https://github.com/networknext/sdk5/). In order for this to work you need clone and interact with this repo over [SSH](https://help.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh).

```bash
git clone git@github.com:networknext/backend.git
git submodule init
git submodule update
```

The tool chain used for development is kept simple to make it easy for any operating system to install and use and work out of the box for POSIX Linux distributions.

- [GCP Cloud SDK](https://cloud.google.com/sdk/docs/quickstarts): needed for the `gsutil` command to publish artifacts
- [Redis](https://redis.io)
- [make](http://man7.org/linux/man-pages/man1/make.1.html)
- [sh](https://linux.die.net/man/1/sh)
- [Go](https://golang.org/dl/#stable) (at least Go 1.13)
- [g++](http://man7.org/linux/man-pages/man1/g++.1.html)
  - [libcurl](https://curl.haxx.se/libcurl/)
  - [libsodium](https://libsodium.gitbook.io)
  - [libpthread](https://www.gnu.org/software/hurd/libpthread.html)
  - [libzmq3-dev](https://zeromq.org/download/)

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

6. Install libzmq3-dev
	Linux:
	`sudo apt install libzmq3-dev`
	Mac:
	`brew install zmq`

7. Install RapidJSON
	Linux:
  `sudo apt install rapidjson-dev`
	Mac:
	`brew install rapidjson`

8. Install g++ version 8
	Linux:
  `sudo apt install g++-8`
	Mac:
	`brew install gcc@8`

9. Install Go (must be 1.13+)
	`cd /usr/local/`
	`sudo curl https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz | sudo tar -zxv`
	Add Go to PATH:
		`echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.profile`
	NOTE: For changes to your `.profile` to reflect in the terminal, sign out and sign back in.
	If you're running WSL, you can stop it by typing `wsl -t <distro>` in Powershell and start it again.

10. Install Redis
	`sudo apt install redis-server`

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

14. Install the Bigtable emulator
	`sudo apt install google-cloud-sdk-bigtable-emulator`
	`sudo apt install google-cloud-sdk-cbt`

15. Install the Pub/Sub emulator
	`sudo apt install google-cloud-sdk-pubsub-emulator`

16. Install SQLite3
	`sudo apt install sqlite3`

	With the sqlite3 package installed no other setup is required to use sqlite3 for unit testing.

17. Run tests to confirm everything is working properly
	`make test`
	`make test-func-parallel`

## Running the "Happy Path"

A good test to see if everything works and is installed is to run the "Happy Path". For this you will need to run the following commands **in separate terminal sessions**.

1. `./next select local`: setup local environment
2. `make dev-relay-gateway`: run the relay gateway
3. `make dev-relay-backend-1`: run the relay backend 1 (requires redis-server)
4. `make dev-relay-backend-2`: run the relay backend 2 (requires redis-server)
5. `make dev-relay-frontend`: run the relay frontend (require redis-server)
6. `make dev-relay`: this will run a reference relay that will talk to the relay gateway. You can also run `make dev-relays` to create 10 relays.
7. `make dev-server-backend4`: run the server backend for sdk4
8. `make dev-server4`: this will run a fake game server for sdk4 and register itself with the server backend
9. `make dev-client4`: this will run a fake game client for sdk4 and request a route from the server which will ask the server backend for a new route for the game client. You can also run `make dev-clients4` to create 10 client sessions.
10. `make dev-portal-cruncher-1`: run portal cruncher 1
11. `make dev-portal-cruncher-2`: run portal cruncher 2
12. `make dev-portal`: this will run the Portal Backend RPC API
13. You will then need to clone the portal repo, https://github.com/networknext/portal, run through its setup, and run `npm run serve`. This will launch the portal at http://127.0.0.1:8080

You should see the fake game server upgrade the clients session and get `(next route)` and `(continue route)` from the server backend which it sends to the fake game client.

Simultaneously you will see the terminal with the relays logging `session created` indicating traffic is passing through relays.

NOTE: In local testing, network next routes are provided immediately, but in practice it will take 5 minutes of relays sending updates before network next routes are generated.

## SQL Storers and the Happy Path

The `FEATURE_POSTGRESQL` environment variable is used to setup SQL storers. If this variable is unset an in_memory storer will be used instead.

### SQLite3

Set the feature environment variable `FEATURE_POSTGRESQL=false` to use sqlite3 as a storer database. The `testdata/sqlite3-empty.sql` file will be loaded by the `NewSQLite3()` function and the `SeedSQLStorage()` function will sideload the database with all the data required to run the Happy Path.

### PostgreSQL

For a more functional test, set the feature environment variable `FEATURE_POSTGRESQL=true` to use PostgreSQL as a storer database. The `SeedSQLStorage()` function will not be used to sideload Happy Path data in this case. There are 2 SQL files in the `testdata` directory which can be used to sideload all the data _prior_ to running the Happy Path. Log into your **local** PostgreSQL server with `psql` and load the SQL files:

```
postgres=> drop database nn; -- if it already exists - this is critical!
postgres=> create database nn;
postgres=> \c nn
nn=> \i pgsql-empty.sql
nn=> \i hp-pgsql-seed.sql
```

At this point your local PostgreSQL server is ready to go. Note: installing and setting up a local PostgreSQL server is beyond the scope of this document.  

## Local Billing and Analytics

It is also possible to locally debug what data is being sent to the `billing` and `analytics` services. To verify that the data being sent to the service is correct:

1. Make sure you have the pubsub emulator installed (instructions are in [Recommended Setup](#Recommended-Setup))
2. Define `PUBSUB_EMULATOR_HOST` in the makefile or in the command (ex. define it as `127.0.0.1:9000`)
3. Along with the rest of the [Happy Path](#Running-the-"Happy-Path"), run the Google Pub/Sub Emulator with `gcloud beta emulators pubsub start --project=local --host-port=127.0.0.1:9000`
3. Run `make dev-billing` or `make dev-analytics`
	- NOTE: the server backend has to be running prior to starting local billing to create the Pub/Sub topic
	- NOTE: the analytics pusher has to be running prior to starting local analytics to create the Pub/Sub topics

This will use a local implementation to print out the entry to the console window.

## SDK

The [`SDK4`](./sdk4) is shipped to customers to use in their game client and server implementations ([`SDK5`](./sdk5) is on the way). The client and server here are slim reference implementations so we can use the SDK locally.

- [`cmd/server4`](./cmd/server4)
- [`cmd/client4`](./cmd/client4)
- [`cmd/server5`](./cmd/server5)
- [`cmd/client5`](./cmd/client5)

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

To run the unit tests, run `make test`. This will run unit tests for all backend components.
Because there are some remote services such as GCP that the backend components talk to, not all unit tests can be run without gcloud emulators or certain environment variables set. If the requirements for each of unit tests aren't met, they will be skipped.
Here are the requirements to run each of the GCP related unit tests:

Stackdriver Metrics:
Add the environment variable `GOOGLE_PROJECT_ID` to your makefile. Set it to a GCP project you have credentials to (ex. `network-next-v3-dev`).
Add the environment variable `GOOGLE_APPLICATION_CREDENTIALS` to your makefile. Set it to the file path of your credentials file (ex. `$(CURRENT_DIR)/testdata/v3-dev-creds.json`).

Pub/Sub:
Install the gcloud pubsub emulator: (Note that the emulator needs a Java Runtime Environment version 1.7 or higher installed and added to PATH)
`gcloud components install beta`
`gcloud components install pubsub-emulator`
or
`sudo apt install google-cloud-sdk-pubsub-emulator`

    Add the environment variable `PUBSUB_EMULATOR_HOST` to your makefile with the local address of the emulator (ex. `127.0.0.1:9000`).

Bigtable:
Install the gcloud bigtable emulator:
`gcloud components install beta`
`gcloud components install bigtable`
or
`sudo apt install google-cloud-sdk-bigtable-emulator`

    Add the environment variable `BIGTABLE_EMULATOR_HOST` to your makefile with the local address of the emulator (ex. `localhost:8086`).

## Functional Tests

In addition to unit tests, the system also take advantage of functional tests that run real world scenarios to make sure that all of the components are working properly.
To run the functional tests for SDK4, run `make test-func4`, or more preferably, `make test-func4-parallel`, since the func tests take a long time to run in series. Functional tests for SDK5 can be run using `make test-func5` and `make test-func5-parallel`.

