<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Planning your production relay fleet

Hello operator, this section is your guide to planning the relay fleet for your game. This is the most fun part of Network Next. I hope you enjoy :)

## 1. Put a relay in each datacenter where you host servers

To accelerate traffic, Network Next needs a "destination relay" as the last hop when delivering traffic to a server in a datacenter.

So, the very first step when planning your relay fleet is to make sure there is at least one relay in every datacenter where host game servers and you want traffic to be accelerated.

Make sure that you add all datacenters where you plan to accelerate traffic into the list of datacenter in terraform for your buyer:

```
locals {
  your_buyer_datacenters = [
    "google.iowa.1",
    "google.iowa.2",
    "google.iowa.3",
    "google.iowa.6"
  ]
}

resource "networknext_buyer_datacenter_settings" your_buyer {
  count = length(local.your_buyer_datacenters)
  buyer_id = networknext_buyer.your_buyer.id
  datacenter_id = networknext_datacenter.datacenters[local.your_buyer_datacenters].id
  enable_acceleration = true
}
```

Network Next uses this list of enabled datacenters per-buyer to only optimize traffic to the datacenters where game servers reside. This is a significant reduction in the amount of work the optimizer needs to do, because route optimization process is otherwise O(n^3) where n is the number of relays in your relay fleet, and now it is ~O(n^2*m) where m is the number of destination relays.

## 2. In each location where you host servers add 20-30 additional relays from other suppliers

What you are doing now is effectively setting up alternative routes from clients to your servers across different networks.

For example, if you host in AWS, spinning up google cloud relays in the same location means that if the google cloud network is better performing, lower latency, or just less congested at any point in time to the AWS network for a client, then the traffic will be steered through google's network on the way to your game server in AWS.

Start with cloud relays first as they are easy to spin up and down with little effort. Once you have exhausted cloud options, start looking into bare metal relays in the same city. It is recommended that you setup bare metal relays with a 10G NIC, and a plan that provides sufficient sustained traffic. It is little value adding a relay if it can only carry < 1gbps of traffic.

Some high quality providers we have used in the past:

* https://i3d.net - excellent bare metal backed by their own private backbone
* https://stackpath.com - stackpath is a CDN that now has their own edge compute where you can run relays
* https://hivelocity.net - they have good connectivity and locations around the world
* https://www.hetzner.com - particularly good in EU
* https://performive.com - previously known as "total server solutions"
* https://oneqode.com - excellent in Asia-Pacific and Australia
* https://www.serversaustralia.com.au - excellent in Australia
* https://www.tencentcloud.com - worth looking into around the world and especially for Asia
* https://vultr.com - they have a low quality terraform provider and cloud, but their bare metal can be useful in certain locations.
* https://www.colocrossing.com - excellent performance around the world and a very good price. Historically they have hosted ESL game servers.
* https://azure.com - not currently supported by our terraform provider, but it can easily be added so let us know if there is interest.
* https://velia.net - some good locations and worth adding
* https://www.latitude.sh - some good locations especially in South America
* https://deploy.equinix.com - used to be know as packet.com, now acquired by equinix, expensive but good locations

Do your own research and of course there are more to try. I recommend you work on a per-city basis and spin up as many different providers in the location, and then over time as you see certain providers performing better than others (carrying more accelerate traffic), then you can whittle down and select the n best providers per-location that you desire. I recommend a minimum of 10 per-location of best providers.

## 3. Deploy 20-30 relays in major cities around the world in regions where you have players

Here you are trying to catch players going on to your relay fleet close to where they are, even if they are playing on servers that are not immediately close to them, in the same country or the same state. You are also assisting cross location and cross region play to select the best backbone to transit from the player's location to the server location. The more relays you have, the more transit options from the player to the server that you unlock.

Some of these locations will already be covered by places where you are hosting servers. In that case, ignore the these locations, you've already set them up in step 1.

Essential locations in North America:

* Los Angeles
* Dallas
* Miami
* Virginia / Washington DC
* San Jose / Silicon Valley / SF
* New York / Newark NJ
* Chicago / Iowa
* Seattle

Secondary locations in USA that should be evaluated on a case by case basis depending on your player distribution:

* Atlanta
* South Carolina
* Denver
* Ohio
* St Louis
* Oregon

Primary locations in Europe:

* Frankfurt
* Amsterdam
* London

Secondary locations in Europe:

* Moscow
* Helsinki
* Stockholm
  
South America essential locations:

* Sao Paulo
* Santiago Chile

Secondary locations in South America:

* Lima, Peru
* Buenos Aires

Primary locations in APAC:

* Singapore
* Hong Kong
* Taiwan
* Tokyo
* Seoul

Secondary locations in APAC:

* Sydney
* Perth
* Melbourne
* Brisbane

## 4. Iterate

The general process is to deploy 30 relays in each location near players, let it run for a month, look at per-relay accelerated traffic in a given location (use Bigquery) and rank relays from most to least traffic carried.

If you wish to optimize cost, reduce relays down to the n relays per-location with the most traffic carried across the last month with real player traffic.

[Back to main documentation](../README.md)
