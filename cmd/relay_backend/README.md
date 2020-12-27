# Relay Backend

The Relay Backend is responsible for:

1. Listening for new relays that come online and initialize them to be apart of set of available relays for routing
2. Accepting stat updates (RTT, jitter, packet loss) from all relays and store their stats
3. Generate a cost then route matrix of optimized routes that are available which the server backend pulls a copy of

### Load & Scalability

Status: **LOW-MEDIUM**
Scalability: **Vertically**

1. Relays send HTTP requests every 1 second
2. Computes the route matrix of all available routes every 1 second

### To Run

Run `make dev-relay-backend`

### Logging

There is a separation of backend logs and relay endpoint logs. Backend logs are controlled with `BACKEND_LOG_LEVEL`, and relay endpoints logs are controller with `RELAY_LOG_LEVEL`.

Levels are cumulative so if you set `BACKEND_LOG_LEVEL=info` you will get `error` and `warn` too.

The default setting is `warn`. To override this you can set your own value by doing `make BACKEND_LOG_LEVEL=debug dev-relay-backend`.

### Environment Variables

#### Required to Run

- `ENV`: the environment the service is running in. Should be either local, dev, staging, or prod
- `RELAY_ROUTER_PRIVATE_KEY`: the private key of the router used to encrypt routes

#### Required for All Functionality

#### Required for Cloud Environment

- `GOOGLE_PROJECT_ID`: The GCP project ID for the environment the service is being run in

#### Optional

- `BACKEND_LOG_LEVEL`: one of `none`, `error`, `warn`, `info`, `debug`. Default `warn`
- `RELAY_LOG_LEVEL`: one of `none`, `error`, `warn`, `info`, `debug`. Default `warn`

#### Optional for Cloud Environment

- `ENABLE_STACKDRIVER_LOGGING`: Whether or not to enable StackDriver Logging. Default `false`
- `ENABLE_STACKDRIVER_PROFILER`: Whether or not to enable StackDriver Profiler. Default `false`
- `ENABLE_STACKDRIVER_METRICS`: Whether or not to enable StackDriver Metrics. Default `false`
- `GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL`: The frequency at which to write metrics to StackDriver. Default `1m`
- `GOOGLE_FIRESTORE_SYNC_INTERVAL`: The frequency at which to check if the Firestore database needs to be synced. Default `10s`
- `COST_MATRIX_INTERVAL`: The frequency at which to generate the cost and route matrices. Default `1s`
- `PING_STATS_PUBLISH_INTERVAL`: The frequency at which to publish relay pings to Google PubSub. Default `1m`
- `RELAY_STATS_PUBLISH_INTERVAL`: The frequency at which to publish relay stats to Google PubSub. Default `10s`
- `MATRIX_BUFFER_SIZE`: The maximum size of the cost or route matrix buffer during serialization. Default `100000`
- `PORT`: The port to listen for HTTP requests. Default `30000`

### Endpoints

- `/relay_init`: When a relay starts up, the first thing it will do is call this endpoint with its information
- `/relay_update`: After a relay has confirmation of successful initialization, it will keep sending this endpoint stats about its network timings
  - Within the response body will be a list of all other relays to ping and gather stats on
- `/relay_stats`: Takes the backend's relay map and turns it into a binary format for the portal to unmarshal so we can query relay stats using the operator tool
