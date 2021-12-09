# Relay Backend

The Relay Backend is responsible for:

1. Receiving routable relays from the Relay Gateway
2. Accepting stat updates (RTT, jitter, packet loss) from all relays and store their stats
3. Generating a cost then route matrix of optimized routes that are available via the `/cost_matrix` and `/route_matrix` endpoints
4. Storing the timestamp of the last calculated cost and route matrix in Redis

### Load & Scalability

Status: **LOW-MEDIUM**
Scalability: **Vertically**

1. Relays send HTTP requests every 1 second, which are batched by the Relay Gateway and sent to all Relay Backends when the batch threshold is met
2. Computes the route matrix of all available routes every 1 second

### To Run

Run `make dev-relay-backend-1` and/or`make dev-relay-backend-2`

### Logging

Debug logs are available by setting `NEXT_DEBUG_LOGS=1`. Otherwise, only error logs will be outputed.

### Environment Variables

#### Required to Run

- `ENV`: the environment the service is running in. Should be either local, dev, staging, or prod
- `RELAY_ROUTER_MAX_JITTER`: the max jitter value to consider a relay unroutable within the lookback period
- `RELAY_ROUTER_MAX_PACKET_LOSS`: the max packet loss value to consider a relay unroutable within the lookback period
- `RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE`: the bandwidth limit for a relay as a function of the relay's NIC Speed to determine if the relay is at max capacity
- `RELAY_BACKEND_ADDRESSES`: the list of valid relay backend addresses
- `MATRIX_STORE_ADDRESS`: the redis database address to store the relay backend live metadata

#### Required for Cloud Environment

- `GOOGLE_PROJECT_ID`: The GCP project ID for the environment the service is being run in

#### Optional
- `PORT`: The port to listen for HTTP requests. Default `30001`
- `ENABLE_STACKDRIVER_LOGGING`: Whether or not to enable StackDriver Logging. Default `false`
- `ENABLE_STACKDRIVER_PROFILER`: Whether or not to enable StackDriver Profiler. Default `false`
- `ENABLE_STACKDRIVER_METRICS`: Whether or not to enable StackDriver Metrics. Default `false`
- `GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL`: The frequency at which to write metrics to StackDriver. Default `1m`
- `COST_MATRIX_INTERVAL`: The frequency at which to generate the cost and route matrices. Default `1s`
- `MATRIX_BUFFER_SIZE`: The maximum size of the cost or route matrix buffer during serialization. Default `100000`
- `FEATURE_MATRIX_CLOUDSTORE`: The feature flag for uploading the route matrix to Google Cloud Storage. Default `false`
- `FEATURE_ENABLE_PPROF`: The feature flag for enabling the pprof http endpoint. Default `false`
- `BIN_SYNC_INTERVAL`: The frequency to sync the database file from disk. Default `1m`
- `BIN_PATH`: The full file path for the location of the database file on disk. Default `./database.bin`
- `MATRIX_STORE_PASSWORD`: The password for the redis database address. Default `""`
- `MATRIX_STORE_MAX_IDLE_CONNS`: The max number of idle redis pool connections to the redis database. Default `5`
- `MATRIX_STORE_MAX_ACTIVE_CONNS`: The max number of active redis pool connections to the redis database. Default `5`
- `MATRIX_STORE_READ_TIMEOUT`: The max time to read from the redis database. Default `250ms`
- `MATRIX_STORE_WRITE_TIMEOUT`: The max time to write to the redis database. Default `250ms`
- `MATRIX_STORE_EXPIRE_TIMEOUT`: The max time for a key to stay alive in the redis database before expiring. Default `5s`

### Endpoints
- `/health`: After starting without failure, this handler responds with a 200 OK.
- `/version`: Provides the build time and commit hash of the relay backend.
- `/database_version`: Provides the creator, creation time, and env of the current database file.
- `/relay_update`: The endpoint for the relay gateway to send batched relay update requests to.
- `/route_matrix`: The endpoint for accessing the latest route matrix.
- `/relay_dashboard_data`: Provides stats between relays in an easy to digest format for the Admin tool.
- `/relay_dashboard_analysis`: Provides a JSON analysis of the relay dashboard data for the Admin tool.
- `/status`: Provides the latest metrics relevant to the relay backend's operations.
- `/dest_relays`: Provides a list of destination relays.
- `/debug/vars`: The endpoint for viewing all metrics during local testing.
- `/relays`: Provides a list of all relays in a CSV format.
- `/cost_matrix`: The endpoint for accessing the latest cost matrix. 
