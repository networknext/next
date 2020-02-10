# Relay

The Relay is responsible for:

1. Routing packets across other relays in the network to and from the game client and server
2. Sending stats updates to the relay_backend that are optimized for route generation

### To Run

Run `make dev-server-backend`

### Environment Variables

#### Required

- `RELAY_ADDRESS`: The address of this relay
- `RELAY_BACKEND_HOSTNAME`: The address of the relay backend
- `RELAY_PRIVATE_KEY`: The private key of each relay.
- `RELAY_PUBLIC_KEY`: The public key of each relay generated with the private key.
- `RELAY_ROUTER_PUBLIC_KEY`: The public key of the router.