# Web Portal

The Web Portal is responsible for:

1. Listening for new relays that come online and initialize them to be apart of set of available relays for routing
2. Accepting stat updates (RTT, Jitter, Packet Loss) from all relays and store their stats
3. Generate a cost then route matrix of optimized routes that are available which the Server Backend pulls a copy of
4. Keeps an updated relay location list using Redis to find nearby relays so the Server Backend can find relays near game clients

### To Run

Run `make dev-portal`

### Logging

Levels are cumulative so if you set `BACKEND_LOG_LEVEL=info` you will get `error` and `warn` too.

The default setting is `warn` when running `make dev-portal`. To override this you can set your own value by doing `make BACKEND_LOG_LEVEL=debug dev-portal`.

### Environment Variables

#### Required

- `PORT`: the port to run the web server on.
- `REDIS_HOST_RELAYS`: address of the Redis server you want to connect to retrieve relay information.
- `ROUTE_MATRIX_URI`: URI to the route matrix, either the URL to the relay backend's route matrix endpoint or a local route matrix file.
- `BASIC_AUTH_USERNAME`: the username needed to login to the web portal.
- `BASIC_AUTH_PASSWORD`: the password needed to login to the web portal.

#### Optional

- `BACKEND_LOG_LEVEL`: one of `none`, `error`, `warn`, `info`, `debug`

### What it does

- Displays a route matrix analysis by pulling the route matrix from the relay backend
- Displays a list of relays from the Redis server
