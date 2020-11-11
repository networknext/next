
How it works
============

The SDK has two main components. *next_client_t* and *next_server_t*.

These components integrate with your game and replace how you send and receive packets between your game client and server:

.. image:: images/client_server.png

You enable monitoring and acceleration for a player by *upgrading* that player's session on your server.

Once a session is upgraded, the SDK looks for a Network Next route with lower latency or packet loss every 10 seconds.

If we find one, we steer that player's traffic across the network next route instead of the default path across the internet.

If the public internet is already good enough, or we can't find a route that's significantly better, the SDK keeps sending packets across the public internet.

If for any reason, the SDK cannot communicate with the Network Next backend, it falls back to sending packets across the public internet.
