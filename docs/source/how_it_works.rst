
How it works
============

The SDK has two main components. A client and a server.

These components are integrated in your game and replace how you send and receive packets:

(diagram)

On the server, you *enable* a player for monitoring and acceleration by *upgrading* that player's session.

(diagram)

Network Next has its own network protocol that runs in parallel with your game network protocol. 

This protocol measures network performance between the client and server and communicates this up to the network next backend.

(diagram)

The backend looks for a Network Next route for upgraded sessions on our marketplace once every 10 seconds.

(diagram)

When we find a route that has lower packet loss or latency than the public internet, we steer traffic for that session away from the public internet, and across the network next route on our network of private networks.

(diagram)

Players for whom the public internet is already good enough, or when we can't find a route that's significantly, continue to send packets across the public internet.

(diagram)

If for any reason, the SDK cannot communicate with the Network Next backend, it falls back seamlessly and just sends packets across the public internet.

(diagram).

This way no matter what happens, your players can always play your game.
