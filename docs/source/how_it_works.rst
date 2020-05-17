
How it works
============

The SDK has two main components. A client and a server.

These components are integrated in your game and replace how you send and receive packets between game client and server.

(diagram)

Network Next has its own network protocol that runs in parallel with you game network protocol.

(diagram)

On the server, you *enable* a player for monitoring and acceleration by *upgrading* that player's session.

(diagram)

The server component of our SDK then communicates with the Network Next backend, and looks for a Network Next route on our marketplace for that player, once every 10 seconds.

(diagram)

When we find a network next route that has lower packet loss or latency than the public internet route for that player, we steer your game traffic between the client and server away from the public internet, and across the network next route on our network of private networks.

(diagram)

Players for whom the public internet is already good enough, or when we can't find Network Next route that's significantly better, continue to send packets across the public internet.

(diagram)

If for any reason, the SDK cannot communicate with the Network Next backend, it falls back seamlessly and sends packets across the public internet.

(diagram).

This way no matter what happens, your players can always play your game.
