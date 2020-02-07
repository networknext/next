# Server Backend

The Server Backend is responsible for:

1. Listening for new game servers that come online to allow client sessions for
2. Pulls the Route Matrix from the Relay Backend so each running instance of the Server Backend has a local/fast copy of routes
3. Accepts client sessions from the game server to find relays near the client, find a route, and send it back to the game server
4. Records metrics and billing information so that the routes being served are billed accordingly

### To Run

Run `make dev-server-backend`

### Environment Variables

#### Required

- `ROUTE_MATRIX_URI`: a URL or local file path to a route matrix binary
- `MAXMIND_DB_URI`: local path to a `.mmdb` file for IP lookups

#### Optional

- `REDIS_HOST`: address of the Redis server you want to connect to, uses the in-memory version if not supplied or invalid
- `CONFIGSTORE_HOST`: address to configstore, uses the in-memory version if not supplied