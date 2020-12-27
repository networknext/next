# Server Backend

The Server Backend is responsible for:

1. Listening for new game servers that come online
2. Pulls the route matrix from the relay backend
3. Accepts client sessions from the game server to find relays near the client, find a route, and send it back to the game server
4. Send route information to the portal cruncher so that the sessions can be viewed in the portal
5. Records metrics and billing information so that the routes being served are billed accordingly

### Load & Scalability

Status: **HIGH**
Scalability: **Horizontally, Load-Balanced**

1. Game clients and game servers send UDP packets every 10 seconds
2. The only network calls made during these requests are to the portal cruncher to send portal data and to Google PubSub to send billing information, everything else is done in memory or in the background across the network to build a local cache

### To Run

Run `redis-server` in one terminal window
Run `make dev-server-backend` in another terminal window

### Logging

Levels are cumulative so if you set `BACKEND_LOG_LEVEL=info` you will get `error` and `warn` too.

The default setting is `warn`. To override this you can set your own value by doing `make BACKEND_LOG_LEVEL=debug dev-server-backend`.

### Environment Variables

#### Required to Run

- `ENV`: the environment the service is running in. Should be either local, dev, staging, or prod
- `SERVER_BACKEND_PRIVATE_KEY`: the private key of the server_backend to sign responses
- `RELAY_ROUTER_PRIVATE_KEY`: the private key of the router used to encrypt routes

#### Required for All Functionality

- `ROUTE_MATRIX_URI`: a URL or local file path to a route matrix binary
- `MAXMIND_CITY_DB_URI`: a URL to the Maxmind City DB
- `MAXMIND_ISP_DB_URI`: a URL to the Maxmind ISP DB
- `FIRESTORE_EMULATOR_HOST`: an IP and port to connect to a running firestore emulator 
- `PUBSUB_EMULATOR_HOST`: an IP and port to connect to a running pubsub emulator 

#### Required for Cloud Environment

- `GOOGLE_PROJECT_ID`: The GCP project ID for the environment the service is being run in
- `PORTAL_CRUNCHER_HOST`: a TCP hostname for the portal cruncher. Make sure this is an internal IP address
- `REDIS_HOST_MULTIPATH_VETO`: a redis IP and port to connect to store multipath veto data

#### Optional

- `BACKEND_LOG_LEVEL`: one of `none`, `error`, `warn`, `info`, `debug`. Default `warn`
- `MAX_NEAR_RELAYS`: The maximum number of near relays to send down to a client. Cannot be greater than `32`. Default `32`

#### Optional for Cloud Environment

- `ENABLE_STACKDRIVER_LOGGING`: Whether or not to enable StackDriver Logging. Default `false`
- `ENABLE_STACKDRIVER_PROFILER`: Whether or not to enable StackDriver Profiler. Default `false`
- `ENABLE_STACKDRIVER_METRICS`: Whether or not to enable StackDriver Metrics. Default `false`
- `GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL`: The frequency at which to write metrics to StackDriver. Default `1m`
- `GOOGLE_FIRESTORE_SYNC_INTERVAL`: The frequency at which to check if the Firestore database needs to be synced. Default `10s`
- `MAXMIND_SYNC_DB_INTERVAL`: The frequency at which to sync the Maxmind databases. Default `24h`
- `ROUTE_MATRIX_SYNC_INTERVAL`: The frequency at which to pull the route matrix from the relay backend. Default `1s`
- `BILLING_CLIENT_COUNT`: The number of billing routines to run concurrently. Default `1`
- `BILLING_BATCHED_MESSAGE_COUNT`: The number of messages to batch before publishing to PubSub. Default `100`
- `BILLING_BATCHED_MESSAGE_MIN_BYTES`: The minimum byte threshold for publishing to PubSub. Default `1024`
- `POST_SESSION_THREAD_COUNT`: The number of goroutines to allocate for post session handling. Default `1000`
- `POST_SESSION_BUFFER_SIZE`: The size of the post session buffer in bytes. Default `1000000`
- `POST_SESSION_PORTAL_MAX_RETRIES`: The number of times to retry on a post session portal update before dropping the portal data. Default `10`
- `POST_SESSION_PORTAL_SEND_BUFFER_SIZE`: The size of the ZeroMQ buffer for post session data. Default `1000000`
- `MULTIPATH_VETO_SYNC_FREQUENCY`: The frequency at which to sync multipath veto data from redis. Default `10s`
- `NUM_THREADS`: The number of goroutines to allocate for UDP packet processing. Default `1`
- `READ_BUFFER`: The size of the read buffer for a socket connection. Default `100000`
- `WRITE_BUFFER`: The size of the write buffer for a socket connection. Default `100000`
- `UDP_PORT`: The port to listen for UDP packets on from the game server. Default `40000`
- `HTTP_PORT`: The port to listen for HTTP requests. Default `40001`
