<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This repo contains the Network Next backend.

# Setup on MacOS

1. Install brew from https://brew.sh

	`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

2. Install dependencies

	`brew install golang redis libsodium postgresql@14`

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

8. Clone repo and cd into it

	`git clone git@github.com:networknext/backend.git`

	`cd backend`

9. Select local environment

	`next select local`

10. Build everything and run unit tests

	`make`

	You should see output like:

```console
gaffer@macbook backend % make build
dist/func_test_sdk5
dist/relay
dist/func_backend
dist/libnext5.so
dist/analytics
dist/magic_backend
dist/client
dist/test
dist/raspberry_client
dist/func_client
dist/raspberry_server
dist/server
dist/api
dist/map_cruncher
dist/func_test_backend
dist/portal_cruncher
dist/func_server
dist/raspberry_backend
dist/server_backend
dist/relay_gateway
dist/relay_backend
./run test
?   	github.com/networknext/backend/modules/admin	[no test files]
?   	github.com/networknext/backend/modules/constants	[no test files]
?   	github.com/networknext/backend/modules/database	[no test files]
?   	github.com/networknext/backend/modules/envvar	[no test files]
ok  	github.com/networknext/backend/modules/common	0.210s
ok  	github.com/networknext/backend/modules/core	0.382s
ok  	github.com/networknext/backend/modules/crypto	0.379s
ok  	github.com/networknext/backend/modules/encoding	0.255s
ok  	github.com/networknext/backend/modules/handlers	0.306s
ok  	github.com/networknext/backend/modules/messages	0.223s
ok  	github.com/networknext/backend/modules/packets	0.885s
ok  	github.com/networknext/backend/modules/portal	0.281s
```

11. Run happy path

	`run happy-path`

	You should see output like:

```console
gaffer@macbook backend % run happy-path

don't worry. be happy.

starting api:

   run api

verifying api ... OK

starting relay backend services:

   run magic-backend
   run relay-gateway
   run relay-backend
   run relay-backend HTTP_PORT=30002

verifying magic backend ... OK
verifying relay gateway ... OK
verifying relay backend 1 ... OK
verifying relay backend 2 ... OK

starting relays:

   run relay
   run relay RELAY_PORT=2001
   run relay RELAY_PORT=2002
   run relay RELAY_PORT=2003
   run relay RELAY_PORT=2004

verifying relays ... OK
verifying relay gateway sees relays ... OK
verifying relay backend 1 sees relays ... OK
verifying relay backend 2 sees relays ... OK

starting server backend:

   run server-backend

verifying server backend ... OK

starting portal cruncher:

   run portal-cruncher
   run portal-cruncher HTTP_PORT=40013

verifying portal cruncher 1 ... OK
verifying portal cruncher 2 ... OK

starting map cruncher:

   run map-cruncher
   run map-cruncher HTTP_PORT=40101

verifying map cruncher 1 ... OK
verifying map cruncher 2 ... OK

starting analytics:

   run analytics
   run analytics HTTP_PORT=40002

verifying analytics 1 ... OK
verifying analytics 2 ... OK

waiting for leader election

    analytics ... OK
    map cruncher ... OK
    relay backend ... OK

starting client and server:

   run client
   run server

verifying server ... OK
verifying client ... OK

post validation:

verifying leader election in relay backend ... OK
verifying leader election in analytics ... OK
verifying leader election in analytics ... OK
verifying leader election in map cruncher ... OK
verifying portal cruncher received session update messages ... OK
verifying portal cruncher received server update messages ... OK
verifying portal cruncher received relay update messages ... OK
verifying portal cruncher received near relay update messages ... OK
verifying map cruncher received map update messages ... OK

*** SUCCESS! ***

```

# Setup on Linux (Ubuntu 22.04 LTS)

1. Install dependencies

	`sudo apt install build-essential redis-server postgresql libcurl4-openssl-dev pkg-config -y`

2. Install libsodium

    `wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18-stable.tar.gz && tar -zxf libsodium-1.0.18-stable.tar.gz && cd libsodium-stable && ./configure && make -j && make check && sudo make install && cd ~`

3. Install latest golang

	Find the latest Linux golang download here: https://go.dev/doc/install

	Then do this, with the latest download URL:

	`wget https://go.dev/dl/go1.20.1.linux-amd64.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go*.tar.gz`

