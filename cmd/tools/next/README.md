![Network Next](https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg)

# Network Next Operator Tool

Run `./next` in the root of the repo for available commands.

## SSH

To SSH: `next ssh [identifier]`

SSH into a remote device. You must set the SSH key before attempting to connect to a device, otherwise you will get denied.

To set the SSH key: `next ssh key [path to key file]`
- You can't use '~' in the path directly, it must be expanded by the shell first. Or in other words don't quote the argument

## Relay

### Enable

To Enable a relay: `next relay enable [relay name]`

The tool will SSH into the specified relay and start the relay service. If it is already running the tool will do nothing.

### Disable

To Disable a relay: `next relay disable [relay name]`

First the tool will update the relay's state in Firestore to the Disabled state. Then it will SSH into a relay, stop the relay service, and end the session. If the service is already stopped the tool will do nothing.

### Update

To Update a relay: `next relay update [relay name]...`

The tool will perform several actions to update a relay.

First you must build the relay binary and ensure it is located at `dist/relay`

Then disable the relay using the [disable](#Disable) functionality

Then you use the tool. You must have the desired environment set via the [relay env](#Env) setting. The tool will query the relay(s) from Firestore and then perform an update on those selected. If any fail for any reason the program will quit and not continue to the next. While updating it will re-generate public and private keys for the relay and set them both in the relay's environment file and publish the changes to Firestore as well.
