<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Create your own buyer

## 1. Generate new buyer keypair

Go to the console and type:

```
cd ~/next && go run sdk/keygen/keygen.go
```

You will see output like this:

```
gaffer@macbook next % go run sdk/keygen/keygen.go

Welcome to Network Next!

This is your public key:

    SPeLMXdfJRtK3E2rEX7L9JIFxpn+cykxtuWAUCZVLbAEcwFrc0oVoQ==

This is your private key:

    SPeLMXdfJRt83tjKOYXbR0JyLdbuaGH7GpK21oTalLITqCOdBVzZ40rcTasRfsv0kgXGmf5zKTG25YBQJlUtsARzAWtzShWh

IMPORTANT: Save your private key in a secure place and don't share it with anybody, not even us!
```

This is your new buyer keypair. The public key can be safely shared with anybody and embedded in your client. The private key should be known only by the server, and not your players.

Save your keys in your secrets directory as `buyer_public_key.txt` and `buyer_private_key.txt` and then back up your secrets directory somewhere.

## 2. Create a new buyer in the database

Edit the file `~/next/terraform/dev/relays/main.tf`

Pick a name for your game. You need a short code for your game, with only a-z characters, and a longer name that is more descriptive.

I'm going to call my game "helsinki", "Helsinki, Finland"

At the bottom, add the following text, replacing "helsinki" and "Helsinki, Finland" with your own buyer code and name:

```
# ----------------------------------------------------------------------------------------

# ==============
# HELSINKI BUYER
# ==============

locals {
  helsinki_public_key = file("~/secrets/buyer_public_key.txt")
  helsinki_datacenters = [
    "google.iowa.1",
    "google.iowa.2",
    "google.iowa.3",
    "google.iowa.6"
  ]
}

resource "networknext_route_shader" helsinki {
  name = "helsinki"
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 2
  route_switch_threshold = 5
  acceptable_packet_loss_instant = 0.25
  acceptable_packet_loss_sustained = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}

resource "networknext_buyer" helsinki {
  name = "Helsinki, Finland"
  code = "helsinki"
  debug = true
  live = true
  route_shader_id = networknext_route_shader.helsinki.id
  public_key_base64 = local.buyer_public_key
}

resource "networknext_buyer_datacenter_settings" helsinki {
  count = length(local.buyer_datacenters)
  buyer_id = networknext_buyer.helsinki.id
  datacenter_id = networknext_datacenter.datacenters[local.buyer_datacenters].id
  enable_acceleration = true
}

# ----------------------------------------------------------------------------------------
```

## 3. Customize your route shader

The route shader is configured as follows:

* `acceptable_latency = 50`. Do not accelerate any player when their latency is already below 50ms. This is good enough and not worth accelerating below. This is a recommended value to start with.
  
* `latency_reduction_threshold = 10`. Do not accelerate a player unless we can find a latency reduction of _at least_ 10 milliseconds.

* `route_select_threshold = 2`. This finds the absolute lowest latency route within 2ms of the best route available. This helps load balance across multiple routes that are close enough together due to fluctuation, instead of everybody going down a route that is temporarily 1 millisecond faster than another.

* `route_switch_threshold = 5`. Hold the current Network Next route, unless a better route is available with at least 5ms lower latency than the current route. Don't set this too low, or the route will flap around every 10 seconds. In the future, I recommend that you increase this to 10ms, but right now 5ms is fine.

* `acceptable_packet_loss_instant = 0.25`. If packet loss > 0.25% occurs in any 10 second period, accelerate the player to reduce packet loss. This catches packet loss spikes.

* `acceptable_packet_loss_sustained = 0.1`. If packet loss > 0.1% occurs for 30 seconds, accelerate the player to reduce packet loss. This captures packet loss that is lower in intensity, but is sustained over a longer period.

* `bandwidth_envelope_up_kbps = 256`. This is the maximum bandwidth in kilobits per-second, kilobits, not kilobytes, sent from client to server. We don't need to change this, but later on when you setup your own buyer you should adjust it to the maximum bandwidth your client will send up to to the server. If the client sends more bandwidth than this, it will not be accelerated (eg. during a load screen, this is typically OK).

* `bandwidth_envelope_down_kbps = 256`. The bandwidth down from server to client in kilobits per-second. Again, we don't need to change this yet.

## 4. Commit the database to the backend

...

## 5. Verify your buyer exists in the portal

...

Up next: [Run your own client and server](run_your_own_client_and_server.cpp).
