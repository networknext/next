
How it works
============

The SDK has two main components. *next_client_t* and *next_server_t*.

These components replace how you send and receive UDP packets between your game client and server:

.. image:: images/client_server.png

Network Next works by monitoring all your players and only accelerating players when find a significant reduction in latency or packet loss.

You enable monitoring for a player by *upgrading* that player's session on your server. 

Once a player is upgraded, the SDK does the following logic every 10 seconds:

1. Looks for a Network Next route with lower latency or packet loss than the public internet
2. If we find one, steer that player's traffic across the network next route instead of the default internet route.
3. Otherwise, keep sending packets across the public internet, because it's already good enough.

Typically, we can provide siginificant improvements for your player base, by accelerating just the 10 or 20% of your players having the worst experiences at any time. 

This way we target your spend effectively towards the players that need it the most.
