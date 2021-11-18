# Portal Backend

The Portal Backend is responsible for:

1. Providing a single RPC API for data management (Relays, Datacenters, Buyers, etc.)
2. Serve the static front end JavaScript Portal UI

### Load & Scalability

Status: **HIGH**
Scalability: **Vertically**

1. Serves the necessary information to the Portal Frontend (i.e. top sessions, map points, etc.)
2. Fulfills requests from the Next and Admin tools for management of relays, datacenters, etc.

### To Run

Run `redis-server` in one terminal window
Run `make dev-portal`

### Logging

Be default, only error logs will be outputted to console.   

Debug logs are controlled by the `NEXT_DEBUG_LOGS` environment variable.

### Environment Variables

#### Required
- `ENV`: the environment the service is running in. Should be either local, dev, staging, or prod
- `LOOKER_SECRET`: the secret for accessing the Looker API
- `LOOKER_HOST`: the hostname for the Looker API
- `GITHUB_ACCESS_TOKEN`: the token to access Github API via OAuth 2.0 to read relase notes
- `SLACK_WEBHOOK_URL`: the Slack webhook URL to send messages notifications to Slack channels
- `SLACK_CHANNEL`: the Slack channel to publish messages to
- `AUTH0_ISSUER`: the Auth0 issuer for the Auth0 client
- `AUTH0_DOMAIN`: the Auth0 domain for the Auth0 client
- `AUTH0_CLIENTID`: the Auth0 client ID for the Auth0 client
- `LOOKER_API_CLIENT_ID`: the client ID for the Looker API
- `LOOKER_API_CLIENT_SECRET`: the client secret for the Looker API
- `PORT`: the port to run the HTTP server on
- `ANALYTICS_MIG`: the name of the Analytics MIG to get the service's status
- `ANALYTICS_PUSHER_URI`: the URI of the Analytics Pusher VM to get the service's status
- `API_URI`: the URI of the API load balancer to get the service's status
- `BILLING_MIG`: the name of the Billing MIG to get the service's status
- `PORTAL_BACKEND_MIG`: the name of the Portal Backend MIG to get the service's status
- `PORTAL_CRUNCHER_URI`: the URI of a single Portal Cruncher VM to get the service's status
- `RELAY_FRONTEND_URI`: the URI of the Relay Frontend load balancer to get the service's status
- `RELAY_GATEWAY_URI`: the URI of the Relay Gateway load balancer to get the service's status
- `RELAY_PUSHER_URI`: the URI of the Relay Pusher VM to get the service's status
- `SERVER_BACKEND_MIG`: the name of the Server Backend MIG to get the service's status
- `VANITY_URI`: the URI of a single Vanity Metric VM to get the service's status
- `MONDAY_API_KEY`: the API key to access information from Monday

#### Optional
- `GOOGLE_PROJECT_ID`: the GCP project ID for the environment the service is being run in. Default `""`
- `REDIS_HOSTNAME`: the hostname for the redis database. Default `127.0.0.1:6379`
- `REDIS_PASSWORD`: the password to connect to the redis database. Default `""`
- `REDIS_MAX_IDLE_CONNS`: the max number of idle redis pool connections. Default `5`
- `REDIS_MAX_ACTIVE_CONNS`: the max number of active redis pool connections. Default `64`
- `FEATURE_BIGTABLE`: the feature flag to viewing historical user sessions through the User Tool). Default `false`
- `BIGTABLE_EMULATOR_HOST`: the hostname for the Bigtable emulator. Should only be set for local development.
- `BIGTABLE_INSTANCE_ID`: the instance ID of the Bigtable instance for the current project. Default `localhost:8086`
- `BIGTABLE_TABLE_NAME`: the table name for the historical sessions table in Bigtable. Default `""`
- `BIGTABLE_CF_NAME`: the column family name for the Bigtable table. Default `""`
- `ENABLE_STACKDRIVER_PROFILER`: whether or not to enable StackDriver Profiler. Default `false`
- `ENABLE_STACKDRIVER_METRICS`: whether or not to enable StackDriver Metrics. Default `false`
- `GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL`: the frequency at which to write metrics to StackDriver. Default `1m`
- `HUBSPOT_API_KEY`: the Hubspot API Key to send create entries in Hubspot. Default `""`
- `AUTH0_CLIENTSECRET`: the Auth0 client secret for the Auth0 client. Default `""`
- `AUTH0_CERT_INTERVAL`: the frequency to refresh the Auth0 certificate. Default `10m`
- `SESSION_MAP_INTERVAL`: the frequency to generate map points per buyer. Default `1s`
- `RELEASE_NOTES_INTERVAL`: the frequency to fetch release notes from Github. Default `30s`
- `PINGDOM_URI`: the URI of the Pingdom VM to get the service's status. Default `""`
- `RELAY_FORWARDER_URI`: the URI of the Relay Forwarder VM to get the service's status. Default `""`
- `ALLOWED_ORIGINS`: the domains allowed for middleware CORS. Default `""`
- `HTTP_TIMEOUT`: the timeout for all HTTP requests. Default `40s`
- `FEATURE_ENABLE_PPROF`: the feature flag to enable the pprof http endpoint. Default `false`
