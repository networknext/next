# General Info

The Relay is responsible for:

1. Routing packets across other relays in the network to and from the game client and server
2. Sending stats updates to the relay_backend that are optimized for route generation

This is the current production relay. It is lacking in performance and will eventually be replaced with the relay within the "relay_new" directory. This code was copied from the "next" repo. The only difference from the next repo's version is that this relay can talk to the current backend via http and is not restricted to udp.

# Pre-build setup

## Dependencies

`sudo apt install libcurl4 libcurl4-openssl-dev libsodium23 libsodum-dev`

- `cURL`: For HTTP communication.
- `libsodium`: Encryption/decryption/signature/etc... library.
- `rapidjson`\*: A fast json parsing library.
- `sparsehash`\*: A hashmap library from Google.
- `concurrentqueue`\*: A threadsafe queue implementation.
- `miniz`\*: Lightweight zlib c implementation.

\* Already within the repo under the "deps" directory.

## Environment Variables

### Required

- `RELAYADDRESS`: The address the other relays and sdk should talk to.
  - Example `RELAYADDRESS='127.0.0.1:1234'`
- `RELAYPUBLICKEY`: The public key of the relay encoded in base64.
  - Example `RELAYPUBLICKEY='9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14='`
- `RELAYPRIVATEKEY`: The corresponding private key encoded in base64.
  - Example `RELAYPRIVATEKEY='lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8='`
- `RELAYROUTERPUBLICKEY`: The router's public key encoded in base64, used to encrypt data for relay verification.
  - Example `RELAYROUTERPUBLICKEY='SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y='`
- `RELAYBACKENDHOSTNAME`: The backend's hostname
   - Example `RELAYBACKENDHOSTNAME='http://localhost:30000'`
- `RELAYNAME`: The name of the relay. Should be the firestore id.
- `RELAYMASTER`: Only the hostname of the backend
  - Example `RELAYMASTER=relays.v3-dev.networknext.com`
- `RELAYUPDATEKEY`: A signing key encoded in base64
- `RELAYPORT`: The port the relay receives data on
- `RELAYSPEED`: One of '100M', '1G', or '10G'. Used to specify the speed of the relay's nic.

### Optional

- `RELAYDEV`: Enable the relay's dev mode. Value should be 0 or 1. 0 == disabled, 1 == enabled.

# Building and running a single relay at a time

To use build.sh,

First, \
`build.sh build`

Then you can run up to two relays using the script. To run the first relay, \
`build.sh run one`

Or to run the second \
`build.sh run two`

You can add a third, fourth, ..., Nth, if you'd like, it was done like this just to get it working.
