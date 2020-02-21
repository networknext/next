# Server Backend

The Server Backend is responsible for:

1. Listening for new game servers that come online to allow client sessions for
2. Pulls the Route Matrix from the Relay Backend so each running instance of the Server Backend has a local/fast copy of routes
3. Accepts client sessions from the game server to find relays near the client, find a route, and send it back to the game server
4. Records metrics and billing information so that the routes being served are billed accordingly

### To Run

Run `make dev-server-backend`

### Logging

Levels are cumulative so if you set `BACKEND_LOG_LEVEL=info` you will get `error` and `warn` too.

The default setting is `warn` when running `make dev-relay-backend` and `make dev-server-backend`. To override this you can set your own value by doing `make BACKEND_LOG_LEVEL=debug dev-relay-backend` and `make BACKEND_LOG_LEVEL=debug dev-server-backend`.

### Environment Variables

#### Required

- `BACKEND_LOG_LEVEL`: one of `none`, `error`, `warn`, `info`, `debug`
- `SERVER_BACKEND_PUBLIC_KEY`: the public key of the server_backend
- `SERVER_BACKEND_PRIVATE_KEY`: the private key of the server_backend to sign responses
- `RELAY_ROUTER_PUBLIC_KEY`: the public key of the router
- `RELAY_ROUTER_PRIVATE_KEY`: the private key of the router used to encrypt routes
- `ROUTE_MATRIX_URI`: a URL or local file path to a route matrix binary
- `MAXMIND_DB_URI`: local path to a `.mmdb` file for IP lookups
- `REDIS_HOST`: address of the Redis server you want to connect to, uses the in-memory version if not supplied or invalid

#### Optional

- `GCP_CREDENTIALS`: JSON blob or path to a .json file for the GCP credentials needed
- `GCP_METRICS_PROJECT`: The name of the StackDriver project that will monitor metrics
- `GCP_METRICS_CLUSTER_LOCATION`: The location of the kubernetes cluster that will monitor metrics
- `GCP_METRICS_CLUSTER_NAME`: The name of the kubernetes cluster that will monitor metrics
- `GCP_METRICS_POD_NAME`: The name of the kubernetes pod that will monitor metrics
- `GCP_METRICS_CONTAINER_NAME`: The name of the kubernetes container that will monitor metrics
- `GCP_METRICS_NAMESPACE_NAME`: The name of the kubernetes namespace that will monitor metrics

#### IMPORTANT

You must have at least `GCP_CREDENTIALS` and `GCP_METRICS_PROJECT` set to push metrics to StackDriver. Other `GCP_METRICS_` vars define the kubernetes container to monitor them from. If they are not set, it will fallback to GCE, and then to global.
