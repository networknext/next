![Network Next](https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg)

# Network Next Operator Tool

Run `./next` in the root of the repo for available commands.

## SSH

SSH into a remote device. You must set the SSH key before attempting to connect to a device, otherwise you will get denied.

To SSH: `next ssh [identifier]`

To set the SSH key: `next ssh key [path to key file]`
- You can't use '~' in the path directly, it must be expanded by the shell first. Or in other words don't quote the argument

## Disable

First the tool will update the relay's state in Firestore to the Disabled state. Then it will SSH into a relay, stop the relay service, and end the session.

To Disable a relay: `next disable [relay name]`
