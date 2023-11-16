<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# How does Network Next work?

Network Next comes in three parts:

1. An SDK (or UE5 plugin) that integrates with your game client and server
2. A backend that runs in Google Cloud that performs route optimization
3. A fleet of relays (software routers) that game traffic is sent across when we accelerate a player

## 1. SDK

The SDK operates by taking over UDP packet send and receive for your game client and server. This way when we detect that a player has high latency or packet loss, we can fix it by steering UDP packets through the relay fleet instead sending them directly from the client to the server IP address and vice versa. This is how Network Next is able to _undo_ bad Internet routing decisions and *force* your player's traffic to go through the lowest latency and packet loss route.

In addition to steering player traffic, the SDK pings nearby relays at the start of each match (according to ip2location for each player), to find the lowest latency, jitter and packet loss _initial hop cost_ on to the Network Next relay fleet for each player. Thus, we are able to plan traffic routing from the client in any ISP around the world, to servers running in datacenters anywhere in the world, while knowing the quality of service (Qos) for all possible routes end-to-end between the client and server. We can even adjust the player route dynamically as network conditions change.

The Network Next SDK is written in C-like C++ and supports all common platforms: Windows, Mac, Linux, iOS, PS4, PS5, XBox One, XBox Series X and Nintendo Switch. It works with all hosting providers with special datacenter autodetect support for Google Cloud, AWS and Multiplay. The only requirement is that you send and receive UDP packets to implement your game protocol and that you have a client/server architecture - peer-to-peer is not supported.

For games that are using UE5 we have plugin that provides a drop-in NetDriver replacement.

## 2. The backend

The Network Next backend runs in Google Cloud and analyzes ping data between relays to find the set of all optimal paths between all relays. This "route matrix" is updated once per-second inside the relay backend, and it provides a constant time lookup for the best routes from one relay to another (even on the internet backbone there are significant optimizations to be found). Combined with the near relay pings for _initial hop_ cost, we can optimize globally from any client in the world to any server in the world, provided that there are relays near the player, sufficient density and diversity of relays suppliers between the player and the server, and a _destination relay_ in the datacenter where the game server is running.

The backend performs this global optimization once per-second in the _relay_backend_ component. The _server_backend_ component uses this _route matrix_ updated once per-second, combined with the near relay cost to find the set of best routes for each player, once every 10 seconds. The portal shows player network performance in real-time and is driven by a redis cluster instance that data is written to from the _server backend_. Network performance data is also written to bigquery every 10 seconds per-player for later analysis.

The backend is load tested to scale up to 1,000 relays, 2.5M servers and 25M CCU, with a typical acceleration rate of 10% of players at any time. Out of the box, without further scaling, the production backend is able to handle 1M CCU and trivially scale up to 10M CCU without significant changes. With additional redis provisioning it can hit 25M+ CCU, but this is not necessary -- the peak CCU for fortnite was 2.5M CCU. 

It is multi-zone for high availability and setup with regional load balancers in google cloud. It supports seamless deploys without disruption to all backend services, and if for any reason the backend is down, the Network Next SDK automatically falls back to non-accelerated mode for your players, so that gameplay is never disrupted.

## 3. The relay fleet

...
