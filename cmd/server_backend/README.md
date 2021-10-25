# Server Backend

The Server Backend is responsible for:

1. Listening for new game servers that come online
2. Pulls the route matrix from the relay frontend
3. Accepts client sessions from the game server to find relays near the client, find a route, and send it back to the game server
4. Send route information to the portal cruncher so that the sessions can be viewed in the portal
5. Records metrics and billing information so that the routes being served are billed accordingly

### Load & Scalability

Status: **HIGH**
Scalability: **Horizontally, Load-Balanced**

1. Game clients and game servers send UDP packets every 10 seconds
2. The route matrix and associated database information is acquired from the Relay Frontend every second
2. The only outgoing network calls made during these requests are to the portal cruncher to send portal data and to Google Pub/Sub to send billing information, everything else is done in memory or in the background across the network to build a local cache

### To Run

Run `redis-server` in one terminal window
Run `make dev-server-backend` in another terminal window

### Logging

Debug logs are available by setting `NEXT_DEBUG_LOGS=1`. Otherwise, only error logs will be outputed.

### Environment Variables

#### Required to Run

- `ENV`: the environment the service is running in. Should be either local, dev, staging, or prod
- `SERVER_BACKEND_PRIVATE_KEY`: the private key of the server_backend to sign responses
- `RELAY_ROUTER_PRIVATE_KEY`: the private key of the router used to encrypt routes
- `MAXMIND_CITY_DB_FILE`: the path to where the MaxMind IP2Location City database file is located 
- `MAXMIND_ISP_DB_FILE`: the path to where the MaxMind IP2Location ISP database file is located 
- `ROUTE_MATRIX_URI`: the URI of the route matrix

#### Required for All Functionality

- `FEATURE_BILLING2`: the feature flag for enabling billing2 entries to be written and submitted
- `PUBSUB_EMULATOR_HOST`: an IP and port to connect to a running pubsub emulator
- `FEATURE_VANITY_METRIC`: the feature flag for enabling sending vanity metrics


#### Required for Cloud Environment

- `GOOGLE_PROJECT_ID`: The GCP project ID for the environment the service is being run in
- `PORTAL_CRUNCHER_HOSTS`: a TCP hostname for the portal cruncher VM. Make sure this is an internal IP address
- `FEATURE_VANITY_METRIC_HOSTS`: a TCP hostname for the vanity metric VM. Make sure this is an internal IP address
- `REDIS_HOST_MULTIPATH_VETO`: a redis IP and port to connect to store multipath veto data

#### Optional

- `MAX_NEAR_RELAYS`: The maximum number of near relays to send down to a client. Cannot be greater than `32`. Default `32`
- `ENABLE_STACKDRIVER_LOGGING`: Whether or not to enable StackDriver Logging. Default `false`
- `ENABLE_STACKDRIVER_PROFILER`: Whether or not to enable StackDriver Profiler. Default `false`
- `ENABLE_STACKDRIVER_METRICS`: Whether or not to enable StackDriver Metrics. Default `false`
- `GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL`: The frequency at which to write metrics to StackDriver. Default `1m`
- `MAXMIND_SYNC_DB_INTERVAL`: The frequency at which to sync the Maxmind databases. Default `1m`
- `ROUTE_MATRIX_SYNC_INTERVAL`: The frequency at which to pull the route matrix from the relay backend. Default `1s`
- `BILLING_CLIENT_COUNT`: The number of billing routines to run concurrently. Default `1`
- `BILLING_BATCHED_MESSAGE_COUNT`: The number of messages to batch before publishing to PubSub. Default `100`
- `BILLING_BATCHED_MESSAGE_MIN_BYTES`: The minimum byte threshold for publishing to PubSub. Default `1024`
- `FEATURE_BILLING2_TOPIC_NAME`: The topic name for the Google Pub/Sub billing. Default `billing2`
- `POST_SESSION_THREAD_COUNT`: The number of goroutines to allocate for post session handling. Default `1000`
- `POST_SESSION_BUFFER_SIZE`: The size of the post session buffer in bytes. Default `1000000`
- `POST_SESSION_PORTAL_MAX_RETRIES`: The number of times to retry on a post session portal update before dropping the portal data. Default `10`
- `POST_SESSION_PORTAL_SEND_BUFFER_SIZE`: The size of the ZeroMQ buffer for post session data. Default `1000000`
- `REDIS_HOST_MULTIPATH_VETO`: The hostname for the redis multipath veto database. Default `""`
- `REDIS_PASSWORD_MULTIPATH_VETO`: The password for the redis multipath veto database. Default `""`
- `REDIS_MAX_IDLE_CONNS_MULTIPATH_VETO`: The max number of idle redis pool connections. Default `5`
- `REDIS_MAX_ACTIVE_CONNS_MULTIPATH_VETO`: The max number of active redis pool connections. Default `64`
- `MULTIPATH_VETO_SYNC_FREQUENCY`: The frequency at which to sync multipath veto data from redis. Default `10s`
- `NUM_THREADS`: The number of goroutines to allocate for UDP packet processing. Default `1`
- `READ_BUFFER`: The size of the read buffer for a socket connection. Default `100000`
- `WRITE_BUFFER`: The size of the write buffer for a socket connection. Default `100000`
- `UDP_PORT`: The port to listen for UDP packets on from the game server. Default `40000`
- `HTTP_PORT`: The port to listen for HTTP requests. Default `40001`
- `FEATURE_VANITY_METRIC_POST_SEND_BUFFER_SIZE`: The size of the ZeroMQ buffer for vanity metrics. Default `1000000`
- `FEATURE_VANITY_METRIC_POST_MAX_RETRIES`: The number of times to retry on a post session vanity update before dropping the vanity data. Default `10`
- `FEATURE_ENABLE_PPROF`: The feature flag for enabling the pprof http endpoint. Default `false`
- `MATRIX_STALE_DURATION`: The amount of time before a route matrix is considered stale. Default `20s`
- `METADATA_SYNC_INTERVAL`: The frequency at which to sync the metadata value for shutting down the HTTP server for connection drain. Default `1m`
- `CONNECTION_DRAIN_METADATA_FIELD`: The key for the metadata value for connection drain. Default `connection-drain`
- `JWT_AUDIENCE`: The JWT audience necessary for accessing private HTTP endpoints. Default `""`
- `AUTH0_CERT_INTERVAL`: The frequency at which to refresh the Auth0 certifiacte. Default `10m`
