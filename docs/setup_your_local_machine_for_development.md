<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup for Development

In this section you will setup your local development environment so you can build the Network Next source, run unit tests, and run the "happy path" to verify the system is working correctly.

Setup depends on whether you are developing on Linux or MacOS. Development on Windows is not supported.

# Setup on Linux (Ubuntu 22.04 LTS)

1. Install dependencies

	`sudo apt install build-essential redis-server postgresql libcurl4-openssl-dev pkg-config -y`

2. Install libsodium

    `wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.19-stable.tar.gz && tar -zxf libsodium-1.0.19-stable.tar.gz && cd libsodium-stable && ./configure && make -j && make check && sudo make install && sudo ldconfig && cd ~`

3. Install latest golang

	Find the latest Linux golang download here: https://go.dev/doc/install

	Then do this, with the latest download URL for your platform:

	`wget https://go.dev/dl/go1.20.1.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go*.tar.gz`

4. Add . and go to your path

	Modify ~/.profile to include:

	`echo "export PATH=/usr/local/go/bin:$PATH:." >> ~/.profile`

	Then source it:

	`source ~/.profile`

5. Configure postgres

   `psql -U postgres -h localhost -c "CREATE USER developer; ALTER USER developer WITH SUPERUSER;"`

   If this step fails, you may need to work around the default postgres account requiring logging in as the user postgres.

   See https://blog.jcharistech.com/2022/09/07/logging-into-postgresql-without-password-prompt/ for a workaround. Apply the "trust" setting to the 127.0.0.1 and ::1 lines in pg_hba.conf then `sudo systemctl restart postgresql` and you should be able to run the step above without being prompted for a password.   

7. Go to the next directory at the command line

    `cd ~/next`

8. Select local environment

   `next select local`

9. Build and run unit tests

   `make`

You should see output like:

```console
root@linux:~/next# make build
dist/func_test_terraform
dist/func_test_sdk
dist/func_test_database
dist/relay-debug
dist/func_test_api
dist/func_backend
dist/func_test_portal
dist/func_test_backend
dist/ip2location
dist/load_test_relays
dist/load_test_servers
dist/load_test_sessions
dist/relay_gateway
dist/magic_backend
dist/raspberry_backend
dist/server_cruncher
dist/api
dist/session_cruncher
dist/func_test_relay
dist/libnext.so
dist/soak_test_relay
dist/test
dist/raspberry_server
dist/client
dist/server
dist/func_client
dist/raspberry_client
dist/server_backend
dist/relay_backend
dist/func_server
./run test

?   	github.com/networknext/next/modules/admin	[no test files]
?   	github.com/networknext/next/modules/constants	[no test files]
ok  	github.com/networknext/next/modules/common	(cached)
?   	github.com/networknext/next/modules/database	[no test files]
?   	github.com/networknext/next/modules/envvar	[no test files]
ok  	github.com/networknext/next/modules/core	(cached)
ok  	github.com/networknext/next/modules/crypto	(cached)
ok  	github.com/networknext/next/modules/encoding	(cached)
?   	github.com/networknext/next/modules/ip2location	[no test files]
ok  	github.com/networknext/next/modules/handlers	(cached)
ok  	github.com/networknext/next/modules/messages	(cached)
ok  	github.com/networknext/next/modules/packets	(cached)
ok  	github.com/networknext/next/modules/portal	(cached)
```

11. Run happy path

    `run happy-path`

You should output like:

```console
root@linux:~/next# run happy-path

don't worry. be happy.

starting session cruncher:

   run session-cruncher

verifying session cruncher ... OK

starting server cruncher:

   run server-cruncher

verifying server cruncher ... OK
starting api:

   run api

verifying api ... OK

starting relay backend services:

   run magic-backend
   run relay-gateway
   run relay-backend

verifying magic backend ... OK
verifying relay gateway ... OK
verifying relay backend ... OK

starting relays:

   run relay RELAY_PORT=2000 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2001 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2002 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2003 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2004 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1

verifying relays ... OK
verifying relay gateway sees relays ... OK
verifying relay backend sees relays ... OK

starting server backend:

   run server-backend

verifying server backend ... OK

waiting for leader election

    relay backend ... OK

starting client and server:

   run client
   run server

verifying server ... OK
verifying client ... OK

*** SUCCESS! ***

```

