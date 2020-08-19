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

Run `make JWT_AUDIENCE="oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n" dev-portal`

### Logging

Levels are cumulative so if you set `BACKEND_LOG_LEVEL=info` you will get `error` and `warn` too.

The default setting is `warn` when running `make JWT_AUDIENCE="oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n" dev-portal`. To override this you can set your own value by doing `make BACKEND_LOG_LEVEL=debug JWT_AUDIENCE="oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n" dev-portal`.

### Environment Variables

#### Required

- `PORT`: the port to run the web server on.
- `ROUTE_MATRIX_URI`: URI to the route matrix, either the URL to the relay backend's route matrix endpoint or a local route matrix file.
- `BASIC_AUTH_USERNAME`: the username needed to login to the web portal.
- `BASIC_AUTH_PASSWORD`: the password needed to login to the web portal.
- `REDIS_HOST_TOP_SESSIONS`: redis instance to save top sessions.
- `REDIS_HOST_SESSION_MAP`: redis instance to save map points.
- `REDIS_HOST_SESSION_META`: redis instance to save session metadata.
- `REDIS_HOST_SESSION_SLICES`: redis instance to save session slices.

#### Optional

- `BACKEND_LOG_LEVEL`: one of `none`, `error`, `warn`, `info`, `debug`

#### Structure

`portal.go`:
* Sets up all env variables needed by the portal
* Sets up all services needed by the portal through a JSONRPC server handler
* Serves the static portal files (index.html and portal.js)

`public/index.html`:
* Template for the portal UI

`public/js/portal.js`:
* Consists of a single Vue component, and handlers that organize the logic used to run the Portal UI.
* The Vue component is used to handle showing/hiding workspaces and functionality.
* Handlers are used some what as a state management system and a way of organizing the different parts of the Portal
