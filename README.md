<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

This repo contains the Network Next backend.

# Setup

1. Install brew from https://brew.sh

	`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

2. Install dependencies

	`brew install golang redis libsodium postgresql@14`

3. Start redis

	`brew services start redis`

4. Start postgres

	`brew services start postgresql@14`

5. Add '.' to your path

	Modify .zshrc to include:

	`export PATH=.:/opt/homebrew/bin:$PATH`

6. Clone repo

	`git clone git@github.com:networknext/backend.git`

	`cd backend`

7. Select local environment

	`next select local`

8. Build everything and run unit tests

	`make`

	You should see output like:

	`dist/func_tests_sdk5
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
dist/func_tests_backend
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
ok  	github.com/networknext/backend/modules/portal	0.281s`

9. Run happy path

	`run happy-path`

	You should see output like:

	