<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This repo contains the Network Next backend.

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
	1. Server Backend 4 & 5 (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 0)
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
    7. Magic Frontend (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 30s)
    8. Magic Backend (Rolling Replace, Maximum Surge 8, Maximum Unavailable 0, Minimum Wait Time 30s)
6. The following services can be deployed at any time:
	1. Relay Pusher (`make deploy-relay-pusher-prod`)
	2. Pingdom (`make deploy-pingdom-prod`)


## Development

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

5. Install libzmq3-dev
	Linux:
	`sudo apt install libzmq3-dev`
	Mac:
	`brew install zeromq`

6. Install Go (must be 1.13+)
	`cd /usr/local/`
	`sudo curl https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz | sudo tar -zxv`
	Add Go to PATH:
		`echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.profile`
	NOTE: For changes to your `.profile` to reflect in the terminal, sign out and sign back in.
	If you're running WSL, you can stop it by typing `wsl -t <distro>` in Powershell and start it again.

7. Install Redis
	`sudo apt install redis-server`

8. Clone the repo with an SSH key
	Instructions from `https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent`

	`ssh-keygen -t rsa -b 4096 -C "your_email@example.com"`
	`eval $(ssh-agent -s)`
	`ssh-add <filepath_priv>` Replace <filepath_priv> with the path to your SSH private key (ex. ~/.ssh/id_rsa)
	Copy the contents of your SSH public key (in same directory as public key, ex. ~/.ssh/id_rsa.pub)
	Add the SSH public key to your Github account
		- Login and go to Settings > SSH and GPG Keys > New SSH Key and paste in your key
  `git clone git@github.com:networknext/backend.git`
  `cd <clone_path>` where `<clone_path>` is the directory you cloned the repo to (usually `~/backend`)

9. Run tests to confirm everything is working properly
	`make test`

## Running the "Happy Path"

A good test to see if everything works and is installed is to run the "Happy Path". For this you will need to run the following command:

	`make dev-happy-path`

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

To run the unit tests, run `make test`. 

## Functional Tests

In addition to unit tests, the system also take advantage of functional tests that run real world scenarios to make sure that all of the components are working properly.

To run the functional tests for SDK4, run `make test-func4`. 

To run the functional tests for SDK5, run `make test-func5`. 

(todo: functional tests for backend...)

The functional tests take a long time to run locally, but they automatically run in || via semaphore on every commit.
