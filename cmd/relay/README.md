# General Info

The Relay is responsible for:

1. Routing packets across other relays in the network to and from the game client and server
2. Sending stats updates to the relay_backend that are optimized for route generation

This is the reference relay implementation. This is what will be handed to other suppliers so they may develop optimized relay implementations. Not for production use. Should be a reliable way to determine what the relay is supposed to do.

# Pre-build setup

## Dependencies

`sudo apt install libcurl4 libcurl4-openssl-dev libsodium23 libsodum-dev`

- `cURL`: For HTTP communication.
- `libsodium`: Encryption/decryption/signature/etc... library.

## Environment Variables

### Required

- `RELAY_ADDRESS`: The address the other relays and sdk should talk to.
  - Example `RELAY_ADDRESS='127.0.0.1:1234'`
- `RELAY_PUBLIC_KEY`: The public key of the relay encoded in base64.
  - Example `RELAY_PUBLIC_KEY='9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14='`
- `RELAY_PRIVATE_KEY`: The corresponding private key encoded in base64.
  - Example `RELAY_PRIVATE_KEY='lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8='`
- `RELAY_ROUTER_PUBLIC_KEY`: The router's public key encoded in base64, used to encrypt data for relay verification.
  - Example `RELAY_ROUTER_PUBLIC_KEY='SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y='`
- `RELAY_BACKEND_HOSTNAME`: The backend's hostname
   - Example `RELAY_BACKEND_HOSTNAME='http://localhost:30000'`

# Building and running a single relay

From the project root `make dev-relay`. This will spawn a single relay instance. using the environment variables within the makefile.

# Building and running multiple relays

To supply an adequate amount of routes multiple relays must be launched. To do so with relative ease use the relay spawning script located in `tools/scripts/relay-spawner.sh`. Use the `-h` flag to show help for how to use it.
