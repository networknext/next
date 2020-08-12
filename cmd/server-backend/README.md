# Server Backend

The Server Backend is responsible for:

1. Listening for new game servers that come online to allow client sessions for
2. Pulls the Route Matrix from the Relay Backend so each running instance of the Server Backend has a local/fast copy of routes
3. Accepts client sessions from the game server to find relays near the client, find a route, and send it back to the game server
4. Records metrics and billing information so that the routes being served are billed accordingly

### Load & Scalability

Status: **HIGH**
Scalability: **Horizontally, Load-Balanced**

1. Game clients and game servers send UDP packets every 10 seconds
2. The only network call made during these requests are to Redis, everything else is done in memory or in the background across the network to build a local cache

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
- `REDIS_HOST_TOP_SESSIONS`: redis instance to save top sessions
- `REDIS_HOST_SESSION_MAP`: redis instance to save map points
- `REDIS_HOST_SESSION_META`: redis instance to save session metadata
- `REDIS_HOST_SESSION_SLICES`: redis instance to save session slices

#### Optional

- `GOOGLE_APPLICATION_CREDENTIALS`: Path to a .json file for the GCP credentials needed
- `GOOGLE_PROJECT_ID`: The Google project ID
- `GOOGLE_PUBSUB_TOPIC_BILLING`: The topic ID to use for billing in Google Pub/Sub
- `NUM_UDP_SOCKETS`: Number of udp sockets to create for packet receiving. All on the same port using the SO_REUSEPORT socket opt, defaults to 8
- `USE_THREAD_POOL`: Whether the server backend should use a thread pool over an unrestricted number of goroutines for processing packets. If true, a single pool will be created per udp socket
- `NUM_PACKET_PROCESSING_THREADS`: The number of threads to assign to the packet processing thread pool, defaults to 256
- `PACKET_SELECTION_PERCENT`: The throttle for how many packets to process. For example, setting to 60 would mean we accept only 60% of the incoming udp packets

#### IMPORTANT

Both `GOOGLE_PROJECT_ID` and `GOOGLE_PUBSUB_TOPIC_BILLING` must be set to send billing entries to Google Pub/Sub.
