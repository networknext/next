![Network Next](https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg)

# Network Next Operator Tool

Run `./next` in the root of the repo for available commands.

## SSH

SSH into a remote device.

Usage: `next ssh [identifier]`

### With Relays

To SSH into a relay you must have the public key that grants you access.\
Once you do you must set the `RELAY_SERVER_KEY` environment varialbe to point to it.