Next step: [Setup Prerequisites](setup_prerequisites.md).

# Setup on MacOS

1. Install brew from https://brew.sh

   `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

2. Install dependencies

   `brew install golang redis libsodium pkg-config postgresql@14`

3. Start redis

   `brew services start redis`

4. Start postgres

   `brew services start postgresql@14`

5. Configure postgres

   `psql -c "CREATE USER developer; ALTER USER developer WITH SUPERUSER;"`

   `brew services start postgresql@14`

6. Add . to your path

   Modify ~/.zshrc to include:

   `export PATH=.:/opt/homebrew/bin:$PATH`

7. Setup SSH keys on your Mac for Github

   Follow instructions here: https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent

8. Go to the next directory on the command line

   `cd next`

9. Select local environment

   `next select local`

10. Build everything and run unit tests

   `make`

   You should see output like:

```console
gaffer@macbook next % make build
dist/func_test_terraform
dist/func_test_sdk
dist/func_test_database
dist/relay-debug
dist/func_test_api
dist/func_backend
dist/func_test_portal
dist/func_test_backend
dist/ip2location
dist/load_test_relays
dist/load_test_servers
dist/load_test_sessions
dist/relay_gateway
dist/magic_backend
dist/raspberry_backend
dist/server_cruncher
dist/api
dist/session_cruncher
dist/func_test_relay
dist/libnext.so
dist/soak_test_relay
dist/test
dist/raspberry_server
dist/client
dist/server
dist/func_client
dist/raspberry_client
dist/server_backend
dist/relay_backend
dist/func_server
./run test

?   	github.com/networknext/next/modules/admin	[no test files]
?   	github.com/networknext/next/modules/constants	[no test files]
ok  	github.com/networknext/next/modules/common	(cached)
?   	github.com/networknext/next/modules/database	[no test files]
?   	github.com/networknext/next/modules/envvar	[no test files]
ok  	github.com/networknext/next/modules/core	(cached)
ok  	github.com/networknext/next/modules/crypto	(cached)
ok  	github.com/networknext/next/modules/encoding	(cached)
?   	github.com/networknext/next/modules/ip2location	[no test files]
ok  	github.com/networknext/next/modules/handlers	(cached)
ok  	github.com/networknext/next/modules/messages	(cached)
ok  	github.com/networknext/next/modules/packets	(cached)
ok  	github.com/networknext/next/modules/portal	(cached)
```

11. Run happy path

   `run happy-path`

   You should see output like:

```console
gaffer@macbook next % run happy-path

don't worry. be happy.

starting session cruncher:

   run session-cruncher

verifying session cruncher ... OK

starting server cruncher:

   run server-cruncher

verifying server cruncher ... OK
starting api:

   run api

verifying api ... OK

starting relay backend services:

   run magic-backend
   run relay-gateway
   run relay-backend

verifying magic backend ... OK
verifying relay gateway ... OK
verifying relay backend ... OK

starting relays:

   run relay RELAY_PORT=2000 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2001 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2002 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2003 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1
   run relay RELAY_PORT=2004 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1

verifying relays ... OK
verifying relay gateway sees relays ... OK
verifying relay backend sees relays ... OK

starting server backend:

   run server-backend

verifying server backend ... OK

waiting for leader election

    relay backend ... OK

starting client and server:

   run client
   run server

verifying server ... OK
verifying client ... OK

*** SUCCESS! ***

```

Next step: [setup prerequisites](setup_prerequisites.md).
