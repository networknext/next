
How it works
============

The SDK has two main components. *next_client_t* and *next_server_t*.

These components replace how you send and receive packets between your game client and server:

.. image:: images/client_server.png

Network Next works by monitoring all your players and only accelerating players when we find a significant reduction in latency or packet loss.

You enable monitoring for a player by *upgrading* that player's session on your server. 

Once a player is upgraded, the SDK does the following logic every 10 seconds:

1. Look for a Network Next route with a significant reduction in latency or packet loss vs. the public internet.
2. If we find one, steer that player's traffic across the network next route instead of the default internet route.
3. Otherwise, keep sending packets across the public internet, because it's already good enough.

Typically, we provide significant improvements for your player base, by accelerating just 10 or 20% of your players at any time. 

In trials, we've found that this improves the network performance of over 70% of your players at least once in the trial period.

This means it's not the same set of players getting accelerated all the time.

In other words, we we keep your costs down, while targeting your spend towards players only when they need it the most.

Network Next. Now *you* control the network!
