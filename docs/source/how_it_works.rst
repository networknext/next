
How it works
============

The SDK has two main components. *next_client_t* and *next_server_t*.

These components integrate with your game and replace how you send and receive packets between your game client and server:

.. image:: images/client_server.png

You enable monitoring and acceleration for a player by *upgrading* that player's session on your server.

One monitoring is enabled for a player, the SDK does the following logic every 10 seconds:

1. Look for a Network Next route with lower latency or packet loss
2. If we find one, steer that player's traffic across the network next route instead of the default path across the internet.
3. If the public internet is already good enough, or we can't find a route that's significantly better, send packets across the public internet.
4. If for any reason, the SDK cannot communicate with the Network Next backend, send packets across the public internet.
