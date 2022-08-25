
How it works
============

The SDK has two main components. *next_client_t* and *next_server_t*.

These components replace how you send and receive packets between your game client and server:

.. image:: images/client_server.png

Network Next monitors all your players, but only accelerates players when we find a significant reduction in latency or packet loss.

You enable monitoring for a player by *upgrading* that player's session on your server. 

Once a player is upgraded, the SDK does the following logic every 10 seconds:

1. Look for a Network Next route with a significant reduction in latency or packet loss vs. the public internet.
2. If we find one, steer that player's traffic across the network next route instead of the default internet route.
3. Otherwise, keep sending packets across the public internet, because it's already good enough.

Typically, we provide significant improvements for your player base by accelerating just 5 or 10% of your players at any time. In trials, we've found that this improves the network performance of more than half of players each month :)
