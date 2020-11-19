
How it works
============

The SDK has two main components. *next_client_t* and *next_server_t*.

These components replace how you send and receive UDP packets between your game client and server:

.. image:: images/client_server.png

Network Next works by monitoring all your players and only accelerating players when find a significant reduction in latency or packet loss.

You enable monitoring for a player by *upgrading* that player's session on your server. 

Not all upgraded players are accelerated, but players that aren't upgraded are *never* accelerated.

Once a player is upgraded, monitoring is enabled for that player and the SDK does the following logic every 10 seconds:

1. Looks for a Network Next route with lower latency or packet loss than the public internet
2. If we find one, steer that player's traffic across the network next route instead of the default internet route.
3. Otherwise, keep sending packets across the public internet, because it's already good enough.

This way we target your spend towards only the players that need it the most.
