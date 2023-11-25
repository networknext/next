<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Run a test client

## 1. Disable raspberry clients

By default, Network Next fills the portal with simulated data from 1024 clients. These are the "raspberry" clients. They're named this because we use to run test clients on Raspberry Pi's around the vorld. 

Nowadays, they're just clients running inside a google cloud datacenter, with simulated random locations and ping data.

Let's turn these clients off so we can see some real data in the portal.

Edit `terraform/dev/backend/main.tf`

Search for "raspberry_backend". In the raspberry backend module, change `target_size` to 0:

```
module "raspberry_backend" {

  source = "../../modules/external_http_service"

  service_name = "raspberry-backend"

  startup_script = <<-EOF1
    #!/bin/bash
    gsutil cp ${var.google_artifacts_bucket}/${var.tag}/bootstrap.sh bootstrap.sh
    chmod +x bootstrap.sh
    ./bootstrap.sh -t ${var.tag} -b ${var.google_artifacts_bucket} -a raspberry_backend.tar.gz
    cat <<EOF > /app/app.env
    ENV=dev
    DEBUG_LOGS=1
    REDIS_HOSTNAME="${google_redis_instance.redis_raspberry.host}:6379"
    EOF
    systemctl start app.service
  EOF1

  tag                      = var.tag
  extra                    = var.extra
  machine_type             = "f1-micro"
  project                  = local.google_project_id
  region                   = var.google_region
  zones                    = var.google_zones
  default_network          = google_compute_network.development.id
  default_subnetwork       = google_compute_subnetwork.development.id
  service_account          = local.google_service_account
  tags                     = ["allow-ssh", "allow-http", "allow-https"]
  domain                   = "raspberry-dev.${var.cloudflare_domain}"
  certificate              = google_compute_managed_ssl_certificate.raspberry-dev.id
  target_size              = 0 // 1

  depends_on = [
    module.server_backend
  ]
}
```

Search for "raspberry_client" and "raspberry_server" and make the same change.

Now deploy a new dev release with a tag:

```
git tag dev-003
git push origin dev-003
```

This change modifies the managed instance groups that run the raspberry clients down to zero VM instances, effectively disabling them. 

It will take a few minutes, but when the change is deployed the portal will be empty:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/ee6a485a-2444-48e7-a8fe-5758c5d2df27">

## 2. Run a test client

There is already a test server running in google cloud, under a buyer account called "Test".

You can run a local client, and it will connect to this server automatically:

```
next select dev
run client
```

You can run multiple clients if you want.

When the client runs you should see something like this:

```console
gaffer@macbook next % run client
0.005014: info: platform is mac (wi-fi)
0.005757: info: log level overridden to 4
0.005765: info: buyer public key override: 'fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=='
0.005768: debug: buyer public key is 'fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=='
0.005770: info: found valid buyer public key: 'fJ9R1DqVKevreg+kvqEkFqbAAa54c6BXcgBn+R2GKM1GkFo8QtkUZA=='
0.005773: info: buyer private key override
0.005775: info: found valid buyer private key
0.005778: info: override server backend hostname: 'server-dev.virtualgo.net'
0.005780: info: server backend public key override: hc3/baZ4FYYaknk1heRK345FK3ZOSpfK1PiMmuIGb1w=
0.005782: info: valid server backend public key
0.005783: info: relay backend public key override: 8jOvRZXo57q2kivlnm4nW9Ff6oi9fgHBnWoUJhz4PQQ=
0.005785: info: valid relay backend public key
0.005788: info: client buyer id is eb29953ad4519f7c
0.005847: info: client bound to 0.0.0.0:63860
0.107106: info: client opened session to 34.67.212.136:30000
0.448722: debug: client received upgrade request packet from server
0.448757: debug: client initial magic: c3,a2,54,f4,d8,4a,fd,4a | 2f,a0,45,04,04,63,37,e2 | 16,f3,23,66,03,d1,6f,4e
0.448767: debug: client external address is 93.90.49.137:63860
0.448866: debug: client sent upgrade response packet to server
0.586565: debug: client received upgrade confirm packet from server
0.604316: info: client upgraded to session cd5b286f6b8ca41d
0.784911: debug: client received route update packet from server
--------------------------------------
forcing network next
best route cost is 2147483647
could not find any next routes
not taking network next. no network next route is available
staying direct
--------------------------------------
0.784970: info: client pinging 16 near relays
0.785157: info: client direct route
0.785214: debug: client sent route update ack packet to server
10.817877: debug: client received route update packet from server
--------------------------------------
forcing network next
best route cost is 134
found 7 suitable routes in [134,134] from 16/16 near relays
route diversity is 5
take network next: amazon.stockholm.3 - google.iowa.6 - google.iowa.1
route cost is 137
--------------------------------------
10.817941: info: client near relay pings completed
10.818365: info: client next route
10.818388: info: client multipath enabled
10.818442: debug: client sent route update ack packet to server
10.818458: debug: client sent route request to relay: 16.170.166.141:40000
10.955381: debug: client received route response from relay
10.955415: debug: client network next route is confirmed
20.872392: debug: client received route update packet from server
--------------------------------------
stay on network next: 16/16 source relays are routable
1 dest relay
amazon.stockholm.3 - google.iowa.6 - google.iowa.1
route continued
route cost is 137
--------------------------------------
20.872528: info: client continues route
20.872618: debug: client sent route update ack packet to server
20.885590: debug: client sent continue request to relay: 16.170.166.141:40000
21.020719: debug: client received continue response from relay
21.020761: debug: client continue network next route is confirmed
30.777396: debug: client received route update packet from server
```

