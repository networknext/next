<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Frequently Asked Questions

## What is Network Next?

Network Next is the marketplace for premium network transit. 

## Why use Network Next?

Players have lower latency, fewer hitches, and less packet loss when your game uses Network Next.

## How does Network Next work?

Suppliers compete on performance and price to carry your game's UDP traffic. When the internet is not providing the performance your player needs, Network Next steers your game traffic across this route, changing the route dynamically - _without your players noticing_ - once every 10 seconds when a better or cheaper route becomes available.

## Does Network Next change the player route all the time?

No. Typically Network Next finds a high quality route and sticks with it, however if the current route degrades, or if we find something significantly better, we are able to switch to that route dynamically.

## Do players notice when the route changes?

No. It's completely seamless, but at the same time we do our best to minimize the number of route changes.

## Does Network Next violate Network Neutrality?

No. Routing via bidding process ensures that we remain _supplier neutral_ and always steer traffic across the route with the best performance at the best price, without favoring any one buyer or supplier over another.

## Can suppliers specify a different price depending on the buyer?

No. Suppliers cannot identify buyers on Network Next and therefore cannot provide one price for one buyer, and a different price for another. Suppliers may only compete on performance and price.

## What sort of improvements can Network Next provide?

It depends on the time of day, ISP and location of each player, and on the data center where game server is located, but overall across all players, we are able to provide significant performance improvements for at least 10% of sessions.

## What's the biggest improvement Network Next has provided?

We've seen players with over 1000ms round trip time reduced to less than 50ms and players with almost 100% packet loss reduced to zero. This is not the common case, but these players certainly appreciate it :)

## How much does Network Next cost?

The average price per-gigabyte is 14c right now, but because Network Next typically targets only the bottom 10% that need help at any time, the amortized cost is just 1.4c per-GB.

## Do players have to pay to use Network Next?

No. Our customer is the game developer, not the player.

## How does billing work?

Each month we sum up the total of 10 second bandwidth slices served by all suppliers for a game (often multiple suppliers are involved in each route), and that is the amount we bill to the game developer.

## Do you bill on bandwidth usage or on the amount reserved?

Private networks reserve capacity for each player to ensure the best network quality, so we bill on the bandwidth envelope specified, not on the amount actually used.

## Can there be a different bandwidth envelope up and down?

Yes. Game protocols are often asymmetric, so you may specify different envelopes for client to server traffic and server to client traffic.

## Is the bandwidth envelope per-server, or per-player

Each player has their own bandwidth envelope.

## What happens if I go over the bandwidth envelope?

Packets that go over the bandwidth envelope are not accelerated and just take the public internet.

## Can I change the bandwidth envelope over time?

Not yet, but we're actively working on this feature.
 
## How do I get access to the Network Next portal

Please contact us at sales@networknext.com

## What games use Network Next?

Network Next is used by [Rocket League](https://rocketleague.com) and we accelerate millions of sessions every day. We're also live with several other games we haven't announced yet.

## What suppliers work with Network Next?

We work with top edge computing and infrastructure suppliers such as:

* [Ericsson Edge Gravity](https://edgegravity.ericsson.com)
* [Stackpath](https://stackpath.com)
* [Limelight](https://limelight.com)
* [Packet](https://packet.com)
* [Zenlayer](https://zenlayer.com)
* [Multiplay](https://multiplay.com)
* [I3D](https://i3d.net)
* [Vultr.com](https://vultr.com)
* [GameServers.com](https://gameservers.com)
* [AWS](https://aws.amazon.com)
* [Google](https://cloud.google.com/)
* [Azure](https://azure.microsoft.com)
* [Oracle Cloud](https://cloud.oracle.com/)
* [IBM Cloud](https://ibm.com/cloud)

_And many more we can't talk about yet!_

## Where can I host game servers with Network Next?

We support game servers hosted by:

* [Multiplay](https://multiplay.com)
* [I3D](https://i3d.net)
* [Vultr.com](https://vultr.com)
* [AWS](https://aws.amazon.com)
* [Google](https://cloud.google.com/)
* [Azure](https://azure.microsoft.com)

If you'd like to evaluate Network Next but you host game servers somewhere else, please let us know at sales@networknext.com

## What platforms do you support?

We support:

* PC
* Mac
* Linux
* iOS
* PS4
* Xbox One
* Nintendo Switch

Access to consoles requires proof you are a registered developer for that platform. Please contact sales@networknext.com for more information.

## What if my platform isn't listed above

Please contact sales@networknext.com so we can understand your use case. Thanks!

## What languages are supported?

The Network Next SDK is written in low-level C style C++ so it's easy to call from other languages. Get involved and contribute a binding for your preferred language to our open source SDK! https://github.com/networknext/sdk

## What locations are supported by Network Next?

Network Next is deployed worldwide!

## Do you have an Unreal Engine 4 Plugin?

Yes. Please contact sales@networknext.com 

## Do you have a Unity plugin?

Not yet, but we're actively working on it.

## Can I use Network Next for player hosted servers?

No. We only support acceleration to servers hosted in datacenters with a known public IP.

## Can I use Network Next for my peer-to-peer game?

No. We only support acceleration of client/server traffic between clients and dedicated servers hosted in datacenters.

## Can I use Network Next to accelerate my TCP game?

Not yet, but in the future we may be able to provide benefit via a TCP shim that reduces the impact of packet loss induced head of line blocking. Please contact sales@networknext.com so we can understand your use case.

## Can I use Network Next to accelerate my WebRTC game?

Not yet. We are actively investigating this as an option. Please contact sales@networknext.com

## Will players be disconnected from my game if Network Next goes down?

No. If our SDK cannot talk to our backend for any reason, players automatically fall back to the public internet.

## Are you hiring?

Yes. https://www.networknext.com/careers
