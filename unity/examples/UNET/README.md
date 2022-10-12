# Network Next UNET Integration

UNET is a deprecated multiplayer solution offered by Unity.

To avoid interfering with existing UNET communications, we made the design choice to only modify UNET's High Level API. To use Network Next in games that utilize UNET, we introduced the `NextClientTransport` and `NextServerTransport` classes, which are replacements for the default transport used by Unity's `NetworkManager`. These transports use the `next-unity` plugin to run a Network Next client and server on their own socket independent of UNET's socket, while also supporting all existing UNET functionality.

The UNET Client and Network Manager examples were tested on Unity Engine version 2018.4.36f1.