4. Add . and go to your path

	Modify ~/.profile to include:

	`export PATH=$PATH:/usr/local/go/bin:.`

	Then source it:

	`source ~/.profile`

5. Configure postgres

   `psql -U postgres -h localhost -c "CREATE USER developer; ALTER USER developer WITH SUPERUSER;"`

   then restart postgres:

   `sudo systemctl restart postgresql`

6. Setup SSH keys on your Linux box for Github

   Follow instructions here: https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent

7. Clone repo and cd into it

	`git clone git@github.com:networknext/backend.git`

	`cd backend`

8. Select local environment

   `next select local`

9. Make everything and run tests

	`make`

10. Build everything and run unit tests

	`make`

	You should see output like:

```console
root@linux:~/backend# make build
dist/func_test_sdk5
dist/relay
dist/func_backend
dist/libnext5.so
dist/analytics
dist/magic_backend
dist/client
dist/test
dist/raspberry_client
dist/func_client
dist/raspberry_server
dist/server
dist/api
dist/map_cruncher
dist/func_test_backend
dist/portal_cruncher
dist/func_server
dist/raspberry_backend
dist/server_backend
dist/relay_gateway
dist/relay_backend
./run test
?   	github.com/networknext/backend/modules/admin	[no test files]
?   	github.com/networknext/backend/modules/constants	[no test files]
?   	github.com/networknext/backend/modules/database	[no test files]
?   	github.com/networknext/backend/modules/envvar	[no test files]
ok  	github.com/networknext/backend/modules/common	0.210s
ok  	github.com/networknext/backend/modules/core	0.382s
ok  	github.com/networknext/backend/modules/crypto	0.379s
ok  	github.com/networknext/backend/modules/encoding	0.255s
ok  	github.com/networknext/backend/modules/handlers	0.306s
ok  	github.com/networknext/backend/modules/messages	0.223s
ok  	github.com/networknext/backend/modules/packets	0.885s
ok  	github.com/networknext/backend/modules/portal	0.281s
```

11. Run happy path

    `run happy-path`

    You should see something like:

```console
root@linux:~/backend# run happy-path

don't worry. be happy.

starting api:

   run api

verifying api ... OK

starting relay backend services:

   run magic-backend
   run relay-gateway
   run relay-backend
   run relay-backend HTTP_PORT=30002

verifying magic backend ... OK
verifying relay gateway ... OK
verifying relay backend 1 ... OK
verifying relay backend 2 ... OK

starting relays:

   run relay
   run relay RELAY_PORT=2001
   run relay RELAY_PORT=2002
   run relay RELAY_PORT=2003
   run relay RELAY_PORT=2004

verifying relays ... OK
verifying relay gateway sees relays ... OK
verifying relay backend 1 sees relays ... OK
verifying relay backend 2 sees relays ... OK

starting server backend:

   run server-backend

verifying server backend ... OK

starting portal cruncher:

   run portal-cruncher
   run portal-cruncher HTTP_PORT=40013

verifying portal cruncher 1 ... OK
verifying portal cruncher 2 ... OK

starting map cruncher:

   run map-cruncher
   run map-cruncher HTTP_PORT=40101

verifying map cruncher 1 ... OK
verifying map cruncher 2 ... OK

starting analytics:

   run analytics
   run analytics HTTP_PORT=40002

verifying analytics 1 ... OK
verifying analytics 2 ... OK

waiting for leader election

    analytics ... OK
    map cruncher ... OK
    relay backend ... OK

starting client and server:

   run client
   run server

verifying server ... OK
verifying client ... OK

post validation:

verifying leader election in relay backend ... OK
verifying leader election in analytics ... OK
verifying leader election in analytics ... OK
verifying leader election in map cruncher ... OK
verifying portal cruncher received session update messages ... OK
verifying portal cruncher received server update messages ... OK
verifying portal cruncher received relay update messages ... OK
verifying portal cruncher received near relay update messages ... OK
verifying map cruncher received map update messages ... OK

*** SUCCESS! ***

```
