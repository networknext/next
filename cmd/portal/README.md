# Web Portal

The Web Portal is responsible for:

1. Providing a single RPC API for data management (Relays, Datacenters, Buyers, etc.)
2. Serve the static front end JavaScript Portal UI

### Load & Scalability

Status: **LOW**  
Scalability: **Vertically**

1. Serves the HTML/CSS/JS frontend portal to customers as they log in and browse the map and session information
2. Fulfills requests from the Ops CLI tool for management of relays, datacenters, etc.

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
