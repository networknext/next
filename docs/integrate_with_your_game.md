<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Integrate with your game

This section applies to games that use UDP sockets directly in C or C++. 

If your game is using Unreal Engine, please skip ahead to [Unreal Engine Plugin](unreal_engine_plugin.md).

## 1. Replace UDP socket on client with next_client_t

Create an instance of the next_client_t object on the client and connect to your server.

To upgrade and accelerate players, the client requires the following:

* A valid buyer public key, which is safe to embed in your executable.
* The server backend public key (safe to embed)
* The relay backend public key (safe to embed)
* The server backend hostname.

By default, #if NEXT_DEVELOPMENT, your copy of the SDK has already been configured to point to your dev environment at "server-dev.[yourdomain.com]" for the server backend hostname and has the correct server and relay backend public keys embedded for your dev environment.

If NEXT_DEVELOPMENT is zero or undefined, then by default the SDK points to your production environment at "server.[yourdomain.com]" and has the correct server backend and relay backend public keys for your production environment.

Generally, on the client all that you must do is make sure that NEXT_DEVELOPMENT is defined appropriately when you build the SDK, and then embed your buyer public key. The SDK takes care of the rest!

You can override these defaults using environment variables, or by passing them in via next_config_t to the next_init function on the client, before you create the next_client_t.

For more details, please see the SDK reference here: https://network-next-sdk.readthedocs-hosted.com/en/latest/reference.html

## 2. Replace UDP socket on server with next_server_t

On the server side, your UDP socket is replaced with next_server_t.

To upgrade and accelerate players, the server needs:

* The datacenter the server is located in, passed in in code when the server is created, or via the environment variable (NEXT_DATACENTER). You can set this datacenter to "cloud" when running in Google Cloud or AWS and the SDK will autodetect the datacenter for you. When running in local playtests or by default during development, pass in "local" as the datacenter name.
* The server IP address, passed in in code *or* overridden by environment variable (NEXT_SERVER_ADDRESS).
* A valid buyer private key, which should be passed in via environment variable (NEXT_BUYER_PRIVATE_KEY), or passed in on the command line. It is best not to embed this in your source code. This private key must be kept top secret.
* The server backend public key (safe to embed)
* The relay backend public key (safe to embed)
* The server backend hostname.

As with the client, the SDK is by default setup to point at your dev environment #if NEXT_DEVELOPMENT, else it will point at your production environment.

## 3. Verify that your client sessions show up in the portal

When you have completed integration, your clients should upgrade and show up in the portal under "Sessions" while they are connected to a server. Your servers should show up in the "Server" page while they are running.

The SDK is designed to fall back to unaccelerated behavior when anything goes wrong or is misconfigured. By making sure that you see both your servers and sessions in the portal, even if they are not accelerated, then you know that your integration is working correctly.