Wait a minute for the portal to display the data (it updates once per-minute).

You will see your test clients in the portal, now with real ping data and your ISP name:

![Uploading image.png…]()

Here I am in Helsinki, Finland in the Kamppi shopping center, beating google to their own datacenter with Network Next. This is in dev, which is not a particularly optimized environment, and I'm in an extremely well connected city. Much larger improvements are seen across a typical player base worldwide.

I can click on the session id, and see my session updating in real-time, once every 10 seconds:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/61cacf15-4f8a-46a2-aa23-f07c95810c9d">

## 3. Edit the route shader

The test client has a "route shader" which specifies the criteria for when we should take Network Next.

What we will do now is set a good default route shader, which will only take Network Next if a significant latency reduction is found, or if the user is experiencing significant packet loss.

Edit the file `terraform/dev/relays/main.tf` and search for "TEST BUYER"

You will find this:

```
resource "networknext_route_shader" test {
  name = "test"
  force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss_instant = 0.25
  acceptable_packet_loss_sustained = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}

resource "networknext_buyer" test {
  name = "Test"
  code = "test"
  debug = true
  live = true
  route_shader_id = networknext_route_shader.test.id
  public_key_base64 = var.test_buyer_public_key
}

resource "networknext_buyer_datacenter_settings" test {
  count = length(var.test_datacenters)
  buyer_id = networknext_buyer.test.id
  datacenter_id = networknext_datacenter.datacenters[var.raspberry_datacenters[count.index]].id
  enable_acceleration = true
}
```

This configuration sets up a test buyer, gives it a route shader, and enables acceleration to datacenters included in the "test_datacenters". Yes, Network Next will only accelerate to specific datacenters that you enable, per-buyer. This way the optimization algorithm is much more efficient.

Right now it is configured as:

```
test_datacenters = [
	"google.iowa.1",
	"google.iowa.2",
	"google.iowa.3",
	"google.iowa.6"
]
```

In `terraform/dev/relays/terraform.tfvars`. This means that for the test customer we can run test servers in any of the google.iowa.* datacenters and they will potentially be accelerated.

Now let's edit the route shader. Comment out the line `force_next`. This line forces all players to get accelerated across Network Next, even if there is no improvement.

Now the route shader looks like this:

```
resource "networknext_route_shader" test {
  name = "test"
  // force_next = true
  acceptable_latency = 50
  latency_reduction_threshold = 10
  route_select_threshold = 0
  route_switch_threshold = 5
  acceptable_packet_loss_instant = 0.25
  acceptable_packet_loss_sustained = 0.1
  bandwidth_envelope_up_kbps = 256
  bandwidth_envelope_down_kbps = 256
}
```

This route shader is configured as follows:

* `acceptable_latency = 50`. Do not accelerate any player when their latency is already below 50ms. This is good enough and not worth accelerating below. This is a recommended value to start with.
  
* `latency_reduction_threshold = 10`. Do not accelerate a player unless we can find a latency reduction of _at least_ 10 milliseconds.

* `route_select_threshold = 0`. This finds the absolute lowest latency route, out of all routes available. In the future, it is best to relax this to 5ms, so we load balance across different routes, instead of forcing everybody down the same route.

* `route_switch_threshold = 5`. Hold the current Network Next route, unless a better route is available with at least 5ms lower latency than the current route. Don't set this too low, or the route will flap around every 10 seconds. In the future, I recommend that you increase this to 10ms, but right now 5ms is fine.

* `acceptable_packet_loss_instanct = 0.25`. If packet loss > 0.25% occurs in any 10 second period, accelerate the player to reduce packet loss. This catches packet loss spikes.

* `acceptable_packet_loss_sustained = 0.1`. If packet loss > 0.1% occurs for 30 seconds, accelerate the player to reduce packet loss. This captures packet loss that is lower in intensity, but is sustained over a longer period.

* `bandwidth_envelope_up_kbps = 256`. This is the maximum bandwidth in kilobits per-second, kilobits, not kilobytes, sent from client to server. We don't need to change this, but later on when you setup your own buyer you should adjust it to the maximum bandwidth your client will send up to to the server. If the client sends more bandwidth than this, it will not be accelerated (eg. during a load screen, this is typically OK).

* `bandwidth_envelope_down_kbps = 256`. The bandwidth down from server to client in kilobits per-second. Again, we don't need to change this yet.

Now, deploy the terraform changes to the postgres database:

```
cd ~/next/terraform/dev/relays
terraform apply
```

Now that we have changed the postgres database, we need to commit it to the backend for it to take effect:

```
cd ~/next
next database
next commit
```

Stop your client, then restart it:

```
next client
```

Now you should after a minute or so see your client session in the portal, but this time it is (almost certainly) not accelerated:



This is what we want. Network Next is designed to only accelerate players when a very significant improvement can be found. For most game developers, we have good internet connections and we live in major cities, and most of the time they are good enough. But across an entire game's player base, around 10% of players at any time will be accelerated with the settings above.
