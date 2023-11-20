<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Native C++ SDK

## Introduction

The Network Next SDK integrates with your game's client and server, and takes over sending and receiving UDP packets on behalf of your game. 

It supports the following platforms:

1. Windows
2. Mac
3. Linux
4. iOS
5. PS4
6. PS5
7. XBox One
8. XBox Series X
9. Nintendo Switch

The SDK lives under the "sdk" directory of the main repo, and as part of the `next keygen` and `next config` steps the source code has been customized with your domain names and keypairs already.

When you run SDK examples, they will attempt to connect by default with your development backend instance. Make sure dev is up when you run the SDK examples. 

To connect to prod instances, simply make sure that NEXT_DEVELOPMENT is not defined in your source code, and the SDK code will automatically point at your production backends. For security reasons, no private keys have been embedded in your source code, so you may need to ensure that NEXT_BUYER_PRIVATE_KEY is set on the command line ahead of time when you run the examples. You can find your per-environment private keys under the ~/secrets directory.

To get started, view the documentation for instructions on how to build the SDK, then view the example programs: [Network Next SDK Documentation](https://network-next-sdk.readthedocs-hosted.com/en/latest/index.html)

## Datacenters

....

## Autodetection support

...
